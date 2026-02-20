package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/models"
)

// CoursesHandler manages course and enrollment CRUD.
type CoursesHandler struct {
	db *pgxpool.Pool
}

// NewCoursesHandler creates a CoursesHandler.
func NewCoursesHandler(db *pgxpool.Pool) *CoursesHandler {
	return &CoursesHandler{db: db}
}

// ListMyCourses returns all courses for the current teacher.
func (h *CoursesHandler) ListMyCourses(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT c.id, c.name, c.subject, c.period, c.room, c.academic_year, c.semester, c.is_active,
		       (SELECT COUNT(*)::int FROM enrollments e WHERE e.course_id = c.id AND e.status = 'active') AS enrollment_count
		FROM courses c
		JOIN teachers t ON t.id = c.teacher_id
		WHERE t.user_id = $1 AND c.school_id = $2
		ORDER BY c.name
	`, claims.UserID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		rows.Scan(&c.ID, &c.Name, &c.Subject, &c.Period, &c.Room,
			&c.AcademicYear, &c.Semester, &c.IsActive, &c.EnrollmentCount)
		courses = append(courses, c)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"courses": courses})
}

// GetEnrolledStudents returns students enrolled in a course.
func (h *CoursesHandler) GetEnrolledStudents(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	courseID := chi.URLParam(r, "courseId")
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT s.id, u.first_name, u.last_name, s.student_number, s.grade_level
		FROM enrollments e
		JOIN students s ON s.id = e.student_id
		JOIN users u ON u.id = s.user_id
		WHERE e.course_id = $1 AND e.status = 'active'
		  AND s.school_id = $2
		ORDER BY u.last_name, u.first_name
	`, courseID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type studentRow struct {
		ID            uuid.UUID `json:"id"`
		FirstName     string    `json:"first_name"`
		LastName      string    `json:"last_name"`
		StudentNumber string    `json:"student_number"`
		GradeLevel    string    `json:"grade_level"`
	}

	var students []studentRow
	for rows.Next() {
		var s studentRow
		rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.StudentNumber, &s.GradeLevel)
		students = append(students, s)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"students": students})
}

// CreateCourse creates a new course (admin only).
func (h *CoursesHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		TeacherID    string  `json:"teacher_id" validate:"required,uuid"`
		Name         string  `json:"name" validate:"required,min=1,max=200"`
		Subject      string  `json:"subject" validate:"required,min=1,max=100"`
		Period       string  `json:"period"`
		Room         string  `json:"room"`
		AcademicYear string  `json:"academic_year" validate:"required"`
		Semester     string  `json:"semester"`
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

	var courseID uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO courses (school_id, teacher_id, name, subject, period, room, academic_year, semester)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, claims.SchoolID, req.TeacherID, req.Name, req.Subject,
		nullStr(req.Period), nullStr(req.Room), req.AcademicYear, nullStr(req.Semester),
	).Scan(&courseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"course_id": courseID})
}
