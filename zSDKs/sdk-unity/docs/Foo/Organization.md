# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                | Type                                                 | Required                                             | Description                                          | Example                                              |
| ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- |
| `Name`                                               | *string*                                             | :heavy_check_mark:                                   | N/A                                                  | Acme Corp                                            |
| `Address`                                            | [Address](../Foo/Address.md)                         | :heavy_check_mark:                                   | Inline address object should end up in foo namespace |                                                      |
| `Departments`                                        | List<[Department](../Foo/Department.md)>             | :heavy_minus_sign:                                   | N/A                                                  |                                                      |