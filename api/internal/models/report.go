package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportCard is a generated grade summary for one student and academic period.
type ReportCard struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	StudentID       uuid.UUID  `json:"student_id" db:"student_id"`
	SchoolID        uuid.UUID  `json:"school_id" db:"school_id"`
	AcademicPeriod  string     `json:"academic_period" db:"academic_period"`
	GPA             *float64   `json:"gpa,omitempty" db:"gpa"`
	TeacherComments *string    `json:"teacher_comments,omitempty" db:"teacher_comments"`
	AdminComments   *string    `json:"admin_comments,omitempty" db:"admin_comments"`
	IsFinalized     bool       `json:"is_finalized" db:"is_finalized"`
	PDFURL          *string    `json:"pdf_url,omitempty" db:"pdf_url"`
	GeneratedBy     uuid.UUID  `json:"generated_by" db:"generated_by"`
	GeneratedAt     time.Time  `json:"generated_at" db:"generated_at"`
}

// Document is a generated official document (enrollment cert, attendance letter, etc.).
type Document struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	SchoolID         uuid.UUID  `json:"school_id" db:"school_id"`
	StudentID        uuid.UUID  `json:"student_id" db:"student_id"`
	Type             string     `json:"type" db:"type"`
	VerificationCode string     `json:"verification_code" db:"verification_code"`
	PDFURL           string     `json:"pdf_url" db:"pdf_url"`
	GeneratedBy      uuid.UUID  `json:"generated_by" db:"generated_by"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}
