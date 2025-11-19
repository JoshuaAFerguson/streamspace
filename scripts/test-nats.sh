#!/usr/bin/env bash
#
# test-nats.sh - Test NATS connectivity and event publishing
#
# This script tests NATS server connectivity and can publish test events
# to verify the event-driven architecture is working correctly.
#
# Usage:
#   ./scripts/test-nats.sh                    # Test connectivity
#   ./scripts/test-nats.sh --publish          # Publish test events
#   ./scripts/test-nats.sh --subscribe        # Subscribe to all events
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
NATS_URL="${NATS_URL:-nats://localhost:4222}"
NATS_MONITOR_URL="${NATS_MONITOR_URL:-http://localhost:8222}"

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

Test NATS connectivity and event publishing for StreamSpace.

Options:
    --status        Show NATS server status (default)
    --publish       Publish test events
    --subscribe     Subscribe to all StreamSpace events
    --streams       List JetStream streams
    --consumers     List JetStream consumers
    -h, --help      Show this help message

Environment Variables:
    NATS_URL          NATS server URL (default: nats://localhost:4222)
    NATS_MONITOR_URL  NATS monitoring URL (default: http://localhost:8222)

Examples:
    $(basename "$0")                    # Test connectivity
    $(basename "$0") --publish          # Publish test events
    $(basename "$0") --streams          # Show JetStream streams

EOF
    exit 0
}

# Check if NATS CLI is installed
check_nats_cli() {
    if command -v nats &> /dev/null; then
        return 0
    fi
    return 1
}

# Test basic connectivity via HTTP monitor
test_connectivity() {
    log "Testing NATS connectivity..."
    log_info "Monitor URL: $NATS_MONITOR_URL"

    # Check if NATS monitor is accessible
    if curl -s -o /dev/null -w "%{http_code}" "$NATS_MONITOR_URL/healthz" | grep -q "200"; then
        log_success "NATS server is healthy"
    else
        log_error "Cannot connect to NATS monitor at $NATS_MONITOR_URL"
        log_info "Make sure NATS is running: ./scripts/docker-dev.sh"
        return 1
    fi

    # Get server info
    echo ""
    log "NATS Server Information:"
    if command -v jq &> /dev/null; then
        curl -s "$NATS_MONITOR_URL/varz" | jq '{
            server_id: .server_id,
            version: .version,
            go: .go,
            host: .host,
            port: .port,
            max_connections: .max_connections,
            connections: .connections,
            in_msgs: .in_msgs,
            out_msgs: .out_msgs,
            in_bytes: .in_bytes,
            out_bytes: .out_bytes
        }'
    else
        curl -s "$NATS_MONITOR_URL/varz" | head -20
        log_info "Install jq for formatted output: brew install jq"
    fi

    return 0
}

# Show JetStream info
show_jetstream_info() {
    log "JetStream Information:"

    if ! curl -s -o /dev/null -w "%{http_code}" "$NATS_MONITOR_URL/jsz" | grep -q "200"; then
        log_error "JetStream is not available"
        return 1
    fi

    if command -v jq &> /dev/null; then
        curl -s "$NATS_MONITOR_URL/jsz" | jq '{
            memory: .memory,
            storage: .storage,
            streams: .streams,
            consumers: .consumers,
            messages: .messages,
            bytes: .bytes
        }'
    else
        curl -s "$NATS_MONITOR_URL/jsz"
    fi

    return 0
}

# List streams
list_streams() {
    log "JetStream Streams:"

    if check_nats_cli; then
        nats -s "$NATS_URL" stream list
    else
        # Use HTTP API
        if command -v jq &> /dev/null; then
            curl -s "$NATS_MONITOR_URL/jsz?streams=true" | jq '.account_details[].stream_detail[] | {name: .name, messages: .state.messages, bytes: .state.bytes, consumers: .state.consumer_count}'
        else
            curl -s "$NATS_MONITOR_URL/jsz?streams=true"
        fi
    fi
}

# List consumers
list_consumers() {
    log "JetStream Consumers:"

    if check_nats_cli; then
        nats -s "$NATS_URL" consumer list --all
    else
        log_warning "Install NATS CLI for consumer listing: brew install nats-io/nats-tools/nats"
        curl -s "$NATS_MONITOR_URL/jsz?consumers=true"
    fi
}

# Publish test events
publish_test_events() {
    log "Publishing test events..."

    if ! check_nats_cli; then
        log_error "NATS CLI is required for publishing"
        log_info "Install: brew install nats-io/nats-tools/nats"
        log_info "Or: go install github.com/nats-io/natscli/nats@latest"
        return 1
    fi

    # Test event payload
    local event_id
    event_id=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "test-$(date +%s)")
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Publish session status event
    local session_event
    session_event=$(cat << EOF
{
    "event_id": "${event_id}",
    "timestamp": "${timestamp}",
    "session_id": "test-session-001",
    "status": "running",
    "phase": "Running",
    "url": "http://localhost:3000",
    "pod_name": "test-pod",
    "message": "Test session status event",
    "controller_id": "test-controller"
}
EOF
)

    log_info "Publishing to streamspace.session.status..."
    echo "$session_event" | nats -s "$NATS_URL" publish streamspace.session.status

    # Publish app status event
    local app_event
    app_event=$(cat << EOF
{
    "event_id": "${event_id}-app",
    "timestamp": "${timestamp}",
    "install_id": "test-install-001",
    "status": "ready",
    "template_name": "test-template",
    "message": "Test app status event",
    "controller_id": "test-controller"
}
EOF
)

    log_info "Publishing to streamspace.app.status..."
    echo "$app_event" | nats -s "$NATS_URL" publish streamspace.app.status

    log_success "Test events published"
    echo ""
    log_info "Events should be received by the API subscriber"
}

# Subscribe to events
subscribe_to_events() {
    log "Subscribing to all StreamSpace events..."
    log_info "Press Ctrl+C to stop"
    echo ""

    if ! check_nats_cli; then
        log_error "NATS CLI is required for subscribing"
        log_info "Install: brew install nats-io/nats-tools/nats"
        return 1
    fi

    nats -s "$NATS_URL" subscribe "streamspace.>"
}

# Parse arguments
MODE="status"
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --status)
                MODE="status"
                shift
                ;;
            --publish)
                MODE="publish"
                shift
                ;;
            --subscribe)
                MODE="subscribe"
                shift
                ;;
            --streams)
                MODE="streams"
                shift
                ;;
            --consumers)
                MODE="consumers"
                shift
                ;;
            --jetstream)
                MODE="jetstream"
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

# Main execution
main() {
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}  StreamSpace NATS Test Utility${COLOR_RESET}"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
    echo -e "${COLOR_BLUE}NATS URL:${COLOR_RESET}     $NATS_URL"
    echo -e "${COLOR_BLUE}Monitor URL:${COLOR_RESET}  $NATS_MONITOR_URL"
    echo ""

    parse_args "$@"

    case $MODE in
        status)
            test_connectivity
            echo ""
            show_jetstream_info
            ;;
        publish)
            test_connectivity || exit 1
            echo ""
            publish_test_events
            ;;
        subscribe)
            test_connectivity || exit 1
            echo ""
            subscribe_to_events
            ;;
        streams)
            test_connectivity || exit 1
            echo ""
            list_streams
            ;;
        consumers)
            test_connectivity || exit 1
            echo ""
            list_consumers
            ;;
        jetstream)
            test_connectivity || exit 1
            echo ""
            show_jetstream_info
            ;;
    esac

    echo ""
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    log_success "Test completed"
    echo -e "${COLOR_BOLD}═══════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
}

# Run main function
main "$@"
