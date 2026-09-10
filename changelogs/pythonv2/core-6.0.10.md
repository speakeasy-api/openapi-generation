## core: 6.0.10 - 2026-02-18
### :bug: Bug Fixes
- schemas with inline enum fields named 'tag' no longer get incorrectly renamed to 'TagT' — the generator now only disambiguates when pydantic's Tag import actually conflicts in the same file, falling back to namespace import (pydantic.Tag) instead of mangling the class name *(commit by [@danielkov](https://github.com/danielkov))*