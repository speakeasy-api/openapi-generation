#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

printf 'INFO\tCompiling api-test-service...\n'
mkdir -p "$REPO_ROOT/bin"
go build -C "$REPO_ROOT/services/speakeasy-api-test-service" -o ../../bin/api-test-service ./cmd/server/main.go

bash "$SCRIPT_DIR/boot-services.sh"
bash "$SCRIPT_DIR/wait-for-services.sh"
