# ImageRecord

An image record


## Fields

| Field                     | Setter Type               | Getter Type               | Required                  | Description               | Example                   |
| ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- | ------------------------- |
| `type`                    | *String*                  | *String*                  | :heavy_check_mark:        | Record type discriminator |                           |
| `imageId`                 | *String*                  | *String*                  | :heavy_check_mark:        | N/A                       | image-001                 |
| `caption`                 | @Nullable *String*        | Optional\<*String*>       | :heavy_minus_sign:        | N/A                       | sample caption            |