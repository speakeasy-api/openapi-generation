# Reviewing changes

This guide helps contributors review changes to the generator, templates, tests,
and generated SDKs. A good review checks the stated intent and the downstream
impact, not only the edited lines.

## Start with the pull request intent

Read the pull request description and diff together. The description should
state why the change is needed, what it changes, whether it affects generated
SDK output, and which validation ran. Ask for clarification rather than
assuming an unstated product or compatibility decision.

Keep comments specific: identify the affected file or behavior, explain the
risk, and suggest the smallest useful next step.

## Generated-output changes

Generator and template changes can affect many SDKs. When output changes,
review the relevant generated artifacts as product output:

- check the appropriate tracked review SDK under `zSDKs/`;
- confirm generated code, documentation, examples, naming, and formatting are
  intentional and idiomatic for the target language;
- look for unexpected churn such as unrelated renames, broad formatting, or
  changed defaults; and
- make sure customer-visible generated-output changes have the required
  changelog or versioning update.

If output is expected to change but no review SDK or focused test artifact was
updated, ask the author to explain the coverage or add the appropriate one.

## Tests and specifications

Behavior changes should have focused evidence. Depending on the change, that
may be a test specification or fragment, an overlay, a generator test, a
snapshot, a generated SDK test, or a review SDK rebuild.

- Use `tests/specs/uber.yaml` and its fragments for behavior that should apply
  broadly.
- Use `tests/specs/review.yaml` and `zSDKs/` for compact generated-output
  review surfaces.
- Use overlays only for behavior that is genuinely target- or variant-specific.
- Confirm new test identities are registered when the relevant target workflow
  needs them.

For changes that affect generated output across targets, be cautious when a PR
updates only one target without an explanation.

## Compatibility and rollout

Call out changes that alter public method signatures, types, package layout,
default behavior, serialization, retries, pagination, authentication, or other
visible SDK behavior. Broad or breaking changes should explain their migration
or rollout strategy and retain safe defaults where practical.

## Repository quality

Check that shared behavior belongs in shared generator or template code rather
than being copied across targets. Script changes must remain portable across
macOS and Linux. Do not use GNU-only shell behavior when a portable equivalent
is available.

Review diffs for credentials, customer OpenAPI documents, private repository
URLs, private filesystem paths, and unredacted logs. Do not reproduce sensitive
information in a review comment.

## Before approving

A review should make clear:

1. what the pull request changes;
2. whether generated output and compatibility match that claim;
3. what validation supports the result; and
4. any remaining risk or follow-up.

If a change is ready, say why it is ready. If it is not, request the concrete
missing evidence or correction rather than leaving a vague objection.
