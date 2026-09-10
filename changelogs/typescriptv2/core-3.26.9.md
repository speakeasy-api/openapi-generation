## core: 3.26.9 - 2025-11-20
### :bee: New Features
- forwardCompatibleEnumsByDefault is now configurable via gen.yaml. When true, any enum which is used on a response will be automatically open/forward compatible - i.e. unknown values will be tolerated. Single value enums won't be automatically opened. Individual enums can be controlled with x-speakeasy-unknown-values: allow/disallow. *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*
