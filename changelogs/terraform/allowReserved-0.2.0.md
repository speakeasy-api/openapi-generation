## allowReserved: 0.2.0 - 2026-09-21
### :bee: New Features
- support passing RFC 3986 reserved characters unencoded in query parameters (allowReserved: true) and path parameters (`x-speakeasy-param-encoding-override: allowReserved`). Requires `fixFlags.encodePathParams` to be enabled *(commit by [@2ynn](https://github.com/2ynn))*
