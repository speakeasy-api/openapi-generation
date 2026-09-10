#!/bin/bash
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# Optional version argument can be passed to this script
VERSION=$1

# Run build script first
echo "Running build script..."
if [ -n "$VERSION" ]; then
  # If a version was provided, pass it to build.sh
  ${SCRIPT_DIR}/build.sh "$VERSION"
else
  # Otherwise let build.sh use the commit SHA
  ${SCRIPT_DIR}/build.sh
fi

# Check if build succeeded
if [ $? -eq 0 ]; then
  echo "Build completed successfully, uploading files to GCS..."
  
  # Create temp directory
  mkdir -p /tmp/wasm-release
  
  # Copy files to temp directory
  cp ${SCRIPT_DIR}/assets/wasm/lib.wasm.gz /tmp/wasm-release/engine.latest.wasm.gz
  cp ${SCRIPT_DIR}/assets/wasm/lib.wasm.br /tmp/wasm-release/engine.latest.wasm.br
  cp ${SCRIPT_DIR}/assets/wasm/wasm_exec.js /tmp/wasm-release/wasm_exec.js
  
  # Upload to GCS bucket
  gsutil cp /tmp/wasm-release/engine.latest.wasm.gz gs://speakeasy-static-backend-bucket/engine/
  gsutil cp /tmp/wasm-release/engine.latest.wasm.br gs://speakeasy-static-backend-bucket/engine/
  gsutil cp /tmp/wasm-release/wasm_exec.js gs://speakeasy-static-backend-bucket/engine/
  
  # Set Cache-Control metadata
  gsutil setmeta -h "Cache-Control:public, max-age=300" gs://speakeasy-static-backend-bucket/engine/engine.latest.wasm.gz
  gsutil setmeta -h "Cache-Control:public, max-age=300" gs://speakeasy-static-backend-bucket/engine/engine.latest.wasm.br
  gsutil setmeta -h "Cache-Control:public, max-age=300" gs://speakeasy-static-backend-bucket/engine/wasm_exec.js
  
  echo "Upload completed successfully"
else
  echo "Build failed, upload aborted"
  exit 1
fi
