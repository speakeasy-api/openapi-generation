# RenderAssetResponse


## Fields

| Field                              | Type                               | Required                           | Description                        |
| ---------------------------------- | ---------------------------------- | ---------------------------------- | ---------------------------------- |
| `HTTPMeta`                         | [HTTPMetadata](./httpmetadata.md)  | :heavy_check_mark:                 | N/A                                |
| `AssetResult`                      | [*AssetResult](./assetresult.md)   | :heavy_minus_sign:                 | OK                                 |
| `AssetStream`                      | `*stream.EventStream[AssetStream]` | :heavy_minus_sign:                 | OK                                 |