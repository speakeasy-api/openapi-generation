# Organization

An organization with nested inline schemas in the foo namespace


## Fields

| Field                                                                      | Type                                                                       | Required                                                                   | Description                                                                | Example                                                                    |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `name`                                                                     | *::String*                                                                 | :heavy_check_mark:                                                         | N/A                                                                        | Acme Corp                                                                  |
| `address`                                                                  | [Components::Foo::Address](../../models/shared/address.md)                 | :heavy_check_mark:                                                         | Inline address object should end up in foo namespace                       |                                                                            |
| `departments`                                                              | T::Array<[Components::Foo::Department](../../models/shared/department.md)> | :heavy_minus_sign:                                                         | N/A                                                                        |                                                                            |