## unions: 3.1.8 - 2026-07-17
### :bug: Bug Fixes
- preserve nullable discriminated-union request fields under Pydantic 2.13 by wrapping optional+nullable closed discriminated-union fields in SerializeAsAny, and lift the now-unnecessary Pydantic <2.13 cap *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
