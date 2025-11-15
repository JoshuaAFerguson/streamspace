package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CollaborationSession represents a collaborative session
type CollaborationSession struct {
	ID               string                 `json:"id"`
	SessionID        string                 `json:"session_id"`
	OwnerID          string                 `json:"owner_id"`
	Participants     []CollaborationUser    `json:"participants"`
	Settings         CollaborationSettings  `json:"settings"`
	ActiveUsers      int                    `json:"active_users"`
	ChatEnabled      bool                   `json:"chat_enabled"`
	AnnotationsEnabled bool                 `json:"annotations_enabled"`
	CursorTracking   bool                   `json:"cursor_tracking"`
	Status           string                 `json:"status"` // "active", "paused", "ended"
	CreatedAt        time.Time              `json:"created_at"`
	EndedAt          *time.Time             `json:"ended_at,omitempty"`
}

// CollaborationUser represents a user in a collaborative session
type CollaborationUser struct {
	UserID         string                 `json:"user_id"`
	Username       string                 `json:"username"`
	Role           string                 `json:"role"` // "owner", "presenter", "participant", "viewer"
	Permissions    CollaborationPermissions `json:"permissions"`
	CursorPosition *CursorPosition        `json:"cursor_position,omitempty"`
	IsActive       bool                   `json:"is_active"`
	JoinedAt       time.Time              `json:"joined_at"`
	LastSeenAt     time.Time              `json:"last_seen_at"`
	Color          string                 `json:"color"` // User color for cursor/annotations
}

// CollaborationPermissions defines what a user can do
type CollaborationPermissions struct {
	CanControl     bool `json:"can_control"`      // Can interact with session
	CanAnnotate    bool `json:"can_annotate"`     // Can create annotations
	CanChat        bool `json:"can_chat"`         // Can send messages
	CanInvite      bool `json:"can_invite"`       // Can invite others
	CanManage      bool `json:"can_manage"`       // Can change settings
	CanRecord      bool `json:"can_record"`       // Can start recording
	CanViewOnly    bool `json:"can_view_only"`    // View-only mode
}

// CollaborationSettings defines session behavior
type CollaborationSettings struct {
	FollowMode          string `json:"follow_mode"` // "none", "follow_presenter", "follow_owner"
	MaxParticipants     int    `json:"max_participants"`
	RequireApproval     bool   `json:"require_approval"`
	AllowAnonymous      bool   `json:"allow_anonymous"`
	LockOnPresenter     bool   `json:"lock_on_presenter"`
	AutoMuteJoiners     bool   `json:"auto_mute_joiners"`
	ShowCursorLabels    bool   `json:"show_cursor_labels"`
	EnableHandRaise     bool   `json:"enable_hand_raise"`
}

// CursorPosition represents cursor location
type CursorPosition struct {
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatMessage represents a collaboration chat message
type ChatMessage struct {
	ID          int64                  `json:"id"`
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Message     string                 `json:"message"`
	MessageType string                 `json:"message_type"` // "text", "system", "reaction"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Annotation represents a drawing/annotation on the session
type Annotation struct {
	ID          string                 `json:"id"`
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	Type        string                 `json:"type"` // "line", "arrow", "rectangle", "circle", "text", "freehand"
	Color       string                 `json:"color"`
	Thickness   int                    `json:"thickness"`
	Points      []Point                `json:"points"`
	Text        string                 `json:"text,omitempty"`
	IsPersistent bool                  `json:"is_persistent"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// Point represents a coordinate point
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// CreateCollaborationSession creates a new collaboration session
func (h *Handler) CreateCollaborationSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	userID := c.GetString("user_id")

	var req struct {
		Settings CollaborationSettings `json:"settings"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults if not provided
		req.Settings = CollaborationSettings{
			FollowMode:        "none",
			MaxParticipants:   10,
			RequireApproval:   false,
			AllowAnonymous:    false,
			ShowCursorLabels:  true,
			EnableHandRaise:   true,
		}
	}

	// Verify session ownership
	if !h.canAccessSession(userID, sessionID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Check if collaboration already exists
	var existingID string
	err := h.DB.QueryRow(`
		SELECT id FROM collaboration_sessions
		WHERE session_id = $1 AND status = 'active'
	`, sessionID).Scan(&existingID)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "collaboration already active", "collaboration_id": existingID})
		return
	}

	// Create collaboration session
	collabID := fmt.Sprintf("collab-%s-%d", sessionID, time.Now().Unix())
	err = h.DB.QueryRow(`
		INSERT INTO collaboration_sessions (
			id, session_id, owner_id, settings, chat_enabled,
			annotations_enabled, cursor_tracking, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, collabID, sessionID, userID, toJSONB(req.Settings), true, true, true, "active").Scan(&collabID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create collaboration session"})
		return
	}

	// Add owner as first participant
	ownerPerms := CollaborationPermissions{
		CanControl:   true,
		CanAnnotate:  true,
		CanChat:      true,
		CanInvite:    true,
		CanManage:    true,
		CanRecord:    true,
		CanViewOnly:  false,
	}

	h.DB.Exec(`
		INSERT INTO collaboration_participants (
			collaboration_id, user_id, role, permissions, color, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, collabID, userID, "owner", toJSONB(ownerPerms), "#0066FF", true)

	c.JSON(http.StatusCreated, gin.H{
		"collaboration_id": collabID,
		"session_id":       sessionID,
		"status":           "active",
		"websocket_url":    fmt.Sprintf("wss://%s/api/v1/collaboration/%s/ws", c.Request.Host, collabID),
	})
}

// JoinCollaborationSession allows a user to join a collaboration
func (h *Handler) JoinCollaborationSession(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	var req struct {
		InviteToken string `json:"invite_token"`
	}
	c.ShouldBindJSON(&req)

	// Get collaboration details
	var sessionID, ownerID string
	var settings, status sql.NullString
	err := h.DB.QueryRow(`
		SELECT session_id, owner_id, settings, status
		FROM collaboration_sessions WHERE id = $1
	`, collabID).Scan(&sessionID, &ownerID, &settings, &status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "collaboration not found"})
		return
	}

	if status.String != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collaboration not active"})
		return
	}

	// Parse settings
	var collabSettings CollaborationSettings
	if settings.Valid && settings.String != "" {
		json.Unmarshal([]byte(settings.String), &collabSettings)
	}

	// Check if user has access to session
	if !h.canAccessSession(userID, sessionID) && req.InviteToken == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied - invitation required"})
		return
	}

	// Check if already a participant
	var existingRole string
	h.DB.QueryRow(`
		SELECT role FROM collaboration_participants
		WHERE collaboration_id = $1 AND user_id = $2
	`, collabID, userID).Scan(&existingRole)

	if existingRole != "" {
		// Update to active
		h.DB.Exec(`
			UPDATE collaboration_participants
			SET is_active = true, last_seen_at = $1
			WHERE collaboration_id = $2 AND user_id = $3
		`, time.Now(), collabID, userID)

		c.JSON(http.StatusOK, gin.H{"message": "rejoined successfully", "role": existingRole})
		return
	}

	// Check participant limit
	var participantCount int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM collaboration_participants
		WHERE collaboration_id = $1 AND is_active = true
	`, collabID).Scan(&participantCount)

	if participantCount >= collabSettings.MaxParticipants {
		c.JSON(http.StatusForbidden, gin.H{"error": "collaboration is full"})
		return
	}

	// Default permissions for participants
	participantPerms := CollaborationPermissions{
		CanControl:   true,
		CanAnnotate:  true,
		CanChat:      true,
		CanInvite:    false,
		CanManage:    false,
		CanRecord:    false,
		CanViewOnly:  false,
	}

	// Assign color
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E2"}
	userColor := colors[participantCount%len(colors)]

	// Add participant
	_, err = h.DB.Exec(`
		INSERT INTO collaboration_participants (
			collaboration_id, user_id, role, permissions, color, is_active
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, collabID, userID, "participant", toJSONB(participantPerms), userColor, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join collaboration"})
		return
	}

	// Update participant count
	h.DB.Exec(`
		UPDATE collaboration_sessions
		SET active_users = (SELECT COUNT(*) FROM collaboration_participants WHERE collaboration_id = $1 AND is_active = true)
		WHERE id = $1
	`, collabID)

	// Send system message
	h.DB.Exec(`
		INSERT INTO collaboration_chat (
			collaboration_id, user_id, message, message_type
		) VALUES ($1, $2, $3, $4)
	`, collabID, "system", fmt.Sprintf("User %s joined the session", userID), "system")

	c.JSON(http.StatusOK, gin.H{
		"message":       "joined successfully",
		"role":          "participant",
		"color":         userColor,
		"websocket_url": fmt.Sprintf("wss://%s/api/v1/collaboration/%s/ws", c.Request.Host, collabID),
	})
}

// LeaveCollaborationSession removes a user from collaboration
func (h *Handler) LeaveCollaborationSession(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	// Update participant status
	_, err := h.DB.Exec(`
		UPDATE collaboration_participants
		SET is_active = false, last_seen_at = $1
		WHERE collaboration_id = $2 AND user_id = $3
	`, time.Now(), collabID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave"})
		return
	}

	// Update active user count
	h.DB.Exec(`
		UPDATE collaboration_sessions
		SET active_users = (SELECT COUNT(*) FROM collaboration_participants WHERE collaboration_id = $1 AND is_active = true)
		WHERE id = $1
	`, collabID)

	// Send system message
	h.DB.Exec(`
		INSERT INTO collaboration_chat (
			collaboration_id, user_id, message, message_type
		) VALUES ($1, $2, $3, $4)
	`, collabID, "system", fmt.Sprintf("User %s left the session", userID), "system")

	c.JSON(http.StatusOK, gin.H{"message": "left successfully"})
}

// GetCollaborationParticipants lists all participants
func (h *Handler) GetCollaborationParticipants(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	// Verify user is a participant
	if !h.isCollaborationParticipant(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	rows, err := h.DB.Query(`
		SELECT cp.user_id, u.username, cp.role, cp.permissions, cp.cursor_position,
		       cp.color, cp.is_active, cp.joined_at, cp.last_seen_at
		FROM collaboration_participants cp
		LEFT JOIN users u ON cp.user_id = u.id
		WHERE cp.collaboration_id = $1
		ORDER BY cp.is_active DESC, cp.joined_at ASC
	`, collabID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve participants"})
		return
	}
	defer rows.Close()

	participants := []CollaborationUser{}
	for rows.Next() {
		var p CollaborationUser
		var permissions, cursorPos sql.NullString
		var username sql.NullString

		err := rows.Scan(&p.UserID, &username, &p.Role, &permissions, &cursorPos,
			&p.Color, &p.IsActive, &p.JoinedAt, &p.LastSeenAt)

		if err == nil {
			if username.Valid {
				p.Username = username.String
			}
			if permissions.Valid && permissions.String != "" {
				json.Unmarshal([]byte(permissions.String), &p.Permissions)
			}
			if cursorPos.Valid && cursorPos.String != "" {
				json.Unmarshal([]byte(cursorPos.String), &p.CursorPosition)
			}
			participants = append(participants, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"participants": participants})
}

// UpdateParticipantRole updates a participant's role and permissions
func (h *Handler) UpdateParticipantRole(c *gin.Context) {
	collabID := c.Param("collabId")
	targetUserID := c.Param("userId")
	userID := c.GetString("user_id")

	var req struct {
		Role        string                   `json:"role"`
		Permissions CollaborationPermissions `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user has manage permissions
	if !h.canManageCollaboration(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	// Update participant
	_, err := h.DB.Exec(`
		UPDATE collaboration_participants
		SET role = $1, permissions = $2
		WHERE collaboration_id = $3 AND user_id = $4
	`, req.Role, toJSONB(req.Permissions), collabID, targetUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

// Chat Operations

// SendChatMessage sends a message to the collaboration chat
func (h *Handler) SendChatMessage(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	var req struct {
		Message     string                 `json:"message" binding:"required"`
		MessageType string                 `json:"message_type"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user is a participant with chat permission
	if !h.hasCollaborationPermission(collabID, userID, "can_chat") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Insert message
	var msgID int64
	err := h.DB.QueryRow(`
		INSERT INTO collaboration_chat (
			collaboration_id, user_id, message, message_type, metadata
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, collabID, userID, req.Message, req.MessageType, toJSONB(req.Metadata)).Scan(&msgID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message_id": msgID,
		"sent_at":    time.Now(),
	})
}

// GetChatHistory retrieves chat history
func (h *Handler) GetChatHistory(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	before := c.Query("before") // Message ID to paginate

	// Verify participant
	if !h.isCollaborationParticipant(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	query := `
		SELECT cc.id, cc.collaboration_id, cc.user_id, u.username, cc.message,
		       cc.message_type, cc.metadata, cc.created_at
		FROM collaboration_chat cc
		LEFT JOIN users u ON cc.user_id = u.id
		WHERE cc.collaboration_id = $1
	`
	args := []interface{}{collabID}
	argCount := 2

	if before != "" {
		beforeID, _ := strconv.ParseInt(before, 10, 64)
		query += fmt.Sprintf(" AND cc.id < $%d", argCount)
		args = append(args, beforeID)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY cc.created_at DESC LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve chat"})
		return
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var msg ChatMessage
		var metadata sql.NullString
		var username sql.NullString

		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.UserID, &username, &msg.Message,
			&msg.MessageType, &metadata, &msg.CreatedAt)

		if err == nil {
			if username.Valid {
				msg.Username = username.String
			}
			if metadata.Valid && metadata.String != "" {
				json.Unmarshal([]byte(metadata.String), &msg.Metadata)
			}
			messages = append(messages, msg)
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Annotation Operations

// CreateAnnotation creates a new annotation
func (h *Handler) CreateAnnotation(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	var req Annotation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify annotate permission
	if !h.hasCollaborationPermission(collabID, userID, "can_annotate") {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	// Get session ID
	var sessionID string
	h.DB.QueryRow("SELECT session_id FROM collaboration_sessions WHERE id = $1", collabID).Scan(&sessionID)

	annotationID := fmt.Sprintf("annot-%d", time.Now().UnixNano())
	req.ID = annotationID
	req.SessionID = sessionID
	req.UserID = userID

	// Calculate expiration if not persistent
	var expiresAt *time.Time
	if !req.IsPersistent {
		expires := time.Now().Add(5 * time.Minute)
		expiresAt = &expires
	}

	_, err := h.DB.Exec(`
		INSERT INTO collaboration_annotations (
			id, collaboration_id, session_id, user_id, type, color, thickness,
			points, text, is_persistent, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, annotationID, collabID, sessionID, userID, req.Type, req.Color, req.Thickness,
		toJSONB(req.Points), req.Text, req.IsPersistent, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create annotation"})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// GetAnnotations retrieves active annotations
func (h *Handler) GetAnnotations(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	if !h.isCollaborationParticipant(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, session_id, user_id, type, color, thickness, points, text,
		       is_persistent, created_at, expires_at
		FROM collaboration_annotations
		WHERE collaboration_id = $1 AND (expires_at IS NULL OR expires_at > $2)
		ORDER BY created_at ASC
	`, collabID, time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve annotations"})
		return
	}
	defer rows.Close()

	annotations := []Annotation{}
	for rows.Next() {
		var a Annotation
		var points sql.NullString

		err := rows.Scan(&a.ID, &a.SessionID, &a.UserID, &a.Type, &a.Color, &a.Thickness,
			&points, &a.Text, &a.IsPersistent, &a.CreatedAt, &a.ExpiresAt)

		if err == nil {
			if points.Valid && points.String != "" {
				json.Unmarshal([]byte(points.String), &a.Points)
			}
			annotations = append(annotations, a)
		}
	}

	c.JSON(http.StatusOK, gin.H{"annotations": annotations})
}

// DeleteAnnotation removes an annotation
func (h *Handler) DeleteAnnotation(c *gin.Context) {
	collabID := c.Param("collabId")
	annotationID := c.Param("annotationId")
	userID := c.GetString("user_id")

	// Verify ownership or manage permission
	var ownerID string
	h.DB.QueryRow("SELECT user_id FROM collaboration_annotations WHERE id = $1", annotationID).Scan(&ownerID)

	if ownerID != userID && !h.canManageCollaboration(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	_, err := h.DB.Exec("DELETE FROM collaboration_annotations WHERE id = $1", annotationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete annotation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "annotation deleted"})
}

// ClearAllAnnotations removes all annotations
func (h *Handler) ClearAllAnnotations(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	if !h.canManageCollaboration(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	result, err := h.DB.Exec("DELETE FROM collaboration_annotations WHERE collaboration_id = $1", collabID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear annotations"})
		return
	}

	count, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"message": "annotations cleared", "count": count})
}

// Helper functions

func (h *Handler) isCollaborationParticipant(collabID, userID string) bool {
	var exists bool
	h.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM collaboration_participants
		WHERE collaboration_id = $1 AND user_id = $2)
	`, collabID, userID).Scan(&exists)
	return exists
}

func (h *Handler) canManageCollaboration(collabID, userID string) bool {
	var permissions sql.NullString
	h.DB.QueryRow(`
		SELECT permissions FROM collaboration_participants
		WHERE collaboration_id = $1 AND user_id = $2
	`, collabID, userID).Scan(&permissions)

	if !permissions.Valid {
		return false
	}

	var perms CollaborationPermissions
	json.Unmarshal([]byte(permissions.String), &perms)
	return perms.CanManage
}

func (h *Handler) hasCollaborationPermission(collabID, userID, permission string) bool {
	var permissions sql.NullString
	h.DB.QueryRow(`
		SELECT permissions FROM collaboration_participants
		WHERE collaboration_id = $1 AND user_id = $2 AND is_active = true
	`, collabID, userID).Scan(&permissions)

	if !permissions.Valid {
		return false
	}

	var perms CollaborationPermissions
	json.Unmarshal([]byte(permissions.String), &perms)

	switch permission {
	case "can_chat":
		return perms.CanChat
	case "can_annotate":
		return perms.CanAnnotate
	case "can_control":
		return perms.CanControl
	case "can_invite":
		return perms.CanInvite
	default:
		return false
	}
}

// GetCollaborationStats returns collaboration statistics
func (h *Handler) GetCollaborationStats(c *gin.Context) {
	collabID := c.Param("collabId")
	userID := c.GetString("user_id")

	if !h.isCollaborationParticipant(collabID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	stats := map[string]interface{}{}

	// Participant count
	var totalParticipants, activeParticipants int
	h.DB.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_active = true)
		FROM collaboration_participants WHERE collaboration_id = $1
	`, collabID).Scan(&totalParticipants, &activeParticipants)
	stats["total_participants"] = totalParticipants
	stats["active_participants"] = activeParticipants

	// Message count
	var messageCount int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM collaboration_chat WHERE collaboration_id = $1
	`, collabID).Scan(&messageCount)
	stats["total_messages"] = messageCount

	// Annotation count
	var annotationCount int
	h.DB.QueryRow(`
		SELECT COUNT(*) FROM collaboration_annotations
		WHERE collaboration_id = $1 AND (expires_at IS NULL OR expires_at > $2)
	`, collabID, time.Now()).Scan(&annotationCount)
	stats["active_annotations"] = annotationCount

	// Session duration
	var startTime time.Time
	h.DB.QueryRow("SELECT created_at FROM collaboration_sessions WHERE id = $1", collabID).Scan(&startTime)
	duration := time.Since(startTime)
	stats["duration_seconds"] = int(duration.Seconds())

	c.JSON(http.StatusOK, stats)
}
