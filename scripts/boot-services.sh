#!/usr/bin/env bash
#
# boot-services.sh — Start test services (httpbin + api-test-service).
#
# PORTABILITY: This script must run on both macOS (BSD userland) and Linux
# (GNU userland).  Key differences to keep in mind:
#   - grep: use -E (POSIX ERE) not -P (PCRE, GNU-only; not available on macOS).
#     For extracting a capture group, pipe through cut/sed/awk instead of \K.
#   - sed: use sed 's/re/rep/' without GNU extensions (-r → use -E if needed,
#     but prefer awk for anything non-trivial).
#   - date: GNU date -d is not available on macOS; use python3 if date math
#     beyond basic formatting is needed.
#   When in doubt, test on both platforms before merging.
#
# Port strategy:
#   1. If HTTPBIN_PORT / API_TEST_SERVICE_PORT are set, use those exact ports.
#   2. Otherwise, try the well-known defaults (35123 / 35456) first — these
#      match the server URLs in the OpenAPI specs.
#   3. If the default port is already occupied, fall back to an OS-assigned
#      ephemeral port.  All tests read the port from environment variables
#      (via CommonHelpers) so they work regardless of which port is assigned.
#
# The actual ports are written to .test-ports (sourced by test-target.sh
# and the Makefile) and PIDs/container IDs to .test-pids for cleanup.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PORT_FILE="$REPO_ROOT/.test-ports"
PID_FILE="$REPO_ROOT/.test-pids"
tmp_log=$(mktemp)
trap 'rm -f "$tmp_log"' EXIT

set +e

HTTPBIN_BASE_IMAGE="kennethreitz/httpbin"

# Well-known defaults matching the OpenAPI spec server URLs
DEFAULT_HTTPBIN_PORT=35123
DEFAULT_API_TEST_SERVICE_PORT=35456

# Track IDs for cleanup
HTTPBIN_CONTAINER_ID=""
API_TEST_SERVICE_PID=""

# ── helpers ────────────────────────────────────────────────────────

# Check if a port is available (nothing listening on it)
port_available() {
	local port="$1"
	! curl -s -f -o /dev/null --max-time 1 "http://localhost:$port/" 2>/dev/null
}

# Start a docker container and capture only the container ID (stderr → tempfile)
docker_run_detach() {
	local docker_stderr
	docker_stderr=$(mktemp)
	local cid
	cid=$(docker run --detach "$@" 2>"$docker_stderr")
	local rc=$?
	if [[ $rc -ne 0 ]]; then
		echo -e "ERROR\tdocker run failed (exit $rc):"
		cat "$docker_stderr"
		rm -f "$docker_stderr"
		return 1
	fi
	rm -f "$docker_stderr"
	echo "$cid"
}

# Discover the host port docker assigned to container_port
docker_discover_port() {
	local cid="$1" container_port="$2"
	local mapping="" port=""
	for attempt in $(seq 1 10); do
		mapping=$(docker port "$cid" "$container_port" 2>/dev/null | head -1)
		port="${mapping##*:}"
		if [[ -n "$port" && "$port" != "0" ]]; then
			echo "$port"
			return 0
		fi
		if ! docker ps -q --filter "id=$cid" | grep -q .; then
			echo -e "ERROR\tContainer exited before port was assigned" >&2
			docker logs "$cid" 2>&1 | tail -20 >&2
			docker rm -f "$cid" 2>/dev/null || true
			return 1
		fi
		sleep 1
	done
	echo -e "ERROR\tCould not determine port after 10 attempts" >&2
	echo -e "ERROR\tLast docker port output: '$mapping'" >&2
	echo -e "ERROR\tContainer status: $(docker inspect --format='{{.State.Status}}' "$cid" 2>/dev/null)" >&2
	docker logs "$cid" 2>&1 | tail -20 >&2
	docker rm -f "$cid" 2>/dev/null || true
	return 1
}

# ── httpbin ──────────────────────────────────────────────────────────

start_httpbin() {
	# Determine port: explicit env var → default → ephemeral fallback
	local requested_port="${HTTPBIN_PORT:-$DEFAULT_HTTPBIN_PORT}"

	# Check if already healthy on that port
	if curl -s -f -o /dev/null --max-time 2 "http://localhost:$requested_port/anything" 2>/dev/null; then
		HTTPBIN_PORT="$requested_port"
		echo -e "INFO\thttpbin already healthy on port $HTTPBIN_PORT"
		return 0
	fi

	# Try the requested port
	echo -e "INFO\tStarting httpbin on port $requested_port..."
	local cid
	cid=$(docker_run_detach -p "$requested_port":80 "$HTTPBIN_BASE_IMAGE")
	if [[ $? -ne 0 ]]; then
		# Port likely occupied — fall back to ephemeral
		echo -e "WARN\tPort $requested_port unavailable, falling back to ephemeral port..."
		cid=$(docker_run_detach -p 0:80 "$HTTPBIN_BASE_IMAGE")
		if [[ $? -ne 0 ]]; then
			echo -e "ERROR\tFailed to start httpbin container"
			return 1
		fi
		HTTPBIN_CONTAINER_ID="$cid"

		HTTPBIN_PORT=$(docker_discover_port "$cid" 80)
		if [[ $? -ne 0 || -z "$HTTPBIN_PORT" ]]; then
			echo -e "ERROR\tFailed to discover ephemeral httpbin port"
			return 1
		fi
	else
		HTTPBIN_CONTAINER_ID="$cid"
		HTTPBIN_PORT="$requested_port"
	fi

	# Wait for healthy (up to 15s)
	for i in $(seq 1 15); do
		sleep 1
		if curl -s -f -o /dev/null --max-time 2 "http://localhost:$HTTPBIN_PORT/anything" 2>/dev/null; then
			echo -e "INFO\thttpbin started on port $HTTPBIN_PORT (container: ${HTTPBIN_CONTAINER_ID:0:12})"
			return 0
		fi
		if ! docker ps -q --filter "id=$HTTPBIN_CONTAINER_ID" | grep -q .; then
			echo -e "ERROR\thttpbin container exited unexpectedly"
			return 1
		fi
	done

	echo -e "ERROR\thttpbin not healthy on port $HTTPBIN_PORT after 15s"
	docker rm -f "$HTTPBIN_CONTAINER_ID" 2>/dev/null || true
	return 1
}

# ── speakeasy-api-test-service ───────────────────────────────────────

start_api_test_service() {
	local binary="$REPO_ROOT/bin/api-test-service"

	if [[ ! -x "$binary" ]]; then
		echo -e "ERROR\tapi-test-service binary not found at $binary"
		echo -e "ERROR\tRun 'make build-api-test-service' first"
		return 1
	fi

	# Determine port: explicit env var → default → ephemeral fallback (0)
	local bind_port="${API_TEST_SERVICE_PORT:-$DEFAULT_API_TEST_SERVICE_PORT}"

	# Check if already healthy on that port
	if curl -s -f -o /dev/null --max-time 2 "http://localhost:$bind_port/ping" 2>/dev/null; then
		API_TEST_SERVICE_PORT="$bind_port"
		echo -e "INFO\tspeakeasy-api-test-service already healthy on port $bind_port"
		return 0
	fi

	echo -e "INFO\tStarting speakeasy-api-test-service (bind=:$bind_port)..."
	HTTPBIN_PORT=$HTTPBIN_PORT "$binary" -b "$bind_port" >"$tmp_log" 2>&1 &
	local pid=$!
	API_TEST_SERVICE_PID=$pid

	# Wait for LISTEN_PORT=<N> in output (up to 15s — binary is pre-built)
	local actual_port=""
	for i in $(seq 1 15); do
		sleep 1
		if ! kill -0 "$pid" 2>/dev/null; then
			# If binding the default port failed, retry with ephemeral
			if [[ "$bind_port" != "0" ]]; then
				echo -e "WARN\tPort $bind_port unavailable, falling back to ephemeral port..."
				>"$tmp_log"
				HTTPBIN_PORT=$HTTPBIN_PORT "$binary" -b 0 >"$tmp_log" 2>&1 &
				pid=$!
				API_TEST_SERVICE_PID=$pid
				bind_port=0
				continue
			fi
			echo -e "ERROR\tapi-test-service process died. Output:"
			cat "$tmp_log"
			return 1
		fi
		# grep -Eo + cut: POSIX-compatible on both macOS (BSD grep) and Linux.
		# Do NOT use grep -oP '\K' — -P (PCRE) is GNU-only and silently
		# produces no output on macOS, causing the 15-second timeout to fire.
		actual_port=$(grep -Eo 'LISTEN_PORT=[0-9]+' "$tmp_log" 2>/dev/null | head -1 | cut -d= -f2 || true)
		if [[ -n "$actual_port" ]]; then
			break
		fi
	done

	if [[ -z "$actual_port" ]]; then
		echo -e "ERROR\tCould not determine api-test-service port after 15s. Output:"
		tail -n 10 "$tmp_log"
		kill "$pid" 2>/dev/null || true
		return 1
	fi

	API_TEST_SERVICE_PORT="$actual_port"

	# Verify health
	for i in $(seq 1 10); do
		if curl -s -f -o /dev/null --max-time 2 "http://localhost:$actual_port/ping" 2>/dev/null; then
			echo -e "INFO\tspeakeasy-api-test-service started on port $actual_port (PID: $pid)"
			return 0
		fi
		sleep 1
	done

	echo -e "ERROR\tspeakeasy-api-test-service not healthy on port $actual_port after 10s"
	kill "$pid" 2>/dev/null || true
	return 1
}

# ── main ─────────────────────────────────────────────────────────────

echo -e "INFO\tStarting test services..."

if ! start_httpbin; then
	echo -e "ERROR\tFailed to start httpbin."
	[[ -n "$HTTPBIN_CONTAINER_ID" ]] && docker rm -f "$HTTPBIN_CONTAINER_ID" 2>/dev/null || true
	exit 1
fi

if ! start_api_test_service; then
	echo -e "ERROR\tFailed to start speakeasy-api-test-service."
	exit 1
fi

cat >"$PORT_FILE" <<EOF
export HTTPBIN_PORT=$HTTPBIN_PORT
export API_TEST_SERVICE_PORT=$API_TEST_SERVICE_PORT
export HTTPBIN_URL=http://localhost:$HTTPBIN_PORT
export API_TEST_SERVICE_URL=http://localhost:$API_TEST_SERVICE_PORT
EOF

cat >"$PID_FILE" <<EOF
HTTPBIN_CONTAINER_ID=$HTTPBIN_CONTAINER_ID
API_TEST_SERVICE_PID=$API_TEST_SERVICE_PID
EOF

echo -e "INFO\tTest services started:"
echo -e "    \t  - httpbin -> http://localhost:$HTTPBIN_PORT"
echo -e "    \t  - speakeasy-api-test-service -> http://localhost:$API_TEST_SERVICE_PORT"
echo -e "INFO\tPort configuration written to $PORT_FILE"
