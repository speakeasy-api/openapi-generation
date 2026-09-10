# Address

Inline address object should end up in foo namespace


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `Street`                                                        | `string`                                                        | :heavy_check_mark:                                              | N/A                                                             | 123 Main St                                                     |
| `City`                                                          | `string`                                                        | :heavy_check_mark:                                              | N/A                                                             | Springfield                                                     |
| `Location`                                                      | [*foo.Location](../foo/location.md)                             | :heavy_minus_sign:                                              | Deeply nested inline object should also end up in foo namespace |                                                                 |