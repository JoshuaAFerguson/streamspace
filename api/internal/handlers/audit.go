package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streamspace/streamspace/api/internal/db"
)

// AuditLogHandler handles audit log queries
type AuditLogHandler struct {
	db *db.Database
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(database *db.Database) *AuditLogHandler {
	return &AuditLogHandler{
		db: database,
	}
}

// AuditLogEntry represents a single audit log entry
type AuditLogEntry struct {
	ID           int                    `json:"id"`
	UserID       string                 `json:"userId,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	IPAddress    string                 `json:"ipAddress,omitempty"`
}

// ListAuditLogs returns audit logs with advanced filtering
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	ctx := context.Background()

	// Parse query parameters
	userID := c.Query("user_id")
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	ipAddress := c.Query("ip_address")

	// Pagination
	limit := 100 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Build query dynamically
	query := `
		SELECT id, user_id, action, resource_type, resource_id, changes, timestamp, ip_address
		FROM audit_log
		WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	// Add filters
	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}

	if resourceType != "" {
		query += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, resourceType)
		argIdx++
	}

	if resourceID != "" {
		query += fmt.Sprintf(" AND resource_id = $%d", argIdx)
		args = append(args, resourceID)
		argIdx++
	}

	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, action)
		argIdx++
	}

	if ipAddress != "" {
		query += fmt.Sprintf(" AND ip_address = $%d", argIdx)
		args = append(args, ipAddress)
		argIdx++
	}

	// Date range filters
	if startDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, startDate); err == nil {
			query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
			args = append(args, parsedDate)
			argIdx++
		}
	}

	if endDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, endDate); err == nil {
			query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
			args = append(args, parsedDate)
			argIdx++
		}
	}

	// Count total before pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS filtered", query)
	var total int
	err := h.db.DB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count audit logs"})
		return
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Collect results
	logs := []AuditLogEntry{}
	for rows.Next() {
		var entry AuditLogEntry
		var changesJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Action,
			&entry.ResourceType,
			&entry.ResourceID,
			&changesJSON,
			&entry.Timestamp,
			&entry.IPAddress,
		)
		if err != nil {
			continue
		}

		// Parse changes JSON
		if len(changesJSON) > 0 {
			var changes map[string]interface{}
			if err := json.Unmarshal(changesJSON, &changes); err == nil {
				entry.Changes = changes
			}
		}

		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"filters": gin.H{
			"user_id":       userID,
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"action":        action,
			"start_date":    startDate,
			"end_date":      endDate,
			"ip_address":    ipAddress,
		},
	})
}

// GetAuditLogStats returns statistics about audit logs
func (h *AuditLogHandler) GetAuditLogStats(c *gin.Context) {
	ctx := context.Background()

	// Get stats by action type
	actionStatsQuery := `
		SELECT action, COUNT(*) as count
		FROM audit_log
		WHERE timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY action
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := h.db.DB().QueryContext(ctx, actionStatsQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get action stats"})
		return
	}
	defer rows.Close()

	actionStats := []map[string]interface{}{}
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err == nil {
			actionStats = append(actionStats, map[string]interface{}{
				"action": action,
				"count":  count,
			})
		}
	}

	// Get stats by user (top 10 most active)
	userStatsQuery := `
		SELECT user_id, COUNT(*) as count
		FROM audit_log
		WHERE timestamp >= NOW() - INTERVAL '30 days'
		  AND user_id IS NOT NULL
		  AND user_id != ''
		GROUP BY user_id
		ORDER BY count DESC
		LIMIT 10
	`

	rows2, err := h.db.DB().QueryContext(ctx, userStatsQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}
	defer rows2.Close()

	userStats := []map[string]interface{}{}
	for rows2.Next() {
		var userID string
		var count int
		if err := rows2.Scan(&userID, &count); err == nil {
			userStats = append(userStats, map[string]interface{}{
				"userId": userID,
				"count":  count,
			})
		}
	}

	// Get total count
	var totalCount int
	err = h.db.DB().QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_log
	`).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	// Get recent count (last 24 hours)
	var recentCount int
	err = h.db.DB().QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_log
		WHERE timestamp >= NOW() - INTERVAL '24 hours'
	`).Scan(&recentCount)
	if err != nil {
		recentCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"totalLogs":      totalCount,
		"recentLogs24h":  recentCount,
		"topActions":     actionStats,
		"topUsers":       userStats,
	})
}

// GetUserAuditLogs returns audit logs for a specific user
func (h *AuditLogHandler) GetUserAuditLogs(c *gin.Context) {
	ctx := context.Background()
	userID := c.Param("userId")

	// Pagination
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get total count
	var total int
	err := h.db.DB().QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_log WHERE user_id = $1
	`, userID).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count logs"})
		return
	}

	// Get logs
	query := `
		SELECT id, user_id, action, resource_type, resource_id, changes, timestamp, ip_address
		FROM audit_log
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.DB().QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := []AuditLogEntry{}
	for rows.Next() {
		var entry AuditLogEntry
		var changesJSON []byte

		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Action,
			&entry.ResourceType,
			&entry.ResourceID,
			&changesJSON,
			&entry.Timestamp,
			&entry.IPAddress,
		)
		if err != nil {
			continue
		}

		// Parse changes JSON
		if len(changesJSON) > 0 {
			var changes map[string]interface{}
			if err := json.Unmarshal(changesJSON, &changes); err == nil {
				entry.Changes = changes
			}
		}

		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"userId": userID,
	})
}
