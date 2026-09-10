#!/usr/bin/env bash

set -euo pipefail

# shellcheck source=scripts/utils.sh
source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)/utils.sh"
load_local_env

# Bound concurrent package builds by default so this aggregate suite remains
# reliable on fresh developer machines and standard CI runners. Callers can
# raise or lower either value when they have measured capacity for it.
GO_TEST_PARALLELISM="${GO_TEST_PARALLELISM:-4}"
GO_TEST_TEST_PARALLELISM="${GO_TEST_TEST_PARALLELISM:-2}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-10m}"

for setting in GO_TEST_PARALLELISM GO_TEST_TEST_PARALLELISM; do
    value="${!setting}"
    if ! [[ "$value" =~ ^[1-9][0-9]*$ ]]; then
        printf '%s must be a positive integer; got %s\n' "$setting" "$value" >&2
        exit 2
    fi
done

log_dir="$(mktemp -d "${TMPDIR:-/tmp}/openapi-generation-test-generator.XXXXXX")"
log_file="$log_dir/test.log"
cleanup() {
    status=$?
    if ((status == 0)); then
        rm -rf "$log_dir"
    else
        printf 'Raw aggregate test output retained at %s\n' "$log_file" >&2
    fi
    return "$status"
}
trap cleanup EXIT

printf 'Running aggregate generator tests with -p %s, -parallel %s, and -timeout %s\n' "$GO_TEST_PARALLELISM" "$GO_TEST_TEST_PARALLELISM" "$GO_TEST_TIMEOUT"

go test -v -json -p "$GO_TEST_PARALLELISM" -parallel "$GO_TEST_TEST_PARALLELISM" -timeout "$GO_TEST_TIMEOUT" \
    ./tests/... \
    ./internal/... \
    ./pkg/... \
    ./cmd/... \
    ./changelogs/... 2>&1 | tee "$log_file" | go tool gotestfmt -hide all
