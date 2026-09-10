#!/bin/bash

set -euo pipefail

ROOT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

retired_paths=(
  cmd/snapshot-test-executor
  cmd/snapshot-test-workflow
  cmd/workflow
  cmd/regression-test
  scripts/regen-snapshot-tests.sh
  .github/disabled-workflows
  .github/workflows/composite/snapshot-test
  .github/workflows/composite/snapshot-test-build
  .github/workflows/composite/customer-regression-test
  .claude/commands/snapshot-test.md
  .claude/commands/forced-flag-pr.md
)

for path in "${retired_paths[@]}"; do
  if [[ -e "$path" ]]; then
    echo "retired private snapshot path must not exist: $path" >&2
    exit 1
  fi
done

retired_workflow_names=(
  regression-test
  snapshot-test
  sync-snapshot-test
)
for name in "${retired_workflow_names[@]}"; do
  for extension in yml yaml; do
    if [[ -e ".github/workflows/$name.$extension" ]]; then
      echo "retired private snapshot workflow must not exist: .github/workflows/$name.$extension" >&2
      exit 1
    fi
  done
done

if compgen -G '.github/workflows/snapshot-test-*.yaml' > /dev/null ||
   compgen -G '.github/workflows/snapshot-test-*.yml' > /dev/null; then
  echo "generated private snapshot workflows must not exist" >&2
  exit 1
fi

retired_reference_pattern='cmd/(snapshot-test-executor|snapshot-test-workflow|regression-test)|scripts/regen-snapshot-tests|\.github/(disabled-workflows|workflows/(regression-test|snapshot-test|sync-snapshot-test))|\.github/workflows/composite/(snapshot-test|snapshot-test-build|customer-regression-test)|\.claude/commands/(snapshot-test|forced-flag-pr)'

if matches=$(git grep -I -nE "$retired_reference_pattern" -- . \
  ':(exclude)CHANGELOG.md' \
  ':(exclude)changerecord.md' \
  ':(exclude)scripts/check-no-private-snapshot-machinery.sh'); then
  printf '%s\n' "$matches" >&2
  echo "active reference to retired private snapshot machinery" >&2
  exit 1
fi

retained_paths=(
  .github/workflows/snapshot-dispatch.yml
  scripts/test-snapshot-dispatch.mjs
  internal/snapshotdispatch/workflow_test.go
)
for path in "${retained_paths[@]}"; do
  if [[ ! -f "$path" ]]; then
    echo "retained generic snapshot artifact must exist: $path" >&2
    exit 1
  fi
done

snapshot_test_count=$(git ls-files 'pkg/generate/snapshots/*_test.go' 'pkg/generate/snapshots/**/*_test.go' | wc -l | tr -d '[:space:]')
if [[ "$snapshot_test_count" -lt 92 ]]; then
  echo "expected at least 92 generic generator snapshot tests, found $snapshot_test_count" >&2
  exit 1
fi

echo "private snapshot machinery guard passed"
