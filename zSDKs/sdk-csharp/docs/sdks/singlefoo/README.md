# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [GetSingleNamespaceFooPet](#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [CreateSingleNamespaceFooPet](#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

## GetSingleNamespaceFooPet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.NamespaceTests.SingleFoo.GetSingleNamespaceFooPetAsync();

// handle response
```

### Response

**[GetSingleNamespaceFooPetResponse](../../Models/GetSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## CreateSingleNamespaceFooPet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

Speakeasy.OpenAPI.Foo.Pet req = new Speakeasy.OpenAPI.Foo.Pet() {
    Id = "pet-foo-123",
    Name = "Fluffy",
    Species = "cat",
};

var res = await sdk.NamespaceTests.SingleFoo.CreateSingleNamespaceFooPetAsync(req);

// handle response
```

### Parameters

| Parameter                                     | Type                                          | Required                                      | Description                                   |
| --------------------------------------------- | --------------------------------------------- | --------------------------------------------- | --------------------------------------------- |
| `request`                                     | [Speakeasy.OpenAPI.Foo.Pet](../../Foo/Pet.md) | :heavy_check_mark:                            | The request object to use for the request.    |

### Response

**[CreateSingleNamespaceFooPetResponse](../../Models/CreateSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |