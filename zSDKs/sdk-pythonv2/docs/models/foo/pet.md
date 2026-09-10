# Pet

A pet in the foo namespace


## Fields

| Field                                   | Type                                    | Required                                | Description                             | Example                                 |
| --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- | --------------------------------------- |
| `id`                                    | *str*                                   | :heavy_check_mark:                      | N/A                                     | pet-foo-123                             |
| `name`                                  | *str*                                   | :heavy_check_mark:                      | N/A                                     | Fluffy                                  |
| `species`                               | *str*                                   | :heavy_check_mark:                      | The species of the pet (e.g., dog, cat) | cat                                     |
| `models`                                | *Optional[str]*                         | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `request`                               | *Optional[str]*                         | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `operations`                            | *Optional[str]*                         | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `errors`                                | *Optional[str]*                         | :heavy_minus_sign:                      | A field name that often collides        |                                         |
| `utils`                                 | *Optional[str]*                         | :heavy_minus_sign:                      | A field name that often collides        |                                         |