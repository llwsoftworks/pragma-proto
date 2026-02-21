// Package shortid generates cryptographically random, URL-safe short identifiers.
//
// Each ID is 8 characters from the base62 alphabet (0-9, A-Z, a-z),
// yielding 62^8 (~2.18 × 10^14) possible values. Collisions are handled
// at the database level via UNIQUE constraints; callers retry on conflict.
//
// Zero external dependencies — uses only crypto/rand from the Go stdlib.
package shortid

import (
	"crypto/rand"
	"fmt"
)

// Length is the number of characters in a generated short ID.
const Length = 8

// alphabet contains exactly 62 URL-safe characters: digits, uppercase, lowercase.
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generate returns a cryptographically random 8-character base62 string.
// It reads 8 bytes from crypto/rand and maps each to a character in the
// 62-char alphabet using rejection sampling to eliminate modulo bias.
//
// The rejection threshold is 248 (62×4), so each byte has at most a
// 248/256 = 96.9% acceptance rate. The probability of needing more than
// one pass for all 8 bytes is negligible.
func Generate() (string, error) {
	buf := make([]byte, Length)
	for i := range buf {
		for {
			var b [1]byte
			if _, err := rand.Read(b[:]); err != nil {
				return "", fmt.Errorf("shortid: crypto/rand: %w", err)
			}
			// Rejection sampling: discard values >= 248 to avoid modulo bias.
			// 248 = 62 * 4, so values 0–247 map uniformly to 0–61.
			if b[0] >= 248 {
				continue
			}
			buf[i] = alphabet[b[0]%62]
			break
		}
	}
	return string(buf), nil
}

// Validate returns true if s is exactly 8 characters from the base62 alphabet.
func Validate(s string) bool {
	if len(s) != Length {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}
