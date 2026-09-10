# Testing guide

This guide helps contributors select the smallest test that proves a change.
Core test and service commands use `mise`; target and variant commands remain
on the root Makefile while their task migration is in progress. See [Development setup](../SETUP.md)
for required toolchains and [Troubleshooting](troubleshooting.md) when a local
check fails.

## Choose the smallest relevant check

Run the narrowest command that exercises the behavior you changed:

| Change                                 | Command                       | What it covers                                                                                             |
| -------------------------------------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------- |
| OpenAPI validation or validation rules | `mise run test:validate`      | Validates the bundled OpenAPI test specifications.                                                         |
| Go generator implementation            | `mise run test:generator`     | Runs the aggregate Go test suite across generator, command, package, and test code.                        |
| One generated target and test variant  | `TARGET=primary make test-go` | Builds the Go SDK for the `primary` variant, starts required local services, and runs its generated tests. |
| Target template TypeScript             | `make check-template-go`      | Type-checks the Go target's templates.                                                                     |
| Tracked generated review output        | `TARGET=review make build-go` | Regenerates the Go review SDK under `zSDKs/`.                                                              |
| Repository-wide matrix                 | `make test`                   | Runs aggregate Go tests, specification validation, all SDK targets, and README checks.                     |

Replace `go` with the affected target. In commands such as `TARGET=primary
make test-go`, `TARGET` is the historical name for the test **variant**;
`go` is the generated **target**. Variant configuration lives in
`tests/config/<variant>/<target>/`.

The full matrix requires multiple target runtimes and Docker. Start with the
smallest relevant check, and state in a pull request which checks you ran and
which were unavailable.

## Generated SDK tests and local services

SDK test commands start `httpbin` and `speakeasy-api-test-service`
automatically. You can start them without selecting a target when investigating
a service-dependent failure:

```bash
mise run services:test-boot
```

The scripts try the documented default ports (`35123` for `httpbin` and
`35456` for the API test service), then use an available port if necessary.
They save the selected ports in `.test-ports`, which target test commands read
automatically. To request fixed ports, set `HTTPBIN_PORT` and/or
`API_TEST_SERVICE_PORT` before starting the services.

If you change `services/speakeasy-api-test-service/`, rebuild and restart the
services before rerunning target tests:

```bash
mise run services:restart
# ... run the affected target test ...
mise run services:stop
```

`mise run services:stop` also removes the local service state files. Do not commit
`.test-ports`, `.test-pids`, service logs, or generated test output.

## Add coverage

Choose the test artifact that reflects the scope of the behavior:

- Add a focused, kebab-case OpenAPI fragment under
  `tests/specs/fragments/uber/` for shared coverage, or
  `tests/specs/fragments/<variant>/` for a specific variant. Include a
  language-neutral description of the construct being tested.
- Use an overlay in `tests/overlays/` only when behavior is genuinely
  variant- or target-specific.
- Add runtime assertions under `templates/templates/<target>/tests/` when the
  generated SDK behavior needs them. Additional handwritten tests use the
  target's existing `_additional` naming convention.
- Add `servers: [{url: http://localhost:35456}]` to an operation only when the
  behavior requires the API test service rather than the default test server.
- Register the test in `internal/features/tests.go` and run
  `go generate ./internal/features/...` when adding a new feature test ID.

Read [Contributing](../CONTRIBUTING.md#adding-test-coverage) for the required
fragment and target-coverage workflow.

## Review SDKs and snapshots

Use a tracked review SDK when a template change needs a compact,
reviewable generated-output diff. Regenerate only the affected target where
possible:

```bash
TARGET=review make build-go
```

Inspect the resulting `zSDKs/` diff before requesting review. For template
changes, also run `npm run format` and `make check-template-<target>`. When
adding, deleting, or moving template files, run `go generate ./...` to update
`templates/perms.go`.

The snapshot companion is intentionally private. Public pull requests receive
only aggregate, public-safe snapshot status. Snapshot execution for a fork
requires maintainer approval of that pull request's exact head SHA. Never put
customer documents, private repository URLs or paths, credentials, or private
snapshot logs in a fixture, issue, pull request, or test output.

## Before requesting review

Run `git diff --check`, inspect generated output deliberately, and report the
focused validation you ran. `mise run lint` is the broader repository gate, but it
writes formatting, fragment-security, and template-permission updates before
checking the tree; inspect its diff and retain only expected changes.
