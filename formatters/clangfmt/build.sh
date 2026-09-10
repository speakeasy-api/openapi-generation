#!/usr/bin/env bash
#
# Build clang-format as a standalone WASM binary (WASI-compatible) using Emscripten.
#
# Uses the same proven approach as wasm-fmt/clang-format: Emscripten's emcmake
# builds LLVM's clang-format library into WASM. We use -s STANDALONE_WASM=1 to
# produce a WASI-compatible binary that runs via wazero (stdin -> format -> stdout).
#
# Prerequisites:
#   - Emscripten SDK installed and activated (source emsdk_env.sh)
#   - cmake (3.24+)
#   - ninja
#   - A host C/C++ compiler (clang or gcc)
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/../../internal/format"
BUILD_DIR="${SCRIPT_DIR}/build"
LLVM_SRC_DIR="${BUILD_DIR}/llvm-src"
EMCC_BUILD_DIR="${BUILD_DIR}/emcc"

LLVM_VERSION="22.1.0-rc3"

# Source emsdk if needed
if ! command -v emcc &>/dev/null; then
    if [ -f "${HOME}/emsdk/emsdk_env.sh" ]; then
        source "${HOME}/emsdk/emsdk_env.sh"
    else
        echo "ERROR: Emscripten (emcc) not found. Install emsdk first." >&2
        exit 1
    fi
fi

echo "=== Building clang-format WASM (standalone/WASI) ==="
echo "  LLVM version: ${LLVM_VERSION}"
echo "  Emscripten: $(emcc --version 2>/dev/null | head -1)"
echo "  Build dir: ${BUILD_DIR}"
echo ""

# ---------- Download LLVM source if needed ----------
if [ ! -d "${LLVM_SRC_DIR}/llvm" ]; then
    LLVM_TARBALL="${BUILD_DIR}/llvm-project-${LLVM_VERSION}.src.tar.xz"
    if [ ! -f "${LLVM_TARBALL}" ]; then
        echo ">>> Downloading LLVM ${LLVM_VERSION} source..."
        mkdir -p "${BUILD_DIR}"
        curl -L -o "${LLVM_TARBALL}" \
            "https://github.com/llvm/llvm-project/releases/download/llvmorg-${LLVM_VERSION}/llvm-project-${LLVM_VERSION}.src.tar.xz"
    fi
    echo ">>> Extracting LLVM source..."
    mkdir -p "${LLVM_SRC_DIR}"
    tar xf "${LLVM_TARBALL}" -C "${LLVM_SRC_DIR}" --strip-components=1
fi

# ---------- Configure with Emscripten ----------
if [ ! -f "${EMCC_BUILD_DIR}/build.ninja" ]; then
    echo ">>> Configuring with emcmake..."
    mkdir -p "${EMCC_BUILD_DIR}"

    # Host compiler for native tablegen tools (emcmake handles cross-compilation)
    export CC=$(which clang 2>/dev/null || which gcc)
    export CXX=$(which clang++ 2>/dev/null || which g++)

    emcmake cmake -G Ninja \
        -S "${LLVM_SRC_DIR}/llvm" \
        -B "${EMCC_BUILD_DIR}" \
        -DCMAKE_BUILD_TYPE=MinSizeRel \
        -DLLVM_TARGETS_TO_BUILD="" \
        -DLLVM_ENABLE_PROJECTS="clang" \
        -DLLVM_ENABLE_BACKTRACES=OFF \
        -DLLVM_ENABLE_CRASH_OVERRIDES=OFF \
        -DLLVM_ENABLE_TERMINFO=OFF \
        -DLLVM_ENABLE_ZLIB=OFF \
        -DLLVM_ENABLE_ZSTD=OFF \
        -DLLVM_ENABLE_THREADS=OFF \
        -DLLVM_ENABLE_PIC=OFF \
        -DLLVM_INCLUDE_BENCHMARKS=OFF \
        -DLLVM_INCLUDE_EXAMPLES=OFF \
        -DLLVM_INCLUDE_TESTS=OFF \
        -DLLVM_INCLUDE_UTILS=OFF \
        -DCLANG_ENABLE_OBJC_REWRITER=OFF \
        -DCLANG_ENABLE_STATIC_ANALYZER=OFF \
        -DCLANG_ENABLE_ARCMT=OFF \
        -DCMAKE_C_FLAGS="-fno-rtti" \
        -DCMAKE_CXX_FLAGS="-fno-rtti"
else
    echo ">>> Using cached emcmake configuration"
fi

# ---------- Build clang-format libraries ----------
echo ""
echo ">>> Building clang-format libraries..."
ninja -C "${EMCC_BUILD_DIR}" clangBasic clangFormat clangRewrite clangToolingCore

echo ""
echo ">>> Linking standalone WASM binary..."

# Include paths
LLVM_INCLUDE="${LLVM_SRC_DIR}/llvm/include"
CLANG_INCLUDE="${LLVM_SRC_DIR}/clang/include"
LLVM_BUILD_INCLUDE="${EMCC_BUILD_DIR}/include"
CLANG_BUILD_INCLUDE="${EMCC_BUILD_DIR}/tools/clang/include"

# Collect all static libraries built by ninja
LIBS=$(find "${EMCC_BUILD_DIR}/lib" -name "*.a" | sort)

# Build the standalone WASM binary.
# -s STANDALONE_WASM=1 produces a WASI-compatible binary (no JS glue needed).
# -s FILESYSTEM=0 avoids the Emscripten FS layer; we only use stdin/stdout.
em++ \
    -Os \
    -fno-exceptions \
    -fno-rtti \
    -I"${LLVM_INCLUDE}" \
    -I"${CLANG_INCLUDE}" \
    -I"${LLVM_BUILD_INCLUDE}" \
    -I"${CLANG_BUILD_INCLUDE}" \
    -s STANDALONE_WASM=1 \
    -s ALLOW_MEMORY_GROWTH=1 \
    -o "${EMCC_BUILD_DIR}/clang-format-wasi.wasm" \
    "${SCRIPT_DIR}/src/main.cc" \
    ${LIBS}

WASM_BIN="${EMCC_BUILD_DIR}/clang-format-wasi.wasm"
WASM_SIZE=$(du -h "${WASM_BIN}" | cut -f1)

echo ""
echo ">>> Output: ${WASM_BIN} (${WASM_SIZE})"
echo ">>> Copying to ${OUTPUT_DIR}/clangfmt.wasm"
cp "${WASM_BIN}" "${OUTPUT_DIR}/clangfmt.wasm"

echo ""
echo "Done!"
