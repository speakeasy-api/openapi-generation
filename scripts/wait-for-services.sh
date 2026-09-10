#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PORT_FILE="$REPO_ROOT/.test-ports"

if [[ ! -f "$PORT_FILE" ]]; then
  printf 'ERROR\tMissing %s. Run ./scripts/boot-services.sh first.\n' "$PORT_FILE"
  exit 1
fi

source "$PORT_FILE"

attempt=0
until curl -s -f -o /dev/null "http://localhost:$HTTPBIN_PORT/anything"; do
  attempt=$((attempt + 1))
  if ((attempt >= 30)); then
    printf 'ERROR\tTimed out waiting for httpbin after 60s\n' >&2
    exit 1
  fi
  sleep 2
  printf 'INFO\tWaiting for httpbin: http://localhost:%s\n' "$HTTPBIN_PORT"
done

attempt=0
until curl -s -f -o /dev/null "http://localhost:$API_TEST_SERVICE_PORT/ping"; do
  attempt=$((attempt + 1))
  if ((attempt >= 30)); then
    printf 'ERROR\tTimed out waiting for speakeasy-api-test-service after 60s\n' >&2
    exit 1
  fi
  sleep 2
  printf 'INFO\tWaiting for speakeasy-api-test-service: http://localhost:%s\n' "$API_TEST_SERVICE_PORT"
done

printf 'INFO\tServices ready: httpbin=%s api-test=%s\n' "$HTTPBIN_PORT" "$API_TEST_SERVICE_PORT"
