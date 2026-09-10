#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

while IFS= read -r -d '' module; do
  module_dir="$(dirname "$module")"
  printf 'go mod tidy in %s\n' "$module_dir"
  (
    cd "$module_dir"
    go mod tidy
  )
done < <(
  find "$REPO_ROOT" \
    -path "$REPO_ROOT/zSDKs" -prune -o \
    -path "$REPO_ROOT/testSDKs" -prune -o \
    -path "$REPO_ROOT/testusages" -prune -o \
    -path "$REPO_ROOT/cmd/wasm/internal" -prune -o \
    -name go.mod -print0
)
