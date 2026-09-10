## unions: 3.1.1 - 2025-12-08
### :bee: New Features
- simplify syntax of discriminated unions when discriminator const is already present on the underlying types *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*
- Support for `applyUnionDiscriminators: true` configurable via gen.yaml. When true, discriminator values will be applied as const fields onto union member types. If the type is used outside of a union, the const will not be applied.

