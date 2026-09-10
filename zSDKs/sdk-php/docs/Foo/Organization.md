# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                | Type                                                 | Required                                             | Description                                          | Example                                              |
| ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- | ---------------------------------------------------- |
| `name`                                               | *string*                                             | :heavy_check_mark:                                   | N/A                                                  | Acme Corp                                            |
| `address`                                            | [Foo\Address](../foo/Address.md)                     | :heavy_check_mark:                                   | Inline address object should end up in foo namespace |                                                      |
| `departments`                                        | array<[Foo\Department](../foo/Department.md)>        | :heavy_minus_sign:                                   | N/A                                                  |                                                      |