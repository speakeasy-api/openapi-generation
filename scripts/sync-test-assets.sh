#!/usr/bin/env bash

set -e

# This script will sync any extra files that need to exist in a given test SDK
# for a given language. The assets must live under:
# `/tests/config/<language>/<test group>`
#
# Because .go files can trip up build/test/lint in this repository, assets with
# this extension can be named *.go.hidden to hide them from tooling and these
# will be renamed to *.go after syncing in the target test SDK.

if ! grep -q "module github.com/speakeasy-api/openapi-generation/v2" "./go.mod"; then
    echo "This script must be run from the root of the openapi-generation repository."
    exit 1
fi

if [ -z "$1" ]; then
    echo "Error: target (e.g. pythonv2) is required"
    exit 1
fi

if [ -z "$2" ]; then
    echo "Error: test sdk id (e.g. primary) is required"
    exit 1
fi

asset_path="./tests/config/$2/$1"
if [ ! -d "$asset_path" ]; then
    exit 0
fi

if ! command -v rsync &> /dev/null; then
    echo "rsync is not installed. Please install it and try again."
    exit 1
fi

test_sdk_path="./testSDKs/sdk-$1-$2" 
mkdir -p "$test_sdk_path"

rsync -a "$asset_path/" "$test_sdk_path/"
echo 'Renaming *.go.hidden to *.go'
find "$test_sdk_path" -type f -name "*.go.hidden" -exec sh -c 'mv "$0" "${0%.hidden}"' {} \;
echo "Synced $asset_path into $test_sdk_path"