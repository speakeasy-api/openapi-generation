# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                  | Type                                                   | Required                                               | Description                                            | Example                                                |
| ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ |
| `name`                                                 | *str*                                                  | :heavy_check_mark:                                     | N/A                                                    | Acme Corp                                              |
| `address`                                              | [foo.Address](../../models/foo/address.md)             | :heavy_check_mark:                                     | Inline address object should end up in foo namespace   |                                                        |
| `departments`                                          | List[[foo.Department](../../models/foo/department.md)] | :heavy_minus_sign:                                     | N/A                                                    |                                                        |