# CLI Template Notes

## Keep `README.md` current

When making changes anywhere under `templates/templates/cli/`, treat `README.md` as a required maintenance surface.

Update `templates/templates/cli/README.md` in the same change whenever you modify or add:

- user-facing commands or flags
- output formats or error behavior
- interactive mode behavior (`explore`, auto-prompting, `--no-interactive`, auth/configure forms)
- agent mode behavior (`--agent-mode`, auto-detection, structured errors, TOON default output format)
- config/env/keyring precedence or storage behavior
- request input behavior (`--body`, stdin merge, bytes/base64/file support)
- binary response behavior (`--output-file`, `--output-b64`, TTY enforcement)
- machine-readable docs surfaces (`--usage`, grouped help)
- generator config knobs in `config.ts` that affect generated CLI behavior

The goal is for `templates/templates/cli/README.md` to stay the maintainer-facing source of truth for what the CLI generator actually does, so external/product documentation can be updated from it.

## Documentation standard

Do not document features based on memory. Verify behavior from the current template/runtime code before updating the README.

When behavior is nuanced, document:

- the relevant files
- user-visible behavior
- precedence rules / defaults
- important edge cases
- any generator config gates

## Scope reminder

All functional changes for the CLI generator belong in `templates/templates/cli/`.
Do not edit generated output directly.
