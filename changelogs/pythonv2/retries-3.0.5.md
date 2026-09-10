## retries: 3.0.5 - 2026-04-15
### :bug: Bug Fixes
- retry mechanism now correctly retries on error status codes (5XX, 429, etc.) instead of raising immediately on the first failure *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
