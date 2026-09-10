# PR title: `type(scope): description` — e.g. `fix(python): handle optional parameters correctly`
<!-- Public text policy: this PR's title, body, comments, and commits are public. Do not name customers, their products/specs/APIs, private paths, or internal process; describe the change so it stands alone. See .claude/skills/public-repo-communication/SKILL.md -->

<!--
  Valid types: chore, fix, feat
  Valid scopes: typescript, python, go, java, csharp, php, ruby, unity, terraform, postman, mcp, mockserver, cli, all
  Multiple scopes: fix(python,go): description
  Chore with no scope: chore: description
-->

<!-- Link a public issue when one exists. Do not include private-tracker IDs or links. -->

## Why

<!-- What user or generator behavior needs to change? -->

## What changed

<!-- Summarize code, templates, generated review SDKs, and any expected output changes. -->

## Testing

<!-- List focused checks run and checks intentionally not run, with reasons. -->

## Public-safety check

- [ ] This change contains no credentials, customer documents, private repository URLs, private filesystem paths, or unredacted private logs.
- [ ] Title, body, comments, and commit messages name no customers or customer-derived identifiers, private paths or trackers, or workflow provenance, and are understandable without private context (`.claude/skills/public-repo-communication/SKILL.md`).
- [ ] Generated fixtures and review SDK changes are public-safe.
- [ ] I reviewed `git diff --check`.
