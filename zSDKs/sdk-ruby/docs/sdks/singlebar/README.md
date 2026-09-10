# NamespaceTests.SingleBar

## Overview

### Available Operations

* [get_single_namespace_bar_pet](#get_single_namespace_bar_pet) - Get Single Namespace Bar Pet

## get_single_namespace_bar_pet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.namespace_tests.single_bar.get_single_namespace_bar_pet

unless res.pet.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetSingleNamespaceBarPetResponse)](../../models/operations/getsinglenamespacebarpetresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |