# StreamSpace Quick Start Guide

Get StreamSpace up and running in under 10 minutes.

**Last Updated**: 2025-11-15
**Version**: v1.0.0

---

## Prerequisites

Before you begin, ensure you have:

- **Kubernetes cluster** (1.19+)
  - k3s recommended for self-hosting
  - Minimum: 4 CPU cores, 16GB RAM, 100GB storage
- **kubectl** configured with cluster access
- **Helm 3.0+** installed
- **Storage provisioner** with ReadWriteMany support (NFS recommended)
- **PostgreSQL database** (can be deployed by StreamSpace)

**Optional but recommended**:
- **Authentik** or **Keycloak** for SSO authentication
- **MetalLB** or cloud LoadBalancer for ingress

---

## Installation

### Option 1: Quick Install (Helm - Recommended)

```bash
# 1. Create namespace
kubectl create namespace streamspace

# 2. Install StreamSpace
helm install streamspace ./chart -n streamspace

# 3. Get the UI URL
kubectl get svc -n streamspace streamspace-ui
```

Access the UI at the LoadBalancer IP or configure ingress.

### Option 2: Manual Install

```bash
# 1. Clone repository
git clone https://github.com/yourusername/streamspace.git
cd streamspace

# 2. Deploy CRDs
kubectl apply -f manifests/crds/

# 3. Deploy configuration
kubectl apply -f manifests/config/

# 4. Deploy application templates
kubectl apply -f manifests/templates/

# 5. Install via Helm
helm install streamspace ./chart -n streamspace
```

---

## First Steps

### 1. Access the Web UI

```bash
# Get the UI service URL
kubectl get svc -n streamspace streamspace-ui

# Or use port-forward for testing
kubectl port-forward -n streamspace svc/streamspace-ui 8080:80
```

Open your browser to: `http://localhost:8080`

### 2. Create Your First User

**Using Local Authentication**:

```bash
# Create user via API or web UI
# Default admin credentials are set during installation
# See chart/values.yaml for configuration
```

**Using SSO** (Authentik/Keycloak):

Configure OIDC or SAML in `chart/values.yaml`:

```yaml
auth:
  mode: oidc  # or saml
  oidc:
    enabled: true
    providerURL: https://auth.example.com
    clientID: streamspace
    clientSecret: YOUR_SECRET
```

### 3. Launch Your First Session

**Via Web UI**:
1. Log in to StreamSpace
2. Browse the **Application Catalog**
3. Click **Launch** on any application (e.g., Firefox)
4. Wait ~30 seconds for session to start
5. Access your session in the browser

**Via kubectl**:

```bash
# Create a Firefox session
kubectl apply -f - <<EOF
apiVersion: stream.space/v1alpha1
kind: Session
metadata:
  name: my-firefox
  namespace: streamspace
spec:
  user: john
  template: firefox-browser
  state: running
  resources:
    memory: 2Gi
    cpu: 1000m
  persistentHome: true
  idleTimeout: 30m
EOF

# Check session status
kubectl get sessions -n streamspace

# Get session URL
kubectl get session my-firefox -n streamspace -o jsonpath='{.status.url}'
```

### 4. Verify Installation

```bash
# Check all pods are running
kubectl get pods -n streamspace

# Check controller logs
kubectl logs -n streamspace deploy/streamspace-controller -f

# Check available templates
kubectl get templates -n streamspace

# List all sessions
kubectl get sessions -n streamspace
```

---

## Configuration

### Basic Settings

Edit `chart/values.yaml` before installation:

```yaml
# Controller configuration
controller:
  config:
    hibernation:
      enabled: true
      defaultIdleTimeout: 30m

# Resource defaults
resources:
  defaultMemory: 2Gi
  defaultCPU: 1000m

# Storage
storage:
  className: nfs-client
  defaultHomeSize: 50Gi

# Ingress
ingress:
  enabled: true
  hostname: streamspace.example.com
  className: traefik
```

### PostgreSQL Configuration

**Using external PostgreSQL**:

```yaml
postgresql:
  enabled: false
  externalHost: postgres.example.com
  externalPort: 5432
  database: streamspace
  username: streamspace
  password: YOUR_SECURE_PASSWORD  # Use secrets in production!
```

**Using bundled PostgreSQL** (development only):

```yaml
postgresql:
  enabled: true
  postgresPassword: CHANGE_ME  # ⚠️ Change this!
```

⚠️ **PRODUCTION WARNING**: Always use a secure password and proper secret management (Sealed Secrets, External Secrets Operator, or SOPS).

### Authentication Configuration

**Local Authentication** (default):

```yaml
auth:
  mode: local
  jwtSecret: YOUR_SECRET_KEY  # Generate with: openssl rand -base64 32
```

**SAML 2.0 SSO**:

```yaml
auth:
  mode: saml
  saml:
    enabled: true
    provider: okta  # okta, azuread, keycloak, authentik, auth0, generic
    metadataURL: https://your-idp.com/metadata
    entityID: streamspace
    spCertPath: /path/to/sp-cert.pem
    spKeyPath: /path/to/sp-key.pem
```

See [docs/SAML_SETUP.md](docs/SAML_SETUP.md) for detailed SAML configuration.

**OIDC OAuth2**:

```yaml
auth:
  mode: oidc
  oidc:
    enabled: true
    provider: keycloak  # keycloak, okta, auth0, google, azuread, github, gitlab, generic
    providerURL: https://auth.example.com
    clientID: streamspace
    clientSecret: YOUR_SECRET
    redirectURI: https://streamspace.example.com/auth/oidc/callback
```

---

## Common Tasks

### View All Sessions

```bash
# All sessions
kubectl get sessions -n streamspace

# User's sessions
kubectl get sessions -n streamspace -l user=john

# Running sessions only
kubectl get ss -n streamspace --field-selector spec.state=running
```

### Hibernate a Session

```bash
kubectl patch session my-firefox -n streamspace \
  --type merge -p '{"spec":{"state":"hibernated"}}'
```

### Wake a Session

```bash
kubectl patch session my-firefox -n streamspace \
  --type merge -p '{"spec":{"state":"running"}}'
```

### Delete a Session

```bash
kubectl delete session my-firefox -n streamspace
```

### View Available Templates

```bash
# List all templates
kubectl get templates -n streamspace

# Filter by category
kubectl get tpl -n streamspace -l category="Web Browsers"

# Get template details
kubectl describe template firefox-browser -n streamspace
```

### Check Resource Usage

```bash
# Pod resource usage
kubectl top pods -n streamspace

# Session resource usage
kubectl get sessions -n streamspace -o wide
```

### View Logs

```bash
# Controller logs
kubectl logs -n streamspace deploy/streamspace-controller -f

# API logs
kubectl logs -n streamspace deploy/streamspace-api -f

# Session pod logs
kubectl logs -n streamspace <pod-name>
```

---

## Monitoring

### Access Grafana

```bash
# Port forward to Grafana
kubectl port-forward -n observability svc/grafana 3000:80

# Open http://localhost:3000
# Default credentials: admin/admin
```

**Available Dashboards**:
- Session Overview - Active/hibernated sessions, resource usage
- User Activity - Logins, launches, session duration
- Cluster Capacity - Resource utilization, queue depth
- API Performance - Request rates, latency, errors

### View Prometheus Metrics

```bash
# Port forward to controller metrics
kubectl port-forward -n streamspace deploy/streamspace-controller 8080:8080

# Query metrics
curl http://localhost:8080/metrics | grep streamspace
```

**Key Metrics**:
- `streamspace_active_sessions_total` - Active sessions
- `streamspace_hibernated_sessions_total` - Hibernated sessions
- `streamspace_session_starts_total` - Session creation counter
- `streamspace_resource_usage_bytes` - Resource consumption

See [FEATURES.md](FEATURES.md#observability-metrics) for complete metrics list.

---

## Troubleshooting

### Sessions Not Starting

```bash
# Check session status
kubectl describe session <name> -n streamspace

# Check controller logs
kubectl logs -n streamspace deploy/streamspace-controller -f

# Check pod status
kubectl get pods -n streamspace -l session=<name>

# Check events
kubectl get events -n streamspace --sort-by=.metadata.creationTimestamp
```

**Common Issues**:
- **Image pull errors**: Check image name and registry access
- **PVC mount errors**: Verify NFS provisioner is working
- **Resource limits**: Check node capacity

### Hibernation Not Working

```bash
# Verify hibernation is enabled
kubectl get cm -n streamspace streamspace-config -o yaml | grep hibernation

# Check lastActivity timestamp
kubectl get session <name> -n streamspace -o jsonpath='{.status.lastActivity}'

# Check hibernation controller logs
kubectl logs -n streamspace deploy/streamspace-controller -f | grep -i hibernation
```

### Cannot Access Session URL

```bash
# Check ingress
kubectl get ingress -n streamspace

# Check ingress controller
kubectl get pods -n kube-system -l app.kubernetes.io/name=traefik

# Check service
kubectl get svc -n streamspace -l session=<name>

# Test connectivity
kubectl port-forward -n streamspace svc/<service-name> 3000:3000
# Access http://localhost:3000
```

### PVC Stuck in Pending

```bash
# Check PVC status
kubectl describe pvc home-<username> -n streamspace

# Check storage class
kubectl get storageclass

# Verify NFS provisioner
kubectl get pods -n kube-system | grep nfs
```

**Common Fixes**:
- Install NFS provisioner
- Verify NFS server is accessible
- Check storage class exists and is default

---

## Next Steps

### Learn More

- **[Features Guide](FEATURES.md)** - Complete list of all features
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture and design
- **[User Guide](docs/USER_GUIDE.md)** - End-user documentation
- **[Admin Guide](docs/ADMIN_GUIDE.md)** - Administrator documentation
- **[Plugin Development](PLUGIN_DEVELOPMENT.md)** - Build custom plugins

### Production Deployment

- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment instructions
- **[Security Guide](docs/SECURITY.md)** - Security best practices
- **[SAML Setup](docs/SAML_SETUP.md)** - SAML 2.0 SSO configuration
- **[AWS Deployment](docs/AWS_DEPLOYMENT.md)** - AWS-specific deployment

### Advanced Features

- **Plugin System** - Extend functionality with plugins
- **Webhooks** - Integrate with external services (16 event types)
- **Compliance** - SOC2, HIPAA, GDPR frameworks
- **Collaboration** - Real-time chat, annotations, screen sharing
- **Scheduling** - Automate session start/stop times

---

## Getting Help

### Documentation

- **README**: [README.md](README.md) - Project overview
- **Roadmap**: [ROADMAP.md](ROADMAP.md) - Development roadmap
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

### Community

- **GitHub Issues**: https://github.com/yourusername/streamspace/issues
- **GitHub Discussions**: https://github.com/yourusername/streamspace/discussions
- **Discord**: https://discord.gg/streamspace (coming soon)

### Support

- **Email**: support@streamspace.io
- **Security Issues**: security@streamspace.io

---

## Uninstall

```bash
# Uninstall Helm release
helm uninstall streamspace -n streamspace

# Delete CRDs (⚠️ this will delete all sessions and templates)
kubectl delete crd sessions.stream.space
kubectl delete crd templates.stream.space

# Delete namespace
kubectl delete namespace streamspace
```

⚠️ **WARNING**: This will delete all user data and sessions. Back up important data first.

---

**Welcome to StreamSpace!** 🚀

For questions or feedback, visit our [GitHub repository](https://github.com/yourusername/streamspace).
