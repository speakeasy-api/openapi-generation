# ShapeRequest

Request body with a discriminated union shape field


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `name`                               | *str*                                | :heavy_check_mark:                   | Name for the shape                   |
| `shape`                              | [models.Shape](../models/shape.md)   | :heavy_check_mark:                   | A discriminated union of shape types |
| `description`                        | *Optional[str]*                      | :heavy_minus_sign:                   | Optional description                 |