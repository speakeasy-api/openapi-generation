#!/usr/bin/env bash

# profile-target.sh
#
# Profiles code generation for a given target and test group (variant),
# producing CPU, memory, and Goja pprof files.
#
# Usage:
#   ./scripts/profile-target.sh <target> <group>
#
# Arguments:
#   target - Target language/type (e.g. go, pythonv2, typescriptv2, terraform)
#   group  - Test group / variant (e.g. primary, secondary)
#
# Output:
#   Profiles are saved to ./profiles/<target>-<group>/
#
# Examples:
#   ./scripts/profile-target.sh go primary
#   ./scripts/profile-target.sh typescriptv2 secondary

set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
ensure_license_election

if [[ $# -lt 2 ]]; then
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset} {bold}{blue}<group>{reset}"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}  The target language (e.g. go, pythonv2, terraform)"
    print_formatted "  {bold}{green}group:{reset}   The test group / variant (e.g. primary, secondary)"
    exit 1
fi

readonly TARGET="$1"
readonly GROUP="$2"
readonly PROFILE_DIR="./profiles/${TARGET}-${GROUP}"

# Determine directory names like build-target.sh does
if [[ "${TARGET}" == "mcp-"* ]]; then
    SDK_NAME="${TARGET}-${GROUP}"
elif [[ "${TARGET}" == "terraform" ]]; then
    SDK_NAME="terraform-provider-${GROUP}"
else
    SDK_NAME="sdk-${TARGET}-${GROUP}"
fi

readonly OUTDIR="./testSDKs/${SDK_NAME}"
readonly SPEC="${OUTDIR}/openapi.yaml"

if [[ ! -f "${SPEC}" ]]; then
    print_formatted "{grey}Test spec not found, building it first...{reset}"
    EXTRA_ARGS="" LOG_OUTPUT="" "${SCRIPT_DIR}/build-target.sh" "${TARGET}" "${GROUP}"
fi

mkdir -p "${PROFILE_DIR}"

print_msg_with_ctx "📊 {green}{bold}Profiling Generation{reset}" "target=${TARGET}|group=${GROUP}"
run_cmd go run cmd/generate/main.go -s "${SPEC}" -o "${OUTDIR}" -l "${TARGET}" -t "${GROUP}" -profile "${PROFILE_DIR}"

echo ""
print_formatted "{green}{bold}Profiling complete!{reset} Files saved to {cyan}${PROFILE_DIR}/{reset}"
echo ""
print_formatted "{bold}View profiles in your browser:{reset}"
print_formatted "  {cyan}CPU:{reset}    go tool pprof -http=: ${PROFILE_DIR}/cpu.pprof"
print_formatted "  {cyan}Memory:{reset} go tool pprof -http=: ${PROFILE_DIR}/mem.pprof"
print_formatted "  {cyan}Goja:{reset}   go tool pprof -http=: ${PROFILE_DIR}/goja.pprof"
