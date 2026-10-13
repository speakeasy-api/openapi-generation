## core: 0.9.5 - 2026-10-13
### :bug: Bug Fixes
- reject a malformed --jq expression as a usage error before any request is sent, and fail generation on a declared projection that does not compile; generated CLIs pin the generator's gojq version *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
