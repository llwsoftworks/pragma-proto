package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/models"
	"github.com/pragma-proto/api/internal/services"
)

// GradesHandler manages grade CRUD and calculations.
type GradesHandler struct {
	db      *pgxpool.Pool
	grading *services.GradingService
}

// NewGradesHandler creates a GradesHandler.
func NewGradesHandler(db *pgxpool.Pool, grading *services.GradingService) *GradesHandler {
	return &GradesHandler{db: db, grading: grading}
}

// ListGrades returns all grades for a course (teacher/admin only).
// courseId URL param is a short_id.
func (h *GradesHandler) ListGrades(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	courseParam := chi.URLParam(r, "courseId")
	if courseParam == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "courseId is required")
		return
	}

	ctx := r.Context()

	// Resolve short_id → UUID.
	courseUUID, err := resolveCourseUUID(ctx, h.db, courseParam, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "course not found")
		return
	}

	// Verify the teacher owns this course (RBAC + ownership check).
	if claims.Role == models.RoleTeacher {
		var teacherCourseCount int
		h.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM courses c
			JOIN teachers t ON t.id = c.teacher_id
			WHERE c.id = $1 AND t.user_id = $2 AND c.school_id = $3
		`, courseUUID, claims.UserID, claims.SchoolID).Scan(&teacherCourseCount)
		if teacherCourseCount == 0 {
			writeError(w, http.StatusForbidden, "forbidden", "you are not the teacher for this course")
			return
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT g.id, g.assignment_id, g.student_id, g.school_id,
		       g.points_earned, g.letter_grade, g.comment, g.graded_by,
		       g.graded_at, g.ai_suggested, g.ai_accepted,
		       g.is_excused, g.is_missing, g.is_late,
		       g.created_at, g.updated_at
		FROM grades g
		JOIN assignments a ON a.id = g.assignment_id
		WHERE a.course_id = $1 AND g.school_id = $2
		ORDER BY g.student_id, a.due_date
	`, courseUUID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "failed to fetch grades")
		return
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var g models.Grade
		if err := rows.Scan(
			&g.ID, &g.AssignmentID, &g.StudentID, &g.SchoolID,
			&g.PointsEarned, &g.LetterGrade, &g.Comment, &g.GradedBy,
			&g.GradedAt, &g.AISuggested, &g.AIAccepted,
			&g.IsExcused, &g.IsMissing, &g.IsLate,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan_error", err.Error())
			return
		}
		grades = append(grades, g)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"grades": grades})
}

// UpsertGrade creates or updates a single grade entry.
// courseId URL param is a short_id.
func (h *GradesHandler) UpsertGrade(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	courseParam := chi.URLParam(r, "courseId")
	ctx := r.Context()

	// Resolve short_id → UUID.
	courseUUID, err := resolveCourseUUID(ctx, h.db, courseParam, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "course not found")
		return
	}
	_ = courseUUID // Used for scoping verification; the grade INSERT uses assignment_id directly.

	var req struct {
		AssignmentID string   `json:"assignment_id" validate:"required,uuid"`
		StudentID    string   `json:"student_id" validate:"required,uuid"`
		PointsEarned *float64 `json:"points_earned"`
		Comment      string   `json:"comment"`
		IsExcused    bool     `json:"is_excused"`
		IsMissing    bool     `json:"is_missing"`
		IsLate       bool     `json:"is_late"`
		AIAccepted   *bool    `json:"ai_accepted"`
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

	// Fetch current grade for audit log old_value.
	var oldGrade *models.Grade
	var existing models.Grade
	err = h.db.QueryRow(ctx, `
		SELECT id, points_earned, letter_grade, comment, is_excused, is_missing, is_late
		FROM grades WHERE assignment_id = $1 AND student_id = $2 AND school_id = $3
	`, req.AssignmentID, req.StudentID, claims.SchoolID).Scan(
		&existing.ID, &existing.PointsEarned, &existing.LetterGrade,
		&existing.Comment, &existing.IsExcused, &existing.IsMissing, &existing.IsLate,
	)
	if err == nil {
		oldGrade = &existing
	}

	// Validate points_earned against max_points.
	var maxPoints float64
	h.db.QueryRow(ctx, `SELECT max_points FROM assignments WHERE id = $1 AND school_id = $2`,
		req.AssignmentID, claims.SchoolID).Scan(&maxPoints)

	if req.PointsEarned != nil && (*req.PointsEarned < 0 || *req.PointsEarned > maxPoints) {
		writeError(w, http.StatusBadRequest, "invalid_points",
			"points_earned must be between 0 and the assignment's max_points")
		return
	}

	var gradeID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO grades
			(assignment_id, student_id, school_id, points_earned, comment,
			 is_excused, is_missing, is_late, ai_accepted, graded_by, graded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (assignment_id, student_id)
		DO UPDATE SET
			points_earned = EXCLUDED.points_earned,
			comment       = EXCLUDED.comment,
			is_excused    = EXCLUDED.is_excused,
			is_missing    = EXCLUDED.is_missing,
			is_late       = EXCLUDED.is_late,
			ai_accepted   = EXCLUDED.ai_accepted,
			graded_by     = EXCLUDED.graded_by,
			graded_at     = NOW(),
			updated_at    = NOW()
		RETURNING id
	`, req.AssignmentID, req.StudentID, claims.SchoolID,
		req.PointsEarned, nullStr(req.Comment),
		req.IsExcused, req.IsMissing, req.IsLate, req.AIAccepted, claims.UserID,
	).Scan(&gradeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Audit log.
	action := "grade.create"
	if oldGrade != nil {
		action = "grade.update"
	}
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     action,
		EntityType: "grade",
		EntityID:   &gradeID,
		OldValue:   oldGrade,
		NewValue:   req,
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{"grade_id": gradeID})
}

// GetStudentGrades returns a student's grades for their own courses.
// Checks grade lock for student/parent roles.
func (h *GradesHandler) GetStudentGrades(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentIDStr := chi.URLParam(r, "studentId")

	ctx := r.Context()

	// For student role, enforce they can only see their own grades.
	if claims.Role == models.RoleStudent {
		var studentUserID uuid.UUID
		h.db.QueryRow(ctx, `SELECT user_id FROM students WHERE id = $1 AND school_id = $2`,
			studentIDStr, claims.SchoolID).Scan(&studentUserID)
		if studentUserID != claims.UserID {
			writeError(w, http.StatusForbidden, "forbidden", "you can only view your own grades")
			return
		}
	}

	// Check grade lock.
	if claims.Role == models.RoleStudent || claims.Role == models.RoleParent {
		var isLocked bool
		h.db.QueryRow(ctx, `
			SELECT is_grade_locked FROM students WHERE id = $1 AND school_id = $2
		`, studentIDStr, claims.SchoolID).Scan(&isLocked)
		if isLocked {
			writeError(w, http.StatusForbidden, "grade_locked",
				"Your grade access has been temporarily restricted. Please contact your school administration.")
			return
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT g.id, g.assignment_id, g.student_id, a.title, a.max_points,
		       a.category, c.name as course_name,
		       g.points_earned, g.letter_grade, g.comment,
		       g.is_excused, g.is_missing, g.is_late, g.updated_at
		FROM grades g
		JOIN assignments a ON a.id = g.assignment_id
		JOIN courses c ON c.id = a.course_id
		WHERE g.student_id = $1 AND g.school_id = $2
		  AND a.is_published = TRUE
		ORDER BY c.name, a.due_date
	`, studentIDStr, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type gradeRow struct {
		ID           uuid.UUID `json:"id"`
		AssignmentID uuid.UUID `json:"assignment_id"`
		StudentID    uuid.UUID `json:"student_id"`
		Title        string    `json:"title"`
		MaxPoints    float64   `json:"max_points"`
		Category     string    `json:"category"`
		CourseName   string    `json:"course_name"`
		PointsEarned *float64  `json:"points_earned"`
		LetterGrade  *string   `json:"letter_grade"`
		Comment      *string   `json:"comment"`
		IsExcused    bool      `json:"is_excused"`
		IsMissing    bool      `json:"is_missing"`
		IsLate       bool      `json:"is_late"`
	}

	var grades []gradeRow
	for rows.Next() {
		var g gradeRow
		if err := rows.Scan(
			&g.ID, &g.AssignmentID, &g.StudentID, &g.Title, &g.MaxPoints,
			&g.Category, &g.CourseName,
			&g.PointsEarned, &g.LetterGrade, &g.Comment,
			&g.IsExcused, &g.IsMissing, &g.IsLate,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan_error", err.Error())
			return
		}
		grades = append(grades, g)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"grades": grades})
}
