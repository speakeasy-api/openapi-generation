#!/usr/bin/env bash
set -euo pipefail
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

OUTDIR="./zSDKs/terraform-provider-testing"



if [ ! -d "${OUTDIR}/internal/provider" ]; then
    print_formatted "{red}Directory ${OUTDIR}/internal/provider does not exist. The directory should exist{reset}"
    exit 1
fi
print_formatted "{blue}Running tests in ${OUTDIR}/internal/provider{reset}"
(cd "${OUTDIR}/internal/provider" && TF_ACC=1 go test -v ./...) || exit_code=$?
if [ ${exit_code:-0} -ne 0 ]; then
    print_formatted "{red}Tests failed{reset}"
    print_formatted "{red}To Debug a particular test in the provider you can run similar command -> TF_LOG=DEBUG TF_ACC=1 go test -C ./zSDKs/terraform-provider-testing/internal/provider  -run TestImportMultipleIDAcronymResourceImportValidation{reset}"
    exit 1
fi
print_formatted "{green}Provider tests passed{reset}"



print_formatted "{blue}Running tests in ./internal/tfmockserver{reset}"
if [ ! -d "${OUTDIR}/internal/tfmockserver" ]; then
    print_formatted "{red}Directory ${OUTDIR}/internal/tfmockserver does not exist. The directory should exist{reset}"
    exit 1
fi
print_formatted "{blue}Running tests in ${OUTDIR}/internal/tfmockserver{reset}"
(cd "${OUTDIR}/internal/tfmockserver" && TF_ACC=1 go test -v  ./...) || exit_code=$?
if [ ${exit_code:-0} -ne 0 ]; then
    print_formatted "{red}Tests failed{reset}"
    exit 1
fi
print_formatted "{green}Tfmockserver tests passed{reset}"

print_formatted "{green}All tests passed{reset}"
