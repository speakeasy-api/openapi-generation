# ChatModelRequest


## Fields

| Field                                  | Type                                   | Required                               | Description                            | Example                                |
| -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- | -------------------------------------- |
| `model`                                | *T.nilable(::String)*                  | :heavy_minus_sign:                     | Model to use                           |                                        |
| `prompt`                               | *::String*                             | :heavy_check_mark:                     | N/A                                    | What is the largest city in the world? |
| `stream`                               | *T.nilable(T::Boolean)*                | :heavy_minus_sign:                     | Stream events as they arrive           |                                        |
| `max_tokens`                           | *T.nilable(::Integer)*                 | :heavy_minus_sign:                     | Maximum number of tokens to generate   |                                        |
| `temperature`                          | *T.nilable(::Float)*                   | :heavy_minus_sign:                     | Sampling temperature                   |                                        |