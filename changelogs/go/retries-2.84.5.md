## retries: 2.84.5 - 2026-09-01
### :bug: Bug Fixes
- a context canceled between retryable responses surfaces as the cancellation instead of the retryable error, which the caller could otherwise report as a successful response *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
