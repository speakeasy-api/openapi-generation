# Contributing

Thanks for improving OpenAPI Generation. This guide explains how to make a focused, reviewable change while the repository is being prepared for public release.

## Before you start

1. Read [README.md](README.md) and complete the relevant portion of [SETUP.md](SETUP.md).
2. Search existing issues and pull requests before starting a larger change.
3. Keep one behavioral change or closely related maintenance task in each pull request.
4. Never include credentials, customer OpenAPI documents, private repository URLs, private filesystem paths, or unredacted logs in a pull request, issue, fixture, or commit message.

The public issue tracker, security contact, Code of Conduct, support route, and trademark policy will be added only after their approved policy text is available. Until then, use the repository's existing maintainer contact route for questions that cannot be discussed publicly.

## Contributor License Agreement

Read the [CLA](CLA.md) before contributing. It permits Speakeasy API to distribute
contributions under AGPL-3.0 and separate commercial licences while contributors
retain copyright ownership. When the CLA check requests your signature, follow
the instructions in its pull request comment. Signatures are recorded separately
from the source branch.

## Pull requests

Use a Conventional Commit-style title:

```text
fix(go): handle optional parameters correctly
feat(python,go): add a retry behavior
chore: update a development dependency
```

Use the pull request template to describe:

- **Why** the change is needed;
- **What changed**, including generated-output impact; and
- **Testing** you ran or intentionally did not run.

Link an issue when one exists, but an external contributor does not need access to any internal tracker. Do not state that a change affects a customer; describe the technical behavior instead.

## Changelogs and generated review SDKs

If a change does not alter generated output, add `[skip changelog]` to the pull request description or apply the `skip changelog` label so CI skips the changelog requirement. If a change modifies generated output, add a changelog entry:

```bash
make changelog <feature> <target> "fix: describe the generated-output change"
TARGET=review make build-<target>
```

For example:

```bash
make changelog pagination go "fix: preserve pagination parameters"
TARGET=review make build-go
```

The changelog command validates the feature and target, creates a `.changesets/*.yaml` file, and the merge process applies the version and changelog updates. Do not hand-write changesets, `features.ts`, or changelog entries that the command manages.

For cross-target changes, supply comma-separated features and targets. Use `all` only when every target is affected:

```bash
make changelog pagination,unions go,typescriptv2 "fix: preserve union pagination output"
TARGET=review make build-go
TARGET=review make build-typescriptv2
```

The `zSDKs/` directory contains small tracked review SDKs. Regenerate only the relevant target when possible. Review the resulting diff carefully: it should demonstrate the intended generated-output change without unrelated churn.

## Tests

Use the [testing guide](docs/testing.md) to select the smallest relevant check
and understand local service, review SDK, and snapshot-test behavior. The most
common commands are:

```bash
# Go implementation changes.
mise run test:generator

# OpenAPI validation changes.
mise run test:validate

# One target and variant.
TARGET=primary make test-go

# One target's template type-check.
make check-template-go
```

SDK tests start `httpbin` and `speakeasy-api-test-service` automatically. If you edit `services/speakeasy-api-test-service/`, run `mise run services:restart` before rerunning a target test.

### Adding test coverage

For behavior that should apply across targets or configurations:

1. Add a focused, kebab-case fragment under `tests/specs/fragments/uber/` or `tests/specs/fragments/<variant>/`.
2. Include an operation or schema description explaining the OpenAPI construct being tested in language-neutral terms.
3. Use `servers: [{url: http://localhost:35456}]` only when the test requires `speakeasy-api-test-service` behavior.
4. Register a test ID in `internal/features/tests.go`, then run `go generate ./internal/features/...`.
5. Add generated-test workflow coverage when the target supports it; otherwise add a target-specific additional test.
6. Remove the test ID from the target's `isTestSkipped` output only for targets that implement the test.

Use an overlay only when the test is genuinely target- or variant-specific. If a new test reveals a compatibility problem in another target, document it in the pull request and open a public issue when issue intake is available; do not add private tracker references to public source files.

### Snapshots and review SDKs

Use a review SDK when a template change needs a compact, inspectable generated-output diff. Use a snapshot test only when focused generator or target tests cannot cover the case.

The snapshot companion remains private by design. Fork pull requests require maintainer approval of the exact head commit before snapshots run. The public pull request receives only aggregate status and a public-safe comment; customer data, private snapshot logs, and private paths must not be exposed.

## Templates and generated files

- Target templates are in `templates/templates/<target>/`; shared template code is in `templates/templates/common/`.
- Prefer `templateFile()` for generated code so the target formatter runs. Reserve `writeFile()` for deliberately unformatted 1:1 content.
- When adding, deleting, or moving template files, run `go generate ./...` to update `templates/perms.go`.
- After modifying files under `templates/templates/`, run `npm run format` and `make check-template-<target>`.
- Run `git diff --check` before requesting review.

## Lint and validation

`mise run lint` is the repository-wide gate. It installs npm dependencies, updates generated permission data and fragment security blocks, applies formatting, runs `golangci-lint`, checks module tidiness, and type-checks templates. It may modify the working tree, so inspect the diff afterward.

For a narrow change, report the focused tests you ran and any unavailable toolchains. The full matrix requires multiple language runtimes and Docker; do not claim it passed unless it actually ran.

## Reviewing changes

Use [Review guidance](docs/reviewing.md) when reviewing a pull request. It explains how to assess generated-output impact, expected tests and review SDKs, changelog requirements, portability, and sensitive information exposure.

## Review expectations

Keep commits readable and avoid mixing refactors, formatting churn, generated output, and behavior changes without a reason. Reviewers should be able to answer:

- Which OpenAPI input or generator behavior changed?
- Which targets are affected?
- How did the change preserve or intentionally alter generated output?
- What validation provides confidence?
- Does the diff expose private, customer-specific, or credential-bearing information?

Maintainer assignment, branch protections, merge queue operation, and release decisions are repository-governance topics. They are intentionally outside this contributor guide until the approved public maintainer policy is available.
