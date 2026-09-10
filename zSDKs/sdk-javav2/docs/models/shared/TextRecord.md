# TextRecord

A text record request


## Fields

| Field                              | Setter Type                        | Getter Type                        | Required                           | Description                        | Example                            |
| ---------------------------------- | ---------------------------------- | ---------------------------------- | ---------------------------------- | ---------------------------------- | ---------------------------------- |
| `type`                             | *String*                           | *String*                           | :heavy_check_mark:                 | Record type discriminator |                                    |
| `text`                        | *String*                           | *String*                           | :heavy_check_mark:                 | N/A                                | sample text                        |
| `note`                        | @Nullable *String*                 | Optional\<*String*>                | :heavy_minus_sign:                 | N/A                                | sample note                        |
