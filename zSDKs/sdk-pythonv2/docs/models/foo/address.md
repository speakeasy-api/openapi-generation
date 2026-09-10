# Address

Inline address object should end up in foo namespace


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `street`                                                        | *str*                                                           | :heavy_check_mark:                                              | N/A                                                             | 123 Main St                                                     |
| `city`                                                          | *str*                                                           | :heavy_check_mark:                                              | N/A                                                             | Springfield                                                     |
| `location`                                                      | [Optional[foo.Location]](../../models/foo/location.md)          | :heavy_minus_sign:                                              | Deeply nested inline object should also end up in foo namespace |                                                                 |