package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/pragma-proto/api/config"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/database"
	"github.com/pragma-proto/api/internal/handlers"
	apimiddleware "github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/services"
)

func main() {
	// Load config — panics on missing required env vars.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Init database.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Run migrations (SQL files are embedded in the binary at compile time).
	if err := db.RunMigrations(context.Background()); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// Init JWT service.
	jwtSvc, err := auth.NewJWTService(cfg.JWTPrivateKey, cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}

	// Init services.
	gradingSvc := services.NewGradingService()
	pdfSvc := services.NewPDFService()
	emailSvc := services.NewEmailService(cfg.ResendAPIKey, cfg.EmailFromAddr)
	aiSvc := services.NewAIService(cfg.ClaudeAPIKey, cfg.ClaudeModel)
	_ = gradingSvc
	_ = pdfSvc
	_ = emailSvc
	_ = aiSvc

	// Storage service (R2).
	storageSvc, err := services.NewStorageService(
		cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey,
		cfg.R2BucketName, cfg.R2Endpoint,
	)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	// Verification service (school secret = encryption root key).
	verificationSvc := services.NewVerificationService(cfg.EncryptionRootKey)

	// Init handlers.
	authH := handlers.NewAuthHandler(db.Pool, jwtSvc)
	gradesH := handlers.NewGradesHandler(db.Pool, gradingSvc)
	assignmentsH := handlers.NewAssignmentsHandler(db.Pool, storageSvc)
	adminH := handlers.NewAdminHandler(db.Pool, emailSvc)
	dashboardH := handlers.NewDashboardHandler(db.Pool)
	aiH := handlers.NewAIHandler(db.Pool, aiSvc)
	documentsH := handlers.NewDocumentsHandler(db.Pool, pdfSvc, storageSvc, verificationSvc, cfg.FrontendOrigin)
	digitalIDH := handlers.NewDigitalIDHandler(db.Pool, storageSvc, verificationSvc, cfg.FrontendOrigin)
	scheduleH := handlers.NewScheduleHandler(db.Pool)
	reportsH := handlers.NewReportsHandler(db.Pool, pdfSvc, storageSvc, gradingSvc)
	coursesH := handlers.NewCoursesHandler(db.Pool)
	studentsH := handlers.NewStudentsHandler(db.Pool)

	// Build router.
	r := chi.NewRouter()

	// Global middleware.
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(apimiddleware.SecurityHeaders)
	r.Use(apimiddleware.CORS(cfg.FrontendOrigin))

	// Health check (no auth).
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Healthy(r.Context()); err != nil {
			http.Error(w, `{"status":"unhealthy"}`, http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Public: document and ID verification.
	r.Get("/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		// Dispatch to digital ID or document verification based on code format.
		// In production, the code prefix distinguishes the type.
		documentsH.VerifyDocument(w, r)
	})

	// Auth routes (no JWT required, but rate limited).
	r.Group(func(r chi.Router) {
		r.Use(apimiddleware.RateLimitLogin)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/register", authH.Register)
	})

	// Authenticated routes.
	authMiddleware := auth.Middleware(jwtSvc)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(apimiddleware.TenantMiddleware)
		r.Use(apimiddleware.RateLimitGeneral)
		r.Use(apimiddleware.AuditMiddleware(db.Pool))

		// Auth: MFA verify, logout.
		r.Post("/auth/mfa/verify", authH.VerifyMFA)
		r.Post("/auth/logout", authH.Logout)

		// Dashboard.
		r.Get("/dashboard", dashboardH.GetDashboard)

		// Grades.
		r.Route("/courses/{courseId}/grades", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Get("/", gradesH.ListGrades)
			r.Post("/", gradesH.UpsertGrade)
		})
		r.Route("/students/{studentId}/grades", func(r chi.Router) {
			r.Get("/", gradesH.GetStudentGrades)
		})

		// Assignments.
		r.Route("/assignments", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Get("/", assignmentsH.ListAssignments)
			r.Post("/", assignmentsH.CreateAssignment)
		})

		// Students self-service (any authenticated student can look up their own record).
		r.Get("/students/me", studentsH.GetMyRecord)
		r.Route("/assignments/{assignmentId}/attachments", func(r chi.Router) {
			r.Get("/", assignmentsH.ListAttachments)
			r.With(apimiddleware.RequireRoles("teacher", "admin", "super_admin")).
				With(apimiddleware.RateLimitFileUpload).
				Post("/upload-url", assignmentsH.RequestUploadURL)
		})

		// AI.
		r.Route("/ai", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Use(apimiddleware.RateLimitAI)
			r.Post("/grading-assistant", aiH.GradingAssistant)
			r.Post("/report-comment", aiH.ReportComment)
		})

		// Admin: grade locks.
		r.Route("/admin", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("admin", "super_admin"))

			r.Get("/students", adminH.ListStudents)
			r.Post("/students/{studentId}/lock", adminH.LockGrade)
			r.Delete("/students/{studentId}/lock", adminH.UnlockGrade)
			r.Post("/grade-locks/bulk", adminH.BulkLockGrades)
		})

		// Documents (rate limited per spec: 5/day).
		r.Route("/documents", func(r chi.Router) {
			r.Use(apimiddleware.RateLimitDocumentGeneration)
			r.Post("/", documentsH.GenerateDocument)
		})

		// Digital IDs.
		r.Route("/students/{studentId}/digital-id", func(r chi.Router) {
			r.Get("/", digitalIDH.GetStudentID)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/", digitalIDH.IssueStudentID)
		})
		r.With(apimiddleware.RequireRoles("admin", "super_admin")).
			Delete("/digital-ids/{idId}", digitalIDH.RevokeStudentID)

		// Schedule.
		r.Route("/schedule", func(r chi.Router) {
			r.Get("/", scheduleH.ListSchedule)
			r.Post("/", scheduleH.CreateScheduleBlock)
			r.Delete("/{blockId}", scheduleH.DeleteScheduleBlock)
		})

		// Report cards.
		r.Route("/reports", func(r chi.Router) {
			r.With(apimiddleware.RequireRoles("teacher", "admin", "super_admin")).
				Post("/", reportsH.GenerateReportCard)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/batch", reportsH.BatchGenerateReports)
		})
		r.Route("/students/{studentId}/reports", func(r chi.Router) {
			r.Get("/", reportsH.ListReportCards)
		})

		// Courses.
		r.Route("/courses", func(r chi.Router) {
			r.Get("/mine", coursesH.ListMyCourses)
			r.Get("/{courseId}", coursesH.GetCourse)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/", coursesH.CreateCourse)
		})
		r.Route("/courses/{courseId}/students", func(r chi.Router) {
			r.Get("/", coursesH.GetEnrolledStudents)
		})
	})

	// Start server.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server: listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-stop
	log.Println("server: shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server: forced shutdown: %v", err)
	}
	log.Println("server: stopped")
}
