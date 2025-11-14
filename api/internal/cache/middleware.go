package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriter is a custom response writer that captures the response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// CacheMiddleware returns a Gin middleware for caching GET requests
func CacheMiddleware(cache *Cache, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip if caching is disabled
		if !cache.IsEnabled() {
			c.Next()
			return
		}

		// Generate cache key from request path and query params
		cacheKey := generateCacheKey(c.Request.URL.RequestURI())

		// Try to get cached response
		var cachedResp CachedResponse
		if err := cache.Get(c.Request.Context(), cacheKey, &cachedResp); err == nil {
			// Cache hit - return cached response
			for key, value := range cachedResp.Headers {
				c.Header(key, value)
			}
			c.Header("X-Cache", "HIT")
			c.Data(cachedResp.StatusCode, "application/json", []byte(cachedResp.Body))
			c.Abort()
			return
		}

		// Cache miss - capture the response
		writer := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		c.Next()

		// Only cache successful responses
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// Capture headers
			headers := make(map[string]string)
			for key := range c.Writer.Header() {
				headers[key] = c.Writer.Header().Get(key)
			}

			// Store in cache
			resp := CachedResponse{
				StatusCode: c.Writer.Status(),
				Headers:    headers,
				Body:       writer.body.String(),
			}

			// Set cache asynchronously to avoid blocking the response
			go func() {
				_ = cache.Set(c.Request.Context(), cacheKey, resp, ttl)
			}()

			c.Header("X-Cache", "MISS")
		}
	}
}

// generateCacheKey creates a consistent cache key from the request URI
func generateCacheKey(uri string) string {
	hash := sha256.Sum256([]byte(uri))
	return fmt.Sprintf("response:%s", hex.EncodeToString(hash[:]))
}

// InvalidateCacheMiddleware clears related cache entries after mutations
func InvalidateCacheMiddleware(cache *Cache, pattern string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only invalidate on successful mutations
		if c.Request.Method != http.MethodGet && c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			if cache.IsEnabled() {
				// Invalidate asynchronously
				go func() {
					_ = cache.DeletePattern(c.Request.Context(), pattern)
				}()
			}
		}
	}
}

// CacheControl middleware adds cache control headers to responses
func CacheControl(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only add cache headers for GET requests
		if c.Request.Method == http.MethodGet {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
		} else {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		}
		c.Next()
	}
}
