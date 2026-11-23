# PostFileWithEncodingRequest


## Fields

| Field                                                             | Type                                                              | Required                                                          | Description                                                       |
| ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- |
| `File`                                                            | [PostFileWithEncodingFile](./PostFileWithEncodingFile.md)         | :heavy_check_mark:                                                | The file to upload (supports CSV, PNG, JPEG, or PDF)              |
| `Attachment`                                                      | [Attachment](./Attachment.md)                                     | :heavy_minus_sign:                                                | An optional binary attachment                                     |
| `FileName`                                                        | *string*                                                          | :heavy_minus_sign:                                                | Optional custom file name                                         |
| `FilePurpose`                                                     | *string*                                                          | :heavy_minus_sign:                                                | Purpose of the file upload                                        |
| `Metadata`                                                        | [PostFileWithEncodingMetadata](./PostFileWithEncodingMetadata.md) | :heavy_minus_sign:                                                | JSON metadata about the file                                      |