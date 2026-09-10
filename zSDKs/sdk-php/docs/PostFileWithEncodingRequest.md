# PostFileWithEncodingRequest


## Fields

| Field                                                              | Type                                                               | Required                                                           | Description                                                        |
| ------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `file`                                                             | [PostFileWithEncodingFile](./PostFileWithEncodingFile.md)          | :heavy_check_mark:                                                 | The file to upload (supports CSV, PNG, JPEG, or PDF)               |
| `attachment`                                                       | [?Attachment](./Attachment.md)                                     | :heavy_minus_sign:                                                 | An optional binary attachment                                      |
| `fileName`                                                         | *?string*                                                          | :heavy_minus_sign:                                                 | Optional custom file name                                          |
| `filePurpose`                                                      | *?string*                                                          | :heavy_minus_sign:                                                 | Purpose of the file upload                                         |
| `metadata`                                                         | [?PostFileWithEncodingMetadata](./PostFileWithEncodingMetadata.md) | :heavy_minus_sign:                                                 | JSON metadata about the file                                       |