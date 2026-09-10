# RequestTimeoutError

A spec-defined error that collides with the built-in RequestTimeoutError in httpclienterrors.ts


## Fields

| Field                                                                       | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `code`                                                                      | *T.nilable(::Integer)*                                                      | :heavy_minus_sign:                                                          | N/A                                                                         |
| `message`                                                                   | *T.nilable(::String)*                                                       | :heavy_minus_sign:                                                          | N/A                                                                         |
| `raw_response`                                                              | [Faraday::Response](https://www.rubydoc.info/gems/faraday/Faraday/Response) | :heavy_minus_sign:                                                          | Raw HTTP response; suitable for custom response parsing                     |