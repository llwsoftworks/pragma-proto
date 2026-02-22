package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/models"
)

// DashboardHandler aggregates data for role-specific dashboards.
type DashboardHandler struct {
	db *pgxpool.Pool
}

// NewDashboardHandler creates a DashboardHandler.
func NewDashboardHandler(db *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetDashboard returns the appropriate dashboard data based on the user's role.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()

	switch claims.Role {
	case models.RoleTeacher:
		h.teacherDashboard(w, r, claims)
	case models.RoleParent:
		h.parentDashboard(w, r, claims)
	case models.RoleStudent:
		h.studentDashboard(w, r, claims)
	case models.RoleSuperAdmin:
		h.superAdminDashboard(w, r, claims)
	case models.RoleAdmin:
		h.adminDashboard(w, r, claims)
	default:
		writeError(w, http.StatusForbidden, "unknown_role", "")
	}
	_ = ctx
}

func (h *DashboardHandler) teacherDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
	ctx := r.Context()
	today := time.Now().Weekday()
	dayOfWeek := int(today)

	// Today's schedule.
	type scheduleBlock struct {
		CourseID   *uuid.UUID `json:"course_id"`
		CourseName string     `json:"course_name"`
		StartTime  string     `json:"start_time"`
		EndTime    string     `json:"end_time"`
		Room       *string    `json:"room"`
		Label      *string    `json:"label"`
		Color      *string    `json:"color"`
	}

	schedRows, _ := h.db.Query(ctx, `
		SELECT sb.course_id, COALESCE(c.name, ''), sb.start_time::text, sb.end_time::text, sb.room, sb.label, sb.color
		FROM schedule_blocks sb
		LEFT JOIN courses c ON c.id = sb.course_id
		WHERE sb.user_id = $1 AND sb.day_of_week = $2 AND sb.school_id = $3
		ORDER BY sb.start_time
	`, claims.UserID, dayOfWeek, claims.SchoolID)

	var schedule []scheduleBlock
	if schedRows != nil {
		defer schedRows.Close()
		for schedRows.Next() {
			var sb scheduleBlock
			schedRows.Scan(&sb.CourseID, &sb.CourseName, &sb.StartTime, &sb.EndTime, &sb.Room, &sb.Label, &sb.Color)
			schedule = append(schedule, sb)
		}
	}

	// Alerts: ungraded assignments.
	var ungradedCount int
	h.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM assignments a
		JOIN courses c ON c.id = a.course_id
		JOIN teachers t ON t.id = c.teacher_id
		WHERE t.user_id = $1 AND a.school_id = $2
		  AND a.due_date < NOW()
		  AND EXISTS (
		    SELECT 1 FROM enrollments e WHERE e.course_id = a.course_id AND e.status = 'active'
		    AND NOT EXISTS (
		      SELECT 1 FROM grades g WHERE g.assignment_id = a.id AND g.student_id = e.student_id
		    )
		  )
	`, claims.UserID, claims.SchoolID).Scan(&ungradedCount)

	// Recent grade activity.
	type recentActivity struct {
		AssignmentTitle string  `json:"assignment_title"`
		CourseName      string  `json:"course_name"`
		AveragePercent  float64 `json:"average_percent"`
		GradedCount     int     `json:"graded_count"`
		TotalCount      int     `json:"total_count"`
	}

	actRows, _ := h.db.Query(ctx, `
		SELECT a.title, c.name,
		       COALESCE(AVG(g.points_earned / a.max_points * 100), 0)::float,
		       COUNT(g.id)::int,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id AND e.status = 'active')::int
		FROM assignments a
		JOIN courses c ON c.id = a.course_id
		JOIN teachers t ON t.id = c.teacher_id
		LEFT JOIN grades g ON g.assignment_id = a.id
		WHERE t.user_id = $1 AND a.school_id = $2
		GROUP BY a.id, a.title, c.name
		ORDER BY a.updated_at DESC
		LIMIT 5
	`, claims.UserID, claims.SchoolID)

	var activity []recentActivity
	if actRows != nil {
		defer actRows.Close()
		for actRows.Next() {
			var ra recentActivity
			actRows.Scan(&ra.AssignmentTitle, &ra.CourseName, &ra.AveragePercent, &ra.GradedCount, &ra.TotalCount)
			activity = append(activity, ra)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":                  "teacher",
		"today_schedule":        schedule,
		"ungraded_assignments":  ungradedCount,
		"recent_grade_activity": activity,
	})
}

func (h *DashboardHandler) parentDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
	ctx := r.Context()

	// Get linked children — return short_id for URL-friendly references.
	rows, _ := h.db.Query(ctx, `
		SELECT s.short_id, u.first_name, u.last_name, s.grade_level, s.is_grade_locked,
		       ps.can_view_grades
		FROM parent_students ps
		JOIN students s ON s.id = ps.student_id
		JOIN users u ON u.id = s.user_id
		WHERE ps.parent_id = $1 AND ps.school_id = $2
	`, claims.UserID, claims.SchoolID)

	type childSummary struct {
		StudentID     string `json:"student_id"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		GradeLevel    string `json:"grade_level"`
		IsGradeLocked bool   `json:"is_grade_locked"`
		CanViewGrades bool   `json:"can_view_grades"`
	}

	var children []childSummary
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var c childSummary
			rows.Scan(&c.StudentID, &c.FirstName, &c.LastName, &c.GradeLevel, &c.IsGradeLocked, &c.CanViewGrades)
			children = append(children, c)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":     "parent",
		"children": children,
	})
}

func (h *DashboardHandler) studentDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
	ctx := r.Context()

	// Get the student record — return short_id for URL-friendly references.
	var studentShortID string
	var isLocked bool
	h.db.QueryRow(ctx, `SELECT short_id, is_grade_locked FROM students WHERE user_id = $1 AND school_id = $2`,
		claims.UserID, claims.SchoolID).Scan(&studentShortID, &isLocked)

	response := map[string]interface{}{
		"role":           "student",
		"student_id":     studentShortID,
		"is_grade_locked": isLocked,
	}

	if isLocked {
		response["grade_message"] = "Your grade access has been temporarily restricted. Please contact your school administration."
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *DashboardHandler) adminDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
	ctx := r.Context()

	var totalStudents, totalTeachers, lockedStudents int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE school_id = $1 AND enrollment_status = 'active'`, claims.SchoolID).Scan(&totalStudents)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM teachers WHERE school_id = $1`, claims.SchoolID).Scan(&totalTeachers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE school_id = $1 AND is_grade_locked = TRUE`, claims.SchoolID).Scan(&lockedStudents)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":            "admin",
		"total_students":  totalStudents,
		"total_teachers":  totalTeachers,
		"locked_students": lockedStudents,
	})
}

func (h *DashboardHandler) superAdminDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims) {
	ctx := r.Context()

	var totalSchools, totalUsers, totalStudents, totalTeachers, totalLockedStudents int
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM schools`).Scan(&totalSchools)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = TRUE`).Scan(&totalUsers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE enrollment_status = 'active'`).Scan(&totalStudents)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM teachers`).Scan(&totalTeachers)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE is_grade_locked = TRUE`).Scan(&totalLockedStudents)

	// Recent audit activity across all schools.
	type recentAudit struct {
		Action    string    `json:"action"`
		UserEmail string    `json:"user_email"`
		SchoolName string   `json:"school_name"`
		CreatedAt time.Time `json:"created_at"`
	}

	auditRows, _ := h.db.Query(ctx, `
		SELECT al.action, COALESCE(u.email, ''), COALESCE(s.name, ''), al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		LEFT JOIN schools s ON s.id = al.school_id
		ORDER BY al.created_at DESC
		LIMIT 10
	`)

	var recentActivity []recentAudit
	if auditRows != nil {
		defer auditRows.Close()
		for auditRows.Next() {
			var ra recentAudit
			auditRows.Scan(&ra.Action, &ra.UserEmail, &ra.SchoolName, &ra.CreatedAt)
			recentActivity = append(recentActivity, ra)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":                  "super_admin",
		"total_schools":         totalSchools,
		"total_users":           totalUsers,
		"total_students":        totalStudents,
		"total_teachers":        totalTeachers,
		"total_locked_students": totalLockedStudents,
		"recent_activity":       recentActivity,
	})
}
