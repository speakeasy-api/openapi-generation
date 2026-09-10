# ~~Obsolete~~

> [!WARNING]
> This SDK is **DEPRECATED**

## Overview

A subSDK in which all operations are deprecated.

### Available Operations

* [~~Deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [GetRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

## ~~Deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `GetRequestBodyFlattenedAway` instead.

### Example Usage

<!-- UsageSnippet language="unity" operationID="deprecated1" method="get" path="/deprecated" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.Obsolete.Deprecated1Async())
{
    // handle response
}


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