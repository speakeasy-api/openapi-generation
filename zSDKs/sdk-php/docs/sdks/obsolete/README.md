# ~~Obsolete~~

> [!WARNING]
> This SDK is **DEPRECATED**

## Overview

A subSDK in which all operations are deprecated.

### Available Operations

* [~~deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

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



$response = $sdk->obsolete->deprecated1(

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