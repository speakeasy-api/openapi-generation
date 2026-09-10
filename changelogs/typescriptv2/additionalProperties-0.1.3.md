## additionalProperties: 0.1.3 - 2025-12-19
### :bee: New Features
- Added `flatAdditionalProperties` config option. When enabled, models with `additionalProperties: true` use an index signature (`[k: string]: unknown`) instead of a nested `additionalProperties` field. Extra properties are accessed directly on the object. Defaults to `false`. **Note:** This only applies to `additionalProperties: true` (any type). Typed `additionalProperties` (e.g. `additionalProperties: { type: string }`) retain their nested field structure to preserve type information.
