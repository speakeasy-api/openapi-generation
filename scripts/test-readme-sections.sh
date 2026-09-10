#!/usr/bin/env bash

# This script tests README section insertion, removal and editing.
# For each {"$i-spec.yaml, "$i-input.md", "$i-output.md"} triplet in the $outDir directory:
# 1. $i-input.md is copied to .README.md (.gitignored)
# 2. Standalone README generation is ran using "$i-spec.yaml" and README.md as input
# 3. The updated README.md file is compared to the expected "$i-output.md"
# 4. If files differ, the test fails and diffs are printed.

# Usage: ./testreadme.sh [lang]
# lang: language to use for standalone README generation (default: go)
#
# Warnings:
# - inputs/outputs are currently configured for Go, and may not work for all languages.

lang=${1:-"go"}
outDir="testreadme/$lang"
fileName=".README.md"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
TEST_DIR="$SCRIPT_DIR/../$outDir"

if [[ ! -d $TEST_DIR ]]; then
    echo -e "ERROR\tREADME section tests are not yet implemented in $outDir"
    exit 1
fi

TEST_FILE="$TEST_DIR/$fileName"

for input in "$TEST_DIR"/*-input.md; do
    inputName=$(basename -- "$input")
    id="${inputName%-input.md}"
    expected="$TEST_DIR/$id-output.md"
    spec="$TEST_DIR/$id-spec.yaml"

    cp $input $TEST_FILE
    go run cmd/generatereadme/main.go -s $spec -o $outDir -l $lang -f $fileName --headers-only

    diffs=$(diff -u $TEST_FILE $expected)
    if [[ -n "$diffs" ]]; then
        echo -e "ERROR\tUnexpected changes found in $TEST_FILE"
        echo "$diffs"
        exit 1
    fi
done
