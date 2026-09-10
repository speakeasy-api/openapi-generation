# ShapeRequest

Request body with a discriminated union shape field


## Fields

| Field                                 | Setter Type                           | Getter Type                           | Required                              | Description                           |
| ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- |
| `name`                                | *String*                              | *String*                              | :heavy_check_mark:                    | Name for the shape                    |
| `shape`                               | [Shape](../../models/shared/Shape.md) | [Shape](../../models/shared/Shape.md) | :heavy_check_mark:                    | A discriminated union of shape types  |
| `description`                         | @Nullable *String*                    | Optional\<*String*>                   | :heavy_minus_sign:                    | Optional description                  |