#!/usr/bin/env bash

# profile-review-sdk.sh
#
# Profiles code generation for a review SDK target, producing CPU, memory,
# and Goja pprof files.
#
# Usage:
#   ./scripts/profile-review-sdk.sh <target>
#
# Arguments:
#   target - Target language (e.g. go, pythonv2, typescriptv2, csharp)
#
# Output:
#   Profiles are saved to ./profiles/<target>-review/
#
# Examples:
#   ./scripts/profile-review-sdk.sh go
#   ./scripts/profile-review-sdk.sh pythonv2

set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

if [[ $# -lt 1 ]] || [[ -z "$1" ]]; then
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset}"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}  The target language (e.g. go, pythonv2, typescriptv2)"
    exit 1
fi

ensure_license_election

readonly TARGET="$1"
readonly OUTDIR="./zSDKs/sdk-${TARGET}"
readonly PROFILE_DIR="./profiles/${TARGET}-review"

which yq > /dev/null || (print_formatted "{red}Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq{reset}"; exit 1)

mkdir -p "${OUTDIR}/.speakeasy"
cp "./tests/config/review/${TARGET}/.speakeasy/gen.yaml" "${OUTDIR}/.speakeasy/gen.yaml"
if [ -e "${OUTDIR}/.speakeasy/gen.lock" ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' "${OUTDIR}/.speakeasy/gen.lock" ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock "${OUTDIR}/.speakeasy/gen.lock"
fi

mkdir -p "${PROFILE_DIR}"

print_msg_with_ctx "📊 {green}{bold}Profiling SDK Generation{reset}" "target=${TARGET}|variant=review"
run_cmd go run cmd/generate/main.go -s ./tests/specs/review.yaml -o "${OUTDIR}" -l "${TARGET}" -profile "${PROFILE_DIR}"

echo ""
print_formatted "{green}{bold}Profiling complete!{reset} Files saved to {cyan}${PROFILE_DIR}/{reset}"
echo ""
print_formatted "{bold}View profiles in your browser:{reset}"
print_formatted "  {cyan}CPU:{reset}    go tool pprof -http=: ${PROFILE_DIR}/cpu.pprof"
print_formatted "  {cyan}Memory:{reset} go tool pprof -http=: ${PROFILE_DIR}/mem.pprof"
print_formatted "  {cyan}Goja:{reset}   go tool pprof -http=: ${PROFILE_DIR}/goja.pprof"
