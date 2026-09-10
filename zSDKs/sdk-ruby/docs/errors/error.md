# Error

A not-so-long multi-line
error model description.



## Fields

| Field                                                                       | Type                                                                        | Required                                                                    | Description                                                                 | Example                                                                     |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `error`                                                                     | *::String*                                                                  | :heavy_check_mark:                                                          | N/A                                                                         | some example error                                                          |
| `code`                                                                      | *::Integer*                                                                 | :heavy_check_mark:                                                          | N/A                                                                         | 2                                                                           |
| `raw_response`                                                              | [Faraday::Response](https://www.rubydoc.info/gems/faraday/Faraday/Response) | :heavy_minus_sign:                                                          | Raw HTTP response; suitable for custom response parsing                     |                                                                             |