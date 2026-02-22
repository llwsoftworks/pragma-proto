package handlers

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
)

// StudentsHandler handles student self-service lookups.
type StudentsHandler struct {
	db *pgxpool.Pool
}

// NewStudentsHandler creates a StudentsHandler.
func NewStudentsHandler(db *pgxpool.Pool) *StudentsHandler {
	return &StudentsHandler{db: db}
}

// GetMyRecord returns the students table row for the currently logged-in student.
// Used by the id-card page to resolve the student's DB id from their user JWT.
func (h *StudentsHandler) GetMyRecord(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()

	var s struct {
		ID               string    `json:"id"`
		StudentNumber    string    `json:"student_number"`
		GradeLevel       string    `json:"grade_level"`
		EnrollmentStatus string    `json:"enrollment_status"`
		IsGradeLocked    bool      `json:"is_grade_locked"`
		EnrollmentDate   time.Time `json:"enrollment_date"`
	}

	err := h.db.QueryRow(ctx, `
		SELECT short_id, student_number, grade_level, enrollment_status, is_grade_locked, enrollment_date
		FROM students
		WHERE user_id = $1 AND school_id = $2
	`, claims.UserID, claims.SchoolID).Scan(
		&s.ID, &s.StudentNumber, &s.GradeLevel,
		&s.EnrollmentStatus, &s.IsGradeLocked, &s.EnrollmentDate,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "student record not found")
		return
	}

	writeJSON(w, http.StatusOK, s)
}
