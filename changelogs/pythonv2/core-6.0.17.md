## core: 6.0.17 - 2026-03-05
### :bug: Bug Fixes
- fix model serializer dropping aliased fields when `model_dump()` is called without `by_alias=True` by falling back to field name lookup
