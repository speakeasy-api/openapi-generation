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

        Deprecated1Response res = sdk.tag1().deprecated1()
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

## auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="java" operationID="auth" method="get" path="/auth" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.AuthResponse;
import org.openapis.openapi.models.operations.AuthSecurity;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        AuthResponse res = sdk.tag1().auth()
                .security(AuthSecurity.builder()
                    .accessToken(System.getenv().getOrDefault("ACCESS_TOKEN", ""))
                    .build())
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                                      | Type                                                                                           | Required                                                                                       | Description                                                                                    |
| ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `security`                                                                                     | [org.openapis.openapi.models.operations.AuthSecurity](../../models/operations/AuthSecurity.md) | :heavy_check_mark:                                                                             | The security requirements to use for the request.                                              |

### Response

**[AuthResponse](../../models/operations/AuthResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## listTest1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="java" operationID="listTest1" method="get" path="/test1/{page}" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Exception {

        SDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                    .myApiKey(MyApiKey.builder()
                        .myApiKey("<value>")
                        .build())
                    .build())
            .build();


        sdk.tag1().listTest1()
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param")
                .callAsStreamUnwrapped()
                .forEach((ResultArray item) -> {
                   // handle item
                });

    }
}
```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `page`                                                                                                                                                                                                                                      | *long*                                                                                                                                                                                                                                      | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `queryParam1`                                                                                                                                                                                                                               | @Nullable *String*                                                                                                                                                                                                                          | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `queryParam2`                                                                                                                                                                                                                               | [QueryParam2](../../models/operations/QueryParam2.md)                                                                                                                                                                                       | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `headerParam1`                                                                                                                                                                                                                              | *String*                                                                                                                                                                                                                                    | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `serverURL`                                                                                                                                                                                                                                 | *String*                                                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | An optional server URL to use.                                                                                                                                                                                                              | http://localhost:8080                                                                                                                                                                                                                       |

### Response

**[ListTest1Response](../../models/operations/ListTest1Response.md)**

### Errors

| Error Type                                | Status Code                               | Content Type                              |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| models/errors/BadRequestResponseException | 400                                       | application/json                          |
| models/errors/Error                       | 500                                       | application/json                          |
| models/errors/SDKException                | 4XX, 5XX                                  | \*/\*                                     |

## postFileWithEncoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="java" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.File;
import org.openapis.openapi.models.operations.PostFileWithEncodingRequest;
import org.openapis.openapi.models.operations.PostFileWithEncodingResponse;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder()
            .build();

        PostFileWithEncodingRequest req = PostFileWithEncodingRequest.builder()
                .file(File.builder()
                    .fileName("example.file")
                    .content(Blob.from(Paths.get("example.file")))
                    .build())
                .build();

        PostFileWithEncodingResponse res = sdk.tag1().postFileWithEncoding()
                .request(req)
                .call();

    }
}
```

### Parameters

| Parameter                                                                             | Type                                                                                  | Required                                                                              | Description                                                                           |
| ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| `request`                                                                             | [PostFileWithEncodingRequest](../../models/operations/PostFileWithEncodingRequest.md) | :heavy_check_mark:                                                                    | The request object to use for the request.                                            |

### Response

**[PostFileWithEncodingResponse](../../models/operations/PostFileWithEncodingResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/Error        | 415                        | application/json           |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |