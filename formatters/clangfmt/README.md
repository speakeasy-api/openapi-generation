# clangfmt WASM Wrapper

Minimal [WASI](https://wasi.dev/) wrapper around [clang-format](https://clang.llvm.org/docs/ClangFormat.html) for use as an embedded WASM formatter in the code generation pipeline.

The wrapper reads C# source from stdin, formats it with clang-format's Microsoft style, and writes the result to stdout. It is compiled to standalone WASM via [Emscripten](https://emscripten.org/) and embedded into the Go binary via `//go:embed`.

## Prerequisites

- **Emscripten SDK** (emsdk) installed and activated:
  ```sh
  git clone https://github.com/emscripten-core/emsdk.git ~/emsdk
  cd ~/emsdk && ./emsdk install 3.1.51 && ./emsdk activate 3.1.51
  source ~/emsdk/emsdk_env.sh
  ```
- **CMake** (3.24+) and **Ninja**:
  ```sh
  apt install cmake ninja-build
  ```
- **Host C/C++ compiler** (clang or gcc) for building native tablegen tools.

## Building

```sh
./build.sh
```

This downloads the LLVM 22.1.0-rc3 source (if not cached), builds the required clang-format libraries via Emscripten, links a standalone WASM binary, and copies it to `internal/format/clangfmt.wasm`.

The build is cached — subsequent runs skip download and cmake configuration.

## How it works

The C++ source in `src/main.cc` is a ~50 line stdin/stdout bridge that calls the clang-format library API (`clang::format::reformat`). The build uses Emscripten's `-s STANDALONE_WASM=1` flag to produce a WASI-compatible binary without JavaScript glue.

Key build flags:
- `-s STANDALONE_WASM=1` — WASI-compatible binary (no JS runtime needed)
- `-s FILESYSTEM=0` — no Emscripten FS layer (stdin/stdout only)
- `-s ALLOW_MEMORY_GROWTH=1` — dynamic memory for large files
- `-Os -fno-exceptions -fno-rtti` — size optimizations (~2.3 MB output)

At runtime, the Go side (`internal/format/clangfmt.go`) uses [wazero](https://wazero.io/) to execute the WASM module with WASI support, piping C# source through stdin/stdout. Emscripten's `env` module imports are stubbed as no-ops since only stdio is used.

## Style

The formatter uses clang-format's **Microsoft** style as the base with overrides to match `dotnet format` conventions:

```yaml
{BasedOnStyle: Microsoft, NamespaceIndentation: All, ColumnLimit: 0, AllowShortIfStatementsOnASingleLine: WithoutElse, IndentCaseLabels: true, BreakBeforeCloseBracketFunction: true}
```

- **BasedOnStyle: Microsoft** — Allman braces, 4-space indentation, idiomatic C# spacing
- **NamespaceIndentation: All** — indent contents inside `namespace {}` (Microsoft default is `None`, but `dotnet format` indents)
- **ColumnLimit: 0** — don't reflow/wrap lines (`dotnet format` has no line length limit)
- **AllowShortIfStatementsOnASingleLine: WithoutElse** — preserve single-line `if` statements (matches `dotnet format`)
- **IndentCaseLabels: true** — indent `case` labels inside `switch` (matches `dotnet format`)
- **BreakBeforeCloseBracketFunction: true** — closing `)` on its own line for multi-line params (LLVM 22+, matches `dotnet format`)

The style string is passed as `argv[1]` from the Go side, and the filename hint as `argv[2]`.

### Formatting behavior

Compared to `dotnet format` output, clang-format produces **identical** results for most code. The remaining known divergences are documented as skipped tests in `internal/format/clangfmt_test.go`:

| Divergence | Description |
| --- | --- |
| Empty `{}` expanded | `{}` becomes `{\n}` on separate lines (Allman style) |
| Enum attribute joining | `[Attr]\nValue,` becomes `[Attr] Value,` on one line |
| Space before indexer | `)[0]` becomes `) [0]` (C# parser limitation) |

### What matches dotnet format

| Feature | Description |
| --- | --- |
| Allman braces | K&R `if (...) {` converted to Allman `if (...)\n{` |
| Keyword spacing | `foreach(` to `foreach (`, `if(` to `if (` |
| Extra space removal | `var  x` to `var x` |
| Generic type spacing | `Dictionary <string>` to `Dictionary<string>` |
| Namespace indentation | Contents inside `namespace {}` properly indented |
| Single-line if preserved | `if (x == null) throw ...;` stays on one line |
| Switch case indentation | `case` labels indented inside `switch` blocks |
| Closing paren placement | `)` on its own line for multi-line params (LLVM 22+) |
| Constructor initializer | `) : base(message)` stays on one line |
| Trailing whitespace | Removed from all lines |
