package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// TokenHasher handles secure token generation and hashing
type TokenHasher struct {
	bcryptCost int
}

// NewTokenHasher creates a new token hasher
func NewTokenHasher() *TokenHasher {
	return &TokenHasher{
		bcryptCost: bcrypt.DefaultCost, // Cost 10 for good security/performance balance
	}
}

// GenerateSecureToken generates a cryptographically secure random token
// Returns the plain token (for giving to user) and the hashed token (for storage)
func (t *TokenHasher) GenerateSecureToken(length int) (plainToken string, hashedToken string, err error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64 for the plain token
	plainToken = base64.URLEncoding.EncodeToString(bytes)

	// Hash the token for storage
	hashedToken, err = t.HashToken(plainToken)
	if err != nil {
		return "", "", err
	}

	return plainToken, hashedToken, nil
}

// HashToken hashes a token using bcrypt for secure storage
// bcrypt is intentionally slow to prevent brute force attacks
func (t *TokenHasher) HashToken(token string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(token), t.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyToken verifies a plain token against a hashed token
func (t *TokenHasher) VerifyToken(plainToken, hashedToken string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(plainToken))
	return err == nil
}

// HashTokenSHA256 provides a faster hash for session tokens where lookup speed is critical
// Use this for session tokens that need fast validation
// Note: Less secure than bcrypt for password-like tokens, but acceptable for session tokens
func (t *TokenHasher) HashTokenSHA256(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// VerifyTokenSHA256 verifies a token against a SHA256 hash
func (t *TokenHasher) VerifyTokenSHA256(plainToken, hashedToken string) bool {
	computedHash := t.HashTokenSHA256(plainToken)
	return computedHash == hashedToken
}

// GenerateSessionToken generates a session-specific token
// Returns plain token and SHA256 hash (faster for session validation)
func (t *TokenHasher) GenerateSessionToken() (plainToken string, hashedToken string, err error) {
	// 32 bytes = 256 bits of entropy
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	plainToken = base64.URLEncoding.EncodeToString(bytes)
	hashedToken = t.HashTokenSHA256(plainToken)

	return plainToken, hashedToken, nil
}

// GenerateAPIToken generates an API token (uses bcrypt for better security)
// Returns plain token and bcrypt hash
func (t *TokenHasher) GenerateAPIToken() (plainToken string, hashedToken string, err error) {
	// 48 bytes = 384 bits of entropy for long-lived tokens
	bytes := make([]byte, 48)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate API token: %w", err)
	}

	plainToken = base64.URLEncoding.EncodeToString(bytes)

	// Use bcrypt for API tokens (they're long-lived and need stronger protection)
	hashedToken, err = t.HashToken(plainToken)
	if err != nil {
		return "", "", err
	}

	return plainToken, hashedToken, nil
}
