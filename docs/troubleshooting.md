# Troubleshooting

This guide covers common local-development issues. It does not replace the
project's security, support, or incident policies, which will be published with
approved contact and policy text.

## Start with a focused command

Prefer the smallest command that reproduces the problem:

```bash
# Validate an OpenAPI document without generating output.
go run ./cmd/validate/main.go -s path/to/openapi.yaml

# Generate without compiling the target output.
go run ./cmd/generate/main.go -s path/to/openapi.yaml -o /tmp/sdk -l go --skip-compile

# Re-run one target and test variant.
TARGET=primary make test-go

# Type-check one target's templates.
make check-template-go
```

`TARGET` is the historical name for the test variant in these Make commands;
`go` is the generated target.

## Toolchain is missing or the wrong version

Start with [Development setup](../SETUP.md). Install only the language toolchains
required by the target you are testing. Target-specific version requirements are
recorded in `tests/config/<variant>/<target>/.speakeasy/gen.yaml`.

Node.js and npm are needed for template formatting and TypeScript checks.
Docker is needed for SDK test services. `mise run lint` requires `golangci-lint`.

## A generated SDK does not compile

First rerun generation with `--skip-compile` to distinguish generation from the
target compiler or package tooling:

```bash
go run ./cmd/generate/main.go -s path/to/openapi.yaml -o /tmp/sdk -l go --skip-compile
```

Then install the target's toolchain and repeat without `--skip-compile`. If the
issue is target-specific, run the smallest corresponding test command, such as
`TARGET=primary make test-go`, and inspect the generated output in its temporary
test directory.

## SDK tests cannot reach local services

SDK test commands start `httpbin` and `speakeasy-api-test-service` automatically.
If you changed `services/speakeasy-api-test-service/`, rebuild and restart it:

```bash
mise run services:restart
```

The service scripts use their documented default ports and record the selected
ports in `.test-ports`. If a default port is occupied, the scripts fall back to
an ephemeral port. To use different fixed ports, set `HTTPBIN_PORT` and/or
`API_TEST_SERVICE_PORT` before starting the test command. Stop services with:

```bash
mise run services:stop
```

## A template change does not appear in a review SDK

Rebuild the affected target's tracked review SDK:

```bash
TARGET=review make build-go
```

After modifying files under `templates/templates/`, run `npm run format` and
`make check-template-<target>`. When adding, moving, or deleting template files,
also run `go generate ./...` so `templates/perms.go` stays current.

## Lint changes files or reports unrelated changes

`mise run lint` intentionally updates fragment security blocks, generated template
permissions, and formatting before running checks. Inspect the resulting diff
and keep only expected changes. For a narrowly scoped change, run focused tests
first and record any full-suite check you did not run.

## Existing SDK regeneration cannot find its schema or target

Use `regen` against the generated SDK directory:

```bash
go run ./cmd/regen/main.go ./path/to/sdk
```

It infers the target from `.speakeasy/gen.yaml` and the schema from
`workflow.yaml` or an `openapi*` file. Override inference when necessary:

```bash
go run ./cmd/regen/main.go -s path/to/openapi.yaml -l go ./path/to/sdk
```

Use repeated `--set key=value` flags to patch `gen.yaml` before regeneration.

## Need execution details

The generator supports local tracing and profiling:

```bash
go run ./cmd/generate/main.go \
  -s path/to/openapi.yaml \
  -o /tmp/sdk \
  -l go \
  --trace=file:///tmp/generation-trace.json \
  --profile /tmp/generation-profile
```

See [Tracing](tracing.md) for trace destinations. Do not upload traces,
profiles, generated fixtures, or logs that contain credentials, customer data,
or private filesystem paths.
