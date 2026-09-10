## unions: 2.86.2 - 2025-11-27
### :bee: New Features
- New gen.yaml flag forwardCompatibleUnionsByDefault controls forward compatibility for discriminated unions in responses. 'tagged-only' makes only discriminated unions open to unknown values. 'false' disables forward compatibility. Unknown discriminator values are returned as an object with the discriminator field set to a special 'UNKNOWN' string and a raw field containing the original input. Individual unions can be controlled with x-speakeasy-unknown-values: allow/disallow. *(commit by [@mfbx9da4](https://github.com/mfbx9da4))*
