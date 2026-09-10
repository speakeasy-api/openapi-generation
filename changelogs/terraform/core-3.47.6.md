## core: 3.47.6 - 2026-05-20
### :bug: Bug Fixes
- delegate to json.Unmarshaler for map types in unmarshalValue, fixing unmarshal errors for OptionalNullable fields with complex value types like time.Time *(commit by [@AshGodfrey](https://github.com/AshGodfrey))*
