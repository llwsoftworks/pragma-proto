package models

import (
	"time"

	"github.com/google/uuid"
)

// Role constants.
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleTeacher    = "teacher"
	RoleParent     = "parent"
	RoleStudent    = "student"
)

// User represents a platform user of any role.
type User struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	SchoolID             uuid.UUID  `json:"school_id" db:"school_id"`
	Role                 string     `json:"role" db:"role"`
	Email                string     `json:"email" db:"email"`
	PasswordHash         string     `json:"-" db:"password_hash"`
	FirstName            string     `json:"first_name" db:"first_name"`
	LastName             string     `json:"last_name" db:"last_name"`
	Phone                *string    `json:"phone,omitempty" db:"phone"`
	ProfilePhoto         *string    `json:"profile_photo,omitempty" db:"profile_photo"`
	MFASecret            *string    `json:"-" db:"mfa_secret"`
	MFAEnabled           bool       `json:"mfa_enabled" db:"mfa_enabled"`
	IsActive             bool       `json:"is_active" db:"is_active"`
	FailedLoginAttempts  int        `json:"-" db:"failed_login_attempts"`
	LockedUntil          *time.Time `json:"-" db:"locked_until"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// Student extends users for student-specific data.
type Student struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	SchoolID         uuid.UUID  `json:"school_id" db:"school_id"`
	StudentNumber    string     `json:"student_number" db:"student_number"`
	GradeLevel       string     `json:"grade_level" db:"grade_level"`
	EnrollmentDate   time.Time  `json:"enrollment_date" db:"enrollment_date"`
	EnrollmentStatus string     `json:"enrollment_status" db:"enrollment_status"`
	IsGradeLocked    bool       `json:"is_grade_locked" db:"is_grade_locked"`
	LockReason       *string    `json:"-" db:"lock_reason"` // never exposed to student/parent
	DateOfBirth      *string    `json:"-" db:"date_of_birth"` // encrypted at rest
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`

	// Joined fields from users table.
	Email     string  `json:"email,omitempty"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Photo     *string `json:"profile_photo,omitempty"`
}

// Teacher extends users for teacher-specific data.
type Teacher struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	SchoolID   uuid.UUID `json:"school_id" db:"school_id"`
	Department *string   `json:"department,omitempty" db:"department"`
	Title      *string   `json:"title,omitempty" db:"title"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`

	// Joined fields from users table.
	Email     string  `json:"email,omitempty"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	Photo     *string `json:"profile_photo,omitempty"`
}

// ParentStudent links a parent to a student.
type ParentStudent struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ParentID         uuid.UUID `json:"parent_id" db:"parent_id"`
	StudentID        uuid.UUID `json:"student_id" db:"student_id"`
	SchoolID         uuid.UUID `json:"school_id" db:"school_id"`
	Relationship     string    `json:"relationship" db:"relationship"`
	IsPrimaryContact bool      `json:"is_primary_contact" db:"is_primary_contact"`
	CanViewGrades    bool      `json:"can_view_grades" db:"can_view_grades"`
	CanGenerateDocs  bool      `json:"can_generate_docs" db:"can_generate_docs"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
