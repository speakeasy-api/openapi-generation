## operationTimeout: 0.3.1 - 2026-03-23
### :bug: Bug Fixes
- ensure per-request timeout_ms override is respected by passing USE_CLIENT_DEFAULT to httpx when unset and preserving timeout extensions across hooks *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
