# ChatResponse


## Fields

| Field                                             | Type                                              | Required                                          | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- |
| `HttpMeta`                                        | [HTTPMetadata](../Models/HTTPMetadata.md)         | :heavy_check_mark:                                | N/A                                               |
| `Object`                                          | [ChatResponseBody](../Models/ChatResponseBody.md) | :heavy_minus_sign:                                | A stream containing chat completion tokens        |
| `ChatStream`                                      | *EventStream<ChatStream>*                         | :heavy_minus_sign:                                | A stream containing chat completion tokens        |