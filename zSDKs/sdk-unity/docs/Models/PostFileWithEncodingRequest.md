# PostFileWithEncodingRequest


## Fields

| Field                                                                     | Type                                                                      | Required                                                                  | Description                                                               |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `File`                                                                    | [PostFileWithEncodingFile](../Models/PostFileWithEncodingFile.md)         | :heavy_check_mark:                                                        | The file to upload (supports CSV, PNG, JPEG, or PDF)                      |
| `Attachment`                                                              | [Attachment](../Models/Attachment.md)                                     | :heavy_minus_sign:                                                        | An optional binary attachment                                             |
| `FileName`                                                                | *string*                                                                  | :heavy_minus_sign:                                                        | Optional custom file name                                                 |
| `FilePurpose`                                                             | *string*                                                                  | :heavy_minus_sign:                                                        | Purpose of the file upload                                                |
| `Metadata`                                                                | [PostFileWithEncodingMetadata](../Models/PostFileWithEncodingMetadata.md) | :heavy_minus_sign:                                                        | JSON metadata about the file                                              |