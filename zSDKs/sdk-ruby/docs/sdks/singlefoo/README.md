# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [get_single_namespace_foo_pet](#get_single_namespace_foo_pet) - Get Single Namespace Foo Pet
* [create_single_namespace_foo_pet](#create_single_namespace_foo_pet) - Create Single Namespace Foo Pet

## get_single_namespace_foo_pet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.single_foo.get_single_namespace_foo_pet

unless res.pet.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetSingleNamespaceFooPetResponse)](../../models/operations/getsinglenamespacefoopetresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## create_single_namespace_foo_pet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::Foo::Pet.new(
  id: 'pet-foo-123',
  name: 'Fluffy',
  species: 'cat'
)
res = s.namespace_tests.single_foo.create_single_namespace_foo_pet(request: req)

unless res.pet.nil?
  # handle response
end

```

### Parameters

| Parameter                                          | Type                                               | Required                                           | Description                                        |
| -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- |
| `request`                                          | [Components::Foo::Pet](../../models/shared/pet.md) | :heavy_check_mark:                                 | The request object to use for the request.         |

### Response

**[T.nilable(Operations::V2::Schemas::CreateSingleNamespaceFooPetResponse)](../../models/operations/createsinglenamespacefoopetresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |