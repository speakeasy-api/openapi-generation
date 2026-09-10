# ChatAgentRequest


## Fields

| Field                                | Type                                 | Required                             | Description                          |
| ------------------------------------ | ------------------------------------ | ------------------------------------ | ------------------------------------ |
| `agent`                              | *::String*                           | :heavy_check_mark:                   | N/A                                  |
| `prompt`                             | *::String*                           | :heavy_check_mark:                   | N/A                                  |
| `stream`                             | *T.nilable(T::Boolean)*              | :heavy_minus_sign:                   | Stream events as they arrive         |
| `max_tokens`                         | *T.nilable(::Integer)*               | :heavy_minus_sign:                   | Maximum number of tokens to generate |
| `temperature`                        | *T.nilable(::Float)*                 | :heavy_minus_sign:                   | Sampling temperature                 |