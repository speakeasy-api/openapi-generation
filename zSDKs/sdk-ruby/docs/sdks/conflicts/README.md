# NamespaceTests.Conflicts

## Overview

### Available Operations

* [get_namespace_conflict](#get_namespace_conflict) - Get Namespace Conflict Test
* [put_namespace_conflict](#put_namespace_conflict) - Put Property Name Conflicts Behind
* [create_namespace_conflict](#create_namespace_conflict) - Create Namespace Conflict Test
* [get_triple_namespace_conflict](#get_triple_namespace_conflict) - Get Triple Namespace Conflict Test
* [get_pet_owners](#get_pet_owners) - Get Pet Owners

## get_namespace_conflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references two different Pet types from different namespaces.
The SDK should properly import and alias both Pet types.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.conflicts.get_namespace_conflict

unless res.namespace_conflict_test.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamespaceConflictResponse)](../../models/operations/getnamespaceconflictresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## put_namespace_conflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = nil
res = s.namespace_tests.conflicts.put_namespace_conflict(request: req)

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                               | Type                                                                                    | Required                                                                                | Description                                                                             |
| --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `request`                                                                               | [Components::ObjWithRenamedProperties](../../models/shared/objwithrenamedproperties.md) | :heavy_check_mark:                                                                      | The request object to use for the request.                                              |

### Response

**[T.nilable(Operations::V2::Schemas::PutNamespaceConflictResponse)](../../models/operations/putnamespaceconflictresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## create_namespace_conflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::Foo::Pet.new(
  id: 'pet-foo-123',
  name: 'Fluffy',
  species: 'cat'
)
res = s.namespace_tests.conflicts.create_namespace_conflict(request: req)

unless res.pet.nil?
  # handle response
end

```

### Parameters

| Parameter                                          | Type                                               | Required                                           | Description                                        |
| -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- |
| `request`                                          | [Components::Foo::Pet](../../models/shared/pet.md) | :heavy_check_mark:                                 | The request object to use for the request.         |

### Response

**[T.nilable(Operations::V2::Schemas::CreateNamespaceConflictResponse)](../../models/operations/createnamespaceconflictresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_triple_namespace_conflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.conflicts.get_triple_namespace_conflict

unless res.triple_namespace_conflict_test.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetTripleNamespaceConflictResponse)](../../models/operations/gettriplenamespaceconflictresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_pet_owners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getPetOwners" method="get" path="/petOwners" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.conflicts.get_pet_owners

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetPetOwnersResponse)](../../models/operations/getpetownersresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |