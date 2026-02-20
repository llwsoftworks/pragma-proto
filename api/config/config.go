package config

import (
	"fmt"
	"os"
)

// Config holds all environment-based configuration for the API server.
type Config struct {
	// Server
	Port string
	Env  string

	// Database (PostgreSQL via Neon)
	DatabaseURL string

	// Authentication
	JWTPrivateKey string // Ed25519 private key (PEM or base64)
	JWTPublicKey  string // Ed25519 public key (PEM or base64)

	// Cloudflare R2
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2Endpoint        string

	// Anthropic Claude API
	ClaudeAPIKey string
	ClaudeModel  string

	// Resend (transactional email)
	ResendAPIKey  string
	EmailFromAddr string

	// Frontend
	FrontendOrigin string

	// Encryption: per-school keys are derived from this root key
	EncryptionRootKey string

	// HIBP (Have I Been Pwned) API
	HIBPAPIKey string
}

// Load reads configuration from environment variables.
// Any missing required variable causes an error.
func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		Env:               getEnv("ENV", "development"),
		DatabaseURL:       requireEnv("DATABASE_URL"),
		JWTPrivateKey:     requireEnv("JWT_PRIVATE_KEY"),
		JWTPublicKey:      requireEnv("JWT_PUBLIC_KEY"),
		R2AccountID:       requireEnv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     requireEnv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: requireEnv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      requireEnv("R2_BUCKET_NAME"),
		R2Endpoint:        requireEnv("R2_ENDPOINT"),
		ClaudeAPIKey:      requireEnv("CLAUDE_API_KEY"),
		ClaudeModel:       getEnv("CLAUDE_MODEL", "claude-sonnet-4-5-20250929"),
		ResendAPIKey:      requireEnv("RESEND_API_KEY"),
		EmailFromAddr:     getEnv("EMAIL_FROM_ADDR", "noreply@pragmagrading.com"),
		FrontendOrigin:    requireEnv("FRONTEND_ORIGIN"),
		EncryptionRootKey: requireEnv("ENCRYPTION_ROOT_KEY"),
		HIBPAPIKey:        getEnv("HIBP_API_KEY", ""),
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		// Panic at startup so misconfigured deployments fail fast.
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
