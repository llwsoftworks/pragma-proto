package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
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

var (
	once       sync.Once
	appHandler http.Handler
)

func buildHandler() http.Handler {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	migrationsDir := filepath.Join("internal", "database", "migrations")
	if err := db.RunMigrations(context.Background(), migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	jwtSvc, err := auth.NewJWTService(cfg.JWTPrivateKey, cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("jwt: %v", err)
	}

	gradingSvc := services.NewGradingService()
	pdfSvc := services.NewPDFService()
	emailSvc := services.NewEmailService(cfg.ResendAPIKey, cfg.EmailFromAddr)
	aiSvc := services.NewAIService(cfg.ClaudeAPIKey, cfg.ClaudeModel)

	storageSvc, err := services.NewStorageService(
		cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey,
		cfg.R2BucketName, cfg.R2Endpoint,
	)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	verificationSvc := services.NewVerificationService(cfg.EncryptionRootKey)

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

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(apimiddleware.SecurityHeaders)
	r.Use(apimiddleware.CORS(cfg.FrontendOrigin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Healthy(r.Context()); err != nil {
			http.Error(w, `{"status":"unhealthy"}`, http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	r.Get("/verify/{code}", func(w http.ResponseWriter, r *http.Request) {
		documentsH.VerifyDocument(w, r)
	})

	r.Group(func(r chi.Router) {
		r.Use(apimiddleware.RateLimitLogin)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/register", authH.Register)
	})

	authMiddleware := auth.Middleware(jwtSvc)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(apimiddleware.TenantMiddleware)
		r.Use(apimiddleware.RateLimitGeneral)
		r.Use(apimiddleware.AuditMiddleware(db.Pool))

		r.Post("/auth/mfa/verify", authH.VerifyMFA)
		r.Post("/auth/logout", authH.Logout)

		r.Get("/dashboard", dashboardH.GetDashboard)

		r.Route("/courses/{courseId}/grades", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Get("/", gradesH.ListGrades)
			r.Post("/", gradesH.UpsertGrade)
		})
		r.Route("/students/{studentId}/grades", func(r chi.Router) {
			r.Get("/", gradesH.GetStudentGrades)
		})

		r.Route("/assignments", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Post("/", assignmentsH.CreateAssignment)
		})
		r.Route("/assignments/{assignmentId}/attachments", func(r chi.Router) {
			r.Get("/", assignmentsH.ListAttachments)
			r.With(apimiddleware.RequireRoles("teacher", "admin", "super_admin")).
				With(apimiddleware.RateLimitFileUpload).
				Post("/upload-url", assignmentsH.RequestUploadURL)
		})

		r.Route("/ai", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("teacher", "admin", "super_admin"))
			r.Use(apimiddleware.RateLimitAI)
			r.Post("/grading-assistant", aiH.GradingAssistant)
			r.Post("/report-comment", aiH.ReportComment)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(apimiddleware.RequireRoles("admin", "super_admin"))
			r.Get("/students", adminH.ListStudents)
			r.Post("/students/{studentId}/lock", adminH.LockGrade)
			r.Delete("/students/{studentId}/lock", adminH.UnlockGrade)
			r.Post("/grade-locks/bulk", adminH.BulkLockGrades)
		})

		r.Route("/documents", func(r chi.Router) {
			r.Use(apimiddleware.RateLimitDocumentGeneration)
			r.Post("/", documentsH.GenerateDocument)
		})

		r.Route("/students/{studentId}/digital-id", func(r chi.Router) {
			r.Get("/", digitalIDH.GetStudentID)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/", digitalIDH.IssueStudentID)
		})
		r.With(apimiddleware.RequireRoles("admin", "super_admin")).
			Delete("/digital-ids/{idId}", digitalIDH.RevokeStudentID)

		r.Route("/schedule", func(r chi.Router) {
			r.Get("/", scheduleH.ListSchedule)
			r.Post("/", scheduleH.CreateScheduleBlock)
			r.Delete("/{blockId}", scheduleH.DeleteScheduleBlock)
		})

		r.Route("/reports", func(r chi.Router) {
			r.With(apimiddleware.RequireRoles("teacher", "admin", "super_admin")).
				Post("/", reportsH.GenerateReportCard)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/batch", reportsH.BatchGenerateReports)
		})
		r.Route("/students/{studentId}/reports", func(r chi.Router) {
			r.Get("/", reportsH.ListReportCards)
		})

		r.Route("/courses", func(r chi.Router) {
			r.Get("/mine", coursesH.ListMyCourses)
			r.With(apimiddleware.RequireRoles("admin", "super_admin")).
				Post("/", coursesH.CreateCourse)
		})
		r.Route("/courses/{courseId}/students", func(r chi.Router) {
			r.Get("/", coursesH.GetEnrolledStudents)
		})
	})

	return r
}

// Handler is the Vercel serverless function entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		appHandler = buildHandler()
	})
	appHandler.ServeHTTP(w, r)
}
