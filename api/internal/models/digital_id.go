package models

import (
	"time"

	"github.com/google/uuid"
)

// DigitalID is the student's digital identification card.
type DigitalID struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	StudentID   uuid.UUID  `json:"student_id" db:"student_id"`
	SchoolID    uuid.UUID  `json:"school_id" db:"school_id"`
	IDNumber    string     `json:"id_number" db:"id_number"`
	QRCodeData  string     `json:"qr_code_data" db:"qr_code_data"`
	BarcodeData *string    `json:"barcode_data,omitempty" db:"barcode_data"`
	PhotoURL    *string    `json:"photo_url,omitempty" db:"photo_url"`
	IssuedAt    time.Time  `json:"issued_at" db:"issued_at"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	IsValid     bool       `json:"is_valid" db:"is_valid"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`

	// Joined / populated fields.
	StudentName  string  `json:"student_name,omitempty"`
	GradeLevel   string  `json:"grade_level,omitempty"`
	SchoolName   string  `json:"school_name,omitempty"`
	SchoolLogoURL *string `json:"school_logo_url,omitempty"`
	QRCodeImage  []byte  `json:"qr_code_image,omitempty"` // PNG bytes
}

// VerificationResult is returned by the public /verify endpoint.
type VerificationResult struct {
	Valid        bool    `json:"valid"`
	StudentName  string  `json:"student_name,omitempty"`
	PhotoURL     *string `json:"photo_url,omitempty"`
	DocumentType string  `json:"document_type,omitempty"` // "digital_id" | "document"
	IssuedAt     string  `json:"issued_at,omitempty"`
	ExpiresAt    string  `json:"expires_at,omitempty"`
}
