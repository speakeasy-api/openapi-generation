# ShapeRequest

Request body with a discriminated union shape field


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `Name`                               | `string`                             | :heavy_check_mark:                   | Name for the shape                   |
| `Shape`                              | [Shape](./shape.md)                  | :heavy_check_mark:                   | A discriminated union of shape types |
| `Description`                        | `*string`                            | :heavy_minus_sign:                   | Optional description                 |