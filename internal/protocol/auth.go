package protocol

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// HashPassword creates a Blynk-compatible password hash.
// The algorithm is: base64(SHA256(password + SHA256(lowercase_email)))
func HashPassword(email, password string) string {
	// First hash: SHA256 of lowercase email
	emailHash := sha256.Sum256([]byte(strings.ToLower(email)))

	// Second hash: SHA256 of (password + email_hash)
	combined := sha256.New()
	combined.Write([]byte(password))
	combined.Write(emailHash[:])
	finalHash := combined.Sum(nil)

	// Encode as base64
	return base64.StdEncoding.EncodeToString(finalHash)
}

// Credentials holds authentication information.
type Credentials struct {
	Email        string
	PasswordHash string
	ClientType   string
	BuildNumber  string
	Platform     string
}

// NewCredentials creates credentials with the default client configuration.
func NewCredentials(email, password string) *Credentials {
	return &Credentials{
		Email:        email,
		PasswordHash: HashPassword(email, password),
		ClientType:   "Other",
		BuildNumber:  "12220000",
		Platform:     "Blynk",
	}
}

// LoginParams returns the parameters for a LOGIN message.
func (c *Credentials) LoginParams() []string {
	return []string{
		c.Email,
		c.PasswordHash,
		c.ClientType,
		c.BuildNumber,
		c.Platform,
	}
}
