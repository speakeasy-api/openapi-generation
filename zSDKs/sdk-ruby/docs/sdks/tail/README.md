# Group.SubGroup.Empty.Tail

## Overview

### Available Operations

* [nested_group_op](#nested_group_op) - An operation at the group's deepest level

## nested_group_op

Notice that 'group.flattened' has no operations.


### Example Usage

<!-- UsageSnippet language="ruby" operationID="nestedGroupOp" method="get" path="/group/nested" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.group.sub_group.empty.tail.nested_group_op

if res.status_code == 200
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::NestedGroupOpResponse)](../../models/operations/nestedgroupopresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |