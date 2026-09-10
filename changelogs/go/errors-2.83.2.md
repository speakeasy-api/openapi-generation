## errors: 2.83.2 - 2026-07-17
### :bee: New Features
- add idiomaticMethodCollisionNames option to resolve field names that collide with generated struct methods (e.g. Error()) using idiomatic replacements (e.g. ErrorInfo) instead of a trailing underscore, with sibling-aware fallback to the legacy underscore name *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
