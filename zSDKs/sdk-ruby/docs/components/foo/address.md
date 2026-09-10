# Address

Inline address object should end up in foo namespace


## Fields

| Field                                                                   | Type                                                                    | Required                                                                | Description                                                             | Example                                                                 |
| ----------------------------------------------------------------------- | ----------------------------------------------------------------------- | ----------------------------------------------------------------------- | ----------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| `street`                                                                | *::String*                                                              | :heavy_check_mark:                                                      | N/A                                                                     | 123 Main St                                                             |
| `city`                                                                  | *::String*                                                              | :heavy_check_mark:                                                      | N/A                                                                     | Springfield                                                             |
| `location`                                                              | [T.nilable(Components::Foo::Location)](../../models/shared/location.md) | :heavy_minus_sign:                                                      | Deeply nested inline object should also end up in foo namespace         |                                                                         |