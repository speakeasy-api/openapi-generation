## errors: 0.2.1 - 2026-09-23
### :bug: Bug Fixes
- error responses that fail to decode against their declared schema now return ResponseValidationError carrying the status code, body and raw response *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
