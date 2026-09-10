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

<!-- UsageSnippet language="java" operationID="deprecated1" method="get" path="/deprecated" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.Deprecated1Response;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        Deprecated1Response res = sdk.obsolete().deprecated1()
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                      | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `serverURL`                    | *String*                       | :heavy_minus_sign:             | An optional server URL to use. |

### Response

**[Deprecated1Response](../../models/operations/Deprecated1Response.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |