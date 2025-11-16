package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// EventType represents the type of session event for real-time notifications.
//
// Event types are organized by category:
//   - Lifecycle: created, updated, deleted, state changes
//   - Activity: connected, disconnected, idle, active
//   - Resources: CPU/memory updates, tag changes
//   - Sharing: shared, unshared
//   - Errors: error notifications
//
// Events are sent to subscribed WebSocket clients in real-time,
// enabling the UI to update without polling.
type EventType string

const (
	// Session lifecycle events - fundamental session operations

	// EventSessionCreated is emitted when a new session is created.
	// Data: session details (template, resources, state)
	EventSessionCreated EventType = "session.created"

	// EventSessionUpdated is emitted when session metadata is modified.
	// Data: changed fields (tags, description, etc.)
	EventSessionUpdated EventType = "session.updated"

	// EventSessionDeleted is emitted when a session is deleted.
	// Data: none (session no longer exists)
	EventSessionDeleted EventType = "session.deleted"

	// EventSessionStateChange is emitted when session state transitions.
	// Data: oldState, newState (running→hibernated, etc.)
	EventSessionStateChange EventType = "session.state.changed"

	// Session activity events - connection and usage tracking

	// EventSessionConnected is emitted when a user connects to a session.
	// Data: connectionId, clientIP, userAgent
	EventSessionConnected EventType = "session.connected"

	// EventSessionDisconnected is emitted when a user disconnects.
	// Data: connectionId, duration
	EventSessionDisconnected EventType = "session.disconnected"

	// EventSessionHeartbeat is emitted on periodic heartbeat (optional).
	// Data: timestamp, active connections count
	EventSessionHeartbeat EventType = "session.heartbeat"

	// EventSessionIdle is emitted when session becomes idle.
	// Data: idleDuration (seconds)
	EventSessionIdle EventType = "session.idle"

	// EventSessionActive is emitted when idle session becomes active again.
	// Data: none
	EventSessionActive EventType = "session.active"

	// Session resource events - configuration changes

	// EventSessionResourcesUpdated is emitted when CPU/memory limits change.
	// Data: resources (cpu, memory, storage)
	EventSessionResourcesUpdated EventType = "session.resources.updated"

	// EventSessionTagsUpdated is emitted when session tags are modified.
	// Data: tags array
	EventSessionTagsUpdated EventType = "session.tags.updated"

	// Session sharing events - collaboration

	// EventSessionShared is emitted when session is shared with another user.
	// Data: sharedWith (userID), permissions
	EventSessionShared EventType = "session.shared"

	// EventSessionUnshared is emitted when sharing is revoked.
	// Data: unsharedFrom (userID)
	EventSessionUnshared EventType = "session.unshared"

	// Session error events - problem notifications

	// EventSessionError is emitted when an error occurs.
	// Data: error (error message), code (error code)
	EventSessionError EventType = "session.error"
)

// SessionEvent represents a session-related event sent to WebSocket clients.
//
// Events are JSON-encoded and sent over WebSocket connections to subscribed
// clients. The UI can listen for specific event types and update in real-time.
//
// Event routing:
//   - Events are routed to clients subscribed to the userID
//   - Events are also routed to clients subscribed to the sessionID
//   - A client can be subscribed to multiple users and sessions
//
// Example event:
//
//	{
//	  "type": "session.created",
//	  "sessionId": "user1-firefox",
//	  "userId": "user1",
//	  "timestamp": "2025-01-15T10:30:00Z",
//	  "data": {
//	    "template": "firefox-browser",
//	    "state": "running",
//	    "resources": {"cpu": "2000m", "memory": "4096Mi"}
//	  }
//	}
type SessionEvent struct {
	// Type identifies the event type (e.g., "session.created").
	Type EventType `json:"type"`

	// SessionID is the ID of the session this event relates to.
	SessionID string `json:"sessionId"`

	// UserID is the owner of the session.
	// Events are routed to all clients subscribed to this user.
	UserID string `json:"userId"`

	// Timestamp is when the event occurred (server time).
	Timestamp time.Time `json:"timestamp"`

	// Data contains event-specific payload (optional).
	// Structure depends on event type.
	Data map[string]interface{} `json:"data,omitempty"`
}

// Notifier handles event subscriptions and targeted real-time notifications.
//
// The Notifier implements a pub/sub pattern:
//   - Clients subscribe to user events (all sessions for a user)
//   - Clients subscribe to session events (specific session)
//   - Backend emits events via NotifySessionEvent()
//   - Notifier routes events to subscribed clients
//   - Hub delivers messages over WebSocket
//
// Subscription model:
//   - User subscriptions: Get all events for a user's sessions
//   - Session subscriptions: Get events for a specific session
//   - Clients can have both types of subscriptions simultaneously
//
// Thread safety:
//   - All map access protected by sync.RWMutex
//   - Safe for concurrent subscriptions and notifications
//
// Example usage:
//
//	notifier := NewNotifier(manager)
//
//	// Client subscribes to user events
//	notifier.SubscribeUser(clientID, userID)
//
//	// Backend emits event
//	notifier.NotifySessionCreated(sessionID, userID, data)
//
//	// Event is routed to subscribed clients via WebSocket
type Notifier struct {
	// manager coordinates WebSocket hubs for message delivery.
	manager *Manager

	// mu protects concurrent access to subscription maps.
	mu sync.RWMutex

	// userSubscriptions maps userID to set of subscribed client IDs.
	// userID -> set of client IDs
	// Clients in this map receive all events for the user's sessions.
	userSubscriptions map[string]map[string]bool

	// sessionSubscriptions maps sessionID to set of subscribed client IDs.
	// sessionID -> set of client IDs
	// Clients in this map receive events only for that specific session.
	sessionSubscriptions map[string]map[string]bool

	// clientUsers maps client IDs to their associated userID.
	// clientID -> userID
	// Used for cleanup when client disconnects.
	clientUsers map[string]string
}

// NewNotifier creates a new event notifier
func NewNotifier(manager *Manager) *Notifier {
	return &Notifier{
		manager:              manager,
		userSubscriptions:    make(map[string]map[string]bool),
		sessionSubscriptions: make(map[string]map[string]bool),
		clientUsers:          make(map[string]string),
	}
}

// SubscribeUser subscribes a client to receive events for a specific user
func (n *Notifier) SubscribeUser(clientID, userID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Add to user subscriptions
	if _, exists := n.userSubscriptions[userID]; !exists {
		n.userSubscriptions[userID] = make(map[string]bool)
	}
	n.userSubscriptions[userID][clientID] = true

	// Track client to user mapping
	n.clientUsers[clientID] = userID

	log.Printf("Client %s subscribed to user %s events", clientID, userID)
}

// SubscribeSession subscribes a client to receive events for a specific session
func (n *Notifier) SubscribeSession(clientID, sessionID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.sessionSubscriptions[sessionID]; !exists {
		n.sessionSubscriptions[sessionID] = make(map[string]bool)
	}
	n.sessionSubscriptions[sessionID][clientID] = true

	log.Printf("Client %s subscribed to session %s events", clientID, sessionID)
}

// UnsubscribeClient removes all subscriptions for a client
func (n *Notifier) UnsubscribeClient(clientID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Remove from user subscriptions
	if userID, exists := n.clientUsers[clientID]; exists {
		if clients, exists := n.userSubscriptions[userID]; exists {
			delete(clients, clientID)
			if len(clients) == 0 {
				delete(n.userSubscriptions, userID)
			}
		}
		delete(n.clientUsers, clientID)
	}

	// Remove from session subscriptions
	for sessionID, clients := range n.sessionSubscriptions {
		if clients[clientID] {
			delete(clients, clientID)
			if len(clients) == 0 {
				delete(n.sessionSubscriptions, sessionID)
			}
		}
	}

	log.Printf("Client %s unsubscribed from all events", clientID)
}

// NotifySessionEvent sends a session event to subscribed clients
func (n *Notifier) NotifySessionEvent(event SessionEvent) {
	n.mu.RLock()
	targetClients := make(map[string]bool)

	// Get clients subscribed to this user
	if event.UserID != "" {
		if clients, exists := n.userSubscriptions[event.UserID]; exists {
			for clientID := range clients {
				targetClients[clientID] = true
			}
		}
	}

	// Get clients subscribed to this session
	if clients, exists := n.sessionSubscriptions[event.SessionID]; exists {
		for clientID := range clients {
			targetClients[clientID] = true
		}
	}
	n.mu.RUnlock()

	// No subscribers, skip
	if len(targetClients) == 0 {
		return
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal session event: %v", err)
		return
	}

	// Send to target clients
	n.manager.sessionsHub.mu.RLock()
	sentCount := 0
	for client := range n.manager.sessionsHub.clients {
		if targetClients[client.id] {
			select {
			case client.send <- data:
				sentCount++
			default:
				log.Printf("Failed to send event to client %s (buffer full)", client.id)
			}
		}
	}
	n.manager.sessionsHub.mu.RUnlock()

	log.Printf("Event %s for session %s sent to %d clients", event.Type, event.SessionID, sentCount)
}

// NotifySessionCreated notifies clients when a session is created
func (n *Notifier) NotifySessionCreated(sessionID, userID string, data map[string]interface{}) {
	event := SessionEvent{
		Type:      EventSessionCreated,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data:      data,
	}
	n.NotifySessionEvent(event)
}

// NotifySessionUpdated notifies clients when a session is updated
func (n *Notifier) NotifySessionUpdated(sessionID, userID string, data map[string]interface{}) {
	event := SessionEvent{
		Type:      EventSessionUpdated,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data:      data,
	}
	n.NotifySessionEvent(event)
}

// NotifySessionDeleted notifies clients when a session is deleted
func (n *Notifier) NotifySessionDeleted(sessionID, userID string) {
	event := SessionEvent{
		Type:      EventSessionDeleted,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	n.NotifySessionEvent(event)
}

// NotifySessionStateChange notifies clients when a session changes state
func (n *Notifier) NotifySessionStateChange(sessionID, userID, oldState, newState string) {
	event := SessionEvent{
		Type:      EventSessionStateChange,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"oldState": oldState,
			"newState": newState,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionConnected notifies clients when someone connects to a session
func (n *Notifier) NotifySessionConnected(sessionID, userID string, connectionID string) {
	event := SessionEvent{
		Type:      EventSessionConnected,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"connectionId": connectionID,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionDisconnected notifies clients when someone disconnects from a session
func (n *Notifier) NotifySessionDisconnected(sessionID, userID string, connectionID string) {
	event := SessionEvent{
		Type:      EventSessionDisconnected,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"connectionId": connectionID,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionIdle notifies clients when a session becomes idle
func (n *Notifier) NotifySessionIdle(sessionID, userID string, idleDuration int64) {
	event := SessionEvent{
		Type:      EventSessionIdle,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"idleDuration": idleDuration,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionActive notifies clients when a session becomes active again
func (n *Notifier) NotifySessionActive(sessionID, userID string) {
	event := SessionEvent{
		Type:      EventSessionActive,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	n.NotifySessionEvent(event)
}

// NotifySessionResourcesUpdated notifies clients when session resources are updated
func (n *Notifier) NotifySessionResourcesUpdated(sessionID, userID string, resources map[string]interface{}) {
	event := SessionEvent{
		Type:      EventSessionResourcesUpdated,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"resources": resources,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionTagsUpdated notifies clients when session tags are updated
func (n *Notifier) NotifySessionTagsUpdated(sessionID, userID string, tags []string) {
	event := SessionEvent{
		Type:      EventSessionTagsUpdated,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"tags": tags,
		},
	}
	n.NotifySessionEvent(event)
}

// NotifySessionShared notifies clients when a session is shared with someone
func (n *Notifier) NotifySessionShared(sessionID, ownerUserID, sharedWithUserID string, permissions []string) {
	// Notify owner
	event := SessionEvent{
		Type:      EventSessionShared,
		SessionID: sessionID,
		UserID:    ownerUserID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sharedWith":  sharedWithUserID,
			"permissions": permissions,
		},
	}
	n.NotifySessionEvent(event)

	// Notify the user it was shared with
	eventForSharedUser := SessionEvent{
		Type:      EventSessionShared,
		SessionID: sessionID,
		UserID:    sharedWithUserID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sharedBy":    ownerUserID,
			"permissions": permissions,
		},
	}
	n.NotifySessionEvent(eventForSharedUser)
}

// NotifySessionError notifies clients about session errors
func (n *Notifier) NotifySessionError(sessionID, userID string, errorMsg string) {
	event := SessionEvent{
		Type:      EventSessionError,
		SessionID: sessionID,
		UserID:    userID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"error": errorMsg,
		},
	}
	n.NotifySessionEvent(event)
}

// CloseAll closes all subscriptions (used during shutdown)
func (n *Notifier) CloseAll() {
	n.mu.Lock()
	defer n.mu.Unlock()

	log.Println("Closing all WebSocket subscriptions...")

	// Clear all subscriptions
	n.userSubscriptions = make(map[string]map[string]bool)
	n.sessionSubscriptions = make(map[string]map[string]bool)
	n.clientUsers = make(map[string]string)

	log.Println("All subscriptions closed")
}
