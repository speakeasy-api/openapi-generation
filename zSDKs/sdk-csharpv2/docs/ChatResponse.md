# ChatResponse


## Fields

| Field                                      | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `HttpMeta`                                 | [HTTPMetadata](./HTTPMetadata.md)          | :heavy_check_mark:                         | N/A                                        |
| `Object`                                   | [ChatResponseBody](./ChatResponseBody.md)  | :heavy_minus_sign:                         | A stream containing chat completion tokens |
| `ChatStream`                               | *IAsyncEnumerable<ChatStream>*             | :heavy_minus_sign:                         | A stream containing chat completion tokens |