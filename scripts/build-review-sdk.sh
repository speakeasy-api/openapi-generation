#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

# Use strict RuboCop config for internal builds (see build-target.sh)
export RUBOCOP_OPTS="--config .rubocop-strict.yml"

which yq > /dev/null || (echo "Error: yq is not installed. Please install v4 or above - https://github.com/mikefarah/yq"; exit 1)

ensure_license_election

# check $1 length > 0
if [ ${#1} -eq 0 ]; then
    echo "Error: target is required"
    exit 1
fi

readonly TARGET="$1"

mkdir -p ./zSDKs/sdk-$TARGET
mkdir -p ./zSDKs/sdk-$TARGET/.speakeasy
# Copy gen.yaml
cp ./tests/config/review/$TARGET/.speakeasy/gen.yaml ./zSDKs/sdk-$TARGET/.speakeasy/gen.yaml
# Copy testfiles directory
mkdir -p ./zSDKs/sdk-$TARGET/.speakeasy/testfiles
cp ./tests/tests/review/testfiles/postfiletest.txt ./zSDKs/sdk-$TARGET/.speakeasy/testfiles/postfiletest.txt

tmp_tests="$(mktemp "./zSDKs/sdk-$TARGET/.speakeasy/tests.arazzo.yaml.XXXXXX.tmp")"
cleanup() {
    rm -f "$tmp_tests"
}
trap cleanup EXIT

# Merge tests.arazzo.yaml - source (tests/tests/review) overrides destination (zSDKs)
# but destination-only workflows are preserved
if [ -e ./zSDKs/sdk-$TARGET/.speakeasy/tests.arazzo.yaml ]; then
    yq eval-all '
      # Collect all documents into an array to avoid duplicate output
      [.] | .[0] as $dest | .[1] as $src |
      # Workflows from dest that do not exist in src (by workflowId)
      ($dest.workflows | map(select(.workflowId as $id | ($src.workflows | map(.workflowId) | contains([$id])) | not))) as $unique_dest |
      # For each src workflow, merge with matching dest workflow (src wins)
      ($src.workflows | map(. as $sw |
        ($dest.workflows | map(select(.workflowId == $sw.workflowId)) | .[0]) as $dw |
        ($dw // {}) * $sw
      )) as $merged |
      # Merge top-level keys (src overrides), then set workflows to combined list
      $dest * $src * {"workflows": ($merged + $unique_dest)}
    ' ./zSDKs/sdk-$TARGET/.speakeasy/tests.arazzo.yaml ./tests/tests/review/tests.arazzo.yaml > "$tmp_tests" \
    && mv "$tmp_tests" ./zSDKs/sdk-$TARGET/.speakeasy/tests.arazzo.yaml
else
    cp ./tests/tests/review/tests.arazzo.yaml ./zSDKs/sdk-$TARGET/.speakeasy/tests.arazzo.yaml
fi

# Copy gen.lock, merging if it already exists
if [ -e ./zSDKs/sdk-$TARGET/.speakeasy/gen.lock ]; then
    yq eval-all '. as $item ireduce ({}; . * $item )' ./zSDKs/sdk-$TARGET/.speakeasy/gen.lock ./tests/config/review/gen.lock -i
else
    cp ./tests/config/review/gen.lock ./zSDKs/sdk-$TARGET/.speakeasy/gen.lock
fi

# Copy lint.yaml
cp ./tests/config/review/lint.yaml ./zSDKs/sdk-$TARGET/.speakeasy/lint.yaml

if [ "${GITHUB_ACTIONS:-}" != "true" ]; then
    rm -f ./zSDKs/sdk-$TARGET/README.md || true
fi

rm -rf "./testprojects/$TARGET" || true
# Combine existing extra args with additional args from command line
EXTRA_ARGS="${EXTRA_ARGS:-} ${*:2}"

print_msg_with_ctx "${BUILD_ICON} {green}{bold}Building SDK{reset}" "target=${TARGET}|variant=review|args=${EXTRA_ARGS}"
run_cmd go run cmd/generate/main.go -s ./tests/specs/review.yaml -o ./zSDKs/sdk-$TARGET -l $TARGET --validate-integrity ${EXTRA_ARGS}

# Copy gen.yaml back to source folder after successful build
cp ./zSDKs/sdk-$TARGET/.speakeasy/gen.yaml ./tests/config/review/$TARGET/.speakeasy/gen.yaml
