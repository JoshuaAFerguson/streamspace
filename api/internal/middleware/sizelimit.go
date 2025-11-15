// Package middleware provides HTTP middleware for the StreamSpace API.
// This file implements request size limiting to prevent DoS attacks.
//
// SECURITY ENHANCEMENT (2025-11-14):
// Added request size limits to prevent denial of service via oversized payloads.
//
// Why Request Size Limits are Critical:
// - Prevents memory exhaustion from giant JSON payloads
// - Prevents disk exhaustion from huge file uploads
// - Prevents slow-loris attacks with endless request bodies
// - Forces attackers to use many small requests (easier to detect/rate-limit)
//
// Implementation:
// - Uses http.MaxBytesReader for hard limits (prevents buffer overflow)
// - Checks Content-Length header before processing (fail fast)
// - Skips for GET/HEAD/OPTIONS (no request body)
// - Returns 413 Payload Too Large with informative error message
//
// Limits:
// - Default request body: 10MB (general API endpoints)
// - JSON payloads: 5MB (structured data)
// - File uploads: 50MB (larger files like logs, exports)
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Request Size Limits define maximum allowed payload sizes.
//
// These values balance security (prevent DoS) with usability (allow reasonable uploads).
const (
	// MaxRequestBodySize is the maximum allowed request body size (10MB)
	MaxRequestBodySize int64 = 10 * 1024 * 1024 // 10 MB

	// MaxJSONPayloadSize is the maximum size for JSON payloads (5MB)
	MaxJSONPayloadSize int64 = 5 * 1024 * 1024 // 5 MB

	// MaxFileUploadSize is the maximum size for file uploads (50MB)
	MaxFileUploadSize int64 = 50 * 1024 * 1024 // 50 MB
)

// RequestSizeLimiter limits the size of incoming HTTP requests
// to prevent DoS attacks via oversized payloads
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, HEAD, OPTIONS requests (no body)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get Content-Length header
		contentLength := c.Request.ContentLength

		// Check if Content-Length exceeds limit
		if contentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":      "Request entity too large",
				"message":    "Request body exceeds maximum allowed size",
				"max_size_mb": float64(maxSize) / (1024 * 1024),
			})
			return
		}

		// Wrap the request body with a LimitReader
		// This prevents reading more than maxSize bytes even if Content-Length is lying
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// JSONSizeLimiter limits JSON payload size for API endpoints
func JSONSizeLimiter() gin.HandlerFunc {
	return RequestSizeLimiter(MaxJSONPayloadSize)
}

// FileUploadLimiter limits file upload size
func FileUploadLimiter() gin.HandlerFunc {
	return RequestSizeLimiter(MaxFileUploadSize)
}

// DefaultSizeLimiter uses the default max request body size
func DefaultSizeLimiter() gin.HandlerFunc {
	return RequestSizeLimiter(MaxRequestBodySize)
}
