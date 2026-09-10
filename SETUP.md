# Development setup

This guide describes the toolchains needed to work on this repository. Start with the small set needed for the target you are changing; the full SDK test matrix requires more language runtimes.

## Required baseline

### Go

Install Go `1.26.2`, matching the version declared in [`go.mod`](go.mod), and ensure `go` is on your `PATH`.

Install the linter used by `mise run lint`:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

No global `GOPRIVATE` setting or Git `insteadOf` rewrite is part of the public contributor setup.

### Node.js

Install a current Node.js release and npm. They provide Prettier and TypeScript checks for templates:

```bash
npm install
```

### Docker

Install Docker Engine or Docker Desktop and verify the Docker CLI can start containers. SDK test commands start local `httpbin` and `speakeasy-api-test-service` containers automatically.

### Supporting tools

- [`yq`](https://github.com/mikefarah/yq) is used by repository scripts.
- The [Speakeasy CLI](https://github.com/speakeasy-api/speakeasy) may be useful for testing end-user workflows. Repository test-spec assembly uses the local `cmd/overlay` command and does not require CLI authentication.

## First local checks

From the repository root:

```bash
# Build the generator and validator binaries.
mise run build

# Format template files (writes formatting changes).
npm run format

# Run the full lint suite (also writes formatting and permission updates).
mise run lint

# Run focused generator tests.
mise run test:generator

# Generate a small Go review SDK without compiling it.
go run ./cmd/generate/main.go \
  -s ./tests/specs/basic-http.yaml \
  -o /tmp/generated-sdk \
  -l go \
  --skip-compile
```

`mise run lint` is intentionally a mutating check: it synchronizes fragment security blocks, updates template permissions, runs Prettier, and runs `gofmt` before linting. Review `git diff` afterward and commit only expected changes.

## Target-specific toolchains

Install the runtimes only when building or testing the corresponding targets.

| Target                      | Required tooling                                                       |
| --------------------------- | ---------------------------------------------------------------------- |
| Go                          | Go                                                                     |
| TypeScript / MCP TypeScript | Node.js and npm                                                        |
| Python                      | Python 3.9+ and Rust/Cargo for some dependencies                       |
| PHP                         | PHP 8.2+ and Composer                                                  |
| Ruby                        | Ruby 3+ and Bundler                                                    |
| Java                        | OpenJDK 11 and Gradle                                                  |
| C#                          | .NET SDKs required by the target's `.speakeasy/gen.yaml` configuration |
| Unity                       | Unity Hub and an appropriate Unity editor; optional for most changes   |
| Terraform                   | Terraform CLI                                                          |
| Postman                     | No runtime test target is currently defined                            |

The target configuration under `tests/config/<variant>/<target>/.speakeasy/gen.yaml` is the source of truth for target-specific versions. The full test matrix may require older runtimes for compatibility testing; install those only when reproducing that target's test path.

## Test services

SDK test targets start their dependencies automatically:

```bash
TARGET=primary make test-go
```

The service scripts select ephemeral ports by default and save them in `.test-ports`. To use fixed ports, export `HTTPBIN_PORT` and/or `API_TEST_SERVICE_PORT` before running the command.

If you change `services/speakeasy-api-test-service/`, rebuild and restart it before rerunning SDK tests:

```bash
mise run services:restart
```

Use `mise run services:stop` to stop test services when you are done.

## Template development

- Target templates live under `templates/templates/<target>/`.
- Common template code lives under `templates/templates/common/`.
- Use `make check-template-<target>` to type-check a target's template JavaScript/TypeScript.
- If you add, delete, or move template files, run `go generate ./...` to refresh `templates/perms.go`.
- If a template change modifies generated review output, rebuild the affected review SDK with `TARGET=review make build-<target>`.

## Optional development tools

### Dev Containers

A [dev container](.devcontainer/Dockerfile) can install a broad toolchain. It is optional and may lag the current test matrix; report or fix any drift you find.

### Tracing and profiling

Run Jaeger locally with:

```bash
mise run tracing
```

The generator supports `--trace=stderr`, `--trace=grpc`, or `--trace=file://<path>`, plus `--profile <directory>` for profiling output. Avoid uploading trace or profiling artifacts that contain customer document data.

## Telemetry opt-out

To disable generator telemetry for a shell session:

```bash
export SPEAKEASY_DISABLE_TELEMETRY=true
```

See the [telemetry section in the README](README.md#telemetry) for the implementation-based description of collected fields and the current limitations.
