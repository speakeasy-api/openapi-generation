# ShapeRequest

Request body with a discriminated union shape field


## Fields

| Field                                                                         | Type                                                                          | Required                                                                      | Description                                                                   |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `name`                                                                        | *::String*                                                                    | :heavy_check_mark:                                                            | Name for the shape                                                            |
| `shape`                                                                       | [T.any(Components::Circle, Components::Rectangle)](../models/shared/shape.md) | :heavy_check_mark:                                                            | A discriminated union of shape types                                          |
| `description`                                                                 | *T.nilable(::String)*                                                         | :heavy_minus_sign:                                                            | Optional description                                                          |