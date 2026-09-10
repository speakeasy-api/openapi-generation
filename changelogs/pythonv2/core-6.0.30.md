## core: 6.0.30 - 2026-06-12
### :bug: Bug Fixes
- base64 file inputs no longer drain seekable streams during repeated validation, preventing empty payloads when pydantic retries union members *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
