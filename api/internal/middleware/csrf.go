// Package middleware provides HTTP middleware for the StreamSpace API.
// This file implements CSRF (Cross-Site Request Forgery) protection.
//
// SECURITY ENHANCEMENT (2025-11-14):
// Added CSRF protection using double-submit cookie pattern with constant-time comparison.
//
// CSRF Attack Scenario (Without Protection):
// 1. User logs into StreamSpace (gets session cookie)
// 2. User visits malicious site evil.com
// 3. evil.com contains: <form action="https://streamspace.io/api/delete-account" method="POST">
// 4. Browser automatically sends session cookie with the malicious request
// 5. StreamSpace deletes user's account (thinks it's a legitimate request)
//
// CSRF Protection (Double-Submit Cookie Pattern):
// 1. GET request: Server generates random CSRF token, sends in both cookie AND header
// 2. Client stores header token (JavaScript can read it)
// 3. POST request: Client sends token in both cookie AND custom header
// 4. Server compares: cookie token == header token (using constant-time comparison)
// 5. If match: Request is from legitimate client (evil.com can't read/set custom headers)
// 6. If mismatch: Request is CSRF attack (blocked)
//
// Why This Works:
// - Malicious sites can trigger POST requests (via forms, fetch)
// - Browsers automatically send cookies with requests (even cross-site)
// - BUT: Malicious sites CANNOT read cookies or set custom headers (Same-Origin Policy)
// - So attacker cannot get the token to put in the custom header
//
// Implementation Details:
// - Token: 32 random bytes, base64-encoded (256 bits of entropy)
// - Comparison: Constant-time (prevents timing attacks)
// - Storage: In-memory map with automatic cleanup (24-hour expiry)
// - Exempt: GET, HEAD, OPTIONS requests (safe methods, no state change)
//
// Usage:
//   router.Use(middleware.CSRFProtection())
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRF Constants define CSRF protection configuration.
const (
	// CSRFTokenLength is the length of CSRF tokens in bytes
	CSRFTokenLength = 32

	// CSRFTokenHeader is the HTTP header for CSRF tokens
	CSRFTokenHeader = "X-CSRF-Token"

	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "csrf_token"

	// CSRFTokenExpiry is how long CSRF tokens are valid
	CSRFTokenExpiry = 24 * time.Hour
)

// CSRFStore stores CSRF tokens with expiration
type CSRFStore struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var (
	globalCSRFStore = &CSRFStore{
		tokens: make(map[string]time.Time),
	}
	csrfCleanupOnce sync.Once
)

// generateCSRFToken generates a random CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// addToken adds a token to the store with expiration
func (cs *CSRFStore) addToken(token string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.tokens[token] = time.Now().Add(CSRFTokenExpiry)
}

// validateToken checks if a token is valid and not expired
func (cs *CSRFStore) validateToken(token string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	expiry, exists := cs.tokens[token]
	if !exists {
		return false
	}
	
	// Check if expired
	if time.Now().After(expiry) {
		return false
	}
	
	return true
}

// removeToken removes a token from the store
func (cs *CSRFStore) removeToken(token string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.tokens, token)
}

// cleanup removes expired tokens
func (cs *CSRFStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		for token, expiry := range cs.tokens {
			if now.After(expiry) {
				delete(cs.tokens, token)
			}
		}
		cs.mu.Unlock()
	}
}

// CSRFProtection middleware validates CSRF tokens for state-changing requests
func CSRFProtection() gin.HandlerFunc {
	// Start cleanup goroutine once
	csrfCleanupOnce.Do(func() {
		go globalCSRFStore.cleanup()
	})

	return func(c *gin.Context) {
		// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			// For GET requests, generate and set a CSRF token
			token, err := generateCSRFToken()
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate CSRF token",
				})
				return
			}

			// Store token
			globalCSRFStore.addToken(token)

			// Set token in response header
			c.Header(CSRFTokenHeader, token)

			// Set token in cookie (HttpOnly for security)
			c.SetCookie(
				CSRFCookieName,
				token,
				int(CSRFTokenExpiry.Seconds()),
				"/",
				"",
				true,  // Secure (HTTPS only in production)
				true,  // HttpOnly
			)

			c.Next()
			return
		}

		// For state-changing methods (POST, PUT, DELETE, PATCH), validate CSRF token
		// Get token from header
		headerToken := c.GetHeader(CSRFTokenHeader)
		
		// Get token from cookie
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token missing",
				"message": "CSRF cookie not found",
			})
			return
		}

		// Tokens must match
		if subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token mismatch",
				"message": "CSRF tokens do not match",
			})
			return
		}

		// Validate token exists and is not expired
		if !globalCSRFStore.validateToken(cookieToken) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token invalid",
				"message": "CSRF token has expired or is invalid",
			})
			return
		}

		c.Next()
	}
}

// GetCSRFToken returns the current CSRF token for the request
// Useful for rendering in HTML forms or passing to frontend
func GetCSRFToken(c *gin.Context) string {
	return c.GetHeader(CSRFTokenHeader)
}
