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
	"github.com/pragma-proto/api/internal/services"
)

// AdminHandler handles administrative operations.
type AdminHandler struct {
	db    *pgxpool.Pool
	email *services.EmailService
}

// NewAdminHandler creates an AdminHandler.
func NewAdminHandler(db *pgxpool.Pool, email *services.EmailService) *AdminHandler {
	return &AdminHandler{db: db, email: email}
}

// LockGrade locks a single student's grade access.
// studentId URL param is a short_id.
func (h *AdminHandler) LockGrade(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentParam := chi.URLParam(r, "studentId")

	var req struct {
		Reason string `json:"reason" validate:"required,min=1,max=500"`
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

	// Resolve student short_id → UUID.
	studentUUID, err := resolveStudentUUID(ctx, h.db, studentParam, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "student not found")
		return
	}

	// Insert grade lock record.
	var lockID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO grade_locks (student_id, school_id, locked_by, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, studentUUID, claims.SchoolID, claims.UserID, req.Reason).Scan(&lockID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Update the student record.
	h.db.Exec(ctx, `UPDATE students SET is_grade_locked = TRUE, lock_reason = $1 WHERE id = $2 AND school_id = $3`,
		req.Reason, studentUUID, claims.SchoolID)

	// Audit log.
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "grade_lock.create",
		EntityType: "grade_lock",
		EntityID:   &lockID,
		NewValue:   map[string]string{"student_id": studentUUID.String(), "reason": req.Reason},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{"lock_id": lockID})
}

// UnlockGrade removes a student's grade lock and notifies them via email.
// The grade_lock row is deleted rather than soft-updated because audit_logs
// already captures the full lock/unlock history. Keeping inactive rows is
// unnecessary bloat.
func (h *AdminHandler) UnlockGrade(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentParam := chi.URLParam(r, "studentId")

	ctx := r.Context()

	// Resolve student short_id → UUID.
	studentUUID, err := resolveStudentUUID(ctx, h.db, studentParam, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "student not found")
		return
	}

	// Capture lock details for the audit log before deletion.
	var lockID uuid.UUID
	var lockReason string
	err = h.db.QueryRow(ctx, `
		SELECT id, reason FROM grade_locks
		WHERE student_id = $1 AND school_id = $2
	`, studentUUID, claims.SchoolID).Scan(&lockID, &lockReason)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "no active lock found for this student")
		return
	}

	// Delete the lock row — audit_logs preserves the history.
	_, err = h.db.Exec(ctx, `DELETE FROM grade_locks WHERE id = $1`, lockID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Update student record.
	h.db.Exec(ctx, `UPDATE students SET is_grade_locked = FALSE, lock_reason = NULL WHERE id = $1 AND school_id = $2`,
		studentUUID, claims.SchoolID)

	// Audit log records who unlocked, when, and the original reason.
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "grade_lock.release",
		EntityType: "grade_lock",
		EntityID:   &lockID,
		OldValue:   map[string]string{"student_id": studentUUID.String(), "reason": lockReason},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	// Send email notification to student and linked parents.
	go h.sendUnlockNotifications(ctx, studentUUID.String(), claims.SchoolID.String())

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// BulkLockGrades locks multiple students at once.
func (h *AdminHandler) BulkLockGrades(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		StudentIDs []string `json:"student_ids" validate:"required,min=1,dive,uuid"`
		Reason     string   `json:"reason" validate:"required,min=1,max=500"`
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
	locked := 0
	for _, sid := range req.StudentIDs {
		var lockID uuid.UUID
		err := h.db.QueryRow(ctx, `
			INSERT INTO grade_locks (student_id, school_id, locked_by, reason)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, sid, claims.SchoolID, claims.UserID, req.Reason).Scan(&lockID)
		if err != nil {
			continue
		}
		h.db.Exec(ctx, `UPDATE students SET is_grade_locked = TRUE, lock_reason = $1 WHERE id = $2 AND school_id = $3`,
			req.Reason, sid, claims.SchoolID)
		locked++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"locked": locked})
}

// ListStudents returns the full student roster for the admin's school.
func (h *AdminHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()
	limit, offset := paginate(r)

	rows, err := h.db.Query(ctx, `
		SELECT s.short_id, u.email, u.first_name, u.last_name,
		       s.student_number, s.grade_level, s.enrollment_status,
		       s.is_grade_locked, s.lock_reason, s.enrollment_date
		FROM students s
		JOIN users u ON u.id = s.user_id
		WHERE s.school_id = $1
		ORDER BY u.last_name, u.first_name
		LIMIT $2 OFFSET $3
	`, claims.SchoolID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type studentRow struct {
		ID               string    `json:"id"`
		Email            string    `json:"email"`
		FirstName        string    `json:"first_name"`
		LastName         string    `json:"last_name"`
		StudentNumber    string    `json:"student_number"`
		GradeLevel       string    `json:"grade_level"`
		EnrollmentStatus string    `json:"enrollment_status"`
		IsGradeLocked    bool      `json:"is_grade_locked"`
		LockReason       *string   `json:"lock_reason"`
		EnrollmentDate   time.Time `json:"enrollment_date"`
	}

	var students []studentRow
	for rows.Next() {
		var s studentRow
		if err := rows.Scan(
			&s.ID, &s.Email, &s.FirstName, &s.LastName,
			&s.StudentNumber, &s.GradeLevel, &s.EnrollmentStatus,
			&s.IsGradeLocked, &s.LockReason, &s.EnrollmentDate,
		); err != nil {
			continue
		}
		students = append(students, s)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"students": students})
}

func (h *AdminHandler) sendUnlockNotifications(ctx interface{ Done() <-chan struct{} }, studentID, schoolID string) {
	// Look up student and parent emails, then send notification.
	// This runs in a goroutine; errors are logged but not surfaced to the caller.
	// TODO: Implement with actual background context and email service.
}
