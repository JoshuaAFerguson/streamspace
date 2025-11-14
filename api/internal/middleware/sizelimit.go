package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimit limits the size of incoming request bodies
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set maximum request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		// Handle body read errors
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"error":   "Request too large",
					"message": "Request body exceeds maximum allowed size",
					"max_size_mb": maxBytes / 1024 / 1024,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// StrictSizeLimit provides stricter size limits for specific endpoints
func StrictSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Request too large",
				"message": "Request body exceeds maximum allowed size for this endpoint",
				"max_size_mb": maxBytes / 1024 / 1024,
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
