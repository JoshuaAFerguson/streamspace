# NATS Event Architecture

## Overview

StreamSpace uses NATS as the message broker between the API and platform controllers. This enables:
- Event-driven communication (millisecond latency)
- Multiple platform controllers (Kubernetes, Docker, Hyper-V, vCenter)
- Clean decoupling of API from platform-specific operations
- Scalable and fault-tolerant architecture

## Architecture Diagram

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Web UI    │ ──► │     API      │ ──► │   Database   │
└─────────────┘     └──────┬───────┘     │ (state)      │
                           │             └──────────────┘
                           │ publish
                           ▼
                    ┌──────────────┐
                    │     NATS     │
                    └──────┬───────┘
                           │ subscribe
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌────────────┐  ┌────────────┐  ┌────────────┐
    │    K8s     │  │   Docker   │  │  vCenter   │
    │ Controller │  │ Controller │  │ Controller │
    └────────────┘  └────────────┘  └────────────┘
```

## Subject Naming Convention

Format: `streamspace.<domain>.<action>.<platform?>`

### Core Subjects

| Subject | Description | Publisher | Subscriber |
|---------|-------------|-----------|------------|
| `streamspace.session.create` | Create new session | API | Controllers |
| `streamspace.session.delete` | Delete session | API | Controllers |
| `streamspace.session.hibernate` | Hibernate session | API | Controllers |
| `streamspace.session.wake` | Wake hibernated session | API | Controllers |
| `streamspace.session.status` | Session status update | Controllers | API |
| `streamspace.app.install` | Install application | API | Controllers |
| `streamspace.app.uninstall` | Uninstall application | API | Controllers |
| `streamspace.app.status` | App installation status | Controllers | API |
| `streamspace.template.create` | Create template | Controllers | API |
| `streamspace.template.delete` | Delete template | API | Controllers |
| `streamspace.node.cordon` | Cordon node | API | Controllers |
| `streamspace.node.drain` | Drain node | API | Controllers |
| `streamspace.controller.heartbeat` | Controller health | Controllers | API |

### Platform-Specific Subjects

Controllers subscribe to platform-specific subjects:
- `streamspace.session.create.kubernetes` - K8s controller only
- `streamspace.session.create.docker` - Docker controller only
- `streamspace.session.create.hyperv` - Hyper-V controller only

## Message Payloads

### Session Create Event

```json
{
  "event_id": "uuid",
  "timestamp": "2025-01-15T10:30:00Z",
  "session_id": "uuid",
  "user_id": "user1",
  "template_id": "firefox-browser",
  "platform": "kubernetes",
  "resources": {
    "memory": "2Gi",
    "cpu": "1000m"
  },
  "persistent_home": true,
  "idle_timeout": "30m",
  "metadata": {
    "request_id": "uuid",
    "source_ip": "192.168.1.1"
  }
}
```

### Session Status Event (from Controller)

```json
{
  "event_id": "uuid",
  "timestamp": "2025-01-15T10:30:05Z",
  "session_id": "uuid",
  "status": "running",
  "phase": "Running",
  "url": "https://user1-firefox.streamspace.local",
  "pod_name": "ss-user1-firefox-abc123",
  "message": "Session started successfully",
  "resource_usage": {
    "memory": "512Mi",
    "cpu": "250m"
  }
}
```

### Application Install Event

```json
{
  "event_id": "uuid",
  "timestamp": "2025-01-15T10:30:00Z",
  "install_id": "uuid",
  "catalog_template_id": 42,
  "template_name": "firefox-browser",
  "display_name": "Firefox Web Browser",
  "manifest": "apiVersion: stream.space/v1alpha1\nkind: Template\n...",
  "installed_by": "admin",
  "platform": "kubernetes"
}
```

### Application Status Event (from Controller)

```json
{
  "event_id": "uuid",
  "timestamp": "2025-01-15T10:30:10Z",
  "install_id": "uuid",
  "status": "ready",
  "template_name": "firefox-browser",
  "template_namespace": "streamspace",
  "message": "Template created successfully"
}
```

### Controller Heartbeat

```json
{
  "controller_id": "k8s-controller-1",
  "platform": "kubernetes",
  "timestamp": "2025-01-15T10:30:00Z",
  "status": "healthy",
  "version": "1.0.0",
  "capabilities": ["sessions", "templates", "nodes"],
  "cluster_info": {
    "name": "production",
    "nodes": 5,
    "version": "1.28.0"
  }
}
```

## Database Schema Changes

### New Tables

#### `platform_controllers`
Tracks registered controllers and their capabilities.

```sql
CREATE TABLE platform_controllers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    controller_id VARCHAR(255) UNIQUE NOT NULL,
    platform VARCHAR(50) NOT NULL, -- kubernetes, docker, hyperv, vcenter
    display_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'unknown', -- healthy, unhealthy, unknown
    version VARCHAR(50),
    capabilities JSONB DEFAULT '[]',
    cluster_info JSONB DEFAULT '{}',
    last_heartbeat TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

#### `event_log`
Audit log of all events for debugging and replay.

```sql
CREATE TABLE event_log (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    subject VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    processed_by VARCHAR(255),
    status VARCHAR(50) DEFAULT 'published', -- published, processing, completed, failed
    error_message TEXT
);

CREATE INDEX idx_event_log_subject ON event_log(subject);
CREATE INDEX idx_event_log_status ON event_log(status);
CREATE INDEX idx_event_log_published_at ON event_log(published_at);
```

### Modified Tables

#### `installed_applications`
Add status tracking for async installation.

```sql
ALTER TABLE installed_applications ADD COLUMN IF NOT EXISTS
    install_status VARCHAR(50) DEFAULT 'pending'; -- pending, installing, ready, failed

ALTER TABLE installed_applications ADD COLUMN IF NOT EXISTS
    install_message TEXT;

ALTER TABLE installed_applications ADD COLUMN IF NOT EXISTS
    platform VARCHAR(50) DEFAULT 'kubernetes';
```

#### `sessions` (if exists, or create)
Add platform field for multi-platform support.

```sql
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS
    platform VARCHAR(50) DEFAULT 'kubernetes';

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS
    controller_id VARCHAR(255);
```

## API Changes

### New Endpoints

```
GET  /api/v1/controllers          - List registered controllers
GET  /api/v1/controllers/:id      - Get controller details
GET  /api/v1/platforms            - List available platforms
```

### Modified Endpoints

All session/application endpoints become async:
- `POST /api/v1/sessions` - Returns immediately with `status: pending`
- `POST /api/v1/applications` - Returns immediately with `install_status: pending`

Frontend polls for status updates or uses WebSocket for real-time updates.

## Controller Implementation

### Subscription Pattern

```go
// Each controller subscribes to its platform-specific subjects
func (c *Controller) Subscribe(nc *nats.Conn) error {
    platform := c.Platform // e.g., "kubernetes"

    // Subscribe to platform-specific events
    nc.Subscribe(fmt.Sprintf("streamspace.session.create.%s", platform), c.handleSessionCreate)
    nc.Subscribe(fmt.Sprintf("streamspace.session.delete.%s", platform), c.handleSessionDelete)
    nc.Subscribe(fmt.Sprintf("streamspace.app.install.%s", platform), c.handleAppInstall)

    // Subscribe to broadcast events (all platforms)
    nc.Subscribe("streamspace.session.create", c.handleSessionCreateIfMatches)

    return nil
}
```

### Publishing Status Updates

```go
func (c *Controller) publishSessionStatus(nc *nats.Conn, session *Session) error {
    event := SessionStatusEvent{
        EventID:   uuid.New().String(),
        Timestamp: time.Now(),
        SessionID: session.ID,
        Status:    session.Status,
        Phase:     session.Phase,
        URL:       session.URL,
        Message:   session.Message,
    }

    data, _ := json.Marshal(event)
    return nc.Publish("streamspace.session.status", data)
}
```

## Configuration

### Environment Variables

```bash
# NATS Connection
NATS_URL=nats://localhost:4222
NATS_USER=streamspace
NATS_PASSWORD=secret
NATS_TLS_ENABLED=false

# Controller Registration
CONTROLLER_ID=k8s-controller-1
CONTROLLER_PLATFORM=kubernetes
HEARTBEAT_INTERVAL=30s
```

### Docker Compose Addition

```yaml
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
      - "8222:8222"  # Monitoring
    command: ["--jetstream", "--store_dir", "/data"]
    volumes:
      - nats_data:/data

volumes:
  nats_data:
```

## Error Handling

### Retry Strategy

Controllers implement exponential backoff for failed operations:
- Initial delay: 1 second
- Max delay: 5 minutes
- Max retries: 10

### Dead Letter Queue

Failed events after max retries go to:
`streamspace.dlq.<original-subject>`

### Circuit Breaker

If a controller fails repeatedly, it's marked as unhealthy and removed from routing.

## Monitoring

### NATS Metrics

- `nats_msgs_received_total` - Messages received by subject
- `nats_msgs_published_total` - Messages published by subject
- `nats_pending_msgs` - Messages pending in queue

### Custom Metrics

- `streamspace_events_published_total` - Events published by type
- `streamspace_events_processed_total` - Events processed by controller
- `streamspace_event_latency_seconds` - Time from publish to process
- `streamspace_controller_health` - Controller health status

## Migration Plan

### Phase 1: Add NATS Infrastructure
1. Add NATS to docker-compose
2. Create NATS client wrapper in API
3. Add event publishing alongside existing K8s calls

### Phase 2: Update Controllers
1. Add NATS subscription to K8s controller
2. Implement status publishing
3. Run in parallel with existing direct K8s calls

### Phase 3: Remove K8s from API
1. Remove k8sClient from API handlers
2. Update frontend for async operations
3. Remove ApplicationInstall CRD (no longer needed)

### Phase 4: Add New Controllers
1. Docker controller
2. Hyper-V controller
3. vCenter controller

## Security Considerations

- Use TLS for NATS connections in production
- Implement authentication (user/password or NKey)
- Consider NATS authorization for subject-level permissions
- Encrypt sensitive data in payloads (credentials, tokens)
- Rate limit event publishing to prevent DoS
