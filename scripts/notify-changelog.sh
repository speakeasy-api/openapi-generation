#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)
repo_root=$(dirname "$script_dir")
webhook_url="${CHANGELOG_NOTIFICATION_WEBHOOK_URL:-}"
if [[ -z "$webhook_url" ]]; then
  echo "Changelog notification is not configured; skipping."
  exit 0
fi

if ! previous_commit=$(git -C "$repo_root" rev-parse HEAD^ 2>/dev/null); then
  echo "No previous commit is available; skipping changelog notification."
  exit 0
fi
new_files=$(git -C "$repo_root" diff --name-only --diff-filter=A "$previous_commit" HEAD -- changelogs/ | grep '\.md$' || true)

if [[ -z "$new_files" ]]; then
  echo "No new changelog files to notify."
  exit 0
fi

while IFS= read -r file; do
  if [[ -f "$file" ]]; then
    file_name=$(basename "$file")
    payload=$(printf '{"text":"New changelog entry: %s"}' "$file_name")
    curl --fail --silent --show-error --request POST \
      --header 'Content-Type: application/json' \
      --data "$payload" \
      "$webhook_url"
  fi
done <<< "$new_files"
