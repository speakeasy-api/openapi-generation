# AssetResult

The result of a render request.


## Fields

| Field                                                                                               | Type                                                                                                | Required                                                                                            | Description                                                                                         |
| --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `id`                                                                                                | *T.nilable(::String)*                                                                               | :heavy_minus_sign:                                                                                  | N/A                                                                                                 |
| `status`                                                                                            | [T.nilable(Components::Status)](../models/shared/status.md)                                         | :heavy_minus_sign:                                                                                  | N/A                                                                                                 |
| `steps`                                                                                             | T::Array<[T.any(Components::AssetNoteStep, Components::AssetOutputStep)](../models/shared/step.md)> | :heavy_minus_sign:                                                                                  | N/A                                                                                                 |