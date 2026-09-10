# GetAssetResponse


## Fields

| Field                                     | Type                                      | Required                                  | Description                               |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| `HttpMeta`                                | [HTTPMetadata](../Models/HTTPMetadata.md) | :heavy_check_mark:                        | N/A                                       |
| `AssetResult`                             | [AssetResult](../Models/AssetResult.md)   | :heavy_minus_sign:                        | OK                                        |
| `AssetStatusStream`                       | *EventStream<AssetStatusStream>*          | :heavy_minus_sign:                        | OK                                        |