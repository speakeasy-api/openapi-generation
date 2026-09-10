# Tag1

## Overview

The first tag.

### Available Operations

* [~~deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
* [auth](#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [listTest1](#listtest1) - Get Test1
* [postFileWithEncoding](#postfilewithencoding) - Post File With Encoding

## ~~deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `getRequestBodyFlattenedAway` instead.

### Example Usage

<!-- UsageSnippet language="php" operationID="deprecated1" method="get" path="/deprecated" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->tag1->deprecated1(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                      | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `$serverURL`                   | *string*                       | :heavy_minus_sign:             | An optional server URL to use. |

### Response

**[?Deprecated1Response](../../Deprecated1Response.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="php" operationID="auth" method="get" path="/auth" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();


$requestSecurity = new OpenAPI\AuthSecurity(
    accessToken: '<YOUR_ACCESS_TOKEN_HERE>',
);

$response = $sdk->tag1->auth(
    security: $requestSecurity
);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                         | Type                                              | Required                                          | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- |
| `security`                                        | [AuthSecurity](../../AuthSecurity.md)             | :heavy_check_mark:                                | The security requirements to use for the request. |

### Response

**[?AuthResponse](../../AuthResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## listTest1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="php" operationID="listTest1" method="get" path="/test1/{page}" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setQueryParam1('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();



$responses = $sdk->tag1->listTest1(
    page: 100,
    queryParam2: OpenAPI\QueryParam2::One,
    headerParam1: 'some example header param'

);


foreach ($responses as $response) {
    if ($response->statusCode === 200) {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `page`                                                                                                                                                                                                                                      | *int*                                                                                                                                                                                                                                       | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `queryParam2`                                                                                                                                                                                                                               | [QueryParam2](../../QueryParam2.md)                                                                                                                                                                                                         | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `headerParam1`                                                                                                                                                                                                                              | *string*                                                                                                                                                                                                                                    | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `queryParam1`                                                                                                                                                                                                                               | *?string*                                                                                                                                                                                                                                   | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `$serverURL`                                                                                                                                                                                                                                | *string*                                                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | An optional server URL to use.                                                                                                                                                                                                              | http://localhost:8080                                                                                                                                                                                                                       |

### Response

**[?ListTest1Response](../../ListTest1Response.md)**

### Errors

| Error Type                          | Status Code                         | Content Type                        |
| ----------------------------------- | ----------------------------------- | ----------------------------------- |
| OpenAPI\BadRequestResponseException | 400                                 | application/json                    |
| OpenAPI\ErrorsError                 | 500                                 | application/json                    |
| OpenAPI\SDKException                | 4XX, 5XX                            | \*/\*                               |

## postFileWithEncoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="php" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\PostFileWithEncodingRequest(
    file: new OpenAPI\PostFileWithEncodingFile(
        fileName: 'example.file',
        content: file_get_contents('example.file');,
    ),
);

$response = $sdk->tag1->postFileWithEncoding(
    request: $request
);

if ($response->object !== null) {
    // handle response
}
```

### Parameters

| Parameter                                                                   | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `$request`                                                                  | [OpenAPI\PostFileWithEncodingRequest](../../PostFileWithEncodingRequest.md) | :heavy_check_mark:                                                          | The request object to use for the request.                                  |

### Response

**[?PostFileWithEncodingResponse](../../PostFileWithEncodingResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\ErrorsError  | 415                  | application/json     |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |