## core: 3.15.32 - 2026-11-23
### :bug: Bug Fixes
- deserialize an explicit null into a nullable free-form object field instead of throwing, so union candidates and collection items carrying a null free-form object resolve instead of being rejected *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
