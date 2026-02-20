package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

// VerificationService generates and validates HMAC-signed verification codes
// for documents and digital IDs.
type VerificationService struct {
	schoolSecret string
}

// NewVerificationService creates a VerificationService bound to a school's secret.
func NewVerificationService(schoolSecret string) *VerificationService {
	return &VerificationService{schoolSecret: schoolSecret}
}

// GenerateCode creates an HMAC-SHA256 verification code for a document or ID.
// The code encodes: documentID + schoolID to ensure cross-school codes are invalid.
func (s *VerificationService) GenerateCode(documentID, schoolID uuid.UUID) string {
	mac := hmac.New(sha256.New, []byte(s.schoolSecret))
	fmt.Fprintf(mac, "%s:%s", documentID.String(), schoolID.String())
	raw := mac.Sum(nil)
	return base64.URLEncoding.EncodeToString(raw)
}

// VerifyCode validates that a code matches the expected value for a document.
func (s *VerificationService) VerifyCode(code string, documentID, schoolID uuid.UUID) bool {
	expected := s.GenerateCode(documentID, schoolID)
	// Constant-time comparison to prevent timing attacks.
	return hmac.Equal([]byte(code), []byte(expected))
}
