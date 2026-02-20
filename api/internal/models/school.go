package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// School represents a tenant (educational institution).
type School struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Address   *string      `json:"address,omitempty" db:"address"`
	LogoURL   *string      `json:"logo_url,omitempty" db:"logo_url"`
	Settings  SchoolSettings `json:"settings" db:"-"`
	SettingsRaw json.RawMessage `json:"-" db:"settings"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// SchoolSettings is stored as JSONB in the schools table.
type SchoolSettings struct {
	// Grading
	GradingScale    []LetterGradeMapping `json:"grading_scale,omitempty"`
	CategoryWeights map[string]float64   `json:"category_weights,omitempty"` // e.g., {"test": 0.4, "homework": 0.2}

	// Branding
	PrimaryColor   string `json:"primary_color,omitempty"`
	SecondaryColor string `json:"secondary_color,omitempty"`
	FontFamily     string `json:"font_family,omitempty"`

	// AI
	AIEnabled bool `json:"ai_enabled"`

	// Document templates
	DocumentTemplates map[string]string `json:"document_templates,omitempty"`

	// Report card
	ReportCardTemplate string `json:"report_card_template,omitempty"`

	// Signature image URL (R2 key) for official documents
	SignatureImageURL string `json:"signature_image_url,omitempty"`
	SignatoryName     string `json:"signatory_name,omitempty"`
	SignatoryTitle    string `json:"signatory_title,omitempty"`
}
