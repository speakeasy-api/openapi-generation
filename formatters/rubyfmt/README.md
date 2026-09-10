# rubyfmt WASM Wrapper

Minimal [WASI](https://wasi.dev/) wrapper around [rubyfmt](https://github.com/fables-tales/rubyfmt) for use as an embedded WASM formatter in the code generation pipeline.

The wrapper reads Ruby source from stdin, formats it with `rubyfmt::format_buffer`, and writes the result to stdout. It is compiled to `wasm32-wasip1` and embedded into the Go binary via `//go:embed`.

## Prerequisites

- **Rust toolchain** with the `wasm32-wasip1` target:
  ```sh
  rustup target add wasm32-wasip1
  ```
- **WASI SDK** (provides the C sysroot needed to compile ruby-prism):
  ```sh
  curl -LO https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-25/wasi-sdk-25.0-x86_64-linux.tar.gz
  sudo tar xf wasi-sdk-25.0-x86_64-linux.tar.gz -C /opt
  sudo ln -sf /opt/wasi-sdk-25.0-x86_64-linux /opt/wasi-sdk
  ```
- **libclang-dev** (for bindgen):
  ```sh
  apt install libclang-dev
  ```

## Building

```sh
./build.sh
```

This compiles the WASM binary and copies it to `internal/format/rubyfmt.wasm`.

The script expects the WASI SDK at `/opt/wasi-sdk` by default. Override with:

```sh
WASI_SDK_PATH=/path/to/wasi-sdk ./build.sh
```

## How it works

The `Cargo.toml` pulls `rubyfmt` as a git dependency from the upstream repo (the library isn't published to crates.io). The Rust code in `src/main.rs` is a ~20 line stdin/stdout bridge. The release profile is tuned for size (`opt-level = "s"`, LTO, strip) producing a ~1.7 MB WASM binary.

At runtime, the Go side (`internal/format/rubyfmt.go`) uses [wazero](https://wazero.io/) to execute the WASM module with WASI support, piping Ruby source through stdin/stdout.
