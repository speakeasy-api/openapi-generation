# Architecture

OpenAPI Generation turns an OpenAPI document and target configuration into a
language-specific SDK or related artifact.

## Generation flow

1. **Load and validate the document.** `cmd/generate` reads the OpenAPI document
   passed with `-s`. The standalone `cmd/validate` command can validate a document
   without generating an SDK.
2. **Build a target-neutral model.** The generator resolves the OpenAPI document
   into an SDK-oriented abstract syntax tree (AST). This gives target templates a
   consistent representation of operations, models, authentication, errors, and
   other API concepts.
3. **Load target configuration and templates.** The selected target (`-l`) maps to
   `templates/templates/<target>/`. Its TypeScript configuration and helper code
   determine supported features, file layout, and target-specific behavior.
4. **Render and format output.** Templates create files in the requested output
   directory. Generated code is passed through the target formatting pipeline
   where one is configured.
5. **Optionally compile and validate output.** Generation normally runs the
   target's configured compilation checks. Use `--skip-compile` only when the
   target compiler or package tooling is unavailable, or when iteration requires
   generation-only feedback.

The direct generator command creates `.speakeasy/gen.yaml` in a new output
directory. It records configuration for later regeneration.

## Repository layout

| Path                                                   | Responsibility                                                                                     |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------- |
| `cmd/`                                                 | Standalone commands for generation, validation, overlays, regeneration, and development workflows. |
| `internal/generate/`                                   | Generator orchestration and generation options.                                                    |
| `internal/ast/`                                        | Target-neutral SDK AST construction.                                                               |
| `internal/configuration/` and `internal/targetconfig/` | Generator and target configuration handling.                                                       |
| `internal/format/`                                     | Formatting pipeline and embedded formatter integration.                                            |
| `internal/validation/`                                 | OpenAPI validation and validation reports.                                                         |
| `pkg/`                                                 | Reusable Go packages exposed by the generator module.                                              |
| `templates/templates/<target>/`                        | Target-specific TypeScript templates, configuration, helpers, features, and tests.                 |
| `templates/templates/common/`                          | Template code shared by multiple targets.                                                          |
| `tests/specs/`                                         | Base OpenAPI documents and reusable document fragments.                                            |
| `tests/overlays/`                                      | Variant- and target-specific OpenAPI overlays.                                                     |
| `tests/config/<variant>/<target>/`                     | Generation configuration for test variants and targets.                                            |
| `zSDKs/`                                               | Small tracked review SDKs used to inspect generated-output changes.                                |

## Targets, variants, and tests

A **target** is the generated output type, such as `go`, `pythonv2`, or
`terraform`. A **variant** selects a test configuration, such as `primary`,
`secondary`, or `review`.

The root Makefile uses the historical environment variable name `TARGET` for a
variant in commands such as `TARGET=primary make test-go`; `go` is the target
and `primary` is the variant. Keep this distinction in mind when locating
configuration under `tests/config/<variant>/<target>/`.

For a broadly applicable behavior change, add a focused OpenAPI fragment under
`tests/specs/fragments/`. Use an overlay only when the behavior is truly
variant- or target-specific. Template changes that alter a tracked review SDK
should regenerate only the affected `zSDKs/` target where possible.

See [Generator-Target Interface](target-interface.md) for the TypeScript target
contract and [Contributing](../CONTRIBUTING.md) for test and review guidance.
