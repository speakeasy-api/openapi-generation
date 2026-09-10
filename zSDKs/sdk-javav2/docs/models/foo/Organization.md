# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                             | Setter Type                                                       | Getter Type                                                       | Required                                                          | Description                                                       | Example                                                           |
| ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- |
| `name`                                                            | *String*                                                          | *String*                                                          | :heavy_check_mark:                                                | N/A                                                               | Acme Corp                                                         |
| `address`                                                         | [Address](../../models/shared/Address.md)                         | [Address](../../models/shared/Address.md)                         | :heavy_check_mark:                                                | Inline address object should end up in foo namespace              |                                                                   |
| `departments`                                                     | @Nullable List\<[Department](../../models/shared/Department.md)>  | Optional\<List\<[Department](../../models/shared/Department.md)>> | :heavy_minus_sign:                                                | N/A                                                               |                                                                   |