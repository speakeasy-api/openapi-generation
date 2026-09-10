# NamespaceTests.SingleBar

## Overview

### Available Operations

* [GetSingleNamespaceBarPet](#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

## GetSingleNamespaceBarPet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.NamespaceTests.SingleBar.GetSingleNamespaceBarPetAsync())
{
    // handle response
}


```

### Response

**[GetSingleNamespaceBarPetResponse](../../Models/GetSingleNamespaceBarPetResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |