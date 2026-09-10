# ChatAgentRequest


## Fields

| Field                                | Setter Type                          | Getter Type                          | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `agent`                              | *String*                             | *String*                             | :heavy_check_mark:                   | N/A                                  |
| `prompt`                             | *String*                             | *String*                             | :heavy_check_mark:                   | N/A                                  |
| `stream`                             | @Nullable *boolean*                  | Optional\<*boolean*>                 | :heavy_minus_sign:                   | Stream events as they arrive         |
| `maxTokens`                          | @Nullable *long*                     | Optional\<*long*>                    | :heavy_minus_sign:                   | Maximum number of tokens to generate |
| `temperature`                        | @Nullable *double*                   | Optional\<*double*>                  | :heavy_minus_sign:                   | Sampling temperature                 |