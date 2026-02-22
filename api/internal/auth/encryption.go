package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

// LoginEncryptor decrypts AES-256-GCM encrypted login payloads.
// The SvelteKit frontend encrypts {email, password} before sending to the API.
type LoginEncryptor struct {
	gcm cipher.AEAD
}

// NewLoginEncryptor creates a LoginEncryptor from a base64-encoded 32-byte key.
func NewLoginEncryptor(keyB64 string) (*LoginEncryptor, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("login_encryption: decode key: %w", err)
	}
	if len(key) != 32 {
		return nil, errors.New("login_encryption: key must be exactly 32 bytes (AES-256)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("login_encryption: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("login_encryption: new gcm: %w", err)
	}

	return &LoginEncryptor{gcm: gcm}, nil
}

// Decrypt takes a base64-encoded ciphertext (nonce || ciphertext || tag)
// and returns the plaintext bytes.
func (e *LoginEncryptor) Decrypt(ciphertextB64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("login_encryption: decode ciphertext: %w", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("login_encryption: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("login_encryption: decrypt: %w", err)
	}

	return plaintext, nil
}
