# ChatModelRequest


## Fields

| Field                                  | Setter Type                            | Getter Type                            | Required                               | Description                            | Example                                |
| -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- |
| `model`                                | @Nullable *String*                     | Optional\<*String*>                    | :heavy_minus_sign:                     | Model to use                           |                                        |
| `prompt`                               | *String*                               | *String*                               | :heavy_check_mark:                     | N/A                                    | What is the largest city in the world? |
| `stream`                               | @Nullable *boolean*                    | Optional\<*boolean*>                   | :heavy_minus_sign:                     | Stream events as they arrive           |                                        |
| `maxTokens`                            | @Nullable *long*                       | Optional\<*long*>                      | :heavy_minus_sign:                     | Maximum number of tokens to generate   |                                        |
| `temperature`                          | @Nullable *double*                     | Optional\<*double*>                    | :heavy_minus_sign:                     | Sampling temperature                   |                                        |