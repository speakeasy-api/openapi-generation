## core: 6.1.15 - 2026-11-06
### :bug: Bug Fixes
- base64 file input fields base64-encode raw bytes, bytearray and byte-format memoryview values instead of rejecting non-UTF-8 payloads or sending ASCII payloads unencoded. str values are still sent as already-encoded base64, so pass already-encoded bytes as str (value.decode("ascii")) to avoid encoding them twice *(commit by [@2ynn](https://github.com/2ynn))*
