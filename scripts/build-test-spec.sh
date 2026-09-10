#!/usr/bin/env bash

# build-test-spec.sh
# Simple wrapper around the create_test_spec function from build-target.sh
# to create a test spec for a single target/variant combination
#
# Usage:
#   ./build-test-spec.sh <target> <variant>
#
# Arguments:
#   target    - Target language/type (e.g. go, python, csharp, typescript)
#   variant   - Test variant (e.g. primary, secondary, tertiary)
#
# Examples:
#   ./build-test-spec.sh python primary
#   ./build-test-spec.sh csharp secondary

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/build-target.sh"

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <target> <variant>"
    echo "Example: $0 python primary"
    exit 1
fi

create_test_spec "$1" "$2"
