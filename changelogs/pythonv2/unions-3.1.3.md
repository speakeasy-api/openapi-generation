## unions: 3.1.3 - 2026-02-10
### :bee: New Features
- add forward-compatible open discriminated unions with Unknown fallback models *(commit by [@vishalg0wda](https://github.com/vishalg0wda))*
- Controlled globally via `forwardCompatibleUnionsByDefault: tagged-only` in gen.yaml, or per-union with `x-speakeasy-unknown-values: allow/disallow`.
