# Pet

A pet in the foo namespace


## Fields

| Field                                   | Type                                    | Required                                | Description                             | Example                                 |
| --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- |
| `id`                                    | *::String*                              | :heavy_check_mark:                      | N/A                                     | pet-foo-123                             |
| `name`                                  | *::String*                              | :heavy_check_mark:                      | N/A                                     | Fluffy                                  |
| `species`                               | *::String*                              | :heavy_check_mark:                      | The species of the pet (e.g., dog, cat) | cat                                     |
| `models`                                | *T.nilable(::String)*                   | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `request`                               | *T.nilable(::String)*                   | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `operations`                            | *T.nilable(::String)*                   | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `errors`                                | *T.nilable(::String)*                   | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `utils`                                 | *T.nilable(::String)*                   | :heavy_minus_sign:                      | A field name that often collides        |                                         |