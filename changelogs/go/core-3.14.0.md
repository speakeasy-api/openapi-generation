## core: 3.14.0 - 2026-09-21
### :bug: Bug Fixes
- enabling fixFlags.encodePathParams percent-encodes path parameters in generated URLs: reserved characters such as `/` or `:` are sent as `%2F`/`%3A` instead of raw. Set `x-speakeasy-param-encoding-override: allowReserved` on a path parameter in the spec to keep its reserved characters unencoded *(commit by [@2ynn](https://github.com/2ynn))*
