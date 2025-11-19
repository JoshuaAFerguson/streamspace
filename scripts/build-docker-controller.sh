#!/usr/bin/env bash
#
# build-docker-controller.sh - Build the StreamSpace Docker platform controller
#
# This script builds the Docker controller which handles session management
# on Docker platforms via NATS events.
#
# Usage:
#   ./scripts/build-docker-controller.sh           # Build Docker image
#   ./scripts/build-docker-controller.sh --binary  # Build binary only
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
CONTROLLER_DIR="${PROJECT_ROOT}/docker-controller"
VERSION="${VERSION:-local}"
GIT_COMMIT="${GIT_COMMIT:-$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
BUILD_DATE="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

# Image name
DOCKER_CONTROLLER_IMAGE="streamspace/docker-controller"

# Build mode
BUILD_BINARY_ONLY=false

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

# Show usage
usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS]

Build the StreamSpace Docker platform controller.

Options:
    --binary        Build Go binary only (no Docker image)
    --push          Push image to registry after building
    -h, --help      Show this help message

Environment Variables:
    VERSION         Image tag (default: local)
    REGISTRY        Docker registry prefix (default: none)

Examples:
    $(basename "$0")                    # Build Docker image
    $(basename "$0") --binary           # Build binary only
    VERSION=v1.0.0 $(basename "$0")     # Build with specific version

EOF
    exit 0
}

# Parse arguments
PUSH_IMAGE=false
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --binary)
                BUILD_BINARY_ONLY=true
                shift
                ;;
            --push)
                PUSH_IMAGE=true
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                ;;
        esac
    done
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if [ ! -d "$CONTROLLER_DIR" ]; then
        log_error "Docker controller directory not found: $CONTROLLER_DIR"
        exit 1
    fi

    if [ "$BUILD_BINARY_ONLY" = true ]; then
        if ! command -v go &> /dev/null; then
            log_error "Go is not installed or not in PATH"
            exit 1
        fi
        log_success "Go is available: $(go version)"
    else
        if ! command -v docker &> /dev/null; then
            log_error "Docker is not installed or not in PATH"
            exit 1
        fi

        if ! docker info &> /dev/null; then
            log_error "Docker daemon is not running"
            exit 1
        fi
        log_success "Docker is available"
    fi
}

# Build binary
build_binary() {
    log "Building Docker controller binary..."
    log_info "Version: $VERSION"
    log_info "Commit: $GIT_COMMIT"

    cd "$CONTROLLER_DIR"

    # Download dependencies
    log_info "Downloading dependencies..."
    go mod download

    # Build binary
    log_info "Compiling..."
    CGO_ENABLED=0 go build \
        -ldflags "-X main.version=${VERSION} -X main.commit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}" \
        -o bin/docker-controller \
        ./cmd/main.go

    log_success "Binary built: $CONTROLLER_DIR/bin/docker-controller"
}

# Build Docker image
build_image() {
    log "Building Docker controller image..."
    log_info "Image: ${DOCKER_CONTROLLER_IMAGE}:${VERSION}"
    log_info "Context: $CONTROLLER_DIR"

    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg COMMIT="${GIT_COMMIT}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        -t "${DOCKER_CONTROLLER_IMAGE}:${VERSION}" \
        -t "${DOCKER_CONTROLLER_IMAGE}:latest" \
        -f "${CONTROLLER_DIR}/Dockerfile" \
        "${CONTROLLER_DIR}/"

    log_success "Docker image built successfully"

    # Show image info
    echo ""
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}" | \
        grep -E "REPOSITORY|${DOCKER_CONTROLLER_IMAGE}" || true
}

# Push image
push_image() {
    if [ "$PUSH_IMAGE" = true ]; then
        log "Pushing image to registry..."
        docker push "${DOCKER_CONTROLLER_IMAGE}:${VERSION}"
        docker push "${DOCKER_CONTROLLER_IMAGE}:latest"
        log_success "Image pushed"
    fi
}

# Main execution
main() {
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  Build StreamSpace Docker Controller${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
    echo -e "${COLOR_BLUE}Version:${COLOR_RESET}    ${VERSION}"
    echo -e "${COLOR_BLUE}Commit:${COLOR_RESET}     ${GIT_COMMIT}"
    echo -e "${COLOR_BLUE}Build Date:${COLOR_RESET} ${BUILD_DATE}"
    echo ""

    parse_args "$@"
    check_prerequisites

    if [ "$BUILD_BINARY_ONLY" = true ]; then
        build_binary
    else
        build_image
        push_image
    fi

    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    log_success "Build completed successfully!"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""

    if [ "$BUILD_BINARY_ONLY" = true ]; then
        log_info "Run the binary:"
        echo "  $CONTROLLER_DIR/bin/docker-controller --nats-url=nats://localhost:4222"
    else
        log_info "Run with docker-compose:"
        echo "  ./scripts/docker-dev.sh --with-docker"
        echo ""
        log_info "Or run standalone:"
        echo "  docker run -d \\"
        echo "    -e NATS_URL=nats://host.docker.internal:4222 \\"
        echo "    -v /var/run/docker.sock:/var/run/docker.sock:ro \\"
        echo "    ${DOCKER_CONTROLLER_IMAGE}:${VERSION}"
    fi
    echo ""
}

# Run main function
main "$@"
