package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Publisher handles publishing events to NATS.
type Publisher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	enabled bool
}

// Config holds NATS connection configuration.
type Config struct {
	URL      string
	User     string
	Password string
	TLS      bool
}

// NewPublisher creates a new NATS event publisher.
// If NATS is unavailable, returns a disabled publisher that logs warnings.
func NewPublisher(cfg Config) (*Publisher, error) {
	if cfg.URL == "" {
		cfg.URL = os.Getenv("NATS_URL")
	}
	if cfg.URL == "" {
		log.Println("Warning: NATS_URL not configured, event publishing disabled")
		return &Publisher{enabled: false}, nil
	}

	// Build connection options
	opts := []nats.Option{
		nats.Name("streamspace-api"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Printf("NATS error: %v", err)
		}),
	}

	// Add authentication if configured
	if cfg.User != "" {
		opts = append(opts, nats.UserInfo(cfg.User, cfg.Password))
	}

	// Connect to NATS
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS at %s: %v", cfg.URL, err)
		log.Println("Event publishing disabled - controllers will not receive events")
		return &Publisher{enabled: false}, nil
	}

	log.Printf("Connected to NATS at %s", conn.ConnectedUrl())

	// Try to get JetStream context for persistence (optional)
	js, err := conn.JetStream()
	if err != nil {
		log.Printf("JetStream not available: %v (using core NATS)", err)
	} else {
		// Create streams for durable message delivery
		if err := createStreams(js); err != nil {
			log.Printf("Warning: Failed to create JetStream streams: %v", err)
			log.Println("Events will be published without durability guarantees")
			js = nil
		} else {
			log.Println("JetStream streams configured for durable event delivery")
		}
	}

	return &Publisher{
		conn:    conn,
		js:      js,
		enabled: true,
	}, nil
}

// createStreams creates JetStream streams for durable event delivery.
func createStreams(js nats.JetStreamContext) error {
	streams := []struct {
		name     string
		subjects []string
	}{
		{
			name: "STREAMSPACE_SESSIONS",
			subjects: []string{
				"streamspace.session.>",
			},
		},
		{
			name: "STREAMSPACE_APPS",
			subjects: []string{
				"streamspace.app.>",
			},
		},
		{
			name: "STREAMSPACE_TEMPLATES",
			subjects: []string{
				"streamspace.template.>",
			},
		},
		{
			name: "STREAMSPACE_NODES",
			subjects: []string{
				"streamspace.node.>",
			},
		},
		{
			name: "STREAMSPACE_CONTROLLERS",
			subjects: []string{
				"streamspace.controller.>",
			},
		},
	}

	for _, s := range streams {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:      s.name,
			Subjects:  s.subjects,
			Retention: nats.WorkQueuePolicy, // Messages deleted after acknowledgment
			MaxAge:    24 * time.Hour,       // Keep messages for 24 hours max
			Storage:   nats.FileStorage,     // Persist to disk
			Replicas:  1,                    // Single replica for simplicity
		})
		if err != nil {
			// Stream might already exist, try to update it
			if err.Error() != "stream name already in use" {
				return fmt.Errorf("failed to create stream %s: %w", s.name, err)
			}
		}
	}

	return nil
}

// Close closes the NATS connection.
func (p *Publisher) Close() {
	if p.conn != nil {
		p.conn.Drain()
		p.conn.Close()
	}
}

// IsEnabled returns whether event publishing is enabled.
func (p *Publisher) IsEnabled() bool {
	return p.enabled
}

// Publish publishes an event to the given subject.
func (p *Publisher) Publish(subject string, event interface{}) error {
	if !p.enabled {
		log.Printf("Event publishing disabled, skipping: %s", subject)
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to %s: %w", subject, err)
	}

	log.Printf("Published event to %s", subject)
	return nil
}

// PublishWithPlatform publishes an event to a platform-specific subject.
func (p *Publisher) PublishWithPlatform(subject, platform string, event interface{}) error {
	// Publish to both generic and platform-specific subjects
	if err := p.Publish(subject, event); err != nil {
		return err
	}
	return p.Publish(SubjectWithPlatform(subject, platform), event)
}

// Request publishes a request and waits for a response.
func (p *Publisher) Request(subject string, event interface{}, timeout time.Duration) (*nats.Msg, error) {
	if !p.enabled {
		return nil, fmt.Errorf("event publishing disabled")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.conn.Request(subject, data, timeout)
}

// Subscribe subscribes to a subject with a handler.
func (p *Publisher) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if !p.enabled {
		return nil, fmt.Errorf("event publishing disabled")
	}
	return p.conn.Subscribe(subject, handler)
}

// QueueSubscribe subscribes to a subject with a queue group.
func (p *Publisher) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if !p.enabled {
		return nil, fmt.Errorf("event publishing disabled")
	}
	return p.conn.QueueSubscribe(subject, queue, handler)
}

// Helper methods for publishing specific events

// PublishSessionCreate publishes a session create event.
func (p *Publisher) PublishSessionCreate(ctx context.Context, event *SessionCreateEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectSessionCreate, event.Platform, event)
}

// PublishSessionDelete publishes a session delete event.
func (p *Publisher) PublishSessionDelete(ctx context.Context, event *SessionDeleteEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectSessionDelete, event.Platform, event)
}

// PublishSessionHibernate publishes a session hibernate event.
func (p *Publisher) PublishSessionHibernate(ctx context.Context, event *SessionHibernateEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectSessionHibernate, event.Platform, event)
}

// PublishSessionWake publishes a session wake event.
func (p *Publisher) PublishSessionWake(ctx context.Context, event *SessionWakeEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectSessionWake, event.Platform, event)
}

// PublishAppInstall publishes an application install event.
func (p *Publisher) PublishAppInstall(ctx context.Context, event *AppInstallEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectAppInstall, event.Platform, event)
}

// PublishAppUninstall publishes an application uninstall event.
func (p *Publisher) PublishAppUninstall(ctx context.Context, event *AppUninstallEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectAppUninstall, event.Platform, event)
}

// PublishTemplateCreate publishes a template create event.
func (p *Publisher) PublishTemplateCreate(ctx context.Context, event *TemplateCreateEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectTemplateCreate, event.Platform, event)
}

// PublishTemplateDelete publishes a template delete event.
func (p *Publisher) PublishTemplateDelete(ctx context.Context, event *TemplateDeleteEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectTemplateDelete, event.Platform, event)
}

// PublishNodeCordon publishes a node cordon event.
func (p *Publisher) PublishNodeCordon(ctx context.Context, event *NodeCordonEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectNodeCordon, event.Platform, event)
}

// PublishNodeUncordon publishes a node uncordon event.
func (p *Publisher) PublishNodeUncordon(ctx context.Context, event *NodeUncordonEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectNodeUncordon, event.Platform, event)
}

// PublishNodeDrain publishes a node drain event.
func (p *Publisher) PublishNodeDrain(ctx context.Context, event *NodeDrainEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return p.PublishWithPlatform(SubjectNodeDrain, event.Platform, event)
}

// GetConnection returns the underlying NATS connection.
// Use with caution - prefer using Publisher methods.
func (p *Publisher) GetConnection() *nats.Conn {
	return p.conn
}
