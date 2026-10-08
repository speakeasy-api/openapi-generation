# CLI schema extensions

The [CLI template README](../templates/templates/cli/README.md) documents command
extensions such as `x-speakeasy-cli-commands`, `x-speakeasy-cli-errors` and
`x-speakeasy-cli-catalog`, together with generator configuration and runtime
behavior.

## Custom pattern error messages

Add `x-speakeasy-pattern-error-message` to a string schema to explain a failed
`pattern` check in terms a CLI user can act on:

```yaml
parameters:
  - name: code
    in: path
    required: true
    schema:
      type: string
      pattern: '^[a-z]+$'
      x-speakeasy-pattern-error-message: Use lowercase letters.
```

Passing `--code ABC` without the extension produces:

```text
invalid value for --code: "ABC" does not match the pattern ^[a-z]+$
```

With the extension it produces:

```text
invalid value for --code: "ABC": Use lowercase letters.
```

The flag name and offending value remain in the error. Values keep their existing
quoting and truncation. The message is literal text: percent signs, quotes,
Unicode and line breaks are preserved. All output modes retain
`validation_error`, reason `CLI_VALIDATION`, and exit code `2`. Invalid input is
rejected before any HTTP request, including during `--dry-run`. Valid values and
requests are unchanged. A `maxLength` error takes precedence when both constraints
fail.

The extension must be a string containing at least one non-whitespace character.
Non-string values, empty strings and whitespace-only strings are generator
validation errors. A valid message on a schema without its own `pattern` produces
a warning. Existing reference and composition rules can supply the effective
pattern and inherit a message; the message is used only if the resulting string
schema has an enforced pattern. Pattern and message remain paired during schema
merges: changing a pattern without supplying a message drops the inherited
message and restores the regex explanation. An equal pattern keeps its message,
and an overriding schema's own message takes precedence. Flattening a primitive
`oneOf` or `anyOf` that combines different patterns drops the inherited message;
when the pattern is unchanged, the first member's message is retained. Messages
on primitive `oneOf` or `anyOf` wrappers are not retained; use an `allOf` wrapper
for a message override at the use site. An explicitly empty pattern is valid,
matches all values, and produces no missing-pattern warning. Extension name rewrites through
`x-speakeasy-extension-rewrite` are supported.

This applies wherever the CLI already checks string flag patterns: path, query,
and header parameters, request-body fields exposed as string flags, and
positional aliases for those flags. Referenced schemas work the same way as
inline schemas when the extension is declared on the referenced schema itself.
A message next to `$ref` is not used, including in OpenAPI 3.1. To customize a
reference at its use site, put it in `allOf` and declare the message on that
wrapper. The extension does not add validation to raw `--body` or stdin
JSON, global flags, unchanged defaults, arrays, enum/date/bytes flags, or patterns
outside the CLI's existing portable subset. Such constraints remain with the
server. See the template README's string-constraint notes for that subset.
Operations with cookie parameters remain unsupported.

`--help` and `--usage` retain their descriptions without adding pattern text or
custom messages. Request-body `--schema` retains the real `pattern` and the
extension as authored. The message customizes diagnostics; it does not hide
schema constraints or generated source.

Runtime support is currently CLI-only. Other generated targets retain their
existing behavior.
