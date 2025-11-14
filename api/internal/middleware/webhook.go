package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebhookAuth validates webhook requests using HMAC-SHA256 signatures
type WebhookAuth struct {
	secret []byte
}

// NewWebhookAuth creates a new webhook authentication middleware
func NewWebhookAuth(secret string) *WebhookAuth {
	return &WebhookAuth{
		secret: []byte(secret),
	}
}

// Middleware returns a Gin middleware that validates webhook signatures
// Expects signature in X-Webhook-Signature header as hex-encoded HMAC-SHA256
func (w *WebhookAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get signature from header
		signature := c.GetHeader("X-Webhook-Signature")
		if signature == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing webhook signature",
			})
			c.Abort()
			return
		}

		// Read request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			c.Abort()
			return
		}

		// Restore body for downstream handlers
		c.Request.Body = io.NopCloser(io.Reader(io.MultiReader(
			io.Reader(nil),
		)))

		// Compute HMAC
		mac := hmac.New(sha256.New, w.secret)
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// Compare signatures using constant-time comparison
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid webhook signature",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Sign generates an HMAC-SHA256 signature for the given payload
// This is a helper function for testing or generating signatures
func (w *WebhookAuth) Sign(payload []byte) string {
	mac := hmac.New(sha256.New, w.secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
