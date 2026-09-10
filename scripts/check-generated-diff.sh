#!/usr/bin/env bash
set -euo pipefail

if [ -z "$(git status --porcelain)" ]; then
  exit 0
fi

echo "Unexpected changes after code generation:"
git status --porcelain
git diff
echo "Regenerate with the documented make build command and commit these changes."
exit 1
