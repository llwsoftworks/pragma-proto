package services

import (
	"github.com/pragma-proto/api/internal/models"
)

// GradingService calculates weighted grades and GPA.
type GradingService struct{}

// NewGradingService creates a GradingService.
func NewGradingService() *GradingService {
	return &GradingService{}
}

// CalculateCourseGrade computes a student's overall grade for a course.
// grades and assignments must correspond 1-to-1 (matched by assignment ID).
// categoryWeights maps category name → fractional weight (must sum to 1.0 if provided).
func (s *GradingService) CalculateCourseGrade(
	assignments []models.Assignment,
	grades []models.Grade,
	scale []models.LetterGradeMapping,
	categoryWeights map[string]float64,
) *models.GradeCalculation {
	if len(assignments) == 0 {
		return nil
	}

	// Map grades by assignment ID for fast lookup.
	gradeByAssignment := make(map[string]*models.Grade, len(grades))
	for i := range grades {
		gradeByAssignment[grades[i].AssignmentID.String()] = &grades[i]
	}

	// Accumulate points per category.
	type catAcc struct {
		earned float64
		total  float64
	}
	cats := make(map[string]*catAcc)

	for _, a := range assignments {
		g, ok := gradeByAssignment[a.ID.String()]
		if !ok || g.IsExcused || g.PointsEarned == nil {
			// Skip ungraded or excused.
			continue
		}
		if _, exists := cats[a.Category]; !exists {
			cats[a.Category] = &catAcc{}
		}
		cats[a.Category].earned += *g.PointsEarned * a.Weight
		cats[a.Category].total += a.MaxPoints * a.Weight
	}

	if len(cats) == 0 {
		return nil
	}

	var totalEarned, totalPoints float64

	if len(categoryWeights) > 0 {
		// Weighted by category.
		for cat, acc := range cats {
			weight, ok := categoryWeights[cat]
			if !ok {
				weight = 1.0 / float64(len(cats))
			}
			if acc.total > 0 {
				totalEarned += (acc.earned / acc.total) * weight
				totalPoints += weight
			}
		}
	} else {
		// Simple total-points method.
		for _, acc := range cats {
			totalEarned += acc.earned
			totalPoints += acc.total
		}
	}

	var pct float64
	if totalPoints > 0 {
		pct = (totalEarned / totalPoints) * 100
	}

	letter := s.PercentToLetter(pct, scale)

	return &models.GradeCalculation{
		Percentage:  pct,
		LetterGrade: letter,
		PointsEarned: totalEarned,
		PointsTotal:  totalPoints,
	}
}

// CalculateGPA computes a GPA (4.0 scale) from a slice of course grade calculations.
func (s *GradingService) CalculateGPA(grades []*models.GradeCalculation, scale []models.LetterGradeMapping) float64 {
	if len(grades) == 0 {
		return 0
	}
	var total float64
	for _, g := range grades {
		total += s.LetterToGradePoint(g.LetterGrade, scale)
	}
	return total / float64(len(grades))
}

// PercentToLetter maps a percentage to a letter grade using the school's scale.
func (s *GradingService) PercentToLetter(pct float64, scale []models.LetterGradeMapping) string {
	if len(scale) == 0 {
		scale = models.DefaultLetterGrades
	}
	for _, m := range scale {
		if pct >= m.MinPercent && pct <= m.MaxPercent {
			return m.Letter
		}
	}
	return "F"
}

// LetterToGradePoint returns the grade point value for a letter grade.
func (s *GradingService) LetterToGradePoint(letter string, scale []models.LetterGradeMapping) float64 {
	if len(scale) == 0 {
		scale = models.DefaultLetterGrades
	}
	for _, m := range scale {
		if m.Letter == letter {
			return m.GradePoint
		}
	}
	return 0
}
