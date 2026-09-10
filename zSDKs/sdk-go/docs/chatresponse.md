# ChatResponse


## Fields

| Field                                      | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `HTTPMeta`                                 | [HTTPMetadata](./httpmetadata.md)          | :heavy_check_mark:                         | N/A                                        |
| `Object`                                   | [*ChatResponseBody](./chatresponsebody.md) | :heavy_minus_sign:                         | A stream containing chat completion tokens |
| `ChatStream`                               | `*stream.EventStream[ChatStream]`          | :heavy_minus_sign:                         | A stream containing chat completion tokens |