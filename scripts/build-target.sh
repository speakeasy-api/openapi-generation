#!/usr/bin/env bash

# build-target.sh
#
# This script builds a given target and variants. It handles the full generation
# process including test spec creation, asset syncing, and generation.
#
# The script performs the following steps:
# 1. Validates input arguments and environment
# 2. For each variant:
#    a. Builds test specifications by:
#       - Cleaning up existing test artifacts
#       - Creating target directory structure 
#       - Syncing test assets from ./tests/config/<language>/<test group>
#       - Copying configuration files (gen.yaml, gen.lock)
#       - Building OpenAPI document
#    b. Generates the target using the test spec
#
# Usage:
#   ./build-target.sh <target> <variant...>
#
# Arguments:
#   target    - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant   - One or more variants to generate (e.g. primary, secondary)
#
# Environment Variables:
#   EXTRA_ARGS - Additional arguments to pass to the generator (optional)
#   LOG_OUTPUT - If set, redirects command output to a log file
#
# Examples:
#   # Generate Python SDK for primary and secondary variants
#   ./build-target.sh python primary secondary
#
#   # Generate TypeScript SDK with extra generator args
#   EXTRA_ARGS="--skip-compile" ./build-target.sh typescript primary


set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
ensure_license_election

# Use strict RuboCop config for internal builds to enforce double quotes
# (matching rubyfmt output). The base .rubocop.yml disables this rule so
# customers with single-quote conventions aren't broken.
export RUBOCOP_OPTS="--config .rubocop-strict.yml"

# Prints usage information for the script and exits with code 1
function print_usage() {
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset} {bold}{blue}<variant...>{reset}"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}       The target to generate the SDK for"
    print_formatted "  {bold}{green}variant:{reset}      One or more variants to generate SDKs for"
    print_formatted "\n{bold}Environment Variables:{reset}"
    print_formatted "  {bold}{green}EXTRA_ARGS{reset}    Additional arguments to pass to the generator"
    print_formatted "  {bold}{green}LOG_OUTPUT{reset}    If set, redirects command output to a log file"
    printf "\n"
    exit 1
}

# Derives the SDK output directory name based on target and variant
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant - Testing variant name (e.g. primary, secondary)
# Returns: SDK directory name as string
function get_target_dir_name() {
    local -r target="$1"
    local -r variant="$2"
    
    local dir_name="sdk-${target}-${variant}"

    if [[ "${target}" == "mcp-"* ]]; then
        dir_name="${target}-${variant}"
    elif [[ "${target}" == "terraform" ]]; then
        dir_name="terraform-provider-${variant}"
    fi

    echo "${dir_name}"
}

# Gets the path to test configuration files for a target/variant
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant - Testing variant name (e.g. primary, secondary)
# Returns: Path to test config directory as string
function get_test_config_path() {
    local -r target="$1"
    local -r variant="$2"

    local test_config_path="./tests/config/${variant}/${target}"
    if [[ "${target}" == "terraform" ]]; then
        test_config_path="./tests/config/${variant}/terraform"
    fi

    echo "${test_config_path}"
}

# Validates that script is being run from project root directory
# Exits with code 1 if not in correct directory
function validate_project_root() {
    if ! grep -q "module github.com/speakeasy-api/openapi-generation/v2" "./go.mod"; then
        print_formatted "{red}This script must be run from the root of the openapi-generation repository.{reset}"
        exit 1
    fi
}

# Syncs test assets from source config to test SDK directory
# The assets must live under: `./tests/config/<language>/<test group>`
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant - Testing variant name (e.g. primary, secondary)
function sync_test_assets() {
    local -r target="$1"
    local -r variant="$2"

    print_formatted "{grey}Syncing test assets{reset}"

    validate_project_root

    local -r test_config_path="$(get_test_config_path "${target}" "${variant}")"
    local -r test_sdk_path="./testSDKs/$(get_target_dir_name "${target}" "${variant}")"

    mkdir -p "${test_sdk_path}/.speakeasy"

    if ! command -v rsync &> /dev/null; then
        print_formatted "{red}{bold}rsync{reset} {red}is not installed. Please install it and try again.{reset}"
        exit 1
    fi
    run_cmd rsync -a "${test_config_path}/" "${test_sdk_path}/"

    print_formatted "{grey}Renaming *.go.hidden to *.go{reset}"
    find "${test_sdk_path}" -type f -name "*.go.hidden" -exec sh -c 'mv "$0" "${0%.hidden}"' {} \;
    print_formatted "{grey}Synced ${test_config_path} into ${test_sdk_path}{reset}"
}

# Rewrites the well-known local test service URLs in copied test artifacts when
# the service runner had to select different ports.
# Args:
#   test_sdk_path - Path to the test SDK being assembled
function rewrite_test_service_urls() {
    local -r test_sdk_path="$1"
    local -r httpbin_url="${HTTPBIN_URL:-http://localhost:35123}"
    local -r api_test_service_url="${API_TEST_SERVICE_URL:-http://localhost:35456}"

    if [[ "${httpbin_url}" == "http://localhost:35123" && "${api_test_service_url}" == "http://localhost:35456" ]]; then
        return
    fi

    TEST_SDK_PATH="${test_sdk_path}" \
        HTTPBIN_URL="${httpbin_url}" \
        API_TEST_SERVICE_URL="${api_test_service_url}" \
        python3 - <<'PY'
import os
from pathlib import Path

root = Path(os.environ["TEST_SDK_PATH"])
replacements = {
    "http://localhost:35123": os.environ["HTTPBIN_URL"],
    "http://localhost:35456": os.environ["API_TEST_SERVICE_URL"],
}

for path in root.rglob("*"):
    if not path.is_file() or path.suffix not in {".yaml", ".yml"}:
        continue

    content = path.read_text(encoding="utf-8")
    updated = content
    for default_url, selected_url in replacements.items():
        updated = updated.replace(default_url, selected_url)

    if updated != content:
        path.write_text(updated, encoding="utf-8")
PY
}

# Creates test specification by cleaning directories and syncing required files
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant - Testing variant name (e.g. primary, secondary)
function create_test_spec() {
    local -r target="$1"
    local -r variant="$2"
    local -r target_dir_name="$(get_target_dir_name "${target}" "${variant}")"

    print_formatted "{grey}Building test spec{reset}"

    rm -rf "./testusages/${target}/${variant}/" || true
    rm -rf "./testprojects/${target}/${target_dir_name}/" || true
    rm -rf "./testSDKs/${target_dir_name}" || true
    mkdir -p "./testSDKs/${target_dir_name}"

    sync_test_assets "${target}" "${variant}"

    if [[ -f "./tests/config/${variant}/gen.lock" ]]; then
        cp "./tests/config/${variant}/gen.lock" "./testSDKs/${target_dir_name}/.speakeasy/gen.lock"
    fi

    # Copy test files from variant-specific directory if it exists, otherwise from uber directory
    if [[ -d "./tests/tests/${variant}" ]]; then
        cp -r "./tests/tests/${variant}/"* "./testSDKs/${target_dir_name}/.speakeasy/" 2>/dev/null || true
    else
        cp -r "./tests/tests/uber/"* "./testSDKs/${target_dir_name}/.speakeasy/" 2>/dev/null || true
    fi
    cp -r "./tests/.hooks/${target}/${variant}/"* "./testSDKs/${target_dir_name}/" 2>/dev/null || true

    run_cmd "${SCRIPT_DIR}/build-openapi-document.sh" "${target}" "${variant}"
    rewrite_test_service_urls "./testSDKs/${target_dir_name}"
}

# Generates test target (legacy named testSDKs) for a target/variant combination
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variant - Testing variant name (e.g. primary, secondary)
function generate_test_target() {
    local -r target="$1"
    local -r variant="$2"
    local -r extra_args="${EXTRA_ARGS:-}"
    local -r sdk_name="$(get_target_dir_name "${target}" "${variant}")"

    if [[ ! -d "./tests/config/${variant}/${target}" ]]; then
        print_msg_with_ctx "${SKIP_ICON} {yellow}{bold}Skipping Generation (no ./tests/config/${variant}/${target} directory){reset}" "target=${target}|variant=${variant}"
        return
    fi

    print_msg_with_ctx "${BUILD_ICON} {green}{bold}Building Target{reset}" "target=${target}|variant=${variant}|args=${extra_args}"
    create_test_spec "${target}" "${variant}"

    run_cmd go run cmd/generate/main.go \
        -t "${variant}" \
        -s "./testSDKs/${sdk_name}/openapi.yaml" \
        -o "./testSDKs/${sdk_name}" \
        -l "${target}" \
        ${extra_args}
}

# Main function to build SDKs for a target and its variants
# Args:
#   target - Target to generate (e.g. go, mcp-typescript, terraform)
#   variants - Array of variant names (e.g. primary, secondary)
function build_target() {
    local -r target="$1"
    shift
    local -r variants=("$@")

    for variant in "${variants[@]}"; do
        # TODO: move build-review-sdk logic into this file.
        if [[ "${variant}" == "review" ]]; then
            if [[ "${target}" == "mcp-"* ]]; then
                "${SCRIPT_DIR}/build-review-mcp.sh" "${target}"
            elif [[ "${target}" == "terraform" ]]; then
                "${SCRIPT_DIR}/build-review-terraform.sh"
            else
                "${SCRIPT_DIR}/build-review-sdk.sh" "${target}"
            fi
            continue
        fi
        generate_test_target "${target}" "${variant}"
    done
}

print_divider "Building Targets"

if [[ $# -lt 2 ]]; then
    print_usage
fi

readonly TARGET="$1"
shift
readonly VARIANTS=($@)

build_target "${TARGET}" "${VARIANTS[@]}"
