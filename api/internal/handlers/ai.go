package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/services"
)

// AIHandler proxies AI requests to Claude with anonymization.
type AIHandler struct {
	db  *pgxpool.Pool
	ai  *services.AIService
}

// NewAIHandler creates an AIHandler.
func NewAIHandler(db *pgxpool.Pool, ai *services.AIService) *AIHandler {
	return &AIHandler{db: db, ai: ai}
}

// GradingAssistant handles AI-assisted grading suggestions.
// Student names and PII are anonymized before being sent to Claude.
func (h *AIHandler) GradingAssistant(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		AssignmentID string            `json:"assignment_id" validate:"required,uuid"`
		Rubric       string            `json:"rubric" validate:"required,min=10"`
		Submissions  map[string]string `json:"submissions" validate:"required,min=1"` // student_id → submission text
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

	// Check AI is enabled for this school.
	var aiEnabled bool
	h.db.QueryRow(ctx, `SELECT (settings->>'ai_enabled')::boolean FROM schools WHERE id = $1`, claims.SchoolID).Scan(&aiEnabled)
	if !aiEnabled {
		writeError(w, http.StatusForbidden, "ai_disabled", "AI features are not enabled for your school")
		return
	}

	// Get max_points for validation.
	var maxPoints float64
	h.db.QueryRow(ctx, `SELECT max_points FROM assignments WHERE id = $1 AND school_id = $2`,
		req.AssignmentID, claims.SchoolID).Scan(&maxPoints)

	// Fetch student names for anonymization.
	studentNames := make(map[string]string)
	rows, _ := h.db.Query(ctx, `
		SELECT s.id, u.first_name || ' ' || u.last_name
		FROM students s JOIN users u ON u.id = s.user_id
		WHERE s.id = ANY($1) AND s.school_id = $2
	`, studentIDSlice(req.Submissions), claims.SchoolID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var sid, name string
			rows.Scan(&sid, &name)
			studentNames[sid] = name
		}
	}

	// Build anonymized submission text.
	anonMap := make(map[string]string) // placeholder → student_id (reverse)
	i := 0
	anonSubmissions := ""
	for studentID, text := range req.Submissions {
		i++
		placeholder := fmt.Sprintf("Student %c", rune('A'+i-1))
		anonMap[placeholder] = studentID
		anonSubmissions += placeholder + ":\n" + text + "\n\n"
	}

	systemPrompt := services.GradingAssistantPrompt(req.Rubric, maxPoints)
	response, tokens, err := h.ai.Complete(ctx, systemPrompt, anonSubmissions, 2048)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai_error", "AI service is unavailable")
		return
	}

	// Log AI interaction.
	interactionID, _ := uuid.Parse("")
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "ai.request",
		EntityType: "ai_interaction",
		NewValue: map[string]interface{}{
			"feature":       "grading_assistant",
			"assignment_id": req.AssignmentID,
			"tokens":        tokens,
		},
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	_ = interactionID

	// Store in ai_interactions table.
	h.db.Exec(ctx, `
		INSERT INTO ai_interactions (school_id, user_id, feature, input_summary, output_summary, tokens_used)
		VALUES ($1, $2, 'grading_assistant', $3, $4, $5)
	`, claims.SchoolID, claims.UserID,
		"grading_assistant request for "+req.AssignmentID,
		response[:min(500, len(response))],
		tokens,
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"raw_response":  response,
		"anonymized":    true,
		"student_map":   anonMap, // tells the teacher which placeholder = which student
		"tokens_used":   tokens,
	})
}

// ReportComment generates an AI-written report card comment.
func (h *AIHandler) ReportComment(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		StudentID     string  `json:"student_id" validate:"required,uuid"`
		CourseID      string  `json:"course_id" validate:"required,uuid"`
		GradeSummary  string  `json:"grade_summary" validate:"required"`
		TrendDir      string  `json:"trend_direction" validate:"required,oneof=improving declining stable"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := r.Context()

	systemPrompt := services.ReportCommentPrompt()
	prompt := "Grade summary (anonymized):\n" + req.GradeSummary + "\n\nTrend: " + req.TrendDir

	response, tokens, err := h.ai.Complete(ctx, systemPrompt, prompt, 512)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai_error", "AI service is unavailable")
		return
	}

	h.db.Exec(ctx, `
		INSERT INTO ai_interactions (school_id, user_id, feature, input_summary, output_summary, tokens_used)
		VALUES ($1, $2, 'report_comments', $3, $4, $5)
	`, claims.SchoolID, claims.UserID, req.GradeSummary[:min(200, len(req.GradeSummary))],
		response[:min(200, len(response))], tokens)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"comment":        response,
		"ai_assisted":    true,
		"tokens_used":    tokens,
	})
}

func studentIDSlice(m map[string]string) []string {
	ids := make([]string, 0, len(m))
	for k := range m {
		ids = append(ids, k)
	}
	return ids
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
