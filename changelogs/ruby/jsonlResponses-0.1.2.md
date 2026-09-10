## jsonlResponses: 0.1.2 - 2026-09-01
### :bug: Bug Fixes
- support retry strategy none and operation-level disabled retry overrides; streaming methods buffer JSON and text bodies before decoding them while SSE, JSONL and raw stream bodies stay live; response content-type matching is case-insensitive *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
