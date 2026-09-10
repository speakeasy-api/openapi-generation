## core: 3.15.5 - 2025-11-25
### :bee: New Features
- union discriminators are inferred for oneOfs missing the explicit OpenAPI discriminator mapping. Configure via `inferUnionDiscriminators: true` in gen.yaml *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*
### :recycle: Refactors
- centralize SDK version constants in Constants class to reduce churn *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
