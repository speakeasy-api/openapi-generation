## core: 3.16.2 - 2026-10-16
### :bug: Bug Fixes
- read a body-less success response inside the method when deserialization is skipped, so a per-request timeout no longer cancels the raw body before the caller reads it *(commit by [@2ynn](https://github.com/2ynn))*
