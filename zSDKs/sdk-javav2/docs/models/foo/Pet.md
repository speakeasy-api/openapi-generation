# Pet

A pet in the foo namespace


## Fields

| Field                                   | Setter Type                             | Getter Type                             | Required                                | Description                             | Example                                 |
| --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- |
| `id`                                    | *String*                                | *String*                                | :heavy_check_mark:                      | N/A                                     | pet-foo-123                             |
| `name`                                  | *String*                                | *String*                                | :heavy_check_mark:                      | N/A                                     | Fluffy                                  |
| `species`                               | *String*                                | *String*                                | :heavy_check_mark:                      | The species of the pet (e.g., dog, cat) | cat                                     |
| `models`                                | @Nullable *String*                      | Optional\<*String*>                     | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `request`                               | @Nullable *String*                      | Optional\<*String*>                     | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `operations`                            | @Nullable *String*                      | Optional\<*String*>                     | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `errors`                                | @Nullable *String*                      | Optional\<*String*>                     | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `utils`                                 | @Nullable *String*                      | Optional\<*String*>                     | :heavy_minus_sign:                      | A field name that often collides        |                                         |