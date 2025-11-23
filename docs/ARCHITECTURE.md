<div align="center">

# 🏗️ StreamSpace Architecture

**Version**: v2.0-beta • **Last Updated**: 2025-11-21

[![Status](https://img.shields.io/badge/Status-v2.0--beta-success.svg)](../CHANGELOG.md)

</div>

---

> [!IMPORTANT]
> **v2.0 Architecture Update**
>
> StreamSpace has evolved to a **Control Plane + Agent** architecture. The Control Plane acts as the central management hub, while Agents (Kubernetes, Docker, etc.) execute commands and manage resources on their respective platforms.

## 🧩 System Overview

StreamSpace is a platform-agnostic container streaming platform. It separates the management logic (Control Plane) from the execution logic (Agents), allowing for scalability and multi-platform support.

### High-Level Architecture

```mermaid
graph TD
    User[User / Browser] -->|HTTPS| Ingress[Ingress / Load Balancer]
    Ingress -->|HTTPS| UI[Web UI]
    Ingress -->|HTTPS/WSS| API[Control Plane API]
    
    subgraph "Control Plane"
        UI
        API
        DB[(PostgreSQL)]
        API --> DB
    end
    
    subgraph "Execution Plane (Kubernetes)"
        K8sAgent[K8s Agent]
        K8sAgent <-->|WebSocket| API
        K8sAgent -->|Manage| Pods[Session Pods]
        API -.->|VNC Proxy| K8sAgent
        K8sAgent -.->|Tunnel| Pods
    end

    subgraph "Execution Plane (Docker - Planned)"
        DockerAgent[Docker Agent]
        DockerAgent <-->|WebSocket| API
        DockerAgent -->|Manage| Containers[Session Containers]
    end
```

## 📦 Core Components

### 1. Control Plane (API)

- **Role**: Central brain of the system.
- **Tech**: Go (Gin framework).
- **Responsibilities**:
  - User Authentication & Authorization (SAML, OIDC).
  - Session Management (CRUD).
  - Agent Coordination (WebSocket Hub).
  - VNC Proxying (Secure tunneling).
  - Database Management.

### 2. Execution Agents

- **Role**: Platform-specific executors.
- **Tech**: Go.
- **Types**:
  - **Kubernetes Agent**: Manages Pods, PVCs, Services.
  - **Docker Agent** (v2.1): Manages Containers, Volumes.
- **Responsibilities**:
  - Connect to Control Plane via secure WebSocket.
  - Execute commands (Start, Stop, Hibernate).
  - Report status and metrics (Heartbeats).
  - Tunnel VNC traffic.

### 3. Web UI

- **Role**: User interface.
- **Tech**: React + TypeScript + Material-UI.
- **Features**:
  - Dashboard & Catalog.
  - Session Viewer (noVNC integration).
  - Admin Panel (User, Agent, Plugin management).

### 4. Session Workspaces

- **Role**: The actual user environment.
- **Tech**: Containerized applications (LinuxServer.io images).
- **Features**:
  - KasmVNC for streaming.
  - Persistent home directory.
  - Isolated environment.

## 🔄 Data Flow

### Session Creation

```mermaid
sequenceDiagram
    participant User
    participant API as Control Plane
    participant DB as Database
    participant Agent as K8s Agent
    participant K8s as Kubernetes

    User->>API: POST /api/v1/sessions
    API->>DB: Check Quota & Create Record
    API->>Agent: Send Command (StartSession)
    Agent->>K8s: Create Deployment/Service/PVC
    K8s-->>Agent: Pod Ready (IP: 10.42.x.x)
    Agent->>API: Update Status (Running)
    API-->>User: Session Ready
```

### VNC Streaming (v2.0 Proxy)

```mermaid
sequenceDiagram
    participant User
    participant API as Control Plane Proxy
    participant Agent as K8s Agent
    participant Pod as Session Pod

    User->>API: WebSocket Connect (/api/v1/vnc/:id)
    API->>Agent: Route VNC Traffic
    Agent->>Pod: Port Forward (5900)
    Pod-->>Agent: VNC Data
    Agent-->>API: VNC Data
    API-->>User: VNC Data
```

## 🛡️ Security Architecture

### Authentication

- **SSO**: Authentik, Okta, Azure AD via OIDC/SAML.
- **Tokens**: JWT (Access + Refresh).
- **MFA**: TOTP support.

### Network Security

- **Ingress**: TLS/SSL enforced.
- **Isolation**: Network Policies deny inter-pod traffic by default.
- **Proxy**: All VNC traffic flows through Control Plane (no direct pod access).

### Data Protection

- **Storage**: Per-user PVCs with RBAC.
- **Encryption**: Secrets management for sensitive data.
- **Audit**: Comprehensive logging of all actions.

## 💾 Resource Management

### Quotas

- **Per User**: Max sessions, CPU, Memory.
- **Enforcement**: Checked at API level before command dispatch.

### Hibernation

- **Auto-Scale**: Idle sessions scale to 0 replicas.
- **Wake**: Instant resume on user interaction.
- **Persistence**: PVCs remain mounted/available.

## 🔌 Plugin System

The plugin system allows extending functionality without modifying the core.

- **Types**: Extension, Webhook, Integration, Theme.
- **Storage**: JSONB configuration in database.
- **Events**: Plugins can subscribe to system events (SessionStart, UserLogin, etc.).

---

<div align="center">
  <sub>StreamSpace Architecture Documentation</sub>
</div>
