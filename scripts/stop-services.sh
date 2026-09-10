#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PORT_FILE="$REPO_ROOT/.test-ports"
PID_FILE="$REPO_ROOT/.test-pids"

HTTPBIN_BASE_IMAGE="kennethreitz/httpbin"

# Source port and PID files if available
if [[ -f "$PORT_FILE" ]]; then
    source "$PORT_FILE"
fi
HTTPBIN_PORT="${HTTPBIN_PORT:-35123}"
API_TEST_SERVICE_PORT="${API_TEST_SERVICE_PORT:-35456}"

HTTPBIN_CONTAINER_ID=""
API_TEST_SERVICE_PID=""
if [[ -f "$PID_FILE" ]]; then
    source "$PID_FILE"
fi

echo -e "WARN\tStopping test services..."

# Kill api-test-service: prefer PID, fall back to port-based lsof
if [[ -n "$API_TEST_SERVICE_PID" ]] && kill -0 "$API_TEST_SERVICE_PID" 2>/dev/null; then
    echo -e "INFO\tKilling api-test-service (PID: $API_TEST_SERVICE_PID)"
    kill "$API_TEST_SERVICE_PID" 2>/dev/null || true
else
    lsof -ti:"$API_TEST_SERVICE_PORT" | xargs -r kill 2>/dev/null || true
fi

# Stop httpbin: prefer container ID, fall back to image-based stop
if [[ -n "$HTTPBIN_CONTAINER_ID" ]]; then
    echo -e "INFO\tStopping httpbin container ${HTTPBIN_CONTAINER_ID:0:12}"
    docker rm -f "$HTTPBIN_CONTAINER_ID" 2>/dev/null || true
else
    docker ps -q --filter ancestor="$HTTPBIN_BASE_IMAGE" | xargs -r docker stop
    docker ps -aq --filter ancestor="$HTTPBIN_BASE_IMAGE" | xargs -r docker rm
fi

# Optional: Prune Docker networks
docker network prune -f

# Wait a moment to allow ports to be released
sleep 2

# Clean up state files
rm -f "$PORT_FILE" "$PID_FILE"
