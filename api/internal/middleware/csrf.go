package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFProtection implements CSRF token validation for state-changing operations
type CSRFProtection struct {
	tokens map[string]time.Time // token -> expiration time
	mu     sync.RWMutex
	maxAge time.Duration
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(maxAge time.Duration) *CSRFProtection {
	csrf := &CSRFProtection{
		tokens: make(map[string]time.Time),
		maxAge: maxAge,
	}

	// Start cleanup goroutine
	go csrf.cleanupExpired()

	return csrf
}

// generateToken creates a cryptographically secure random token
func (c *CSRFProtection) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// cleanupExpired removes expired tokens periodically
func (c *CSRFProtection) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for token, expiry := range c.tokens {
			if now.After(expiry) {
				delete(c.tokens, token)
			}
		}
		c.mu.Unlock()
	}
}

// IssueToken generates and stores a new CSRF token
func (c *CSRFProtection) IssueToken(ctx *gin.Context) (string, error) {
	token, err := c.generateToken()
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.tokens[token] = time.Now().Add(c.maxAge)
	c.mu.Unlock()

	// Set token in cookie for SPA applications
	ctx.SetCookie(
		"csrf_token",
		token,
		int(c.maxAge.Seconds()),
		"/",
		"",
		true,  // Secure - only over HTTPS in production
		true,  // HttpOnly - prevent JavaScript access
	)

	return token, nil
}

// ValidateToken checks if a token is valid
func (c *CSRFProtection) ValidateToken(token string) bool {
	c.mu.RLock()
	expiry, exists := c.tokens[token]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		// Clean up expired token
		c.mu.Lock()
		delete(c.tokens, token)
		c.mu.Unlock()
		return false
	}

	return true
}

// Middleware returns a Gin middleware that validates CSRF tokens
// Should be applied to all state-changing routes (POST, PUT, PATCH, DELETE)
func (c *CSRFProtection) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip CSRF validation for safe methods
		if ctx.Request.Method == "GET" || ctx.Request.Method == "HEAD" || ctx.Request.Method == "OPTIONS" {
			ctx.Next()
			return
		}

		// Get token from header (for AJAX requests)
		token := ctx.GetHeader("X-CSRF-Token")

		// Fallback to cookie if header not present
		if token == "" {
			cookie, err := ctx.Cookie("csrf_token")
			if err == nil {
				token = cookie
			}
		}

		// Validate token
		if token == "" || !c.ValidateToken(token) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF validation failed",
				"message": "Invalid or missing CSRF token. Please refresh and try again.",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// IssueTokenHandler returns a handler that issues CSRF tokens
// Should be called on initial page load or login
func (c *CSRFProtection) IssueTokenHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := c.IssueToken(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate CSRF token",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
		})
	}
}
