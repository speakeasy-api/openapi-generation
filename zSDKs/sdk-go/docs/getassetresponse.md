# GetAssetResponse


## Fields

| Field                                    | Type                                     | Required                                 | Description                              |
| ---------------------------------------- | ---------------------------------------- | ---------------------------------------- | ---------------------------------------- |
| `HTTPMeta`                               | [HTTPMetadata](./httpmetadata.md)        | :heavy_check_mark:                       | N/A                                      |
| `AssetResult`                            | [*AssetResult](./assetresult.md)         | :heavy_minus_sign:                       | OK                                       |
| `AssetStatusStream`                      | `*stream.EventStream[AssetStatusStream]` | :heavy_minus_sign:                       | OK                                       |