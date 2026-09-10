## examples: 2.82.6 - 2026-05-15
### :bug: Bug Fixes
- render OpenAPI 'example: null' values as the literal "null" in model docs instead of the Go-formatted '<nil>', which broke MDX renderers and leaked Go internals *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
