# Address

Inline address object should end up in foo namespace


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `street`                                                        | *string*                                                        | :heavy_check_mark:                                              | N/A                                                             | 123 Main St                                                     |
| `city`                                                          | *string*                                                        | :heavy_check_mark:                                              | N/A                                                             | Springfield                                                     |
| `location`                                                      | [?Foo\Location](../foo/Location.md)                             | :heavy_minus_sign:                                              | Deeply nested inline object should also end up in foo namespace |                                                                 |