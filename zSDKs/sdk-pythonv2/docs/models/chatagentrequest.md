# ChatAgentRequest


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `agent`                              | *str*                                | :heavy_check_mark:                   | N/A                                  |
| `prompt`                             | *str*                                | :heavy_check_mark:                   | N/A                                  |
| `stream`                             | *Optional[bool]*                     | :heavy_minus_sign:                   | Stream events as they arrive         |
| `max_tokens`                         | *Optional[int]*                      | :heavy_minus_sign:                   | Maximum number of tokens to generate |
| `temperature`                        | *Optional[float]*                    | :heavy_minus_sign:                   | Sampling temperature                 |