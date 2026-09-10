# ShapeRequest

Request body with a discriminated union shape field


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `name`                               | *string*                             | :heavy_check_mark:                   | Name for the shape                   |
| `shape`                              | [Circle\|Rectangle](./Shape.md)      | :heavy_check_mark:                   | A discriminated union of shape types |
| `description`                        | *?string*                            | :heavy_minus_sign:                   | Optional description                 |