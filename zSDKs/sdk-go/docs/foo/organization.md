# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                | Type                                                 | Required                                             | Description                                          | Example                                              |
| ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- |
| `Name`                                               | `string`                                             | :heavy_check_mark:                                   | N/A                                                  | Acme Corp                                            |
| `Address`                                            | [foo.Address](../foo/address.md)                     | :heavy_check_mark:                                   | Inline address object should end up in foo namespace |                                                      |
| `Departments`                                        | [][foo.Department](../foo/department.md)             | :heavy_minus_sign:                                   | N/A                                                  |                                                      |