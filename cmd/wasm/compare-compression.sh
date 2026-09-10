#!/bin/bash

# Script to compare Brotli vs gzip compression for WASM binaries
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# ANSI color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

get_file_size_bytes() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        stat -f%z "$1"
    else
        stat -c%s "$1"
    fi
}

format_size() {
    local bytes=$1
    local mb=$((bytes / 1024 / 1024))
    local kb=$((bytes / 1024))
    if [ $bytes -gt 1048576 ]; then
        echo "${mb} MB (${bytes} bytes)"
    elif [ $bytes -gt 1024 ]; then
        echo "${kb} KB (${bytes} bytes)"
    else
        echo "${bytes} bytes"
    fi
}

compare_compression() {
    local input_file="$1"

    if [ ! -f "$input_file" ]; then
        echo "Error: Input file $input_file does not exist"
        return 1
    fi

    echo -e "${BLUE}Compression Comparison for: $(basename "$input_file")${NC}"
    echo "================================================"

    # Original file size
    local original_size=$(get_file_size_bytes "$input_file")
    echo -e "Original size: ${YELLOW}$(format_size $original_size)${NC}"
    echo ""

    # Test gzip compression (levels 1-9)
    echo -e "${GREEN}Gzip Compression:${NC}"
    for level in {1..9}; do
        local gz_file="${input_file}.gz${level}"
        gzip -${level} -c "$input_file" > "$gz_file"
        local gz_size=$(get_file_size_bytes "$gz_file")
        local gz_ratio=$(echo "scale=2; $gz_size * 100 / $original_size" | bc)
        printf "  Level %d: %s (%.2f%% of original)\n" "$level" "$(format_size $gz_size)" "$gz_ratio"
        rm "$gz_file"
    done
    echo ""

    # Test brotli compression (levels 0-11)
    echo -e "${GREEN}Brotli Compression:${NC}"

    # Check if brotli is installed
    if ! command -v brotli &> /dev/null; then
        echo "  Error: brotli is not installed. Install with:"
        echo "    macOS: brew install brotli"
        echo "    Ubuntu/Debian: apt-get install brotli"
        echo "    CentOS/RHEL: yum install brotli"
        return 1
    fi

    for level in {0..9}; do
        local br_file="${input_file}.br${level}"
        brotli --quality=$level -c "$input_file" > "$br_file"
        local br_size=$(get_file_size_bytes "$br_file")
        local br_ratio=$(echo "scale=2; $br_size * 100 / $original_size" | bc)
        printf "  Level %d: %s (%.2f%% of original)\n" "$level" "$(format_size $br_size)" "$br_ratio"
        rm "$br_file"
    done
    echo ""
}

# Main execution
if [ $# -eq 0 ]; then
    # If no arguments, try to find the WASM file
    if [ -f "$SCRIPT_DIR/main.wasm" ]; then
        compare_compression "$SCRIPT_DIR/main.wasm"
    elif [ -f "$SCRIPT_DIR/assets/wasm/lib.wasm.gz" ]; then
        echo "Found existing compressed file. Decompressing first..."
        gunzip -c "$SCRIPT_DIR/assets/wasm/lib.wasm.gz" > "$SCRIPT_DIR/temp.wasm"
        compare_compression "$SCRIPT_DIR/temp.wasm"
        rm "$SCRIPT_DIR/temp.wasm"
    else
        echo "No WASM file found. Please build first with:"
        echo "  ./build.sh"
        echo "Or specify a file:"
        echo "  $0 <path-to-wasm-file>"
        exit 1
    fi
else
    compare_compression "$1"
fi