# ChatModelRequest


## Fields

| Field                                  | Type                                   | Required                               | Description                            | Example                                |
| -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- |
| `Model`                                | *string*                               | :heavy_minus_sign:                     | Model to use                           |                                        |
| `Prompt`                               | *string*                               | :heavy_check_mark:                     | N/A                                    | What is the largest city in the world? |
| `Stream`                               | *bool*                                 | :heavy_minus_sign:                     | Stream events as they arrive           |                                        |
| `MaxTokens`                            | *long*                                 | :heavy_minus_sign:                     | Maximum number of tokens to generate   |                                        |
| `Temperature`                          | *double*                               | :heavy_minus_sign:                     | Sampling temperature                   |                                        |