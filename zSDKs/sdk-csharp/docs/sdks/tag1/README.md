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

<!-- UsageSnippet language="csharp" operationID="deprecated1" method="get" path="/deprecated" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Tag1.Deprecated1Async();

// handle response
```

### Parameters

| Parameter                      | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `serverURL`                    | *string*                       | :heavy_minus_sign:             | An optional server URL to use. |

### Response

**[Deprecated1Response](../../Models/Deprecated1Response.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## Auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="csharp" operationID="auth" method="get" path="/auth" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Tag1.AuthAsync(security: new AuthSecurity() {
    AccessToken = "<YOUR_ACCESS_TOKEN_HERE>",
});

// handle response
```

### Parameters

| Parameter                                         | Type                                              | Required                                          | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- |
| `security`                                        | [AuthSecurity](../../Models/AuthSecurity.md)      | :heavy_check_mark:                                | The security requirements to use for the request. |

### Response

**[AuthResponse](../../Models/AuthResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## ListTest1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="listTest1" method="get" path="/test1/{page}" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

ListTest1Response? res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"
);

while(res != null)
{
    // handle items

    res = await res.Next!();
}
```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Page`                                                                                                                                                                                                                                      | *long*                                                                                                                                                                                                                                      | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `QueryParam2`                                                                                                                                                                                                                               | [QueryParam2](../../Models/QueryParam2.md)                                                                                                                                                                                                  | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `HeaderParam1`                                                                                                                                                                                                                              | *string*                                                                                                                                                                                                                                    | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `QueryParam1`                                                                                                                                                                                                                               | *string*                                                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `serverURL`                                                                                                                                                                                                                                 | *string*                                                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | An optional server URL to use.                                                                                                                                                                                                              | http://localhost:8080                                                                                                                                                                                                                       |

### Response

**[ListTest1Response](../../Models/ListTest1Response.md)**

### Errors

| Error Type                  | Status Code                 | Content Type                |
| --------------------------- | --------------------------- | --------------------------- |
| BadRequestResponseException | 400                         | application/json            |
| ErrorsError                 | 500                         | application/json            |
| SDKException                | 4XX, 5XX                    | \*/\*                       |

## PostFileWithEncoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileWithEncodingRequest req = new PostFileWithEncodingRequest() {
    File = new PostFileWithEncodingFile() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.Tag1.PostFileWithEncodingAsync(req);

// handle response
```

### Parameters

| Parameter                                                                  | Type                                                                       | Required                                                                   | Description                                                                |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `request`                                                                  | [PostFileWithEncodingRequest](../../Models/PostFileWithEncodingRequest.md) | :heavy_check_mark:                                                         | The request object to use for the request.                                 |

### Response

**[PostFileWithEncodingResponse](../../Models/PostFileWithEncodingResponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| ErrorsError      | 415              | application/json |
| SDKException     | 4XX, 5XX         | \*/\*            |