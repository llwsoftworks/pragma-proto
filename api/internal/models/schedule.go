package models

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleBlock is a time slot in a user's weekly schedule.
type ScheduleBlock struct {
	ID          uuid.UUID  `json:"-" db:"id"`
	ShortID     string     `json:"id" db:"short_id"`
	SchoolID    uuid.UUID  `json:"school_id" db:"school_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	CourseID    *uuid.UUID `json:"course_id,omitempty" db:"course_id"`
	DayOfWeek   int        `json:"day_of_week" db:"day_of_week"` // 0=Sunday
	StartTime   string     `json:"start_time" db:"start_time"`   // "HH:MM"
	EndTime     string     `json:"end_time" db:"end_time"`       // "HH:MM"
	Room        *string    `json:"room,omitempty" db:"room"`
	Label       *string    `json:"label,omitempty" db:"label"`
	Color       *string    `json:"color,omitempty" db:"color"`
	Semester    *string    `json:"semester,omitempty" db:"semester"`
	IsRecurring bool       `json:"is_recurring" db:"is_recurring"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`

	// Joined fields.
	CourseName string `json:"course_name,omitempty"`
}

// ConflictDetail describes a scheduling conflict.
type ConflictDetail struct {
	ExistingBlockID uuid.UUID `json:"existing_block_id"`
	ConflictType    string    `json:"conflict_type"` // "room" | "teacher"
	DayOfWeek       int       `json:"day_of_week"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	Label           string    `json:"label"`
}
