# NamespaceTests.Types

## Overview

### Available Operations

* [GetNamespaceTypes](#getnamespacetypes) - Get Namespace Types Test
* [GetNamespaceAnimal](#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [GetNamespaceVehicle](#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [GetNamespaceOrganization](#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

## GetNamespaceTypes

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.NamespaceTests.Types.GetNamespaceTypesAsync();

// handle response
```

### Response

**[GetNamespaceTypesResponse](../../Models/GetNamespaceTypesResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetNamespaceAnimal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.NamespaceTests.Types.GetNamespaceAnimalAsync();

// handle response
```

### Response

**[GetNamespaceAnimalResponse](../../Models/GetNamespaceAnimalResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetNamespaceVehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.NamespaceTests.Types.GetNamespaceVehicleAsync();

// handle response
```

### Response

**[GetNamespaceVehicleResponse](../../Models/GetNamespaceVehicleResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |

## GetNamespaceOrganization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="csharp" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.NamespaceTests.Types.GetNamespaceOrganizationAsync();

// handle response
```

### Response

**[GetNamespaceOrganizationResponse](../../Models/GetNamespaceOrganizationResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |