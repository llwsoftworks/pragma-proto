package models

import (
	"time"

	"github.com/google/uuid"
)

// Assignment categories.
const (
	CategoryHomework      = "homework"
	CategoryQuiz          = "quiz"
	CategoryTest          = "test"
	CategoryExam          = "exam"
	CategoryProject       = "project"
	CategoryClasswork     = "classwork"
	CategoryParticipation = "participation"
	CategoryOther         = "other"
)

// Assignment represents a graded task in a course.
type Assignment struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CourseID    uuid.UUID  `json:"course_id" db:"course_id"`
	SchoolID    uuid.UUID  `json:"school_id" db:"school_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	MaxPoints   float64    `json:"max_points" db:"max_points"`
	Category    string     `json:"category" db:"category"`
	Weight      float64    `json:"weight" db:"weight"`
	IsPublished bool       `json:"is_published" db:"is_published"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`

	// Joined when fetching with attachments.
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment is a file attached to an assignment.
type Attachment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	AssignmentID uuid.UUID `json:"assignment_id" db:"assignment_id"`
	SchoolID     uuid.UUID `json:"school_id" db:"school_id"`
	FileName     string    `json:"file_name" db:"file_name"`
	FileKey      string    `json:"file_key" db:"file_key"` // R2 object key
	FileSize     int64     `json:"file_size" db:"file_size"`
	MIMEType     string    `json:"mime_type" db:"mime_type"`
	UploadedBy   uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	Version      int       `json:"version" db:"version"`
	IsCurrent    bool      `json:"is_current" db:"is_current"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Populated on response — never stored.
	DownloadURL string `json:"download_url,omitempty"`
}

// AllowedMIMETypes lists allowed upload MIME types per spec.
var AllowedMIMETypes = map[string]bool{
	"application/pdf":                                                      true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":    true,
	"image/jpeg":  true,
	"image/png":   true,
	"image/gif":   true,
	"audio/mpeg":  true,
	"video/mp4":   true,
}
