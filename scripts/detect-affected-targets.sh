#!/bin/bash

set -e

# detect-affected-targets.sh
# Determines which SDK generation targets are affected by changes in the current branch
# compared to main branch. Outputs ONLY a comma-separated list of targets to stdout.
# If no targets are affected or only CI changes, outputs nothing (empty string).
# All diagnostic messages go to stderr.

# Valid SDK targets
ALL_TARGETS=(
  "cli"
  "typescriptv2"
  "pythonv2"
  "javav2"
  "go"
  "csharp"
  "php"
  "ruby"
  "terraform"
  "unity"
  "postman"
  "mcp-typescript"
)

# Get the base branch to compare against
BASE_BRANCH="${1:-origin/main}"

# Fetch the latest from origin to ensure we have up-to-date refs
git fetch origin --quiet 2>/dev/null || true

# Get list of changed files
CHANGED_FILES=$(git diff --name-only "$BASE_BRANCH"...HEAD 2>/dev/null || echo "")

if [ -z "$CHANGED_FILES" ]; then
  # No changes - output nothing
  exit 0
fi

# Initialize affected targets set
declare -A affected_targets
all_targets_affected=false
any_code_changes=false

# Patterns that affect ALL targets
ALL_TARGET_PATTERNS=(
  "^internal/"
  "^cmd/generate/"
  "^cmd/validate/"
  "^cmd/generatereadme/"
  "^cmd/generateusage/"
  "^templates/templates/common/"
  "^go\.mod$"
  "^go\.sum$"
  "^pkg/"
  "^Makefile$"
)

# Patterns that should NOT trigger any tests (CI infrastructure only)
NO_TEST_PATTERNS=(
  "^\.github/workflows/pr-checks\.yaml$"
  "^\.github/workflows/go-lint\.yml$"
  "^\.github/workflows/test-coverage\.yml$"
  "^\.github/workflows/preprocessing\.yml$"
  "^\.github/workflows/commits\.yml$"
  "^\.github/workflows/test-expiring-todos\.yml$"
  "^\.github/labeler\.yml$"
  "^scripts/lint\.sh$"
  "^CHANGELOG\.md$"
  "^README\.md$"
  "^docs/"
  "^\.gitignore$"
  "^\.golangci\.yml$"
  "^LICENSE$"
)

# Function to check if a file matches any pattern
matches_pattern() {
  local file=$1
  shift
  local patterns=("$@")

  for pattern in "${patterns[@]}"; do
    if echo "$file" | grep -qE "$pattern"; then
      return 0
    fi
  done
  return 1
}

# Analyze each changed file
while IFS= read -r file; do
  [ -z "$file" ] && continue

  # Check if file should not trigger any tests
  if matches_pattern "$file" "${NO_TEST_PATTERNS[@]}"; then
    continue
  fi

  # If we find any file that requires testing, mark it
  any_code_changes=true

  # Check if this change affects all targets
  if matches_pattern "$file" "${ALL_TARGET_PATTERNS[@]}"; then
    all_targets_affected=true
    break
  fi

  # Check for target-specific template changes
  if echo "$file" | grep -qE "^templates/templates/([^/]+)/"; then
    target=$(echo "$file" | sed -E 's|^templates/templates/([^/]+)/.*|\1|')
    # Verify it's a valid target (not "common" which we handle separately)
    if [[ " ${ALL_TARGETS[@]} " =~ " ${target} " ]]; then
      affected_targets["$target"]=1
      # Terraform and cli SDKs are generated from the Go templates, so Go
      # template changes should trigger their tests as well.
      if [ "$target" = "go" ]; then
        affected_targets["terraform"]=1
        affected_targets["cli"]=1
      fi
    fi
  fi

  # Check for test config changes
  if echo "$file" | grep -qE "^tests/config/[^/]+/([^/]+)/"; then
    target=$(echo "$file" | sed -E 's|^tests/config/[^/]+/([^/]+)/.*|\1|')
    if [[ " ${ALL_TARGETS[@]} " =~ " ${target} " ]]; then
      affected_targets["$target"]=1
    fi
  fi

  # Check for changelog changes
  if echo "$file" | grep -qE "^changelogs/([^/]+)/"; then
    target=$(echo "$file" | sed -E 's|^changelogs/([^/]+)/.*|\1|')
    if [[ " ${ALL_TARGETS[@]} " =~ " ${target} " ]]; then
      affected_targets["$target"]=1
    fi
  fi

  # Check for changeset files (affect all targets listed in the changeset)
  if echo "$file" | grep -qE "^\.changesets/.*\.yaml$"; then
    if [ -f "$file" ]; then
      # Extract bare target names from "- targetname" lines under "targets:"
      for cs_target in $(grep -E '^\s*- ' "$file" | sed -E 's/^\s*- //' | tr -d '[:space:]'); do
        if [[ " ${ALL_TARGETS[@]} " =~ " ${cs_target} " ]]; then
          affected_targets["$cs_target"]=1
        fi
      done
    else
      # File was deleted; conservatively test all targets
      all_targets_affected=true
      break
    fi
  fi

  # Check for target-specific workflow changes
  if echo "$file" | grep -qE "^\.github/workflows/test-([^.]+)\.(yml|yaml)$"; then
    target=$(echo "$file" | sed -E 's#^\.github/workflows/test-([^.]+)\.(yml|yaml)$#\1#')
    if [[ " ${ALL_TARGETS[@]} " =~ " ${target} " ]]; then
      affected_targets["$target"]=1
    fi
  fi
done <<< "$CHANGED_FILES"

# Output results to stdout (processable format only)
if [ "$any_code_changes" = false ]; then
  # Only CI/docs changes - output nothing
  exit 0
elif [ "$all_targets_affected" = true ]; then
  # Output all targets as comma-separated list
  IFS=,
  echo "${ALL_TARGETS[*]}"
else
  # Output specific targets as comma-separated list
  if [ ${#affected_targets[@]} -eq 0 ]; then
    # No specific patterns matched but we had code changes - test all to be safe
    IFS=,
    echo "${ALL_TARGETS[*]}"
  else
    # Output affected targets
    targets_list=""
    for target in "${!affected_targets[@]}"; do
      if [ -z "$targets_list" ]; then
        targets_list="$target"
      else
        targets_list="$targets_list,$target"
      fi
    done
    echo "$targets_list"
  fi
fi
