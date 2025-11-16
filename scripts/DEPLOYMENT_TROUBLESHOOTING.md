# Helm Deployment Troubleshooting Guide

## Issue: "Chart.yaml file is missing" Error

If you're encountering the error `Error: INSTALLATION FAILED: Chart.yaml file is missing` when running `./local-deploy.sh`, this guide provides solutions.

## Quick Solution: Use Alternative Deployment Script

We've created an alternative deployment script that packages the chart first, which avoids path resolution issues:

```bash
./scripts/local-deploy-alt.sh
```

This script:
- Packages the Helm chart into a `.tgz` file before installation
- Eliminates path resolution issues
- Follows Helm best practices
- Works consistently across all environments

## Diagnostic Approach: Enhanced Debugging

The main `local-deploy.sh` script now includes comprehensive debugging output:

```bash
./scripts/local-deploy.sh
```

This will show:
1. **Helm version** - Identifies version-specific issues
2. **Chart path** - Fully resolved absolute path
3. **Chart.yaml existence** - Verifies file is accessible
4. **Directory contents** - Lists all chart files
5. **Helm lint results** - Validates chart structure
6. **Chart packaging test** - Tests if Helm can package the chart
7. **Debug output** - Full Helm operation details with `--debug` flag

## Common Causes and Solutions

### 1. Path Resolution Issues

**Symptom:** Chart.yaml exists but Helm can't find it

**Solution:** Use absolute paths (already fixed in latest version)

```bash
# Now uses fully resolved path
CHART_PATH="$(cd "${PROJECT_ROOT}/chart" && pwd)"
```

### 2. Helm Version Compatibility

**Symptom:** Error varies by Helm version

**Solution:** Check your Helm version:

```bash
helm version
```

Recommended: Helm 3.8+ for best compatibility

### 3. Chart Structure Issues

**Symptom:** Helm lint fails

**Solution:** Verify chart structure:

```bash
chart/
├── Chart.yaml          # Chart metadata
├── values.yaml         # Default values
├── templates/          # Kubernetes templates
│   ├── _helpers.tpl   # Template helpers
│   └── *.yaml         # Resource templates
└── crds/              # Custom Resource Definitions
    └── *.yaml
```

### 4. Permission Issues

**Symptom:** Can't read chart files

**Solution:** Check file permissions:

```bash
ls -la chart/
ls -la chart/Chart.yaml
```

All files should be readable (r-- permission).

### 5. Working Directory Issues

**Symptom:** Error when running from different directories

**Solution:** Both scripts now handle this correctly. Run from any directory:

```bash
# From project root
./scripts/local-deploy.sh

# From scripts directory
./local-deploy.sh
```

## Makefile Fixes

If using `make helm-install`, the Makefile now uses absolute paths:

```bash
make helm-install
```

The Makefile will display the resolved chart path for debugging.

## Manual Installation Steps

If both scripts fail, try manual installation:

```bash
# 1. Navigate to project root
cd /path/to/streamspace

# 2. Create namespace
kubectl create namespace streamspace

# 3. Apply CRDs manually
kubectl apply -f chart/crds/

# 4. Package chart
helm package chart/ -d /tmp

# 5. Install from package
helm install streamspace /tmp/streamspace-*.tgz \
  --namespace streamspace \
  --set controller.image.tag=local \
  --set controller.image.pullPolicy=Never \
  --set api.image.tag=local \
  --set api.image.pullPolicy=Never \
  --set ui.image.tag=local \
  --set ui.image.pullPolicy=Never \
  --wait
```

## Verification Steps

After deployment, verify the installation:

```bash
# Check Helm release status
helm status streamspace -n streamspace

# Check pods
kubectl get pods -n streamspace

# Check CRDs
kubectl get crds | grep streamspace

# View Helm release values
helm get values streamspace -n streamspace
```

## Debugging Commands

### Check Helm can access the chart

```bash
helm lint chart/
helm template streamspace chart/ > /tmp/rendered-templates.yaml
helm package chart/ -d /tmp
```

### Check kubectl connectivity

```bash
kubectl cluster-info
kubectl get nodes
kubectl get namespaces
```

### View Helm debug output

```bash
helm install streamspace chart/ \
  --namespace streamspace \
  --dry-run \
  --debug
```

## Getting Help

If you're still experiencing issues:

1. **Capture full output:**
   ```bash
   ./scripts/local-deploy.sh 2>&1 | tee deployment-debug.log
   ```

2. **Check the logs** in `deployment-debug.log` for:
   - Helm version
   - Chart path resolution
   - Lint results
   - Error messages

3. **Try alternative script:**
   ```bash
   ./scripts/local-deploy-alt.sh
   ```

4. **Report issue** with:
   - Helm version (`helm version`)
   - Kubernetes version (`kubectl version`)
   - OS and architecture
   - Full debug output
   - Contents of chart/ directory (`ls -laR chart/`)

## Related Changes

- ✅ Fixed `.helmignore` to remove confusing `!Chart.yaml` line
- ✅ Updated Makefile to use `CHART_PATH` variable with absolute paths
- ✅ Enhanced `local-deploy.sh` with comprehensive debugging
- ✅ Created `local-deploy-alt.sh` as alternative package-based approach

## References

- [Helm Charts Documentation](https://helm.sh/docs/topics/charts/)
- [Helm Install Command](https://helm.sh/docs/helm/helm_install/)
- [Debugging Helm Charts](https://helm.sh/docs/chart_template_guide/debugging/)
