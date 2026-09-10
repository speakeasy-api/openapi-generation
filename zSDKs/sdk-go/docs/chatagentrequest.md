# ChatAgentRequest


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `Agent`                              | `string`                             | :heavy_check_mark:                   | N/A                                  |
| `Prompt`                             | `string`                             | :heavy_check_mark:                   | N/A                                  |
| `Stream`                             | `*bool`                              | :heavy_minus_sign:                   | Stream events as they arrive         |
| `MaxTokens`                          | `*int64`                             | :heavy_minus_sign:                   | Maximum number of tokens to generate |
| `Temperature`                        | `*float64`                           | :heavy_minus_sign:                   | Sampling temperature                 |