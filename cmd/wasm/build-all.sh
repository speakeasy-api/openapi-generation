#!/bin/bash
set -eu
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "$0")" &>/dev/null && pwd)"
readonly REPO_ROOT="$(git rev-parse --show-toplevel)"

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Validate Go version higher or equal to than this.
REQUIRED_GO_VERSION="1.25.9"
CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

if ! echo "$CURRENT_GO_VERSION $REQUIRED_GO_VERSION" | awk '{
    split($1, current, ".")
    split($2, required, ".")
    if ((current[1] > required[1]) || \
        (current[1] == required[1] && current[2] > required[2]) || \
        (current[1] == required[1] && current[2] == required[2] && current[3] >= required[3])) {
        exit 0;
    }
    exit 1
}'; then
    echo -e "${RED}Error: Go version must be $REQUIRED_GO_VERSION or higher, but found $CURRENT_GO_VERSION${NC}"
    exit 1
fi

get_file_size() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        BYTES=$(stat -f%z "$1")
    else
        BYTES=$(stat -c%s "$1")
    fi
    MB=$((BYTES / 1024 / 1024))
    echo "$MB MB"
}

# Function to get raw file size in bytes
get_file_size_bytes() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        stat -f%z "$1"
    else
        stat -c%s "$1"
    fi
}

# Function to print colored or plain text
print_message() {
    local color="$1"
    local message="$2"
    if [ "$USE_COLORS" = true ]; then
        echo -e "${color}${message}${NC}"
    else
        echo "$message"
    fi
}

# Function to build a single WASM binary
build_wasm_binary() {
    local binary_name="$1"
    local version="$2"
    local binary_dir="$SCRIPT_DIR/$binary_name"
    
    if [ ! -d "$binary_dir" ]; then
        print_message "$RED" "Directory $binary_dir does not exist"
        return 1
    fi
    
    print_message "$BLUE" "Building $binary_name with version: $version"
    
    # Prepare output directories
    mkdir -p "$SCRIPT_DIR/assets/wasm"
    
    printf "  Propagating dependencies from parent go.mod... "
    if ! "$SCRIPT_DIR/../../scripts/propagate-replace-directives.sh" "$SCRIPT_DIR/../../go.mod" "$binary_dir/go.mod" > /dev/null 2>&1; then
        printf "Failed\n"
        return 1
    fi
    printf "Done\n"
    
    printf "  Installing dependencies... "
    if ! (cd "$binary_dir" && GOWORK=off go mod tidy) > /dev/null 2>&1; then
        printf "Failed\n"
        return 1
    fi
    printf "Done\n"
    
    printf "  Building WASM binary..."
    if ! (cd "$binary_dir" && GOWORK=off GOOS=js GOARCH=wasm go build -trimpath -ldflags="-X 'main.version=$version'" -o "$binary_dir/$binary_name.wasm"); then
        printf "Failed\n"
        return 1
    fi
    SIZE="$(get_file_size "$binary_dir/$binary_name.wasm")"
    printf " Done (%s)\n" "$SIZE"
    
    printf "  Compressing WASM binary (gzip)..."
    gzip -9 -c "$binary_dir/$binary_name.wasm" > "$binary_dir/$binary_name.wasm.gz"
    GZ_SIZE="$(get_file_size "$binary_dir/$binary_name.wasm.gz")"
    printf " Done (%s)\n" "$GZ_SIZE"
    
    printf "  Compressing WASM binary (brotli)..."
    # Check if brotli is installed
    if ! command -v brotli &> /dev/null; then
        echo ""
        echo "  Warning: brotli is not installed. Skipping Brotli compression."
        echo "  Install with: brew install brotli (macOS) or apt-get install brotli (Ubuntu)"
        rm "$binary_dir/$binary_name.wasm"
    else
        brotli --quality=9 -c "$binary_dir/$binary_name.wasm" > "$binary_dir/$binary_name.wasm.br"
        # Keep the .wasm file for the summary table, but copy to assets for deployment
        cp "$binary_dir/$binary_name.wasm" "$SCRIPT_DIR/assets/wasm/$binary_name.wasm"
        BR_SIZE="$(get_file_size "$binary_dir/$binary_name.wasm.br")"
        printf " Done (%s)\n" "$BR_SIZE"
    fi
    
    # Move outputs
    mv "$binary_dir/$binary_name.wasm.gz" "$SCRIPT_DIR/assets/wasm/$binary_name.wasm.gz"
    if [ -f "$binary_dir/$binary_name.wasm.br" ]; then
        mv "$binary_dir/$binary_name.wasm.br" "$SCRIPT_DIR/assets/wasm/$binary_name.wasm.br"
    fi

    # Success message
    print_message "$GREEN" "  $(realpath "$SCRIPT_DIR/assets/wasm/$binary_name.wasm.gz") built successfully"
    if [ -f "$SCRIPT_DIR/assets/wasm/$binary_name.wasm.br" ]; then
        print_message "$GREEN" "  $(realpath "$SCRIPT_DIR/assets/wasm/$binary_name.wasm.br") built successfully"
    fi
    return 0
}

# Parse arguments
WATCH_MODE=false
USE_COLORS=false
VERSION=""
BINARY=""

for arg in "$@"; do
    case $arg in
        --watch)
            WATCH_MODE=true
            USE_COLORS=true
            ;;
        --binary=*)
            BINARY="${arg#*=}"
            ;;
        *)
            if [ -z "$VERSION" ]; then
                VERSION="$arg"
            fi
            ;;
    esac
done

# Set version if not provided
if [ -z "$VERSION" ]; then
    COMMIT_SHA=$(git rev-parse HEAD | cut -c1-7)
    if [ -n "$(git status --porcelain)" ]; then
        DIRTY_SHA=$(git diff | sha1sum | cut -c1-7)
        VERSION="${COMMIT_SHA}-dirty-${DIRTY_SHA}"
    else
        VERSION="${COMMIT_SHA}"
    fi
fi

# Define the binaries to build
if [ -n "${BINARY:-}" ]; then
    # If a specific binary is requested, build only that one
    BINARIES="$BINARY"
else
    # Build all binaries
    BINARIES="overlay ast full-stack"
fi

# Function to format file size for display
format_size() {
    local bytes=$1
    if [ "$bytes" -lt 1048576 ]; then
        # Less than 1MB, show in KB
        echo $((bytes / 1024))"KB"
    else
        # 1MB or more, show in MB
        echo $((bytes / 1048576))"MB"
    fi
}

# Function to calculate percentage change
calc_percentage() {
    local old_size=$1
    local new_size=$2
    if [ "$old_size" -eq 0 ]; then
        echo "N/A"
    else
        local change=$((((new_size - old_size) * 100) / old_size))
        if [ "$change" -gt 0 ]; then
            echo "+${change}%"
        else
            echo "${change}%"
        fi
    fi
}

# Function to print size summary table
print_size_summary() {
    echo ""
    print_message "$BLUE" "📊 Binary Size Summary"
    echo ""
    printf "%-12s %-12s %-12s %-12s %-12s %-10s\n" "Binary" "Uncompressed" "Gzip" "Brotli" "Gzip vs Orig" "Brotli vs Orig"
    printf "%-12s %-12s %-12s %-12s %-12s %-10s\n" "------" "------------" "----" "------" "----------" "------------"
    
    for binary in $BINARIES; do
        local wasm_file="$SCRIPT_DIR/assets/wasm/$binary.wasm"
        local gz_file="$SCRIPT_DIR/assets/wasm/$binary.wasm.gz"
        local br_file="$SCRIPT_DIR/assets/wasm/$binary.wasm.br"
        
        if [ -f "$wasm_file" ] && [ -f "$gz_file" ] && [ -f "$br_file" ]; then
            local wasm_size=$(get_file_size_bytes "$wasm_file")
            local gz_size=$(get_file_size_bytes "$gz_file")
            local br_size=$(get_file_size_bytes "$br_file")
            
            local wasm_display=$(format_size $wasm_size)
            local gz_display=$(format_size $gz_size)
            local br_display=$(format_size $br_size)
            
            local gz_pct=$(calc_percentage $wasm_size $gz_size)
            local br_pct=$(calc_percentage $wasm_size $br_size)
            
            printf "%-12s %-12s %-12s %-12s %-12s %-10s\n" "$binary" "$wasm_display" "$gz_display" "$br_display" "$gz_pct" "$br_pct"
        fi
    done
    echo ""
}

# Function to build all binaries
build_all() {
    local failed_builds=0
    
    print_message "$GREEN" "Building WASM binaries with version: $VERSION"
    
    # Copy wasm_exec.js once
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$SCRIPT_DIR/assets/wasm/wasm_exec.js"
    
    for binary in $BINARIES; do
        if ! build_wasm_binary "$binary" "$VERSION"; then
            failed_builds=$((failed_builds + 1))
        fi
        echo ""
    done
    
    if [ $failed_builds -eq 0 ]; then
        print_message "$GREEN" "✅ All binaries built successfully!"
        print_size_summary
    else
        print_message "$RED" "❌ $failed_builds binaries failed to build"
        return 1
    fi
}

# Check if watch mode is enabled
if [ "$WATCH_MODE" = true ]; then
    print_message "$GREEN" "Starting watch mode. Press Ctrl+C to stop."
    
    # Check if fswatch is installed
    if ! command -v fswatch &> /dev/null; then
        print_message "$RED" "Error: fswatch is not installed. Please install it with: brew install fswatch"
        exit 1
    fi
    
    # Initial build
    build_all
    
    print_message "$GREEN" "Listening for changes to *.go files in $REPO_ROOT"
    # Change to repo root and watch for changes in Go files
    (cd "$REPO_ROOT" && fswatch -r -e ".*" -i "\\.go$" . | while read -r file; do
        print_message "$GREEN" "Change detected in $file"
        build_all
    done)
else
    # Single build
    build_all
fi
