#!/usr/bin/env bash
#
# local-deploy-alt.sh - Alternative Helm deployment approach
#
# This script packages the chart first, then installs from the package.
# This avoids potential path resolution issues with chart directories.
#

set -euo pipefail

# Colors for output
COLOR_RESET='\033[0m'
COLOR_BOLD='\033[1m'
COLOR_GREEN='\033[32m'
COLOR_YELLOW='\033[33m'
COLOR_BLUE='\033[34m'
COLOR_RED='\033[31m'

# Project configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="${NAMESPACE:-streamspace}"
RELEASE_NAME="${RELEASE_NAME:-streamspace}"
VERSION="${VERSION:-local}"

# Helm chart location
CHART_DIR="${PROJECT_ROOT}/chart"
PACKAGE_DIR="${PROJECT_ROOT}/.helm-packages"

# Helper functions
log() {
    echo -e "${COLOR_BOLD}==>${COLOR_RESET} $*"
}

log_success() {
    echo -e "${COLOR_GREEN}✓${COLOR_RESET} $*"
}

log_error() {
    echo -e "${COLOR_RED}✗${COLOR_RESET} $*" >&2
}

log_info() {
    echo -e "${COLOR_BLUE}→${COLOR_RESET} $*"
}

log_warning() {
    echo -e "${COLOR_YELLOW}⚠${COLOR_RESET} $*"
}

# Main execution
main() {
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  StreamSpace Alternative Deployment${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
    echo -e "${COLOR_BLUE}Strategy:${COLOR_RESET}      Package-first installation"
    echo -e "${COLOR_BLUE}Namespace:${COLOR_RESET}     ${NAMESPACE}"
    echo -e "${COLOR_BLUE}Release:${COLOR_RESET}       ${RELEASE_NAME}"
    echo -e "${COLOR_BLUE}Version:${COLOR_RESET}       ${VERSION}"
    echo ""

    # Create package directory
    log "Creating package directory..."
    mkdir -p "${PACKAGE_DIR}"
    log_success "Package directory ready: ${PACKAGE_DIR}"

    # Package the chart
    log "Packaging Helm chart..."
    log_info "Chart directory: ${CHART_DIR}"

    cd "${PROJECT_ROOT}"
    local package_file=$(helm package "${CHART_DIR}" -d "${PACKAGE_DIR}" --version "${VERSION}" | grep -oE '[^ ]+\.tgz$')

    if [ -z "${package_file}" ] || [ ! -f "${package_file}" ]; then
        log_error "Failed to package chart"
        exit 1
    fi

    log_success "Chart packaged: ${package_file}"

    # Create namespace
    log "Creating namespace: ${NAMESPACE}"
    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
    log_success "Namespace ready"

    # Apply CRDs manually (Helm doesn't upgrade CRDs automatically)
    log "Applying CRDs..."
    kubectl apply -f "${CHART_DIR}/crds/"
    log_success "CRDs applied"

    # Install or upgrade from package
    log "Deploying StreamSpace from package..."

    if helm status "${RELEASE_NAME}" -n "${NAMESPACE}" &> /dev/null; then
        log_info "Release exists, upgrading..."
        helm upgrade "${RELEASE_NAME}" "${package_file}" \
            --namespace "${NAMESPACE}" \
            --set controller.image.tag="${VERSION}" \
            --set controller.image.pullPolicy=Never \
            --set api.image.tag="${VERSION}" \
            --set api.image.pullPolicy=Never \
            --set ui.image.tag="${VERSION}" \
            --set ui.image.pullPolicy=Never \
            --set postgresql.enabled=true \
            --set postgresql.auth.password=streamspace \
            --wait \
            --timeout 5m
    else
        log_info "Installing fresh release..."
        helm install "${RELEASE_NAME}" "${package_file}" \
            --namespace "${NAMESPACE}" \
            --create-namespace \
            --set controller.image.tag="${VERSION}" \
            --set controller.image.pullPolicy=Never \
            --set api.image.tag="${VERSION}" \
            --set api.image.pullPolicy=Never \
            --set ui.image.tag="${VERSION}" \
            --set ui.image.pullPolicy=Never \
            --set postgresql.enabled=true \
            --set postgresql.auth.password=streamspace \
            --wait \
            --timeout 5m
    fi

    log_success "Helm deployment complete"

    # Wait for pods
    log "Waiting for pods to be ready..."
    sleep 5
    kubectl wait --for=condition=ready pod \
        -l app.kubernetes.io/name=streamspace \
        -n "${NAMESPACE}" \
        --timeout=300s 2>/dev/null || log_warning "Some pods may still be starting"

    # Show status
    echo ""
    log "Deployment Status:"
    echo ""
    log_info "Pods:"
    kubectl get pods -n "${NAMESPACE}"
    echo ""
    log_info "Services:"
    kubectl get svc -n "${NAMESPACE}"
    echo ""

    # Access instructions
    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  Access Instructions${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
    log_info "Port-forward UI:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-ui 3000:80"
    echo "  Then access: http://localhost:3000"
    echo ""
    log_info "Port-forward API:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/${RELEASE_NAME}-api 8000:8000"
    echo ""
    log_info "View logs:"
    echo "  kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/component=controller -f"
    echo ""
    log_info "Cleanup:"
    echo "  ./scripts/local-teardown.sh"
    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    log_success "Deployment complete!"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
}

# Run main function
main "$@"
