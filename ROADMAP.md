<div align="center">

# 🗺️ StreamSpace Roadmap

**Current Version**: v2.0-beta • **Last Updated**: 2025-11-21

[![Status](https://img.shields.io/badge/Status-v2.0--beta-success.svg)](CHANGELOG.md)

</div>

---

> [!IMPORTANT]
> **Current Status: v2.0-beta (Integration Testing)**
>
> StreamSpace has completed the major architectural transformation to a multi-platform Control Plane + Agent model. We are currently in the integration testing phase.

## 📅 Release Timeline

```mermaid
gantt
    title StreamSpace Development Roadmap
    dateFormat  YYYY-MM-DD
    section v1.0
    Core Platform       :done,    des1, 2025-10-01, 2025-11-01
    Admin UI            :done,    des2, 2025-11-01, 2025-11-15
    Security Hardening  :done,    des3, 2025-11-01, 2025-11-15
    v1.0 Release        :done,    des4, 2025-11-21, 1d

    section v2.0 (Current)
    Architecture Design :done,    v2_1, 2025-11-21, 1d
    Control Plane       :done,    v2_2, 2025-11-21, 3d
    K8s Agent           :done,    v2_3, 2025-11-21, 3d
    VNC Proxy           :done,    v2_4, 2025-11-21, 2d
    Integration Testing :active,  v2_5, 2025-11-21, 7d
    v2.0 Stable         :         v2_6, after v2_5, 1d

    section Future
    Docker Agent (v2.1) :         v2_7, after v2_6, 14d
    VNC Independence    :         v3_0, 2026-01-01, 60d
```

## 🎯 Priorities

### 1. Integration Testing (Immediate)

- **Focus**: Validate the new v2.0 Control Plane + Agent architecture.
- **Tasks**:
  - [ ] E2E VNC streaming validation
  - [ ] Multi-agent session creation
  - [ ] Failover and reconnection tests
  - [ ] Performance benchmarking

### 2. Test Coverage Expansion (High)

- **Current**: ~70% on new v2.0 code, ~20% overall.
- **Target**: 80%+ overall.
- **Plan**:
  - [ ] Expand API handler tests
  - [ ] Add UI component tests
  - [ ] Increase controller test coverage

### 3. Plugin Implementation (Medium)

- **Current**: Framework complete, 28 stub plugins.
- **Target**: Working implementations for top 10 plugins.
- **Top Plugins**:
  - Calendar, Slack, Teams, Discord, PagerDuty
  - Compliance, DLP, Analytics

### 4. Docker Support (v2.1)

- **Current**: Planned.
- **Target**: Functional Docker Agent.
- **Scope**:
  - Container lifecycle management
  - Local volume management
  - Network configuration

## 🛤️ Detailed Roadmap

### v1.0.0-READY (Completed) ✅

- **Core**: Functional Kubernetes platform
- **Auth**: Complete authentication stack (SAML, OIDC, MFA)
- **Admin**: Full admin UI and configuration
- **Security**: Production-hardened (Audit logs, RBAC, Security headers)

### v2.0-beta (Current) 🚧

- **Architecture**: Multi-platform Control Plane + Agent
- **Connectivity**: Secure VNC Proxy (Firewall-friendly)
- **Agent**: Kubernetes Agent implementation
- **UI**: Real-time agent monitoring

### v2.1 (Planned) 📝

- **Platform**: Docker Agent support
- **Deployment**: Docker Compose support
- **Storage**: Local volume management

### v3.0 (Future) 🔮

- **Streaming**: WebRTC support for lower latency
- **VNC**: Migration to TigerVNC + noVNC (native images)
- **Hardware**: GPU acceleration support
- **Federation**: Multi-cluster support

## 🤝 How to Contribute

We welcome contributions! Here are the high-impact areas:

1. **Testing**: Help us reach our 80% coverage goal.
2. **Plugins**: Pick a stub plugin and implement it.
3. **Documentation**: Improve guides and examples.

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

<div align="center">
  <sub>StreamSpace Roadmap</sub>
</div>
