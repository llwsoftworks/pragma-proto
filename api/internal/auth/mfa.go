package auth

import (
	"fmt"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTPSecret creates a new TOTP secret for a user.
// Returns the secret key and the provisioning URI (used to generate a QR code).
func GenerateTOTPSecret(email, issuer string) (secret, provisioningURI string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", fmt.Errorf("mfa: generate totp key: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTP validates a 6-digit TOTP code against a stored secret.
// Returns true when the code is valid for the current time window.
func VerifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}
