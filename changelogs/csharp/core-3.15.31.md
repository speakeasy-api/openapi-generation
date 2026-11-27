## core: 3.15.31 - 2026-11-19
### :bee: New Features
- add optionalNullableScope writable-models to limit OptionalNullable<T> wrapping to models the SDK can send, with an implicit unwrap to T? in that mode; under this scope, referencing a response-only schema from a request body later wraps it and is source-breaking for code that reads it *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
