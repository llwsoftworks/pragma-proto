package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AIService proxies requests to the Anthropic Claude API.
// All student names and PII are anonymized BEFORE being sent to Claude.
type AIService struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewAIService creates an AIService.
func NewAIService(apiKey, model string) *AIService {
	return &AIService{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1",
		client:  &http.Client{},
	}
}

// claudeRequest is the JSON body sent to the Claude API.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a prompt to Claude and returns the text response and token count.
func (s *AIService) Complete(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (response string, tokensUsed int, err error) {
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	reqBody := claudeRequest{
		Model:     s.model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: systemPrompt + "\n\n" + userPrompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("ai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", 0, fmt.Errorf("ai: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("ai: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("ai: claude returned %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", 0, fmt.Errorf("ai: decode response: %w", err)
	}

	var parts []string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			parts = append(parts, c.Text)
		}
	}

	total := claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens
	return strings.Join(parts, ""), total, nil
}

// AnonymizeStudents replaces student names with anonymized placeholders.
// names is a map of student_id → full_name. Returns the anonymized text
// and a reverse map for de-anonymization.
func AnonymizeStudents(text string, names map[string]string) (anonymized string, reverseMap map[string]string) {
	reverseMap = make(map[string]string)
	anonymized = text
	i := 0
	for _, name := range names {
		i++
		placeholder := fmt.Sprintf("Student %c", rune('A'+i-1))
		reverseMap[placeholder] = name
		anonymized = strings.ReplaceAll(anonymized, name, placeholder)
	}
	return anonymized, reverseMap
}

// DeAnonymize restores original names in AI output using the reverse map.
func DeAnonymize(text string, reverseMap map[string]string) string {
	for placeholder, name := range reverseMap {
		text = strings.ReplaceAll(text, placeholder, name)
	}
	return text
}

// GradingAssistantPrompt builds the system prompt for the grading assistant feature.
func GradingAssistantPrompt(rubric string, maxPoints float64) string {
	return fmt.Sprintf(`You are an expert teacher's assistant helping to grade student assignments objectively.

Rubric:
%s

Maximum points: %.2f

For each student submission provided, respond with a JSON array where each element contains:
- "student": the student identifier (e.g., "Student A")
- "suggested_points": a numeric score from 0 to %.2f
- "reasoning": a brief, professional explanation of the score

Base your grades strictly on the rubric. Do not infer student identity from content.
Never include student names or any PII in your response.`,
		rubric, maxPoints, maxPoints)
}

// StudentInsightsPrompt builds the prompt for the at-risk student detection feature.
func StudentInsightsPrompt() string {
	return `You are an educational data analyst. Analyze the anonymized grade trajectory data provided.

For each student showing a concerning trend, respond with a JSON array of alerts:
- "student": student identifier
- "alert": brief description of the trend
- "severity": "low" | "medium" | "high"

Focus on: consistent grade drops, sudden declines, persistent failing grades, missing assignment patterns.
Only flag genuinely concerning patterns. Do not include students who are performing acceptably.`
}

// ReportCommentPrompt builds the prompt for AI-generated report card comments.
func ReportCommentPrompt() string {
	return `You are an experienced teacher writing professional report card comments.

Given the anonymized student grade summary, attendance data, and trend direction provided, write a single professional, specific, and constructive comment (2-4 sentences) suitable for a parent or guardian to read.

Rules:
- Do not use the student's name or any identifying information
- Be specific about academic performance (e.g., reference the subject area)
- Balance strengths with areas for growth
- Use professional, warm, encouraging language
- Return only the comment text, no JSON wrapper`
}
