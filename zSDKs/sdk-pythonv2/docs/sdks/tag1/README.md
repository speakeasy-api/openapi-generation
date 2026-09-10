# Tag1

## Overview

The first tag.

### Available Operations

* [~~deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away) instead.
* [auth](#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [list_test1](#list_test1) - Get Test1
* [post_file_with_encoding](#post_file_with_encoding) - Post File With Encoding

## ~~deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `get_request_body_flattened_away` instead.

### Example Usage

<!-- UsageSnippet language="python" operationID="deprecated1" method="get" path="/deprecated" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.tag1.deprecated1()

    # Use the SDK ...
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        await sdk.tag1.deprecated1()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |
| `server_url`                                                        | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | An optional server URL to use.                                      |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="python" operationID="auth" method="get" path="/auth" -->
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.tag1.auth(security=speakeasy.new_openapi.AuthSecurity(
        access_token=os.getenv("SPEAKEASY_ACCESS_TOKEN", ""),
    ))

    # Use the SDK ...
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        await sdk.tag1.auth(security=speakeasy.new_openapi.AuthSecurity(
            access_token=os.getenv("SPEAKEASY_ACCESS_TOKEN", ""),
        ))

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `security`                                                          | [models.AuthSecurity](../../authsecurity.md)                        | :heavy_check_mark:                                                  | The security requirements to use for the request.                   |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## list_test1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="python" operationID="listTest1" method="get" path="/test1/{page}" -->
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    query_param1="some example query param",
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

    res = sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

    while res is not None:
        # Handle items

        res = res.next()
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        query_param1="some example query param",
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

        while res is not None:
            # Handle items

            res = res.next()

asyncio.run(main())
```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `query_param2`                                                                                                                                                                                                                              | [models.QueryParam2](../../models/queryparam2.md)                                                                                                                                                                                           | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `page`                                                                                                                                                                                                                                      | *int*                                                                                                                                                                                                                                       | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `header_param1`                                                                                                                                                                                                                             | *str*                                                                                                                                                                                                                                       | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `query_param1`                                                                                                                                                                                                                              | *Optional[str]*                                                                                                                                                                                                                             | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `retries`                                                                                                                                                                                                                                   | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                                                                                                                                                                            | :heavy_minus_sign:                                                                                                                                                                                                                          | Configuration to override the default retry behavior of the client.                                                                                                                                                                         |                                                                                                                                                                                                                                             |
| `server_url`                                                                                                                                                                                                                                | *Optional[str]*                                                                                                                                                                                                                             | :heavy_minus_sign:                                                                                                                                                                                                                          | An optional server URL to use.                                                                                                                                                                                                              | http://localhost:8080                                                                                                                                                                                                                       |

### Response

**[models.ListTest1Response](../../models/listtest1response.md)**

### Errors

| Error Type                     | Status Code                    | Content Type                   |
| ------------------------------ | ------------------------------ | ------------------------------ |
| models.BadRequestResponseError | 400                            | application/json               |
| models.ErrorsError             | 500                            | application/json               |
| models.SDKError                | 4XX, 5XX                       | \*/\*                          |

## post_file_with_encoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="python" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.tag1.post_file_with_encoding(file={
        "file_name": "example.file",
        "content": open("example.file", "rb"),
    })

    # Handle response
    print(res)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.tag1.post_file_with_encoding(file={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                                                     | Type                                                                                          | Required                                                                                      | Description                                                                                   |
| --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `file`                                                                                        | [models.PostFileWithEncodingFile](../../models/postfilewithencodingfile.md)                   | :heavy_check_mark:                                                                            | The file to upload (supports CSV, PNG, JPEG, or PDF)                                          |
| `attachment`                                                                                  | [Optional[models.Attachment]](../../models/attachment.md)                                     | :heavy_minus_sign:                                                                            | An optional binary attachment                                                                 |
| `file_name`                                                                                   | *Optional[str]*                                                                               | :heavy_minus_sign:                                                                            | Optional custom file name                                                                     |
| `file_purpose`                                                                                | *Optional[str]*                                                                               | :heavy_minus_sign:                                                                            | Purpose of the file upload                                                                    |
| `metadata`                                                                                    | [Optional[models.PostFileWithEncodingMetadata]](../../models/postfilewithencodingmetadata.md) | :heavy_minus_sign:                                                                            | JSON metadata about the file                                                                  |
| `retries`                                                                                     | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                              | :heavy_minus_sign:                                                                            | Configuration to override the default retry behavior of the client.                           |

### Response

**[models.PostFileWithEncodingResponse](../../models/postfilewithencodingresponse.md)**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| models.ErrorsError | 415                | application/json   |
| models.SDKError    | 4XX, 5XX           | \*/\*              |