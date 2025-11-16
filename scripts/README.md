# StreamSpace Deployment Scripts

This directory contains scripts for building and deploying StreamSpace locally.

## Quick Start

### 1. Build Local Images

```bash
./scripts/local-build.sh
```

This builds Docker images for the controller, API, and UI.

### 2. Deploy to Local Kubernetes

**Choose the appropriate deployment method based on your Helm version:**

#### If you have Helm v3.19.0 (Docker Desktop users) → Use kubectl

```bash
./scripts/local-deploy-kubectl.sh
```

Helm v3.19.0 has a critical bug that breaks all chart operations. The kubectl-based script bypasses Helm entirely.

#### If you have Helm v3.18.0 or earlier → Use Helm

```bash
./scripts/local-deploy.sh
```

This uses the Helm chart for deployment.

### 3. Check Your Helm Version

```bash
helm version --short
```

- **v3.19.0 or later**: Use `local-deploy-kubectl.sh`
- **v3.18.0 or earlier**: Use `local-deploy.sh`

## Script Descriptions

### local-build.sh

Builds all StreamSpace Docker images locally:
- `streamspace/streamspace-controller:local`
- `streamspace/streamspace-api:local`
- `streamspace/streamspace-ui:local`

**Usage:**
```bash
./scripts/local-build.sh
```

**Options:**
- Set `VERSION=custom` to use a different tag

### local-deploy.sh

Deploys StreamSpace using Helm chart (for Helm v3.18.0 or earlier).

**Usage:**
```bash
./scripts/local-deploy.sh
```

**Environment Variables:**
- `NAMESPACE`: Kubernetes namespace (default: `streamspace`)
- `RELEASE_NAME`: Helm release name (default: `streamspace`)
- `VERSION`: Image tag to use (default: `local`)

**Requirements:**
- Helm v3.18.0 or earlier
- kubectl with cluster access
- Local Docker images built
- Kubernetes cluster (Docker Desktop, k3s, minikube, etc.)

### local-deploy-kubectl.sh

Deploys StreamSpace using raw Kubernetes manifests (Helm-free).

**Usage:**
```bash
./scripts/local-deploy-kubectl.sh
```

**Environment Variables:**
- `NAMESPACE`: Kubernetes namespace (default: `streamspace`)
- `VERSION`: Image tag to use (default: `local`)

**Requirements:**
- kubectl with cluster access
- Local Docker images built
- Kubernetes cluster (Docker Desktop, k3s, minikube, etc.)
- **Does NOT require Helm** - works with any Helm version

**Why This Exists:**
- Helm v3.19.0 (bundled with Docker Desktop) has a critical bug
- Provides a Helm-free alternative for users who can't downgrade
- Uses the same manifests, just applies them directly with kubectl

### local-teardown.sh

Removes StreamSpace deployment and cleans up resources.

**Usage:**
```bash
./scripts/local-teardown.sh
```

## Helm v3.19.0 Issue

**Problem:** Helm v3.19.0 has a critical regression in the chart loader that makes it completely unusable for loading charts from directories. All operations fail:
- `helm lint` → fails
- `helm template` → fails
- `helm package` → fails
- `helm install` → fails

**Affected Users:**
- Docker Desktop users on macOS/Windows (Helm is bundled)
- Anyone who upgraded to Helm v3.19.0

**Solutions:**
1. **Use `local-deploy-kubectl.sh`** (recommended) - bypasses Helm entirely
2. **Downgrade Helm** to v3.18.0 or earlier (if possible)

**For Details:**
See `docs/DEPLOYMENT_TROUBLESHOOTING.md` for comprehensive troubleshooting.

## Common Tasks

### Access the UI

```bash
kubectl port-forward -n streamspace svc/streamspace-ui 3000:80
```

Then open: http://localhost:3000

### Access the API

```bash
kubectl port-forward -n streamspace svc/streamspace-api 8000:8000
```

Then open: http://localhost:8000

### View Logs

```bash
# Controller
kubectl logs -n streamspace -l app.kubernetes.io/component=controller -f

# API
kubectl logs -n streamspace -l app.kubernetes.io/component=api -f

# UI
kubectl logs -n streamspace -l app.kubernetes.io/component=ui -f

# Database
kubectl logs -n streamspace -l app.kubernetes.io/component=database -f
```

### Check Deployment Status

```bash
kubectl get pods -n streamspace
kubectl get svc -n streamspace
kubectl get sessions -n streamspace
kubectl get templates -n streamspace
```

### Clean Up

Using Helm (if you deployed with `local-deploy.sh`):
```bash
./scripts/local-teardown.sh
```

Using kubectl (if you deployed with `local-deploy-kubectl.sh`):
```bash
kubectl delete namespace streamspace
```

## Troubleshooting

### Images Not Found

Build images first:
```bash
./scripts/local-build.sh
```

Verify they exist:
```bash
docker images | grep streamspace
```

### Helm Chart Errors

If using Helm v3.19.0:
```bash
# Switch to kubectl-based deployment
./scripts/local-deploy-kubectl.sh
```

If using older Helm:
```bash
# Check Helm version
helm version --short

# Use Helm deployment
./scripts/local-deploy.sh
```

### Pods Not Starting

Check pod status:
```bash
kubectl get pods -n streamspace
kubectl describe pod <pod-name> -n streamspace
kubectl logs <pod-name> -n streamspace
```

Common issues:
- ImagePullBackOff: Images not built or wrong pullPolicy
- CrashLoopBackOff: Check logs for errors
- Pending: Check resource availability

### Complete Troubleshooting Guide

See: `docs/DEPLOYMENT_TROUBLESHOOTING.md`

## Support

- Documentation: `docs/`
- Issues: https://github.com/streamspace/streamspace/issues
- Discussions: https://github.com/streamspace/streamspace/discussions
