#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

which yq > /dev/null || (echo "Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq"; exit 1)

# check $1 length > 0
if [ ${#1} -eq 0 ]; then
    echo "Error: target is required"
    exit 1
fi

readonly TARGET="$1"
readonly TARGET_DIR="./zSDKs/${TARGET}"
ensure_license_election

mkdir -p "${TARGET_DIR}/.speakeasy"
cp ./tests/config/review/$TARGET/.speakeasy/gen.yaml "${TARGET_DIR}/.speakeasy/gen.yaml"
cp -r ./tests/tests/review/* "${TARGET_DIR}/.speakeasy/"
if [ -e "${TARGET_DIR}/.speakeasy/gen.lock" ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' "${TARGET_DIR}/.speakeasy/gen.lock" ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock "${TARGET_DIR}/.speakeasy/gen.lock"
fi
if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    rm -f "${TARGET_DIR}/README.md" || true
fi

# Combine existing extra args with additional args from command line
EXTRA_ARGS="${EXTRA_ARGS:-} ${*:2}"

print_msg_with_ctx "${BUILD_ICON} {green}{bold}Building SDK{reset}" "target=${TARGET}|variant=review|args=${EXTRA_ARGS}"

run_cmd go run cmd/generate/main.go -s ./tests/specs/review.yaml -o "${TARGET_DIR}" -l $TARGET --validate-integrity ${EXTRA_ARGS}
