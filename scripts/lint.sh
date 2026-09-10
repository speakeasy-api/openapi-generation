#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR="$SCRIPT_DIR/.."

echo "===================================="
echo "Running lint checks..."
echo "===================================="

# Prevent private snapshot/customer regression machinery from returning.
"$SCRIPT_DIR/check-no-private-snapshot-machinery.sh"

# Check dependencies
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint is not installed. Install via..."
    echo "$ go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo "node is not installed. Please install Node.js"
    exit 1
fi

# 0. Sync fragment security blocks
echo ""
echo "===================================="
echo "0. Update fragment security..."
echo "===================================="
"$SCRIPT_DIR/sync-fragment-security.sh"

# 1. Update permissions (from update-permissions.sh)
echo ""
echo "===================================="
echo "1. Updating permissions..."
echo "===================================="
"$SCRIPT_DIR/update-permissions.sh"

# 2. Install npm dependencies
echo ""
echo "===================================="
echo "2. Installing npm dependencies..."
echo "===================================="
(cd "$ROOT_DIR" && npm i)

# 3. Run prettier on templates
echo ""
echo "===================================="
echo "3. Running prettier format..."
echo "===================================="
(cd "$ROOT_DIR" && npm run format)

# 4. Run go fmt
echo ""
echo "===================================="
echo "4. Running go fmt..."
echo "===================================="
(cd "$ROOT_DIR" && find . -name "*.go" \
    -not -path "./zSDKs/*" \
    -not -path "./testSDKs/*" \
    -not -path "./internal/features/features_generated.go" \
    -not -path "./internal/features/tests_generated.go" \
    -not -path "./internal/sanitization/lookup_table.go" \
    -exec gofmt -l -w {} +)

# 5. Run golangci-lint
echo ""
echo "===================================="
echo "5. Running golangci-lint..."
echo "===================================="
(cd "$ROOT_DIR" && golangci-lint run --timeout=10m)

# 6. go mod tidy -diff
echo ""
echo "===================================="
echo "6. Running go mod tidy -diff..."
echo "===================================="
for mod_dir in "$ROOT_DIR" "$ROOT_DIR/cmd/wasm"; do
    echo "go mod tidy -diff in $mod_dir"
    (cd "$mod_dir" && go mod tidy -diff)
done

# 7. Run tsgo on all templates
echo ""
echo "===================================="
echo "7. Running tsgo on all templates..."
echo "===================================="
TEMPLATES=(cli csharp go javav2 mcp-typescript mockserver php postman pythonv2 ruby terraform typescriptv2 unity)
for template in "${TEMPLATES[@]}"; do
    echo "Checking TypeScript in templates/$template..."
    if ! npx tsgo --project "$ROOT_DIR/templates/templates/$template"; then
        echo "🚨 TypeScript check failed for $template"
        exit 1
    fi
done

echo ""
echo "===================================="
echo "✅ All lint checks passed!"
echo "===================================="
