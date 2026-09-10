#!/usr/bin/env bash
set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

OUTDIR="./zSDKs/terraform-provider-testing"
ensure_license_election

which yq > /dev/null || (print_formatted "{red}Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq{reset}"; exit 1)
mkdir -p "${OUTDIR}"
mkdir -p "${OUTDIR}/.speakeasy"
cp "./tests/config/review/terraform/.speakeasy/gen.yaml" "${OUTDIR}/.speakeasy/gen.yaml"
if [ -e "${OUTDIR}/.speakeasy/gen.lock" ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' "${OUTDIR}/.speakeasy/gen.lock" ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock "${OUTDIR}/.speakeasy/gen.lock"
fi
if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    rm -f "${OUTDIR}/README.md" || true
fi

print_msg_with_ctx "${BUILD_ICON} {green}{bold}Building Terraform Provider{reset}" "variant=review"

run_cmd go run cmd/generate/main.go -s ./tests/specs/review-terraform.yaml -o "${OUTDIR}" -l terraform -t review --validate-integrity
provider_output_file=$(mktemp)
(cd "${OUTDIR}" && go mod tidy)
(cd "${OUTDIR}" && go run main.go --debug > "$provider_output_file" 2>&1) &
provider_pid=$!
cleanup() {
    print_formatted "{grey}Cleaning up...{reset}"
    kill $provider_pid
    rm -f "$provider_output_file"
}
trap cleanup EXIT
max_wait=30
start_time=$(date +%s)
while true; do
    if grep -q "TF_REATTACH_PROVIDERS" "$provider_output_file"; then
        break
    fi

    current_time=$(date +%s)
    if (( current_time - start_time >= max_wait )); then
        cat "$provider_output_file"
        print_formatted "{red}Timeout waiting for provider to start{reset}"
        exit 1
    fi

    sleep 1
done
TF_REATTACH_PROVIDERS=$(grep "TF_REATTACH_PROVIDERS" "$provider_output_file" | sed -n 's/.*TF_REATTACH_PROVIDERS=//p' | sed "s/^'//" | sed "s/'$//" | tr -d '\n')
run_cmd echo "TF_REATTACH_PROVIDERS=$TF_REATTACH_PROVIDERS"
export TF_REATTACH_PROVIDERS
run_terraform_validate() {
    local dir=$1
    echo "Running terraform validate in $dir"
    (cd "$dir" && terraform init && terraform validate)
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
      echo "Terraform validate failed in $dir"
      exit 1
    fi
}
for dir in "${OUTDIR}/examples/resources"/* "${OUTDIR}/examples/data-sources"/*; do
    if [ -d "$dir" ]; then
        run_cmd run_terraform_validate "$dir"
    fi
done

${SCRIPT_DIR}/test-review-terraform.sh
