## pagination: 2.83.7 - 2026-09-02
### :bug: Bug Fixes
- operations that combine x-speakeasy-pagination with x-speakeasy-polling no longer generate a Next method: the polling loop runs each attempt under a request-scoped context that a Next closure cannot outlive, so pagination is dropped for polled operations with a generation-time warning and a Next call on such an operation no longer compiles *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
