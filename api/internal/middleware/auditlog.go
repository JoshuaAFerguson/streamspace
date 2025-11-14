package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streamspace/streamspace/api/internal/db"
)

// AuditEvent represents a structured audit log event
type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	Username     string                 `json:"username,omitempty"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	StatusCode   int                    `json:"status_code"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Duration     int64                  `json:"duration_ms"`
	RequestBody  map[string]interface{} `json:"request_body,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger handles structured audit logging
type AuditLogger struct {
	database      *db.Database
	logRequestBody  bool
	logResponseBody bool
	sensitiveFields []string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(database *db.Database, logBodies bool) *AuditLogger {
	return &AuditLogger{
		database:        database,
		logRequestBody:  logBodies,
		logResponseBody: false, // Usually too verbose
		sensitiveFields: []string{"password", "token", "secret", "apiKey", "api_key"},
	}
}

// redactSensitiveData removes sensitive fields from data
func (a *AuditLogger) redactSensitiveData(data map[string]interface{}) map[string]interface{} {
	redacted := make(map[string]interface{})
	for key, value := range data {
		isSensitive := false
		for _, field := range a.sensitiveFields {
			if key == field {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			redacted[key] = "[REDACTED]"
		} else if nested, ok := value.(map[string]interface{}); ok {
			redacted[key] = a.redactSensitiveData(nested)
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// logEvent logs an audit event to the database
func (a *AuditLogger) logEvent(event *AuditEvent) error {
	if a.database == nil {
		return nil // Audit logging disabled
	}

	details, _ := json.Marshal(map[string]interface{}{
		"method":        event.Method,
		"path":          event.Path,
		"status_code":   event.StatusCode,
		"duration_ms":   event.Duration,
		"request_body":  event.RequestBody,
		"response_body": event.ResponseBody,
		"error":         event.Error,
		"metadata":      event.Metadata,
	})

	query := `
		INSERT INTO audit_log (user_id, action, resource_type, resource_id, changes, timestamp, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := a.database.DB().Exec(
		query,
		event.UserID,
		event.Action,
		event.Resource,
		event.ResourceID,
		details,
		event.Timestamp,
		event.IPAddress,
	)

	return err
}

// Middleware returns a Gin middleware that logs all requests
func (a *AuditLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Capture request body if enabled
		var requestBody map[string]interface{}
		if a.logRequestBody && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body

			if len(bodyBytes) > 0 && len(bodyBytes) < 10240 { // Max 10KB
				json.Unmarshal(bodyBytes, &requestBody)
				requestBody = a.redactSensitiveData(requestBody)
			}
		}

		// Create response writer wrapper to capture response
		writer := &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Extract user information from context
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")

		// Determine action and resource from path
		action := c.Request.Method
		resource := c.Request.URL.Path

		// Create audit event
		event := &AuditEvent{
			Timestamp:   startTime,
			UserID:      getUserIDString(userID),
			Username:    getUsernameString(username),
			Action:      action,
			Resource:    resource,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			StatusCode:  c.Writer.Status(),
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Duration:    duration.Milliseconds(),
			RequestBody: requestBody,
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.Error = c.Errors.String()
		}

		// Log the event (async to avoid blocking)
		go a.logEvent(event)
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Helper functions to safely extract user info
func getUserIDString(userID interface{}) string {
	if userID == nil {
		return ""
	}
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

func getUsernameString(username interface{}) string {
	if username == nil {
		return ""
	}
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}
