# NamespaceTests.Conflicts

## Overview

### Available Operations

* [GetNamespaceConflict](#getnamespaceconflict) - Get Namespace Conflict Test
* [PutNamespaceConflict](#putnamespaceconflict) - Put Property Name Conflicts Behind
* [CreateNamespaceConflict](#createnamespaceconflict) - Create Namespace Conflict Test
* [GetTripleNamespaceConflict](#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [GetPetOwners](#getpetowners) - Get Pet Owners

## GetNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references two different Pet types from different namespaces.
The SDK should properly import and alias both Pet types.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.NamespaceTests.Conflicts.GetNamespaceConflictAsync())
{
    // handle response
}


```

### Response

**[GetNamespaceConflictResponse](../../Models/GetNamespaceConflictResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## PutNamespaceConflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="unity" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
```csharp
using Speakeasy.OpenAPI;
using Speakeasy.OpenAPI.Baz;

var sdk = new SDK();

ObjWithRenamedProperties req = ;


using(var res = await sdk.NamespaceTests.Conflicts.PutNamespaceConflictAsync(req))
{
    // handle response
}


```

### Parameters

| Parameter                                                            | Type                                                                 | Required                                                             | Description                                                          |
| -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `request`                                                            | [ObjWithRenamedProperties](../../Models/ObjWithRenamedProperties.md) | :heavy_check_mark:                                                   | The request object to use for the request.                           |

### Response

**[PutNamespaceConflictResponse](../../Models/PutNamespaceConflictResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## CreateNamespaceConflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="unity" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
```csharp
using Speakeasy.OpenAPI;
using Speakeasy.OpenAPI.Foo;

var sdk = new SDK();

Speakeasy.OpenAPI.Foo.Pet req = new Speakeasy.OpenAPI.Foo.Pet() {
    Id = "pet-foo-123",
    Name = "Fluffy",
    Species = "cat",
};


using(var res = await sdk.NamespaceTests.Conflicts.CreateNamespaceConflictAsync(req))
{
    // handle response
}


```

### Parameters

| Parameter                                     | Type                                          | Required                                      | Description                                   |
| --------------------------------------------- | --------------------------------------------- | --------------------------------------------- | --------------------------------------------- |
| `request`                                     | [Speakeasy.OpenAPI.Foo.Pet](../../Foo/Pet.md) | :heavy_check_mark:                            | The request object to use for the request.    |

### Response

**[CreateNamespaceConflictResponse](../../Models/CreateNamespaceConflictResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetTripleNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.NamespaceTests.Conflicts.GetTripleNamespaceConflictAsync())
{
    // handle response
}


```

### Response

**[GetTripleNamespaceConflictResponse](../../Models/GetTripleNamespaceConflictResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetPetOwners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="unity" operationID="getPetOwners" method="get" path="/petOwners" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();


using(var res = await sdk.NamespaceTests.Conflicts.GetPetOwnersAsync())
{
    // handle response
}


```

### Response

**[GetPetOwnersResponse](../../Models/GetPetOwnersResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |