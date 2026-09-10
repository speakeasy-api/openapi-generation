<div align="center">
  <a href="https://www.speakeasy.com/" target="_blank">
    <img
      width="1500"
      height="500"
      alt="Speakeasy"
      src="https://github.com/user-attachments/assets/0e56055b-02a3-4476-9130-4be299e5a39c"
    />
  </a>

  <h1>OpenAPI Generation</h1>

  <p>
    Speakeasy's Go-based generator for producing idiomatic SDKs, Terraform
    providers, MCP servers, Postman collections, and developer tooling from
    OpenAPI documents.
  </p>

  <p>
    <a href="https://spec.openapis.org/oas/v3.1.1.html"><img alt="OpenAPI support" src="https://img.shields.io/badge/OpenAPI-3.0%20%7C%203.1-85EA2D.svg?style=for-the-badge&logo=openapiinitiative" /></a>
    <a href="https://spec.openapis.org/oas/v3.2.0.html"><img alt="OpenAPI 3.2 constructs" src="https://img.shields.io/badge/OpenAPI%203.2-selected%20constructs-85EA2D.svg?style=for-the-badge&logo=openapiinitiative" /></a>
    <a href="https://go.dev/"><img alt="Go version" src="https://img.shields.io/badge/go-1.26.2-00ADD8.svg?style=for-the-badge&logo=go" /></a>
    <br />
    <a href="https://github.com/speakeasy-api/openapi-generation/actions/workflows/test.yml"><img alt="GitHub Actions: Test" src="https://img.shields.io/github/actions/workflow/status/speakeasy-api/openapi-generation/test.yml?style=for-the-badge&label=CI" /></a>
    <a href="#supported-targets"><img alt="Supported targets" src="https://img.shields.io/badge/targets-12-8957E5.svg?style=for-the-badge" /></a>
    <a href="LICENSE"><img alt="AGPL-3.0 license" src="https://img.shields.io/badge/license-AGPL--3.0-blue.svg?style=for-the-badge" /></a>
    <br />
    <a href="https://www.speakeasy.com/"><img alt="Built by Speakeasy" src="https://www.speakeasy.com/assets/badges/built-by-speakeasy.svg" /></a>
  </p>

  <p>
    <a href="SETUP.md"><b>Development setup</b></a> &nbsp;//&nbsp;
    <a href="docs/architecture.md"><b>Architecture</b></a> &nbsp;//&nbsp;
    <a href="docs/cli-tools.md"><b>Command reference</b></a> &nbsp;//&nbsp;
    <a href="docs/testing.md"><b>Testing guide</b></a> &nbsp;//&nbsp;
    <a href="CONTRIBUTING.md"><b>Contributing</b></a> &nbsp;//&nbsp;
    <a href="https://github.com/speakeasy-api/speakeasy"><b>Speakeasy CLI</b></a>
  </p>
</div>

---

## Generate from OpenAPI

This repository is the development home for the generator and its templates. For
normal end-user SDK generation, use the
[Speakeasy CLI](https://github.com/speakeasy-api/speakeasy). To develop the
generator or a target template directly, generate from an OpenAPI document:

```bash
go run ./cmd/generate/main.go \
  -s ./tests/specs/basic-http.yaml \
  -o /tmp/generated-sdk \
  -l go \
  --license agpl-3.0-only \
  --skip-compile
```

`--license agpl-3.0-only` accepts AGPL-3.0-only licensing for the generated
output; commercial customers supply a license token instead. Run `./zero` once
per checkout to record either choice in `.env` (see
[Licensing generated output](#licensing-generated-output)).

The first run creates `.speakeasy/gen.yaml` in the output directory with default
target configuration. Remove `--skip-compile` when the generated target's
compiler and package tooling are installed. See the
[command reference](docs/cli-tools.md) for generator, validation, regeneration,
and test commands.

## Supported targets

The generator supports 12 primary targets:

| SDKs and tooling                                                       | Other outputs                                                  |
| ---------------------------------------------------------------------- | -------------------------------------------------------------- |
| C#, Go, Java, MCP TypeScript, PHP, Python, Ruby, TypeScript, and Unity | CLI applications, Postman collections, and Terraform providers |

Target templates live in `templates/templates/<target>`, with generation and
test configuration in `tests/config/<variant>/<target>/`. OpenAPI 3.0 and 3.1
are supported. The generator also handles selected OpenAPI 3.2 constructs, but
OpenAPI 3.2 is not yet supported as a complete document version.

## Develop locally

Install Go `1.26.2`, Node.js and npm, Docker, and the tooling required by the
targets you plan to test. The full setup guide covers target-specific runtimes
and local test services:

```bash
# Install the repository's documented prerequisites first.
# See SETUP.md for target-specific toolchains.

# Build the generator and validator with the initial mise task.
mise run build

# Run the core contributor checks through mise.
mise run test:generator
mise run test:validate
mise run lint

# Generate and test one target against the primary test variant.
TARGET=primary make test-go

# Check one target's TypeScript templates.
make check-template-go
```

`mise` tasks are being introduced incrementally. Use `make` for target- and
variant-specific commands that
do not yet have a documented `mise` equivalent; nested template and `zSDKs`
Makefiles remain supported for their existing consumers.

SDK tests start local `httpbin` and `speakeasy-api-test-service` containers
automatically. If you change `services/speakeasy-api-test-service/`, run
`mise run services:restart` before rerunning SDK tests. For command details and
common local failures, read the [testing guide](docs/testing.md),
[command reference](docs/cli-tools.md), and [troubleshooting guide](docs/troubleshooting.md).

## How the generator is organised

The generation flow is:

1. Parse and validate the OpenAPI document.
2. Build an SDK-oriented abstract syntax tree (AST).
3. Load the target's templates and TypeScript helpers.
4. Render, format, and optionally compile the generated output.

| Path                            | Purpose                                                                                      |
| ------------------------------- | -------------------------------------------------------------------------------------------- |
| `cmd/`                          | Generator, validation, overlay, and development commands.                                    |
| `internal/`                     | Generator implementation, AST construction, formatting, validation, and supporting services. |
| `pkg/`                          | Public Go packages used by the generator.                                                    |
| `templates/templates/<target>/` | Target-specific templates, helpers, tests, and feature definitions.                          |
| `templates/templates/common/`   | Template code shared across targets.                                                         |
| `tests/specs/`                  | Base OpenAPI test documents and reusable fragments.                                          |
| `tests/overlays/`               | Variant- and target-specific OpenAPI overlays.                                               |
| `tests/config/`                 | Per-variant target generation configuration.                                                 |
| `zSDKs/`                        | Small tracked review SDKs that make generated-output changes reviewable.                     |

Read [Architecture](docs/architecture.md) for the full generation flow and
[Generator–Target Interface](docs/target-interface.md) for the TypeScript
target contract.

## Test generated output

- Add generally applicable OpenAPI coverage with a documented fragment in
  `tests/specs/fragments/`.
- Use an overlay only for behavior that is genuinely target- or
  variant-specific.
- Add runtime coverage under `templates/templates/<target>/tests/` when
  generated behavior needs it.
- Regenerate the relevant `zSDKs/` review SDK when a template change affects
  tracked output.

The repository's snapshot companion is intentionally private. Public pull
requests receive only public-safe aggregate snapshot status; snapshot execution
must never expose customer inputs, private paths, or private logs.

## Telemetry

The generator records usage telemetry for telemetry-eligible generation
contexts. To disable it before running the generator:

```bash
export SPEAKEASY_DISABLE_TELEMETRY=true
```

The current implementation records selected target and template, success or
failure, operating system and architecture, generator and CLI version, run
location, configuration and feature flags, and `.speakeasy/gen.yaml` ignore
rule use. When available, it can also include document- or
configuration-derived metadata such as server URL, support contact details,
document title, OpenAPI version, customer or workspace identifiers, validation
warnings, and validation errors. It does not change an OpenAPI document or
generated output.

This disclosure reflects `internal/analytics/analytics.go`. Review document
content before generating if telemetry remains enabled; direct-mode telemetry
minimisation is tracked separately from this README update.

## Licensing generated output

The repository is licensed under [AGPL-3.0](LICENSE). Generated output carries
the license its caller elects — AGPL-3.0-only, in which case the generator
emits a `LICENSE` and `NOTICE` for Speakeasy-authored material, or commercial,
proven by a license token — and the generator refuses to run without an
election. That notice does not determine the license of content derived solely
from the OpenAPI document supplied for generation.

Commercial and generated-artifact licensing, including rights for existing
artifacts, is described in [LICENSING.md](LICENSING.md).

### License tokens

With a commercial license token, your generated SDK is yours to use, modify,
distribute, and license on your own terms—even though the generator itself is
AGPL-3.0. This includes Speakeasy code copied into the output, and these rights
continue after your token expires. Third-party content, including any supplied
through the input API specification, remains subject to its existing licenses.
See [Licensing](LICENSING.md).

### Local development: `./zero`

Run `./zero` once per checkout (also `mise run zero`). It records your license
election in a gitignored `.env`, which `make`, `mise run`, and the scripts
under `scripts/` load automatically:

- Commercial customers: `speakeasy auth login`, then `./zero`. It fetches a
  signed license token for your workspace (`SPEAKEASY_LICENSE_TOKEN`). Tokens
  expire after 30 days; rerun `./zero` to refresh.
- Open source: `./zero --license agpl-3.0-only` accepts AGPL-3.0-only output
  (`SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only`).
- Non-interactive use (`./zero --agent`, `-y`) never elects AGPL on its own;
  pass `--license` explicitly.

Without an election, `make test-<target>` and the build scripts stop with the
same guidance. `.env.example` documents the variables.

### Repository fixtures are never stamped

Output generated into a directory inside this repository — the tracked
`zSDKs/` review fixtures and the gitignored `testSDKs/` — omits the AGPL
header, `LICENSE`, `NOTICE`, and README badge regardless of the election
(`internal/repofixture`): those files are covered by the repository
[LICENSE](LICENSE), and every contributor, with a commercial token or an AGPL
election, regenerates identical bytes. The election is still required.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Keep a
pull request focused, describe generated-output impact, and run the smallest
relevant checks first. Never include credentials, customer documents, private
repository URLs, private filesystem paths, or unredacted logs in an issue, pull
request, fixture, or commit message.

Contributors must follow the [Contributor License Agreement](CLA.md) signing
process described in [CONTRIBUTING.md](CONTRIBUTING.md). The agreement permits
Speakeasy API to offer contributions under both AGPL-3.0 and separate commercial
licences.

The public security, Code of Conduct, support, trademark, and
enforcement routes require approved policy text and will be added separately.
