package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streamspace/streamspace/api/internal/quota"
)

// QuotaMiddleware enforces resource quotas at the API level
type QuotaMiddleware struct {
	enforcer *quota.Enforcer
}

// NewQuotaMiddleware creates a new quota middleware
func NewQuotaMiddleware(enforcer *quota.Enforcer) *QuotaMiddleware {
	return &QuotaMiddleware{
		enforcer: enforcer,
	}
}

// Middleware provides quota enforcement for all requests
func (q *QuotaMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get username from context (set by auth middleware)
		username, exists := c.Get("username")
		if !exists {
			// Skip quota check for unauthenticated requests
			c.Next()
			return
		}

		// Store enforcer in context for handlers to use
		c.Set("quota_enforcer", q.enforcer)
		c.Set("quota_username", username)

		c.Next()
	}
}

// EnforceSessionCreation is a helper that can be called from session creation handlers
func EnforceSessionCreation(c *gin.Context, requestedCPU, requestedMemory string, requestedGPU int, currentUsage *quota.Usage) error {
	enforcer, exists := c.Get("quota_enforcer")
	if !exists {
		// No enforcer, allow
		return nil
	}

	username, exists := c.Get("quota_username")
	if !exists {
		// No username, allow
		return nil
	}

	quotaEnforcer := enforcer.(*quota.Enforcer)
	usernameStr := username.(string)

	// Parse and validate resource requests
	cpu, memory, err := quotaEnforcer.ValidateResourceRequest(requestedCPU, requestedMemory)
	if err != nil {
		return err
	}

	// Check quotas
	return quotaEnforcer.CheckSessionCreation(c.Request.Context(), usernameStr, cpu, memory, requestedGPU, currentUsage)
}

// GetUserQuota is a gin handler that returns the user's quota limits and current usage
func GetUserQuota(enforcer *quota.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		usernameStr := username.(string)

		// Get user limits
		limits, err := enforcer.GetUserLimits(c.Request.Context(), usernameStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get quota limits",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"limits": limits,
		})
	}
}
