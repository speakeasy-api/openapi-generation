# Formatter Package

This package formats generated source code before it is written to disk. Each language target dispatches to a language-specific formatter. Languages without a formatter pass through unchanged.

## Dispatch

`Format(target, fileName, data)` in `format.go` routes by target and file extension:

| Target                         | Extensions                      | Formatter                                                                                                              | Runtime            |
| ------------------------------ | ------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------ |
| `csharp`, `unity`              | `.cs`                           | [clang-format](https://clang.llvm.org/docs/ClangFormat.html)                                                           | wazero (WASM/WASI) |
| `typescript`, `mcp-typescript` | `.ts`, `.js`                    | [dprint-plugin-typescript](https://github.com/dprint/dprint-plugin-typescript)                                         | wazero (WASM)      |
| `python`                       | `.py`                           | [dprint-plugin-ruff](https://github.com/dprint/dprint-plugin-ruff)                                                     | wazero (WASM)      |
| `java`, `javav2`               | `.java`                         | [dprint-plugin-java](https://github.com/speakeasy-api/dprint-plugin-java)                                              | wazero (WASM)      |
| `ruby`                         | `.rb`, `.rbi`                   | [rubyfmt](https://github.com/fables-tales/rubyfmt)                                                                     | wazero (WASM/WASI) |
| `php`                          | `.php`                          | [dprint-plugin-mago](https://github.com/dprint/dprint-plugin-mago) ([mago](https://github.com/carthage-software/mago)) | wazero (WASM)      |
| `go`, `cli`, `mockserver`      | `.go`                           | `go/format` (stdlib)                                                                                                   | native             |
| `terraform`                    | `.tf`, `.tfvars`, `.tftest.hcl` | [hclwrite](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclwrite)                                                    | native             |

All other targets/extensions return data unchanged.

## Embedded Binaries

The formatters that run via WASM have their binaries embedded into the Go binary at compile time using `//go:embed`:

| File                            | Source                         | How to rebuild                         |
| ------------------------------- | ------------------------------ | -------------------------------------- |
| `clangfmt.wasm`                 | Built from LLVM source         | `cd formatters/clangfmt && ./build.sh` |
| `dprint-plugin-typescript.wasm` | Pre-built from dprint releases | Download from dprint plugin registry   |
| `dprint-plugin-ruff.wasm`       | Pre-built from dprint releases | Download from dprint plugin registry   |
| `dprint-plugin-java.wasm`       | Pre-built from dprint releases | Download from dprint plugin registry   |
| `dprint-plugin-mago.wasm`       | Pre-built from dprint releases | Download from dprint plugin registry   |
| `rubyfmt.wasm`                  | Built from Rust source         | `cd formatters/rubyfmt && ./build.sh`  |

## Formatter Runtimes

### wazero (WASM)

[wazero](https://wazero.io/) is a zero-dependency WebAssembly runtime for Go.

**dprint** (`dprint.go`): The TypeScript, Python, Java, and PHP formatters use the [dprint plugin ABI](https://dprint.dev/plugin-dev/). A single WASM module is instantiated per language at startup and reused for all formatting calls. Communication happens through shared memory buffers and exported functions (`format`, `set_file_path`, `get_formatted_text`, etc.).

**rubyfmt** (`rubyfmt.go`): The Ruby formatter uses the simpler WASI stdin/stdout model. The WASM module is compiled once at startup, then a fresh instance is created per format call (each needs a unique module name). The Rust WASI binary imports `env.__wasi_init_tp` which is stubbed as a no-op. WASI `proc_exit(0)` is treated as success.

### Native

Go files use the standard library `go/format.Source()`. Terraform files use the `hclwrite` package from HashiCorp (inlined MPL-2.0 code from the Terraform CLI).

## WASM Build Tag

Files that depend on wazero are gated with `//go:build !js || !wasm`. When compiling to WASM itself (e.g. for browser use), `dprint_wasm.go` provides stub implementations that return data unchanged.

## Build Projects

Formatter build tooling lives in the `formatters/` directory at the repo root:

- `formatters/rubyfmt/` — Rust/Cargo project for compiling rubyfmt to WASI WASM
