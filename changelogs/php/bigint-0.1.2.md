## bigint: 0.1.2 - 2026-09-02
### :bug: Bug Fixes
- serialize integers larger than PHP_INT_MAX by raising an overflow error instead of silently clamping to PHP_INT_MAX *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
