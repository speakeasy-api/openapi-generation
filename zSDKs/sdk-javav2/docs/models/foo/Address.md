# Address

Inline address object should end up in foo namespace


## Fields

| Field                                                           | Setter Type                                                     | Getter Type                                                     | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `street`                                                        | *String*                                                        | *String*                                                        | :heavy_check_mark:                                              | N/A                                                             | 123 Main St                                                     |
| `city`                                                          | *String*                                                        | *String*                                                        | :heavy_check_mark:                                              | N/A                                                             | Springfield                                                     |
| `location`                                                      | @Nullable [Location](../../models/shared/Location.md)           | Optional\<[Location](../../models/shared/Location.md)>          | :heavy_minus_sign:                                              | Deeply nested inline object should also end up in foo namespace |                                                                 |