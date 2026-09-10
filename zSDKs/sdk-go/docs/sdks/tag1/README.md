# Tag1

## Overview

The first tag.

### Available Operations

* [~~Deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [GetRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
* [Auth](#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [ListTest1](#listtest1) - Get Test1
* [PostFileWithEncoding](#postfilewithencoding) - Post File With Encoding

## ~~Deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `GetRequestBodyFlattenedAway` instead.

### Example Usage

<!-- UsageSnippet language="go" operationID="deprecated1" method="get" path="/deprecated" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.Tag1.Deprecated1(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*Deprecated1Response](../../deprecated1response.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## Auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="go" operationID="auth" method="get" path="/auth" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"os"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.Tag1.Auth(ctx, examplealias.AuthSecurity{
        AccessToken: os.Getenv("SPEAKEASY_ACCESS_TOKEN"),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `security`                                            | [AuthSecurity](../../authsecurity.md)                 | :heavy_check_mark:                                    | The security requirements to use for the request.     |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*AuthResponse](../../authresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## ListTest1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="go" operationID="listTest1" method="get" path="/test1/{page}" -->
```go
package main

import(
	"context"
	"os"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New(
        examplealias.WithQueryParam1("some example query param"),
        examplealias.WithSecurity(examplealias.Security{
            MyAPIKey: &examplealias.MyAPIKey{
                MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
            },
        }),
    )

    res, err := s.Tag1.ListTest1(ctx, 100, examplealias.QueryParam2One, "some example header param")
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        for {
            // handle items

            res, err = res.Next()

            if err != nil {
                // handle error
            }

            if res == nil {
                break
            }
        }
    }
}
```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ctx`                                                                                                                                                                                                                                       | [context.Context](https://pkg.go.dev/context#Context)                                                                                                                                                                                       | :heavy_check_mark:                                                                                                                                                                                                                          | The context to use for the request.                                                                                                                                                                                                         |                                                                                                                                                                                                                                             |
| `page`                                                                                                                                                                                                                                      | `int64`                                                                                                                                                                                                                                     | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `queryParam2`                                                                                                                                                                                                                               | [QueryParam2](../../queryparam2.md)                                                                                                                                                                                                         | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `headerParam1`                                                                                                                                                                                                                              | `string`                                                                                                                                                                                                                                    | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `queryParam1`                                                                                                                                                                                                                               | `*string`                                                                                                                                                                                                                                   | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `opts`                                                                                                                                                                                                                                      | [][examplealias.Option](../../option.md)                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | The options for this request.                                                                                                                                                                                                               |                                                                                                                                                                                                                                             |

### Response

**[*ListTest1Response](../../listtest1response.md), error**

### Errors

| Error Type                           | Status Code                          | Content Type                         |
| ------------------------------------ | ------------------------------------ | ------------------------------------ |
| examplealias.BadRequestResponseError | 400                                  | application/json                     |
| examplealias.ErrorsError             | 500                                  | application/json                     |
| examplealias.SDKError                | 4XX, 5XX                             | \*/\*                                |

## PostFileWithEncoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="go" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"os"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    example, fileErr := os.Open("example.file")
    if fileErr != nil {
        panic(fileErr)
    }

    res, err := s.Tag1.PostFileWithEncoding(ctx, examplealias.PostFileWithEncodingRequest{
        File: examplealias.PostFileWithEncodingFile{
            FileName: "example.file",
            Content: example,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `ctx`                                                               | [context.Context](https://pkg.go.dev/context#Context)               | :heavy_check_mark:                                                  | The context to use for the request.                                 |
| `request`                                                           | [PostFileWithEncodingRequest](../../postfilewithencodingrequest.md) | :heavy_check_mark:                                                  | The request object to use for the request.                          |
| `opts`                                                              | [][examplealias.Option](../../option.md)                            | :heavy_minus_sign:                                                  | The options for this request.                                       |

### Response

**[*PostFileWithEncodingResponse](../../postfilewithencodingresponse.md), error**

### Errors

| Error Type               | Status Code              | Content Type             |
| ------------------------ | ------------------------ | ------------------------ |
| examplealias.ErrorsError | 415                      | application/json         |
| examplealias.SDKError    | 4XX, 5XX                 | \*/\*                    |