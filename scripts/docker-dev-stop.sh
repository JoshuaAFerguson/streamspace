#!/usr/bin/env bash
#
# docker-dev-stop.sh - Stop StreamSpace development environment
#
# This script stops and optionally removes the Docker Compose development environment.
#
# Usage:
#   ./scripts/docker-dev-stop.sh           # Stop services
#   ./scripts/docker-dev-stop.sh --clean   # Stop and remove volumes
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
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yml"

# Options
REMOVE_VOLUMES=false

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

Stop StreamSpace development environment.

Options:
    --clean         Remove volumes (database data will be lost)
    --remove-all    Remove everything including images
    -h, --help      Show this help message

Examples:
    $(basename "$0")               # Stop services, keep data
    $(basename "$0") --clean       # Stop and remove volumes

EOF
    exit 0
}

# Parse arguments
REMOVE_IMAGES=false
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --clean)
                REMOVE_VOLUMES=true
                shift
                ;;
            --remove-all)
                REMOVE_VOLUMES=true
                REMOVE_IMAGES=true
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

# Determine docker compose command
get_compose_cmd() {
    if docker compose version &> /dev/null 2>&1; then
        echo "docker compose"
    else
        echo "docker-compose"
    fi
}

# Stop services
stop_services() {
    local compose_cmd
    compose_cmd=$(get_compose_cmd)

    log "Stopping development environment..."

    cd "$PROJECT_ROOT"

    if [ "$REMOVE_VOLUMES" = true ]; then
        log_warning "Removing volumes (data will be lost)..."
        $compose_cmd -f "$COMPOSE_FILE" --profile docker --profile dev --profile monitoring down -v
    else
        $compose_cmd -f "$COMPOSE_FILE" --profile docker --profile dev --profile monitoring down
    fi

    log_success "Services stopped"
}

# Remove images
remove_images() {
    local compose_cmd
    compose_cmd=$(get_compose_cmd)

    if [ "$REMOVE_IMAGES" = true ]; then
        log "Removing images..."
        cd "$PROJECT_ROOT"
        $compose_cmd -f "$COMPOSE_FILE" --profile docker --profile dev --profile monitoring down --rmi local
        log_success "Images removed"
    fi
}

# Main execution
main() {
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  Stop StreamSpace Development Environment${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""

    parse_args "$@"
    stop_services
    remove_images

    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    log_success "Development environment stopped"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""

    if [ "$REMOVE_VOLUMES" = true ]; then
        log_info "Volumes removed. Database data has been cleared."
    else
        log_info "Volumes preserved. Restart with: ./scripts/docker-dev.sh"
    fi
    echo ""
}

# Run main function
main "$@"
