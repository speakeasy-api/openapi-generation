## openEnums: 0.1.3 - 2026-06-03
### :bug: Bug Fixes
- no-zod open enum types now inline their primitive fallback as `(T & {})` so arbitrary string and number values can be assigned without collapsing literal completions or emitting `Unrecognized<T>` helpers *(commit by [@ThomasRooney](https://github.com/ThomasRooney))*
