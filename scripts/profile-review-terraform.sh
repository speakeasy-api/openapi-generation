#!/usr/bin/env bash
set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
ensure_license_election

OUTDIR="./zSDKs/terraform-provider-testing"
PROFILE_DIR="./profiles/review-terraform"

which yq > /dev/null || (print_formatted "{red}Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq{reset}"; exit 1)
mkdir -p "${OUTDIR}"
mkdir -p "${OUTDIR}/.speakeasy"
cp "./tests/config/review/terraform/.speakeasy/gen.yaml" "${OUTDIR}/.speakeasy/gen.yaml"
if [ -e "${OUTDIR}/.speakeasy/gen.lock" ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' "${OUTDIR}/.speakeasy/gen.lock" ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock "${OUTDIR}/.speakeasy/gen.lock"
fi

mkdir -p "${PROFILE_DIR}"

print_msg_with_ctx "📊 {green}{bold}Profiling Terraform Provider Generation{reset}" "variant=review"
run_cmd go run cmd/generate/main.go -s ./tests/specs/review-terraform.yaml -o "${OUTDIR}" -l terraform -t review -profile "${PROFILE_DIR}"

echo ""
print_formatted "{green}{bold}Profiling complete!{reset} Files saved to {cyan}${PROFILE_DIR}/{reset}"
echo ""
print_formatted "{bold}View profiles in your browser:{reset}"
print_formatted "  {cyan}CPU:{reset}    go tool pprof -http=: ${PROFILE_DIR}/cpu.pprof"
print_formatted "  {cyan}Memory:{reset} go tool pprof -http=: ${PROFILE_DIR}/mem.pprof"
print_formatted "  {cyan}Goja:{reset}   go tool pprof -http=: ${PROFILE_DIR}/goja.pprof"
