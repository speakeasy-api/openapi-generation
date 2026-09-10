#!/usr/bin/env bash
#
# Build rubyfmt as a WASM binary for embedding in the Go formatter.
#
# Prerequisites:
#   - Rust toolchain (rustup) with wasm32-wasip1 target:
#       rustup target add wasm32-wasip1
#   - WASI SDK installed at /opt/wasi-sdk (or set WASI_SDK_PATH):
#       curl -LO https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-25/wasi-sdk-25.0-x86_64-linux.tar.gz
#       sudo tar xf wasi-sdk-25.0-x86_64-linux.tar.gz -C /opt
#       sudo ln -sf /opt/wasi-sdk-25.0-x86_64-linux /opt/wasi-sdk
#   - libclang-dev (for bindgen):
#       apt install libclang-dev
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/../../internal/format"

# Default to /opt/wasi-sdk if not explicitly set
WASI_SDK_PATH="${WASI_SDK_PATH:-/opt/wasi-sdk}"

if [ ! -d "${WASI_SDK_PATH}" ]; then
    echo "ERROR: WASI SDK not found at ${WASI_SDK_PATH}" >&2
    echo "" >&2
    echo "Install it with:" >&2
    echo "  curl -LO https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-25/wasi-sdk-25.0-x86_64-linux.tar.gz" >&2
    echo "  sudo tar xf wasi-sdk-25.0-x86_64-linux.tar.gz -C /opt" >&2
    echo "  sudo ln -sf /opt/wasi-sdk-25.0-x86_64-linux /opt/wasi-sdk" >&2
    echo "" >&2
    echo "Or set WASI_SDK_PATH to your installation directory." >&2
    exit 1
fi

# Detect clang include path from WASI SDK
CLANG_VERSION_DIR=$(ls -1 "${WASI_SDK_PATH}/lib/clang/" | head -1)
CLANG_INCLUDE="${WASI_SDK_PATH}/lib/clang/${CLANG_VERSION_DIR}/include"

echo "Building rubyfmt WASM..."
echo "  WASI SDK: ${WASI_SDK_PATH}"
echo "  Clang includes: ${CLANG_INCLUDE}"

cd "${SCRIPT_DIR}"

# Build for wasm32-wasip1 with WASI SDK for C dependencies (ruby-prism)
env CARGO_NET_GIT_FETCH_WITH_CLI=true \
    "BINDGEN_EXTRA_CLANG_ARGS_wasm32-wasip1=--sysroot=${WASI_SDK_PATH}/share/wasi-sysroot -I${CLANG_INCLUDE}" \
    WASI_SDK_PATH="${WASI_SDK_PATH}" \
    cargo build --target wasm32-wasip1 --release

WASM_BIN="target/wasm32-wasip1/release/rubyfmt-wasm.wasm"
WASM_SIZE=$(du -h "${WASM_BIN}" | cut -f1)

echo "Copying ${WASM_BIN} (${WASM_SIZE}) -> ${OUTPUT_DIR}/rubyfmt.wasm"
cp "${WASM_BIN}" "${OUTPUT_DIR}/rubyfmt.wasm"

echo "Done."
