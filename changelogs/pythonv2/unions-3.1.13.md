## unions: 3.1.13 - 2026-10-28
### :bug: Bug Fixes
- serialize unknown open-union variants as their raw payload when preserveUnknownFields is enabled; model_dump() of an unknown variant no longer includes the discriminator, raw and is_unknown wrapper keys *(commit by [@2ynn](https://github.com/2ynn))*
