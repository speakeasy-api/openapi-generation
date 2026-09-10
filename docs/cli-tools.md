# Command reference

These commands are intended for generator and template development. Run them
from the repository root with `go run ./cmd/<command> ...`. The
[Speakeasy CLI](https://github.com/speakeasy-api/speakeasy) is the normal entry
point for end-user SDK generation.

## Common commands

| Command           | Use it to                                                                        | Example                                                                                          |
| ----------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `generate`        | Generate a new SDK or other target from an OpenAPI document.                     | `go run ./cmd/generate -s spec.yaml -o /tmp/sdk -l go --skip-compile`                            |
| `regen`           | Regenerate an already bootstrapped SDK, inferring the target and schema.         | `go run ./cmd/regen ./path/to/sdk`                                                               |
| `validate`        | Validate an OpenAPI document with the configured ruleset.                        | `go run ./cmd/validate -s spec.yaml`                                                             |
| `overlay`         | Join OpenAPI documents and apply one or more OpenAPI Overlay documents.          | `go run ./cmd/overlay -s base.yaml -join fragment.yaml -overlay overlay.yaml -out combined.yaml` |
| `targettest`      | Run target-specific tests against an existing generated output directory.        | `go run ./cmd/targettest -o /tmp/sdk -t go`                                                      |
| `changelog`       | Create a validated feature or generator changeset for a generated-output change. | `go run ./cmd/changelog pagination go "fix: preserve pagination parameters"`                     |
| `changeset/apply` | Apply pending changesets locally; `--dry-run` reports the planned changes.       | `go run ./cmd/changeset/apply --dry-run`                                                         |
| `check-upgrade`   | Compare an existing SDK's `gen.yaml` with current new-SDK defaults.              | `go run ./cmd/check-upgrade -repo ./path/to/sdk`                                                 |
| `security`        | List target template dependencies or query OSV and npm audit findings.           | `go run ./cmd/security -format summary -target go`                                               |


`generate` accepts `--trace` and `--profile` for local diagnostics. See
[Tracing](tracing.md) for trace destinations. `regen` accepts `--set key=value`
to patch an SDK's `gen.yaml`. Both `generate` and `regen` require an explicit license
election and refuse to run without one: pass `--license agpl-3.0-only` (or set
`SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only`) to accept AGPL-3.0-only output, which stamps
AGPL-3.0-only headers into every file and emits LICENSE/NOTICE. For commercial output, supply a
registry-signed license token with `--license-token <path>` (or the raw token in
`SPEAKEASY_LICENSE_TOKEN`); the token is validated in the same run. `./zero` records either
election in a gitignored `.env` that `make`, `mise run`, and `scripts/` load automatically. See
[License tokens](../README.md#license-tokens).

## Repository commands

Core repository maintenance commands use `mise`. Target- and variant-specific
commands remain on the root Makefile until their task migration lands.

```bash
# Build the generator and validator.
mise run build

# Validate the bundled test specifications.
mise run test:validate

# Run Go generator and package tests.
mise run test:generator

# Run the repository-wide lint and formatting suite (writes files).
mise run lint

# Start, restart, or stop local SDK test services.
mise run services:test-boot
mise run services:restart
mise run services:stop

# Apply changesets or regenerate maintenance metadata (writes files).
mise run changesets:apply
mise run features
mise run permissions

# Run a single language against a test variant.
TARGET=primary make test-go

# Rebuild a tracked review SDK.
TARGET=review make build-go

# Type-check a target's TypeScript templates.
make check-template-go
```

In `TARGET=primary make test-go`, `TARGET` is a historical name for the test
**variant**, while `go` identifies the generated **target**. The available
variant configuration lives in `tests/config/<variant>/<target>/`.

## Other development utilities

The `cmd/` directory also contains tools for feature reporting, standalone
README and usage generation, JSON Schema output, dependency reporting, upgrade
inspection, debugging, and repository maintenance. Use
`go run ./cmd/<name> --help` when a command exposes flags. Not every command is
part of the normal external contributor workflow; prefer the commands above
unless a feature's own tests or documentation direct you elsewhere.

Do not pass credentials, customer documents, private repository paths, or
unredacted production logs to commands or include them in generated fixtures.
