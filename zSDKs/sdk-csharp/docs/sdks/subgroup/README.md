# Group.SubGroup

## Overview

### Available Operations

* [SubGroupOp](#subgroupop) - An operation at the group's top level

## SubGroupOp

An operation at the group's top level

### Example Usage

<!-- UsageSnippet language="csharp" operationID="subGroupOp" method="get" path="/group/subgroup" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Group.SubGroup.SubGroupOpAsync();

// handle response
```

### Response

**[SubGroupOpResponse](../../Models/SubGroupOpResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |