# NamespaceTests.Types

## Overview

### Available Operations

* [get_namespace_types](#get_namespace_types) - Get Namespace Types Test
* [get_namespace_animal](#get_namespace_animal) - Get Namespace Animal (Discriminated Union)
* [get_namespace_vehicle](#get_namespace_vehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [get_namespace_organization](#get_namespace_organization) - Get Namespace Organization (Nested Inline Schemas)

## get_namespace_types

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.types.get_namespace_types

unless res.namespace_types_test.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamespaceTypesResponse)](../../models/operations/getnamespacetypesresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_namespace_animal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.types.get_namespace_animal

unless res.animal.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamespaceAnimalResponse)](../../models/operations/getnamespaceanimalresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_namespace_vehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.types.get_namespace_vehicle

unless res.vehicle.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamespaceVehicleResponse)](../../models/operations/getnamespacevehicleresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_namespace_organization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.types.get_namespace_organization

unless res.organization.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamespaceOrganizationResponse)](../../models/operations/getnamespaceorganizationresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |