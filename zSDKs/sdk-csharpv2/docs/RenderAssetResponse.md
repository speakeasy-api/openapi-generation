# RenderAssetResponse


## Fields

| Field                             | Type                              | Required                          | Description                       |
| --------------------------------- | --------------------------------- | --------------------------------- | --------------------------------- |
| `HttpMeta`                        | [HTTPMetadata](./HTTPMetadata.md) | :heavy_check_mark:                | N/A                               |
| `AssetResult`                     | [AssetResult](./AssetResult.md)   | :heavy_minus_sign:                | OK                                |
| `AssetStream`                     | *IAsyncEnumerable<AssetStream>*   | :heavy_minus_sign:                | OK                                |