package handlers

import (
	"encoding/json"
	"fmt"
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

// SuperAdminHandler handles platform-level operations exclusive to super_admin.
// These routes bypass TenantMiddleware — they operate across all schools.
type SuperAdminHandler struct {
	db    *pgxpool.Pool
	email *services.EmailService
}

// NewSuperAdminHandler creates a SuperAdminHandler.
func NewSuperAdminHandler(db *pgxpool.Pool, email *services.EmailService) *SuperAdminHandler {
	return &SuperAdminHandler{db: db, email: email}
}

// ---------- Platform Stats ----------

// GetPlatformStats returns aggregate stats across all schools.
func (h *SuperAdminHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var totalSchools, totalUsers, totalStudents, totalTeachers, totalLockedStudents int

	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM schools`).Scan(&totalSchools)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = TRUE`).Scan(&totalUsers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE enrollment_status = 'active'`).Scan(&totalStudents)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM teachers`).Scan(&totalTeachers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE is_grade_locked = TRUE`).Scan(&totalLockedStudents)

	// Per-school summary.
	rows, err := h.db.Query(ctx, `
		SELECT s.id, s.name,
		       (SELECT COUNT(*) FROM users u WHERE u.school_id = s.id AND u.is_active = TRUE)::int AS user_count,
		       (SELECT COUNT(*) FROM students st WHERE st.school_id = s.id AND st.enrollment_status = 'active')::int AS student_count,
		       (SELECT COUNT(*) FROM teachers t WHERE t.school_id = s.id)::int AS teacher_count,
		       s.created_at
		FROM schools s
		ORDER BY s.name
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type schoolSummary struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		UserCount    int       `json:"user_count"`
		StudentCount int       `json:"student_count"`
		TeacherCount int       `json:"teacher_count"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var schools []schoolSummary
	for rows.Next() {
		var s schoolSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.UserCount, &s.StudentCount, &s.TeacherCount, &s.CreatedAt); err != nil {
			continue
		}
		schools = append(schools, s)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":                 "super_admin",
		"total_schools":        totalSchools,
		"total_users":          totalUsers,
		"total_students":       totalStudents,
		"total_teachers":       totalTeachers,
		"total_locked_students": totalLockedStudents,
		"schools":              schools,
	})
}

// ---------- School CRUD ----------

// ListSchools returns all schools on the platform.
func (h *SuperAdminHandler) ListSchools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, offset := paginate(r)

	rows, err := h.db.Query(ctx, `
		SELECT s.id, s.name, s.address, s.logo_url, s.settings, s.created_at, s.updated_at,
		       (SELECT COUNT(*) FROM users u WHERE u.school_id = s.id AND u.is_active = TRUE)::int AS user_count,
		       (SELECT COUNT(*) FROM students st WHERE st.school_id = s.id AND st.enrollment_status = 'active')::int AS student_count
		FROM schools s
		ORDER BY s.name
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type schoolRow struct {
		ID           uuid.UUID        `json:"id"`
		Name         string           `json:"name"`
		Address      *string          `json:"address"`
		LogoURL      *string          `json:"logo_url"`
		Settings     json.RawMessage  `json:"settings"`
		CreatedAt    time.Time        `json:"created_at"`
		UpdatedAt    time.Time        `json:"updated_at"`
		UserCount    int              `json:"user_count"`
		StudentCount int              `json:"student_count"`
	}

	var schools []schoolRow
	for rows.Next() {
		var s schoolRow
		if err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.LogoURL, &s.Settings,
			&s.CreatedAt, &s.UpdatedAt, &s.UserCount, &s.StudentCount); err != nil {
			continue
		}
		schools = append(schools, s)
	}

	var total int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM schools`).Scan(&total)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"schools": schools,
		"total":   total,
	})
}

// CreateSchool creates a new school on the platform.
func (h *SuperAdminHandler) CreateSchool(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		Name    string  `json:"name" validate:"required,min=1,max=300"`
		Address *string `json:"address" validate:"omitempty,max=500"`
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
	var schoolID uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO schools (name, address) VALUES ($1, $2) RETURNING id
	`, req.Name, req.Address).Scan(&schoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   schoolID,
		UserID:     &claims.UserID,
		Action:     "school.create",
		EntityType: "school",
		EntityID:   &schoolID,
		NewValue:   map[string]interface{}{"name": req.Name, "address": req.Address},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{"school_id": schoolID})
}

// GetSchool returns a single school's details.
func (h *SuperAdminHandler) GetSchool(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	ctx := r.Context()

	var school struct {
		ID        uuid.UUID       `json:"id"`
		Name      string          `json:"name"`
		Address   *string         `json:"address"`
		LogoURL   *string         `json:"logo_url"`
		Settings  json.RawMessage `json:"settings"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}

	err := h.db.QueryRow(ctx, `
		SELECT id, name, address, logo_url, settings, created_at, updated_at
		FROM schools WHERE id = $1
	`, schoolID).Scan(&school.ID, &school.Name, &school.Address, &school.LogoURL,
		&school.Settings, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "school not found")
		return
	}

	// Counts for this school.
	var totalUsers, totalStudents, totalTeachers, lockedStudents, totalCourses int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE school_id = $1 AND is_active = TRUE`, schoolID).Scan(&totalUsers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE school_id = $1 AND enrollment_status = 'active'`, schoolID).Scan(&totalStudents)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM teachers WHERE school_id = $1`, schoolID).Scan(&totalTeachers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE school_id = $1 AND is_grade_locked = TRUE`, schoolID).Scan(&lockedStudents)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM courses WHERE school_id = $1 AND is_active = TRUE`, schoolID).Scan(&totalCourses)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"school":          school,
		"total_users":     totalUsers,
		"total_students":  totalStudents,
		"total_teachers":  totalTeachers,
		"locked_students": lockedStudents,
		"total_courses":   totalCourses,
	})
}

// UpdateSchool updates a school's name, address, or settings.
func (h *SuperAdminHandler) UpdateSchool(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	schoolID := chi.URLParam(r, "schoolId")

	var req struct {
		Name     *string         `json:"name" validate:"omitempty,min=1,max=300"`
		Address  *string         `json:"address" validate:"omitempty,max=500"`
		LogoURL  *string         `json:"logo_url"`
		Settings json.RawMessage `json:"settings"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := r.Context()

	// Read current state for audit log.
	var oldName string
	var oldAddress *string
	h.db.QueryRow(ctx, `SELECT name, address FROM schools WHERE id = $1`, schoolID).Scan(&oldName, &oldAddress)

	// Build dynamic update. Only update provided fields.
	if req.Name != nil {
		h.db.Exec(ctx, `UPDATE schools SET name = $1, updated_at = NOW() WHERE id = $2`, *req.Name, schoolID)
	}
	if req.Address != nil {
		h.db.Exec(ctx, `UPDATE schools SET address = $1, updated_at = NOW() WHERE id = $2`, *req.Address, schoolID)
	}
	if req.LogoURL != nil {
		h.db.Exec(ctx, `UPDATE schools SET logo_url = $1, updated_at = NOW() WHERE id = $2`, *req.LogoURL, schoolID)
	}
	if req.Settings != nil {
		h.db.Exec(ctx, `UPDATE schools SET settings = $1, updated_at = NOW() WHERE id = $2`, req.Settings, schoolID)
	}

	schoolUUID, _ := uuid.Parse(schoolID)
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   schoolUUID,
		UserID:     &claims.UserID,
		Action:     "school.update",
		EntityType: "school",
		EntityID:   &schoolUUID,
		OldValue:   map[string]interface{}{"name": oldName, "address": oldAddress},
		NewValue:   req,
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DeleteSchool soft-deletes a school by deactivating all its users.
// Schools are not physically deleted to preserve audit trails.
func (h *SuperAdminHandler) DeleteSchool(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	schoolID := chi.URLParam(r, "schoolId")
	ctx := r.Context()

	// Deactivate all users in the school.
	tag, err := h.db.Exec(ctx, `UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE school_id = $1`, schoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Invalidate all sessions for this school.
	h.db.Exec(ctx, `DELETE FROM sessions WHERE school_id = $1`, schoolID)

	schoolUUID, _ := uuid.Parse(schoolID)
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   schoolUUID,
		UserID:     &claims.UserID,
		Action:     "school.deactivate",
		EntityType: "school",
		EntityID:   &schoolUUID,
		NewValue:   map[string]interface{}{"users_deactivated": tag.RowsAffected()},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                 true,
		"users_deactivated":  tag.RowsAffected(),
	})
}

// ---------- User Management (Cross-School) ----------

// ListSchoolUsers returns users for a specific school.
func (h *SuperAdminHandler) ListSchoolUsers(w http.ResponseWriter, r *http.Request) {
	schoolID := chi.URLParam(r, "schoolId")
	ctx := r.Context()
	limit, offset := paginate(r)

	// Optional role filter.
	roleFilter := r.URL.Query().Get("role")

	type userRow struct {
		ID          uuid.UUID  `json:"id"`
		Email       string     `json:"email"`
		Role        string     `json:"role"`
		FirstName   string     `json:"first_name"`
		LastName    string     `json:"last_name"`
		IsActive    bool       `json:"is_active"`
		MFAEnabled  bool       `json:"mfa_enabled"`
		LastLoginAt *time.Time `json:"last_login_at"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	var users []userRow

	if roleFilter != "" {
		rows, err := h.db.Query(ctx, `
			SELECT id, email, role, first_name, last_name, is_active, mfa_enabled, last_login_at, created_at
			FROM users
			WHERE school_id = $1 AND role = $2
			ORDER BY last_name, first_name
			LIMIT $3 OFFSET $4
		`, schoolID, roleFilter, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		defer rows.Close()
		for rows.Next() {
			var u userRow
			if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.FirstName, &u.LastName,
				&u.IsActive, &u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
	} else {
		rows, err := h.db.Query(ctx, `
			SELECT id, email, role, first_name, last_name, is_active, mfa_enabled, last_login_at, created_at
			FROM users
			WHERE school_id = $1
			ORDER BY role, last_name, first_name
			LIMIT $2 OFFSET $3
		`, schoolID, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		defer rows.Close()
		for rows.Next() {
			var u userRow
			if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.FirstName, &u.LastName,
				&u.IsActive, &u.MFAEnabled, &u.LastLoginAt, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})
}

// CreateSuperAdmin creates a new super_admin account.
// Super-admins are not tied to any school (school_id is NULL).
func (h *SuperAdminHandler) CreateSuperAdmin(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=12"`
		FirstName string `json:"first_name" validate:"required,min=1,max=100"`
		LastName  string `json:"last_name" validate:"required,min=1,max=100"`
		Phone     string `json:"phone"`
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

	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "weak_password", err.Error())
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash_error", "")
		return
	}

	ctx := r.Context()
	var userID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO users (school_id, role, email, password_hash, first_name, last_name, phone)
		VALUES (NULL, 'super_admin', $1, $2, $3, $4, $5)
		RETURNING id
	`, req.Email, hash, req.FirstName, req.LastName, nullStr(req.Phone)).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "email_exists", "a super-admin with this email already exists")
		return
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   uuid.Nil,
		UserID:     &claims.UserID,
		Action:     "user.create",
		EntityType: "user",
		EntityID:   &userID,
		NewValue:   map[string]string{"email": req.Email, "role": "super_admin"},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{"user_id": userID})
}

// CreateSchoolUser creates a user within a specific school.
// Cannot create super_admin users — use CreateSuperAdmin for that.
func (h *SuperAdminHandler) CreateSchoolUser(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	schoolID := chi.URLParam(r, "schoolId")

	var req struct {
		Role      string `json:"role" validate:"required,oneof=admin teacher parent student"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=12"`
		FirstName string `json:"first_name" validate:"required,min=1,max=100"`
		LastName  string `json:"last_name" validate:"required,min=1,max=100"`
		Phone     string `json:"phone"`
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

	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "weak_password", err.Error())
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash_error", "")
		return
	}

	ctx := r.Context()
	var userID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO users (school_id, role, email, password_hash, first_name, last_name, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, schoolID, req.Role, req.Email, hash, req.FirstName, req.LastName, nullStr(req.Phone)).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "email_exists", "email already registered at this school")
		return
	}

	// If creating a student or teacher, also create the role-specific record.
	schoolUUID, _ := uuid.Parse(schoolID)
	if req.Role == models.RoleStudent {
		h.db.Exec(ctx, `
			INSERT INTO students (user_id, school_id, student_number, grade_level, enrollment_date, short_id)
			VALUES ($1, $2, $3, $4, NOW(), left(md5(gen_random_uuid()::text), 8))
		`, userID, schoolID, "PENDING", "Unassigned")
	} else if req.Role == models.RoleTeacher {
		h.db.Exec(ctx, `
			INSERT INTO teachers (user_id, school_id) VALUES ($1, $2)
		`, userID, schoolID)
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   schoolUUID,
		UserID:     &claims.UserID,
		Action:     "user.create",
		EntityType: "user",
		EntityID:   &userID,
		NewValue:   map[string]string{"email": req.Email, "role": req.Role, "school_id": schoolID},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{"user_id": userID})
}

// UpdateUserStatus activates or deactivates a user.
func (h *SuperAdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	var req struct {
		IsActive bool `json:"is_active"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := r.Context()

	// Prevent super-admin from deactivating themselves.
	if userID == claims.UserID.String() && !req.IsActive {
		writeError(w, http.StatusBadRequest, "self_deactivation", "cannot deactivate your own account")
		return
	}

	tag, err := h.db.Exec(ctx, `UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`, req.IsActive, userID)
	if err != nil || tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	// If deactivating, invalidate sessions.
	if !req.IsActive {
		h.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	}

	targetUUID, _ := uuid.Parse(userID)
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "user.status_update",
		EntityType: "user",
		EntityID:   &targetUUID,
		NewValue:   map[string]interface{}{"is_active": req.IsActive},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ---------- Platform Audit Logs ----------

// ListPlatformAuditLogs returns audit logs across all schools, with optional filters.
func (h *SuperAdminHandler) ListPlatformAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, offset := paginate(r)

	// Optional filters.
	schoolFilter := r.URL.Query().Get("school_id")
	actionFilter := r.URL.Query().Get("action")

	query := `
		SELECT al.id, al.school_id, s.name AS school_name,
		       al.user_id, COALESCE(u.email, '') AS user_email,
		       al.action, al.entity_type, al.entity_id,
		       al.old_value, al.new_value,
		       al.ip_address, al.user_agent, al.created_at
		FROM audit_logs al
		LEFT JOIN schools s ON s.id = al.school_id
		LEFT JOIN users u ON u.id = al.user_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if schoolFilter != "" {
		query += ` AND al.school_id = $` + itoa(argIdx)
		args = append(args, schoolFilter)
		argIdx++
	}
	if actionFilter != "" {
		query += ` AND al.action LIKE $` + itoa(argIdx)
		args = append(args, actionFilter+"%")
		argIdx++
	}

	query += ` ORDER BY al.created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	type auditRow struct {
		ID         uuid.UUID       `json:"id"`
		SchoolID   uuid.UUID       `json:"school_id"`
		SchoolName string          `json:"school_name"`
		UserID     *uuid.UUID      `json:"user_id"`
		UserEmail  string          `json:"user_email"`
		Action     string          `json:"action"`
		EntityType string          `json:"entity_type"`
		EntityID   *uuid.UUID      `json:"entity_id"`
		OldValue   json.RawMessage `json:"old_value"`
		NewValue   json.RawMessage `json:"new_value"`
		IPAddress  *string         `json:"ip_address"`
		UserAgent  *string         `json:"user_agent"`
		CreatedAt  time.Time       `json:"created_at"`
	}

	var logs []auditRow
	for rows.Next() {
		var l auditRow
		if err := rows.Scan(&l.ID, &l.SchoolID, &l.SchoolName,
			&l.UserID, &l.UserEmail,
			&l.Action, &l.EntityType, &l.EntityID,
			&l.OldValue, &l.NewValue,
			&l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"audit_logs": logs})
}

// itoa converts int to string for building dynamic SQL arg placeholders.
func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
