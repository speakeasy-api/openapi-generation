## core: 0.8.2 - 2026-10-01
### :bug: Bug Fixes
- --dry-run and --debug mask request and response body values marked sensitive (x-speakeasy-param-sensitive or format: password) through the SDK redact package, including whole bodies, recursive schemas and form or multipart fields *(commit by [@2ynn](https://github.com/2ynn))*
