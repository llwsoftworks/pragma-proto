package auth

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the JWT payload fields.
type Claims struct {
	jwt.RegisteredClaims
	UserID   uuid.UUID `json:"uid"`
	SchoolID uuid.UUID `json:"sid"`
	Role     string    `json:"role"`
	Email    string    `json:"email"`
	MFADone  bool      `json:"mfa_done"` // false until TOTP verified
}

// tokenDurations by role, per spec.
var tokenDurations = map[string]time.Duration{
	"super_admin": 15 * time.Minute,
	"admin":       15 * time.Minute,
	"teacher":     15 * time.Minute,
	"parent":      24 * time.Hour,
	"student":     24 * time.Hour,
}

// JWTService signs and validates JWTs using Ed25519.
type JWTService struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewJWTService creates a JWTService from base64-encoded Ed25519 key material.
func NewJWTService(privateKeyB64, publicKeyB64 string) (*JWTService, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("jwt: decode private key: %w", err)
	}
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("jwt: decode public key: %w", err)
	}
	return &JWTService{
		privateKey: ed25519.PrivateKey(privBytes),
		publicKey:  ed25519.PublicKey(pubBytes),
	}, nil
}

// Issue creates a signed JWT for the given user.
// mfaDone should be false for the initial token before TOTP verification.
func (s *JWTService) Issue(userID, schoolID uuid.UUID, role, email string, mfaDone bool) (string, error) {
	dur, ok := tokenDurations[role]
	if !ok {
		dur = 15 * time.Minute
	}

	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(dur)),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   userID,
		SchoolID: schoolID,
		Role:     role,
		Email:    email,
		MFADone:  mfaDone,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(s.privateKey)
}

// Validate parses and validates a JWT string, returning the Claims on success.
func (s *JWTService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	return claims, nil
}

// HashToken hashes a JWT for storage in the sessions table.
// Stored hashes allow session invalidation without decoding tokens.
func HashToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(h[:])
}
