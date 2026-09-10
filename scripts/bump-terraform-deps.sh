#!/usr/bin/env bash
# bump-terraform-deps.sh — Fetch latest tagged versions of terraform-plugin-*
# Go modules and update templates/templates/terraform/config.ts.
#
# Updates both:
#   1. getTemplateDependencies() — directly managed dependency versions
#   2. additionalDependencyMinimumVersions — minimum versions for customer
#      additionalDependencies that must stay compatible with the managed deps
#
# Usage:
#   ./scripts/bump-terraform-deps.sh          # fetch latest & update config.ts
#   ./scripts/bump-terraform-deps.sh --dry-run # show what would change

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_TS="$ROOT_DIR/templates/templates/terraform/config.ts"

# shellcheck source=utils.sh
source "$SCRIPT_DIR/utils.sh"

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

# Modules managed in getTemplateDependencies()
MANAGED_MODULES=(
  "github.com/hashicorp/terraform-plugin-docs"
  "github.com/hashicorp/terraform-plugin-framework"
  "github.com/hashicorp/terraform-plugin-framework-jsontypes"
  "github.com/hashicorp/terraform-plugin-framework-validators"
  "github.com/hashicorp/terraform-plugin-go"
  "github.com/hashicorp/terraform-plugin-log"
)

# Modules managed in additionalDependencyMinimumVersions
ADDITIONAL_DEP_MODULES=(
  "github.com/hashicorp/terraform-plugin-sdk/v2"
  "github.com/hashicorp/terraform-plugin-testing"
)

# fetch_latest_version queries the Go module proxy for the latest version.
fetch_latest_version() {
  local module="$1"
  local url="https://proxy.golang.org/${module}/@latest"
  local version

  version=$(curl -fsSL "$url" | grep -Eo '"Version"[[:space:]]*:[[:space:]]*"[^"]+"' | head -1 | sed 's/.*"v/v/' | sed 's/"$//')

  if [[ -z "$version" ]]; then
    echo "${RED}ERROR:${RESET} Failed to fetch latest version for $module" >&2
    return 1
  fi

  echo "$version"
}

# update_config_version replaces a version string in config.ts for a given module.
# Matches the pattern: version: "vX.Y.Z" on lines following the module name.
update_config_version() {
  local module="$1"
  local new_version="$2"
  local file="$3"

  # For getTemplateDependencies entries: find the module name string, then
  # update the version field on a subsequent line.
  # Pattern: "module": { ... version: "vX.Y.Z" ...
  # We use awk to find the module key and update the next version line.
  local tmp
  tmp=$(mktemp)

  awk -v mod="\"$module\"" -v ver="\"$new_version\"" '
    $0 ~ mod { found=1 }
    found && /version:/ {
      sub(/version: "[^"]*"/, "version: " ver)
      found=0
    }
    { print }
  ' "$file" > "$tmp" && mv "$tmp" "$file"
}

# update_additional_dep_version replaces a version string in the
# additionalDependencyMinimumVersions map.
# Pattern: "module": "vX.Y.Z",
update_additional_dep_version() {
  local module="$1"
  local new_version="$2"
  local file="$3"

  local tmp
  tmp=$(mktemp)

  awk -v mod="\"$module\"" -v ver="\"$new_version\"" '
    $0 ~ mod && /: "v/ {
      sub(/: "v[^"]*"/, ": " ver)
    }
    { print }
  ' "$file" > "$tmp" && mv "$tmp" "$file"
}

changed=0

echo "${BOLD}Fetching latest terraform-plugin-* versions...${RESET}"
echo ""

echo "${CYAN}Managed dependencies (getTemplateDependencies):${RESET}"
for module in "${MANAGED_MODULES[@]}"; do
  latest=$(fetch_latest_version "$module")
  # Extract current version from config.ts by finding the module key line,
  # then printing the next line containing "version:" and extracting the value.
  current=$(awk -v mod="\"$module\"" '
    $0 ~ mod { found=1; next }
    found && /version:/ { print; found=0 }
  ' "$CONFIG_TS" | grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)

  if [[ "$current" == "$latest" ]]; then
    echo "  ${GREEN}✓${RESET} $module ${GREY}$current (up to date)${RESET}"
  else
    echo "  ${YELLOW}↑${RESET} $module ${RED}$current${RESET} → ${GREEN}$latest${RESET}"
    changed=1
    if [[ "$DRY_RUN" == false ]]; then
      update_config_version "$module" "$latest" "$CONFIG_TS"
    fi
  fi
done

echo ""
echo "${CYAN}Additional dependency minimums (additionalDependencyMinimumVersions):${RESET}"
for module in "${ADDITIONAL_DEP_MODULES[@]}"; do
  latest=$(fetch_latest_version "$module")
  # Extract current minimum version from the additionalDependencyMinimumVersions map
  current=$(grep -F "\"$module\"" "$CONFIG_TS" | grep -Eo 'v[0-9]+\.[0-9]+\.[0-9]+' | tail -1)

  if [[ "$current" == "$latest" ]]; then
    echo "  ${GREEN}✓${RESET} $module ${GREY}$current (up to date)${RESET}"
  else
    echo "  ${YELLOW}↑${RESET} $module ${RED}$current${RESET} → ${GREEN}$latest${RESET}"
    changed=1
    if [[ "$DRY_RUN" == false ]]; then
      update_additional_dep_version "$module" "$latest" "$CONFIG_TS"
    fi
  fi
done

echo ""
if [[ "$DRY_RUN" == true ]]; then
  echo "${MAGENTA}Dry run — no files were modified.${RESET}"
elif [[ "$changed" -eq 0 ]]; then
  echo "${GREEN}All terraform-plugin-* dependencies are up to date.${RESET}"
else
  echo "${GREEN}Updated ${CONFIG_TS##"$ROOT_DIR"/}${RESET}"
  echo ""
  echo "Next steps:"
  echo "  1. Run ${BOLD}npm run format${RESET} to apply prettier formatting"
  echo "  2. Run ${BOLD}make check-template-terraform${RESET} to verify TypeScript checks"
  echo "  3. Run ${BOLD}./scripts/build-review-terraform.sh${RESET} to test"
fi
