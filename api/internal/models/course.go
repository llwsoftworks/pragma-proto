package models

import (
	"time"

	"github.com/google/uuid"
)

// Course is a class offered by a school.
type Course struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SchoolID     uuid.UUID `json:"school_id" db:"school_id"`
	TeacherID    uuid.UUID `json:"teacher_id" db:"teacher_id"`
	Name         string    `json:"name" db:"name"`
	Subject      string    `json:"subject" db:"subject"`
	Period       *string   `json:"period,omitempty" db:"period"`
	Room         *string   `json:"room,omitempty" db:"room"`
	AcademicYear string    `json:"academic_year" db:"academic_year"`
	Semester     *string   `json:"semester,omitempty" db:"semester"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Joined fields.
	TeacherName     string `json:"teacher_name,omitempty"`
	EnrollmentCount int    `json:"enrollment_count,omitempty"`
}

// Enrollment links a student to a course.
type Enrollment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	StudentID  uuid.UUID  `json:"student_id" db:"student_id"`
	CourseID   uuid.UUID  `json:"course_id" db:"course_id"`
	SchoolID   uuid.UUID  `json:"school_id" db:"school_id"`
	EnrolledAt time.Time  `json:"enrolled_at" db:"enrolled_at"`
	DroppedAt  *time.Time `json:"dropped_at,omitempty" db:"dropped_at"`
	Status     string     `json:"status" db:"status"`
}
