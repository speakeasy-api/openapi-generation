## allowReserved: 0.1.1 - 2026-08-26
### :bug: Bug Fixes
- no longer honour `allowReserved: true` on path parameters (keyword only applies to query parameters), use `x-speakeasy-param-encoding-override: allowReserved` instead to preserve the previous encoding behaviour *(commit by [@2ynn](https://github.com/2ynn))*
