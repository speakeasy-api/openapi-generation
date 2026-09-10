## core: 0.13.15 - 2026-05-29
### :bug: Bug Fixes
- roll back Go directive to 1.25.9 (from 1.26); staticcheck v0.7.0 builds under go1.25.x and cannot lint a go 1.26 module, breaking generation *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
