# WASM

This package contains various utilities that are compiled to WASM for use in the web browser. The functionality has been split into multiple granular WASM binaries to reduce memory usage in the browser.

## Binaries

### 1. Overlay (`overlay/`)
**Smallest binary** - Contains only overlay functionality with minimal dependencies.
- `CalculateOverlay` - Calculates overlay differences between OpenAPI specs
- `ApplyOverlay` - Applies overlays to OpenAPI specs
- **Dependencies**: jsonpath, openapi-overlay, yaml

### 2. AST (`ast/`)
**Medium binary** - Contains AST serialization functionality.
- `SerializeSandboxAST` - Serializes a simplified AST for an OpenAPI specification
- **Dependencies**: libopenapi, fastAST

### 3. Full stack (`full-stack/`)
**Large binary** - Contains full generation functionality for usage snippets + docs data
- `GenerateUsageSnippets` - Generates usage snippets for a given spec
- `SerializeDocsData` - depends on the full generator AST
- **Dependencies**: Full generation stack

## Building

### Build all binaries:
```bash
pnpm build
# or
./build-all.sh
```

### Build individual binaries:
```bash
pnpm build:overlay
pnpm build:ast
pnpm build:usage-snippets
pnpm build:docsdata
pnpm build:legacy
```

### Build with version:
```bash
./build-all.sh v1.2.3
./build-all.sh --binary=overlay v1.2.3
```

### Watch mode:
```bash
./build-all.sh --watch
```

## Usage

When you load the resulting WASM binary in JavaScript using `WebAssembly.instantiate`, by default Golang's WASM implementation (e.g `wasm_exec.js`) will automatically inject any loaded "entrypoints" into the global scope. So therefore, `SerializeSandboxAST` will be a globally defined method, for example.

Each binary includes a `Healthcheck` function for verifying the binary is loaded correctly.

### Testing

To run all tests, you can run:

```
pnpm test
```

`Vitest` is setup in the `cmd/wasm` directory. `pnpm test` will automatically compile the WASM binary prior to the tests running. `watch` mode and `parallelism` are disabled due to WASM's single threaded nature.

There is the test helper `getWasmFunction` that takes a function name and will initialise the WASM binary for your tests and give you a reference to the method back:

```typescript
describe("foo entrypoint", () => {
  let Foo: (...args: any[]) => Promise<string>;

  beforeAll(async () => {
    // "Foo" corresponds to the name of the entrypoint
    // in the main.go file. e.g:
    // js.Global().Set("Foo", promisify(func(args []js.Value) (string, error) {
    // // Function body!
    // })
    Foo = await getWasmFunction("Foo");
  });

  it("does foo", async () => {
      const bar = await Foo();
      expect(bar).toEqual(1);
  })
})
```

### QoL scripts

There are two scripts in this directory to make life a little easier if you're not using the Vitest test harness directly:

1. **build.sh** - Builds the WASM binary
   - Validates Go version compatibility (requires Go 1.25.3 or higher)
   - Sets a version identifier based on git commit info (or accepts a version argument)
   - Compiles the Go code to WebAssembly
   - Compresses the WASM binary with gzip
   - Copies the WASM binary and required JavaScript files to the `assets` directory
   - Pass `--watch` to run the script in watch mode.

   Usage:
   ```bash
   ./build.sh [version] [--watch]
   ```

2. **publish.sh** - Publishes the WASM binary to Google Cloud Storage:
   - Runs the build script first, optionally passing a version
   - Copies the WASM binary and JavaScript files to a GCS bucket
   - Sets appropriate cache control metadata
   - Makes the files publicly accessible

   Usage:
   ```bash
   ./publish.sh [version]
   ```

The three files output are:

1. `engine.latest.wasm.gz`: a GZip-compressed built generator WASM binary
2. `engine.latest.wasm.br`: a Brotli-compressed built generator WASM binary
3. `wasm_exec.js`: a matching javascript file from the Go runtime that configures the go runtime for wasm (e.g. configures polyfills for fs etc.)
