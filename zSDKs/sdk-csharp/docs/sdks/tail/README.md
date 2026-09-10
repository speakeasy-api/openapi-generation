# Group.SubGroup.Empty.Tail

## Overview

### Available Operations

* [NestedGroupOp](#nestedgroupop) - An operation at the group's deepest level

## NestedGroupOp

Notice that 'group.flattened' has no operations.


### Example Usage

<!-- UsageSnippet language="csharp" operationID="nestedGroupOp" method="get" path="/group/nested" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Group.SubGroup.Empty.Tail.NestedGroupOpAsync();

// handle response
```

### Response

**[NestedGroupOpResponse](../../Models/NestedGroupOpResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |