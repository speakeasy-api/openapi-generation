## retries: 3.1.1 - 2026-08-26
### :bug: Bug Fixes
- close the retried response before sleeping so streaming retries do not leak pooled connections *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
