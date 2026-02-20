package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

// Argon2id parameters per spec: 64MB memory, 3 iterations, 4 parallelism.
var argon2Params = &argon2id.Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword hashes a plaintext password using Argon2id.
func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2Params)
}

// VerifyPassword checks a plaintext password against an Argon2id hash.
// Returns true when the password matches.
func VerifyPassword(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

// ValidatePasswordStrength enforces the minimum password policy.
// Returns a descriptive error if the password is too weak.
func ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}
	return nil
}

// CheckBreachedPassword queries the Have I Been Pwned API (k-anonymity model)
// to determine whether a password has appeared in a known breach.
// Returns true when the password is breached (unsafe).
// A network error is treated as non-fatal: we log it and allow the password.
func CheckBreachedPassword(password string) (bool, error) {
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))
	prefix := hashHex[:5]
	suffix := hashHex[5:]

	resp, err := http.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		// Network failure → don't block the user but surface the error.
		return false, fmt.Errorf("hibp: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("hibp: read body: %w", err)
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], suffix) {
			return true, nil // found in breach database
		}
	}
	return false, nil
}
