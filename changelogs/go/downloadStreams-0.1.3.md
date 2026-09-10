## downloadStreams: 0.1.3 - 2026-09-01
### :bug: Bug Fixes
- raw response-stream bodies take ownership of the per-operation timeout cancel (released on close or when a read finishes), so a timeout bounds the whole download instead of severing the body when the method returns *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
