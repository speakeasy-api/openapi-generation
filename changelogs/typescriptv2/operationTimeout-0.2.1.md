## operationTimeout: 0.2.1 - 2026-05-19
### :bug: Bug Fixes
- AbortSignal is no longer shared across retry attempts, preventing every retry from instantly failing with TimeoutError when timeoutMs is configured *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
