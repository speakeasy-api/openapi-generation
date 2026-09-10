#!/usr/bin/env bash

# profile-spec.sh
#
# Profiles code generation for a given target using an arbitrary OpenAPI spec,
# producing CPU, memory, and Goja pprof files.
#
# Usage:
#   ./scripts/profile-spec.sh <target> <spec>
#
# Arguments:
#   target - Target language/type (e.g. go, pythonv2, typescriptv2, terraform)
#   spec   - Path to an existing OpenAPI Specification document file
#
# Output:
#   Profiles are saved to ./profiles/<basename>-<target>/ where <basename> is
#   derived from the spec filename (e.g. my-api.yaml -> my-api).
#
# Examples:
#   ./scripts/profile-spec.sh go ./my-api.yaml        -> profiles/my-api-go/
#   ./scripts/profile-spec.sh typescriptv2 /tmp/openapi.yaml -> profiles/openapi-typescriptv2/

set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
ensure_license_election

if [[ $# -lt 2 ]]; then
    print_formatted "{bold}{yellow}Usage:{reset} $0 {bold}{blue}<target>{reset} {bold}{blue}<spec>{reset}"
    print_formatted "\n{bold}Arguments:{reset}"
    print_formatted "  {bold}{green}target:{reset}  The target language (e.g. go, pythonv2, terraform)"
    print_formatted "  {bold}{green}spec:{reset}    Path to an OpenAPI Specification document file"
    exit 1
fi

readonly TARGET="$1"
readonly SPEC="$2"

# Derive a directory-safe basename from the spec filename (strip path, extensions, replace unsafe chars)
SPEC_BASENAME="$(basename "${SPEC}")"
SPEC_BASENAME="${SPEC_BASENAME%%.*}"
SPEC_BASENAME="${SPEC_BASENAME//[^a-zA-Z0-9._-]/-}"

readonly OUTDIR="./zSDKs/profile-${SPEC_BASENAME}-${TARGET}"
readonly PROFILE_DIR="./profiles/${SPEC_BASENAME}-${TARGET}"

if [[ ! -f "${SPEC}" ]]; then
    print_formatted "{red}Error: spec file not found: ${SPEC}{reset}"
    exit 1
fi

mkdir -p "${OUTDIR}"
mkdir -p "${PROFILE_DIR}"

print_msg_with_ctx "📊 {green}{bold}Profiling Generation{reset}" "target=${TARGET}|spec=${SPEC}"
run_cmd go run cmd/generate/main.go -s "${SPEC}" -o "${OUTDIR}" -l "${TARGET}" -profile "${PROFILE_DIR}"

echo ""
print_formatted "{green}{bold}Profiling complete!{reset} Files saved to {cyan}${PROFILE_DIR}/{reset}"
echo ""
print_formatted "{bold}View profiles in your browser:{reset}"
print_formatted "  {cyan}CPU:{reset}    go tool pprof -http=: ${PROFILE_DIR}/cpu.pprof"
print_formatted "  {cyan}Memory:{reset} go tool pprof -http=: ${PROFILE_DIR}/mem.pprof"
print_formatted "  {cyan}Goja:{reset}   go tool pprof -http=: ${PROFILE_DIR}/goja.pprof"
