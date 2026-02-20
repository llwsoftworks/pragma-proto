package models

import (
	"time"

	"github.com/google/uuid"
)

// Grade holds a student's grade for a single assignment.
type Grade struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AssignmentID uuid.UUID  `json:"assignment_id" db:"assignment_id"`
	StudentID    uuid.UUID  `json:"student_id" db:"student_id"`
	SchoolID     uuid.UUID  `json:"school_id" db:"school_id"`
	PointsEarned *float64   `json:"points_earned" db:"points_earned"`
	LetterGrade  *string    `json:"letter_grade,omitempty" db:"letter_grade"`
	Comment      *string    `json:"comment,omitempty" db:"comment"`
	GradedBy     *uuid.UUID `json:"graded_by,omitempty" db:"graded_by"`
	GradedAt     *time.Time `json:"graded_at,omitempty" db:"graded_at"`
	AISuggested  *float64   `json:"ai_suggested,omitempty" db:"ai_suggested"`
	AIAccepted   *bool      `json:"ai_accepted,omitempty" db:"ai_accepted"`
	IsExcused    bool       `json:"is_excused" db:"is_excused"`
	IsMissing    bool       `json:"is_missing" db:"is_missing"`
	IsLate       bool       `json:"is_late" db:"is_late"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// GradeCalculation holds the computed summary for a student in a course.
type GradeCalculation struct {
	StudentID    uuid.UUID `json:"student_id"`
	CourseID     uuid.UUID `json:"course_id"`
	Percentage   float64   `json:"percentage"`
	LetterGrade  string    `json:"letter_grade"`
	PointsEarned float64   `json:"points_earned"`
	PointsTotal  float64   `json:"points_total"`
	// Per-category breakdowns.
	CategoryBreakdown map[string]CategoryGrade `json:"category_breakdown"`
}

// CategoryGrade holds the grade summary within a single assignment category.
type CategoryGrade struct {
	Category     string  `json:"category"`
	Percentage   float64 `json:"percentage"`
	PointsEarned float64 `json:"points_earned"`
	PointsTotal  float64 `json:"points_total"`
	Weight       float64 `json:"weight"`
}

// LetterGradeMapping defines how percentage ranges map to letter grades.
// Stored in school settings JSONB and loaded at runtime.
type LetterGradeMapping struct {
	MinPercent float64 `json:"min_percent"`
	MaxPercent float64 `json:"max_percent"`
	Letter     string  `json:"letter"`
	GradePoint float64 `json:"grade_point"` // 4.0 scale
}

// DefaultLetterGrades is the fallback grading scale.
var DefaultLetterGrades = []LetterGradeMapping{
	{MinPercent: 93, MaxPercent: 100, Letter: "A", GradePoint: 4.0},
	{MinPercent: 90, MaxPercent: 93, Letter: "A-", GradePoint: 3.7},
	{MinPercent: 87, MaxPercent: 90, Letter: "B+", GradePoint: 3.3},
	{MinPercent: 83, MaxPercent: 87, Letter: "B", GradePoint: 3.0},
	{MinPercent: 80, MaxPercent: 83, Letter: "B-", GradePoint: 2.7},
	{MinPercent: 77, MaxPercent: 80, Letter: "C+", GradePoint: 2.3},
	{MinPercent: 73, MaxPercent: 77, Letter: "C", GradePoint: 2.0},
	{MinPercent: 70, MaxPercent: 73, Letter: "C-", GradePoint: 1.7},
	{MinPercent: 67, MaxPercent: 70, Letter: "D+", GradePoint: 1.3},
	{MinPercent: 63, MaxPercent: 67, Letter: "D", GradePoint: 1.0},
	{MinPercent: 60, MaxPercent: 63, Letter: "D-", GradePoint: 0.7},
	{MinPercent: 0, MaxPercent: 60, Letter: "F", GradePoint: 0.0},
}
