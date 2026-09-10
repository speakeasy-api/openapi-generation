# Group

## Overview

### Available Operations

* [root_group_op](#root_group_op) - An operation at the group's root level

## root_group_op

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="ruby" operationID="rootGroupOp" method="get" path="/group/root" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.group.root_group_op

if res.status_code == 200
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::RootGroupOpResponse)](../../models/operations/rootgroupopresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |