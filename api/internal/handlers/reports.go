package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/models"
	"github.com/pragma-proto/api/internal/services"
)

// ReportsHandler generates and manages report cards.
type ReportsHandler struct {
	db      *pgxpool.Pool
	pdf     *services.PDFService
	storage *services.StorageService
	grading *services.GradingService
}

// NewReportsHandler creates a ReportsHandler.
func NewReportsHandler(
	db *pgxpool.Pool,
	pdf *services.PDFService,
	storage *services.StorageService,
	grading *services.GradingService,
) *ReportsHandler {
	return &ReportsHandler{db: db, pdf: pdf, storage: storage, grading: grading}
}

// GenerateReportCard generates a PDF report card for a single student.
func (h *ReportsHandler) GenerateReportCard(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		StudentID      string `json:"student_id" validate:"required,uuid"`
		AcademicPeriod string `json:"academic_period" validate:"required,min=3,max=100"`
		TeacherComment string `json:"teacher_comments"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	ctx := r.Context()

	// Fetch student data.
	var student models.Student
	var user models.User
	var school models.School

	h.db.QueryRow(ctx, `
		SELECT s.id, s.student_number, s.grade_level, s.enrollment_status,
		       u.first_name, u.last_name, u.email
		FROM students s JOIN users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.school_id = $2
	`, req.StudentID, claims.SchoolID).Scan(
		&student.ID, &student.StudentNumber, &student.GradeLevel, &student.EnrollmentStatus,
		&user.FirstName, &user.LastName, &user.Email,
	)

	h.db.QueryRow(ctx, `SELECT id, name, logo_url FROM schools WHERE id = $1`, claims.SchoolID).Scan(
		&school.ID, &school.Name, &school.LogoURL,
	)

	// Fetch course grades.
	gradeRows, err := h.db.Query(ctx, `
		SELECT c.name, u.first_name || ' ' || u.last_name,
		       COALESCE(AVG(g.points_earned / a.max_points * 100), 0)::float,
		       COALESCE(
		           CASE WHEN AVG(g.points_earned / a.max_points * 100) >= 93 THEN 'A'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 90 THEN 'A-'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 87 THEN 'B+'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 83 THEN 'B'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 80 THEN 'B-'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 77 THEN 'C+'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 73 THEN 'C'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 70 THEN 'C-'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 67 THEN 'D+'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 63 THEN 'D'
		                WHEN AVG(g.points_earned / a.max_points * 100) >= 60 THEN 'D-'
		                ELSE 'F'
		           END, 'N/A')
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		JOIN teachers t ON t.id = c.teacher_id
		JOIN users u ON u.id = t.user_id
		JOIN assignments a ON a.course_id = c.id AND a.is_published = TRUE
		LEFT JOIN grades g ON g.assignment_id = a.id AND g.student_id = e.student_id
		WHERE e.student_id = $1 AND e.status = 'active' AND e.course_id IN (
		    SELECT id FROM courses WHERE school_id = $2
		)
		GROUP BY c.name, u.first_name, u.last_name
		ORDER BY c.name
	`, req.StudentID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer gradeRows.Close()

	var courseGrades []services.CourseGradeRow
	var gpaSum float64
	var courseCount int
	for gradeRows.Next() {
		var row services.CourseGradeRow
		gradeRows.Scan(&row.CourseName, &row.TeacherName, &row.Percentage, &row.LetterGrade)
		courseGrades = append(courseGrades, row)
		gpaSum += gradePointFromLetter(row.LetterGrade)
		courseCount++
	}

	var gpa float64
	if courseCount > 0 {
		gpa = gpaSum / float64(courseCount)
	}

	data := services.ReportCardData{
		School:          &school,
		Student:         &student,
		StudentUser:     &user,
		AcademicPeriod:  req.AcademicPeriod,
		GPA:             gpa,
		CourseGrades:    courseGrades,
		TeacherComments: req.TeacherComment,
		GeneratedAt:     time.Now(),
	}

	htmlBytes, err := h.pdf.RenderReportCardHTML(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf_error", err.Error())
		return
	}

	reportID := uuid.New()
	key := services.ObjectKey(claims.SchoolID.String(), "reports", reportID.String()+".html")

	if err := h.storage.PutObject(ctx, key, htmlBytes, "text/html"); err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	var rcID uuid.UUID
	h.db.QueryRow(ctx, `
		INSERT INTO report_cards
			(student_id, school_id, academic_period, gpa, teacher_comments, pdf_url, generated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.StudentID, claims.SchoolID, req.AcademicPeriod, gpa, req.TeacherComment, key, claims.UserID).Scan(&rcID)

	downloadURL, _ := h.storage.PresignDownload(ctx, key)

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "report_card.generate",
		EntityType: "report_card",
		EntityID:   &rcID,
		NewValue:   map[string]string{"student_id": req.StudentID, "period": req.AcademicPeriod},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"report_card_id": rcID,
		"gpa":            gpa,
		"download_url":   downloadURL,
	})
}

// BatchGenerateReports generates report cards for multiple students concurrently.
func (h *ReportsHandler) BatchGenerateReports(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		StudentIDs     []string `json:"student_ids" validate:"required,min=1"`
		AcademicPeriod string   `json:"academic_period" validate:"required"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	type result struct {
		StudentID    string `json:"student_id"`
		ReportCardID string `json:"report_card_id,omitempty"`
		Error        string `json:"error,omitempty"`
	}

	results := make([]result, len(req.StudentIDs))
	var wg sync.WaitGroup

	// Semaphore to limit concurrent goroutines.
	sem := make(chan struct{}, 10)

	for i, sid := range req.StudentIDs {
		wg.Add(1)
		go func(idx int, studentID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = result{StudentID: studentID}
			// Individual generation logic (simplified for batch).
			var rcID uuid.UUID
			err := h.db.QueryRow(r.Context(), `
				INSERT INTO report_cards
					(student_id, school_id, academic_period, generated_by)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`, studentID, claims.SchoolID, req.AcademicPeriod, claims.UserID).Scan(&rcID)
			if err != nil {
				results[idx].Error = err.Error()
				return
			}
			results[idx].ReportCardID = rcID.String()
		}(i, sid)
	}

	wg.Wait()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results":  results,
		"total":    len(req.StudentIDs),
		"period":   req.AcademicPeriod,
	})
}

// ListReportCards returns a student's report card history.
// studentId URL param is a short_id.
func (h *ReportsHandler) ListReportCards(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentParam := chi.URLParam(r, "studentId")
	ctx := r.Context()

	// Resolve student short_id → UUID.
	studentUUID, err := resolveStudentUUID(ctx, h.db, studentParam, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "student not found")
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT id, academic_period, gpa, is_finalized, pdf_url, generated_at
		FROM report_cards
		WHERE student_id = $1 AND school_id = $2
		ORDER BY generated_at DESC
	`, studentUUID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	var reports []models.ReportCard
	for rows.Next() {
		var rc models.ReportCard
		rows.Scan(&rc.ID, &rc.AcademicPeriod, &rc.GPA, &rc.IsFinalized, &rc.PDFURL, &rc.GeneratedAt)
		// Generate fresh download URL.
		if rc.PDFURL != nil {
			if url, err := h.storage.PresignDownload(ctx, *rc.PDFURL); err == nil {
				rc.PDFURL = &url
			}
		}
		reports = append(reports, rc)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"report_cards": reports})
}

func gradePointFromLetter(letter string) float64 {
	gp := map[string]float64{
		"A": 4.0, "A-": 3.7,
		"B+": 3.3, "B": 3.0, "B-": 2.7,
		"C+": 2.3, "C": 2.0, "C-": 1.7,
		"D+": 1.3, "D": 1.0, "D-": 0.7,
		"F": 0.0,
	}
	if gp, ok := gp[letter]; ok {
		return gp
	}
	return 0.0
}
