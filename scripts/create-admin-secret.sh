#!/usr/bin/env bash
#
# create-admin-secret.sh - Create Kubernetes secret for admin credentials
#
# This script creates the streamspace-admin-credentials secret with a default
# admin password for initial setup. The password can be changed after deployment.
#

set -euo pipefail

# Colors for output
COLOR_RESET='\033[0m'
COLOR_BOLD='\033[1m'
COLOR_GREEN='\033[32m'
COLOR_YELLOW='\033[33m'
COLOR_BLUE='\033[34m'
COLOR_RED='\033[31m'

# Configuration
NAMESPACE="${NAMESPACE:-streamspace}"
SECRET_NAME="streamspace-admin-credentials"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Password12345}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@streamspace.local}"

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

# Check prerequisites
check_prerequisites() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
}

# Create admin credentials secret
create_admin_secret() {
    log "Creating admin credentials secret..."

    # Check if namespace exists
    if ! kubectl get namespace "${NAMESPACE}" &> /dev/null; then
        log_warning "Namespace ${NAMESPACE} does not exist, creating..."
        kubectl create namespace "${NAMESPACE}"
    fi

    # Check if secret already exists
    if kubectl get secret "${SECRET_NAME}" -n "${NAMESPACE}" &> /dev/null; then
        log_warning "Secret ${SECRET_NAME} already exists in namespace ${NAMESPACE}"
        log_info "To recreate, delete it first:"
        log_info "  kubectl delete secret ${SECRET_NAME} -n ${NAMESPACE}"
        return 0
    fi

    # Create the secret
    kubectl create secret generic "${SECRET_NAME}" \
        -n "${NAMESPACE}" \
        --from-literal=username="${ADMIN_USERNAME}" \
        --from-literal=password="${ADMIN_PASSWORD}" \
        --from-literal=email="${ADMIN_EMAIL}"

    # Add labels to match the Helm chart pattern
    kubectl label secret "${SECRET_NAME}" \
        -n "${NAMESPACE}" \
        app.kubernetes.io/name=streamspace \
        app.kubernetes.io/component=admin \
        app.kubernetes.io/managed-by=kubectl

    log_success "Admin credentials secret created successfully"
    log_info "Secret name: ${SECRET_NAME}"
    log_info "Namespace: ${NAMESPACE}"
    log_info "Username: ${ADMIN_USERNAME}"
    log_info "Email: ${ADMIN_EMAIL}"
    log_warning "Default password is set. Please change it after first login!"
}

# Show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Create Kubernetes secret for StreamSpace admin credentials."
    echo ""
    echo "Options:"
    echo "  -h, --help       Show this help message"
    echo "  -n, --namespace  Kubernetes namespace (default: streamspace)"
    echo "  -p, --password   Admin password (default: Password12345)"
    echo "  -e, --email      Admin email (default: admin@streamspace.local)"
    echo ""
    echo "Environment Variables:"
    echo "  NAMESPACE        Kubernetes namespace"
    echo "  ADMIN_PASSWORD   Admin password"
    echo "  ADMIN_EMAIL      Admin email"
    echo ""
    echo "Examples:"
    echo "  $0                           # Use defaults"
    echo "  $0 -n myspace -p MySecret    # Custom namespace and password"
    echo "  ADMIN_PASSWORD=secret $0     # Use environment variable"
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -p|--password)
                ADMIN_PASSWORD="$2"
                shift 2
                ;;
            -e|--email)
                ADMIN_EMAIL="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Main execution
main() {
    parse_args "$@"

    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  StreamSpace Admin Credentials Setup${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""

    check_prerequisites
    create_admin_secret

    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    log_success "Admin credentials secret setup complete!"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
}

main "$@"
