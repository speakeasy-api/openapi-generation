# ChatModelRequest


## Fields

| Field                                  | Type                                   | Required                               | Description                            | Example                                |
| -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- |
| `model`                                | *Optional[str]*                        | :heavy_minus_sign:                     | Model to use                           |                                        |
| `prompt`                               | *str*                                  | :heavy_check_mark:                     | N/A                                    | What is the largest city in the world? |
| `stream`                               | *Optional[bool]*                       | :heavy_minus_sign:                     | Stream events as they arrive           |                                        |
| `max_tokens`                           | *Optional[int]*                        | :heavy_minus_sign:                     | Maximum number of tokens to generate   |                                        |
| `temperature`                          | *Optional[float]*                      | :heavy_minus_sign:                     | Sampling temperature                   |                                        |