## serverEvents: 0.0.2 - 2026-09-02
### :bug: Bug Fixes
- the embedded Go SDK supports retry strategy none and operation-level disabled retry overrides; SSE/JSONL streams, raw response-stream bodies and JSON bodies left unread by SkipDeserialization take ownership of the per-operation timeout cancel (released on close or when the read ends) instead of it firing when the method returns, and request timeouts no longer cancel streams before iteration *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
