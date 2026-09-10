# Group.SubGroup

## Overview

### Available Operations

* [sub_group_op](#sub_group_op) - An operation at the group's top level

## sub_group_op

An operation at the group's top level

### Example Usage

<!-- UsageSnippet language="ruby" operationID="subGroupOp" method="get" path="/group/subgroup" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.group.sub_group.sub_group_op

if res.status_code == 200
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::SubGroupOpResponse)](../../models/operations/subgroupopresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |