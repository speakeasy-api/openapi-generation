#!/usr/bin/env bash

# profile-review-mcp.sh
#
# Profiles code generation for a review MCP target, producing CPU, memory,
# and Goja pprof files.
#
# Usage:
#   ./scripts/profile-review-mcp.sh <target>
#
# Arguments:
#   target - MCP target (e.g. typescript or mcp-typescript)
#
# Output:
#   Profiles are saved to ./profiles/<target>-review/
#
# Examples:
#   ./scripts/profile-review-mcp.sh typescript

set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

if [[ $# -lt 1 ]] || [[ -z "$1" ]]; then
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset}"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}  The MCP target (e.g. typescript or mcp-typescript)"
    exit 1
fi

ensure_license_election

TARGET="$1"
if [[ "${TARGET}" != "mcp-"* ]]; then
    TARGET="mcp-${TARGET}"
fi
readonly TARGET
readonly TARGET_DIR="./zSDKs/${TARGET}"
readonly PROFILE_DIR="./profiles/${TARGET}-review"

which yq > /dev/null || (print_formatted "{red}Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq{reset}"; exit 1)

mkdir -p "${TARGET_DIR}/.speakeasy"
cp "./tests/config/review/${TARGET}/.speakeasy/gen.yaml" "${TARGET_DIR}/.speakeasy/gen.yaml"
if [ -e "${TARGET_DIR}/.speakeasy/gen.lock" ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' "${TARGET_DIR}/.speakeasy/gen.lock" ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock "${TARGET_DIR}/.speakeasy/gen.lock"
fi

mkdir -p "${PROFILE_DIR}"

print_msg_with_ctx "📊 {green}{bold}Profiling MCP Generation{reset}" "target=${TARGET}|variant=review"
run_cmd go run cmd/generate/main.go -s ./tests/specs/review.yaml -o "${TARGET_DIR}" -l "${TARGET}" -profile "${PROFILE_DIR}"

echo ""
print_formatted "{green}{bold}Profiling complete!{reset} Files saved to {cyan}${PROFILE_DIR}/{reset}"
echo ""
print_formatted "{bold}View profiles in your browser:{reset}"
print_formatted "  {cyan}CPU:{reset}    go tool pprof -http=: ${PROFILE_DIR}/cpu.pprof"
print_formatted "  {cyan}Memory:{reset} go tool pprof -http=: ${PROFILE_DIR}/mem.pprof"
print_formatted "  {cyan}Goja:{reset}   go tool pprof -http=: ${PROFILE_DIR}/goja.pprof"
