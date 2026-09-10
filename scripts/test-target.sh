#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# test-target.sh
#
# This script runs tests for generated SDKs against specified test targets.
#
# Arguments:
#   $1  - Target (e.g. go, python, typescript)
#   $2+ - Variants to test (e.g. primary, secondary, review)
#
# Optional flags:
#   --skip-primary-usage: Skip primary usage tests
#   --skip-standalone-usage: Skip standalone usage tests
#
# Example usage:
#   ./test-target.sh go primary secondary review  # Test all variants with all test types
#   ./test-target.sh python primary --skip-primary-usage  # Skip primary usage tests
#   ./test-target.sh typescript primary review  # Test primary and review variants

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
load_local_env
readonly PROJECT_ROOT="${SCRIPT_DIR}/.."

# Source dynamically allocated ports if available
readonly PORT_FILE="${PROJECT_ROOT}/.test-ports"
if [[ -f "${PORT_FILE}" ]]; then
    source "${PORT_FILE}"
fi
# Export with defaults as fallback
export HTTPBIN_PORT="${HTTPBIN_PORT:-35123}"
export API_TEST_SERVICE_PORT="${API_TEST_SERVICE_PORT:-35456}"
export HTTPBIN_URL="${HTTPBIN_URL:-http://localhost:$HTTPBIN_PORT}"
export API_TEST_SERVICE_URL="${API_TEST_SERVICE_URL:-http://localhost:$API_TEST_SERVICE_PORT}"

function usage() {
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset} {bold}{blue}<variant...>{reset} [{bold}{blue}flags{reset}]"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}  The target to run tests for"
    print_formatted "  {bold}{green}variant:{reset}   One or more variants to test"
    print_formatted "\n{bold}Optional Flags:{reset}"
    print_formatted "  {bold}{green}--skip-primary-usage{reset}      Skip primary usage tests"
    print_formatted "  {bold}{green}--skip-standalone-usage{reset}   Skip standalone usage tests"
    printf "\n"
    exit 1
}

function derive_sdk_target_dir() {
    local -r target="$1"
    local -r variant="$2"

    if [[ "${variant}" == "review" ]]; then
        echo "./zSDKs/sdk-${target}"
        return
    fi

    local dir_name
    if [[ "${target}" == "mcp-"* ]]; then
        dir_name="${target}-${variant}"
    elif [[ "${target}" == "terraform" ]]; then
        dir_name="terraform-provider-${variant}"
    else
        dir_name="sdk-${target}-${variant}"
    fi

    echo "./testSDKs/${dir_name}"
}

function targettest_sdk() {
    local -r target="$1"
    local -r variant="$2"
    local -r sdk_target_dir="$(derive_sdk_target_dir "${target}" "${variant}")"

    if [[ ! -d "${sdk_target_dir}" ]]; then
        print_msg_with_ctx "${SKIP_ICON} {yellow}{bold}Skipping Target Tests{reset}" "target=${target}|variant=${variant}"
        return
    fi
    print_msg_with_ctx "${TARGET_ICON} {blue}{bold}Running Target Tests{reset}" "target=${target}|variant=${variant}"
    run_cmd go run cmd/targettest/main.go -t "${target}" -o "${sdk_target_dir}" || exit 1
}

function test_usage_file() {
    # Test USAGE.md templating (running 'tests/testusage.go' for the primary spec)
    local -r target="$1"
    print_msg_with_ctx "${SNIPPETS_ICON} {blue}{bold}Testing Usage Example Selection{reset}" "target=${target}|variant=primary"
    pushd "${PROJECT_ROOT}" > /dev/null && run_cmd go run ./tests -lang "${target}" -mode usage -group primary || exit 1; popd > /dev/null
}

function test_standalone_usage() {
    # Test standalone usage snippet generation ('generateusage' command)
    local -r target="$1"
    print_msg_with_ctx "${SNIPPETS_ICON} {blue}{bold}Testing Standalone Usage Snippet Generation{reset}" "target=${target}|variant=review"
    run_cmd "${SCRIPT_DIR}/test-standalone-usage.sh" "${target}" || exit 1
}

function test_coverage() {
    local -r target="$1"
    print_msg_with_ctx "${COVERAGE_ICON} {blue}{bold}Testing Coverage{reset}" "target=${target}"
    pushd "${PROJECT_ROOT}" > /dev/null && run_cmd go run ./tests -lang "${target}" -mode coverage || exit 1; popd > /dev/null
}

function test_target() {
    local -r target="$1"
    shift
    local -r variants=("$@")
    local -r skip_primary_usage="${SKIP_PRIMARY_USAGE:-false}"
    local -r skip_standalone_usage="${SKIP_STANDALONE_USAGE:-false}"

    # Run target tests for each variant
    for variant in "${variants[@]}"; do
        targettest_sdk "${target}" "${variant}"
    done

    print_divider "Usage Tests"

    # Run usage tests if not skipped
    if [[ " ${variants[@]} " =~ " primary " ]]; then
        if [[ "${skip_primary_usage}" == "true" ]]; then
            print_msg_with_ctx "${SKIP_ICON} {yellow}{bold} Skipping Usage Example Selection Tests{reset}" "target=${target}"
        else
            test_usage_file "${target}"
        fi
    fi

    # Run standalone usage tests if not skipped
    if [[ " ${variants[@]} " =~ " review " ]]; then
        if [[ "${skip_standalone_usage}" == "true" ]]; then
            print_msg_with_ctx "${SKIP_ICON} {yellow}{bold} Skipping Standalone Usage Snippet Tests{reset}" "target=${target}"
        else
            test_standalone_usage "${target}"
        fi
    fi

    print_divider "Coverage Tests"

    # Run coverage tests if multiple variants specified
    if [[ "${#variants[@]}" -gt 1 ]]; then
        test_coverage "${target}"
    fi
    
    print_formatted "${SUCCESS_ICON} All {bold}{cyan}${target}{reset} tests passed successfully"
}

print_divider "Target Tests"

if [[ $# -lt 1 ]]; then
    usage
fi

# Parse arguments and flags
declare -a variants=()
target="$1"
shift

while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-primary-usage)
            export SKIP_PRIMARY_USAGE=true
            shift
            ;;
        --skip-standalone-usage)
            export SKIP_STANDALONE_USAGE=true
            shift
            ;;
        --help)
            usage
            ;;
        *)
            variants+=("$1")
            shift
            ;;
    esac
done


if [[ "${#variants[@]}" -eq 0 ]]; then
    print_formatted "${SKIP_ICON} {bold}{yellow}Skipping tests for ${target} as no variants were specified{reset}"
    exit 0
fi

test_target "${target}" "${variants[@]}"
