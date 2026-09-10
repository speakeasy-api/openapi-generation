## serverEvents: 1.0.18 - 2026-06-05
### :bug: Bug Fixes
- close the underlying httpx response on all stream-termination paths to prevent connection leaks when an event stream is consumed without a context manager *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
