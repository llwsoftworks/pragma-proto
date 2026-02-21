package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/models"
	"github.com/pragma-proto/api/internal/services"
)

// AssignmentsHandler manages assignment and attachment CRUD.
type AssignmentsHandler struct {
	db      *pgxpool.Pool
	storage *services.StorageService
}

// NewAssignmentsHandler creates an AssignmentsHandler.
func NewAssignmentsHandler(db *pgxpool.Pool, storage *services.StorageService) *AssignmentsHandler {
	return &AssignmentsHandler{db: db, storage: storage}
}

// ListAssignments returns assignments for the teacher's courses (or all for admin).
func (h *AssignmentsHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()

	type assignmentRow struct {
		ID          uuid.UUID  `json:"id"`
		CourseID    uuid.UUID  `json:"course_id"`
		CourseName  string     `json:"course_name"`
		Title       string     `json:"title"`
		Description *string    `json:"description,omitempty"`
		DueDate     *time.Time `json:"due_date,omitempty"`
		MaxPoints   float64    `json:"max_points"`
		Category    string     `json:"category"`
		Weight      float64    `json:"weight"`
		IsPublished bool       `json:"is_published"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}

	// Build query; teachers see only their own courses, admins see all.
	var (
		query string
		args  []interface{}
	)
	if claims.Role == models.RoleTeacher {
		query = `
			SELECT a.id, a.course_id, c.name,
			       a.title, a.description, a.due_date, a.max_points,
			       a.category, a.weight, a.is_published, a.created_at, a.updated_at
			FROM assignments a
			JOIN courses c ON c.id = a.course_id
			JOIN teachers t ON t.id = c.teacher_id
			WHERE t.user_id = $1 AND a.school_id = $2
			ORDER BY a.due_date DESC NULLS LAST, a.created_at DESC`
		args = []interface{}{claims.UserID, claims.SchoolID}
	} else {
		query = `
			SELECT a.id, a.course_id, c.name,
			       a.title, a.description, a.due_date, a.max_points,
			       a.category, a.weight, a.is_published, a.created_at, a.updated_at
			FROM assignments a
			JOIN courses c ON c.id = a.course_id
			WHERE a.school_id = $1
			ORDER BY a.due_date DESC NULLS LAST, a.created_at DESC`
		args = []interface{}{claims.SchoolID}
	}

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	var assignments []assignmentRow
	for rows.Next() {
		var a assignmentRow
		if err := rows.Scan(
			&a.ID, &a.CourseID, &a.CourseName,
			&a.Title, &a.Description, &a.DueDate,
			&a.MaxPoints, &a.Category, &a.Weight,
			&a.IsPublished, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan_error", err.Error())
			return
		}
		assignments = append(assignments, a)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"assignments": assignments})
}

// CreateAssignment creates a new assignment in a course.
func (h *AssignmentsHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		CourseID    string   `json:"course_id" validate:"required,uuid"`
		Title       string   `json:"title" validate:"required,min=1,max=300"`
		Description string   `json:"description"`
		DueDate     *string  `json:"due_date"`
		MaxPoints   float64  `json:"max_points" validate:"required,min=0"`
		Category    string   `json:"category" validate:"required,oneof=homework quiz test exam project classwork participation other"`
		Weight      float64  `json:"weight" validate:"min=0,max=1"`
		IsPublished bool     `json:"is_published"`
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
	if req.Weight == 0 {
		req.Weight = 1.0
	}

	ctx := r.Context()

	// Verify teacher owns this course.
	var teacherCount int
	h.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM courses c JOIN teachers t ON t.id = c.teacher_id
		WHERE c.id = $1 AND t.user_id = $2 AND c.school_id = $3
	`, req.CourseID, claims.UserID, claims.SchoolID).Scan(&teacherCount)
	if teacherCount == 0 && claims.Role == models.RoleTeacher {
		writeError(w, http.StatusForbidden, "forbidden", "you are not the teacher for this course")
		return
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_due_date", "due_date must be RFC3339 format")
			return
		}
		dueDate = &t
	}

	var assignmentID uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO assignments
			(course_id, school_id, title, description, due_date, max_points, category, weight, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, req.CourseID, claims.SchoolID, req.Title, nullStr(req.Description),
		dueDate, req.MaxPoints, req.Category, req.Weight, req.IsPublished,
	).Scan(&assignmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "assignment.create",
		EntityType: "assignment",
		EntityID:   &assignmentID,
		NewValue:   req,
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{"assignment_id": assignmentID})
}

// ListCourseAssignments returns all assignments for a specific course.
// Teachers see all (published + draft) for courses they own; admins see all.
func (h *AssignmentsHandler) ListCourseAssignments(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	courseID := chi.URLParam(r, "courseId")
	if courseID == "" {
		writeError(w, http.StatusBadRequest, "missing_param", "courseId is required")
		return
	}
	ctx := r.Context()

	// Verify teacher owns this course.
	if claims.Role == models.RoleTeacher {
		var count int
		h.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM courses c
			JOIN teachers t ON t.id = c.teacher_id
			WHERE c.id = $1 AND t.user_id = $2 AND c.school_id = $3
		`, courseID, claims.UserID, claims.SchoolID).Scan(&count)
		if count == 0 {
			writeError(w, http.StatusForbidden, "forbidden", "you are not the teacher for this course")
			return
		}
	}

	rows, err := h.db.Query(ctx, `
		SELECT a.id, a.course_id, a.title, a.description, a.due_date,
		       a.max_points, a.category, a.weight, a.is_published,
		       a.created_at, a.updated_at
		FROM assignments a
		WHERE a.course_id = $1 AND a.school_id = $2
		ORDER BY a.due_date DESC NULLS LAST, a.created_at DESC
	`, courseID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type assignmentRow struct {
		ID          uuid.UUID  `json:"id"`
		CourseID    uuid.UUID  `json:"course_id"`
		Title       string     `json:"title"`
		Description *string    `json:"description,omitempty"`
		DueDate     *time.Time `json:"due_date,omitempty"`
		MaxPoints   float64    `json:"max_points"`
		Category    string     `json:"category"`
		Weight      float64    `json:"weight"`
		IsPublished bool       `json:"is_published"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}

	var assignments []assignmentRow
	for rows.Next() {
		var a assignmentRow
		if err := rows.Scan(
			&a.ID, &a.CourseID, &a.Title, &a.Description, &a.DueDate,
			&a.MaxPoints, &a.Category, &a.Weight, &a.IsPublished,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan_error", err.Error())
			return
		}
		assignments = append(assignments, a)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"assignments": assignments})
}

// RequestUploadURL generates a presigned R2 upload URL for a file attachment.
func (h *AssignmentsHandler) RequestUploadURL(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	assignmentID := chi.URLParam(r, "assignmentId")

	var req struct {
		FileName    string `json:"file_name" validate:"required,min=1,max=255"`
		MIMEType    string `json:"mime_type" validate:"required"`
		FileSizeBytes int64 `json:"file_size_bytes" validate:"required,min=1,max=26214400"` // 25MB
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

	if err := services.ValidateMIMEType(req.MIMEType); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_mime_type", err.Error())
		return
	}

	ctx := r.Context()
	fileID := uuid.New()
	key := services.ObjectKey(claims.SchoolID.String(), "attachments",
		assignmentID+"/"+fileID.String()+"-"+req.FileName)

	url, err := h.storage.PresignUpload(ctx, key, req.FileSizeBytes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Pre-register the attachment metadata (confirmed after upload).
	var attachID uuid.UUID
	h.db.QueryRow(ctx, `
		INSERT INTO assignment_attachments
			(assignment_id, school_id, file_name, file_key, file_size, mime_type, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, assignmentID, claims.SchoolID, req.FileName, key, req.FileSizeBytes, req.MIMEType, claims.UserID,
	).Scan(&attachID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"upload_url":    url,
		"attachment_id": attachID,
		"file_key":      key,
	})
}

// ListAttachments returns current attachments for an assignment.
func (h *AssignmentsHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	assignmentID := chi.URLParam(r, "assignmentId")
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT id, file_name, file_key, file_size, mime_type, version, created_at
		FROM assignment_attachments
		WHERE assignment_id = $1 AND school_id = $2 AND is_current = TRUE
		ORDER BY created_at DESC
	`, assignmentID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type attRow struct {
		ID        uuid.UUID `json:"id"`
		FileName  string    `json:"file_name"`
		FileKey   string    `json:"-"`
		FileSize  int64     `json:"file_size"`
		MIMEType  string    `json:"mime_type"`
		Version   int       `json:"version"`
		CreatedAt time.Time `json:"created_at"`
		DownloadURL string  `json:"download_url,omitempty"`
	}

	var attachments []attRow
	for rows.Next() {
		var a attRow
		if err := rows.Scan(&a.ID, &a.FileName, &a.FileKey, &a.FileSize, &a.MIMEType, &a.Version, &a.CreatedAt); err != nil {
			continue
		}
		// Generate presigned download URL.
		if url, err := h.storage.PresignDownload(ctx, a.FileKey); err == nil {
			a.DownloadURL = url
		}
		attachments = append(attachments, a)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"attachments": attachments})
}
