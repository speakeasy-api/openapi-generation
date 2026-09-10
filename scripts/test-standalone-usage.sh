#!/usr/bin/env bash

# Test script for standalone usage generation ('generateusage' command)
#
# Usage: ./test-standalone-usage.sh [target]
#   target: Optional. Specific target to test (e.g. csharp, go, javav2,
#           pythonv2, etc.). If not provided, tests all targets.

target_arg=$1
spec=tests/specs/review.yaml
group=tag1

# Expected operation to be generated based on the 'x-speakeasy-usage-example' extension
operation=listTest1

if [ -n "$target_arg" ]; then
    targets=("${target_arg}")
else
    targets=("csharp" "go" "javav2" "php" "pythonv2" "ruby" "typescriptv2")
fi

tmp_out=$(mktemp -d)
trap 'rm -rf "$tmp_out"' EXIT

for target in "${targets[@]}"; do
    lang=${target%v2}
    echo "Generating Standalone Usage for $target..."
    out="$tmp_out/$target"
    mkdir -p "$out/.speakeasy"
    gen_yaml="tests/config/review/$target/.speakeasy/gen.yaml"
    if [ -f "$gen_yaml" ]; then
        cp "$gen_yaml" "$out/.speakeasy/gen.yaml"
    else
        echo -e "WARN\tMissing review gen.yaml for $target ($gen_yaml); falling back to default config"
    fi
    usage1=$(go run cmd/generateusage/main.go -g $group -s $spec -l $lang -o "$out")

    if [ "${SPEAKEASY_DEBUG}" == "true" ]; then
        echo "$usage1"
    fi

    if ! echo "$usage1" | grep -q "Usage snippet provided for $operation"; then
        echo -e "ERROR\tStandalone usage generation failed!\n$usage1"
        exit 1
    fi
done;
