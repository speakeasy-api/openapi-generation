# Group

## Overview

### Available Operations

* [RootGroupOp](#rootgroupop) - An operation at the group's root level

## RootGroupOp

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="csharp" operationID="rootGroupOp" method="get" path="/group/root" -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

var res = await sdk.Group.RootGroupOpAsync();

// handle response
```

### Response

**[RootGroupOpResponse](../../Models/RootGroupOpResponse.md)**

### Errors

| Error Type   | Status Code  | Content Type |
| ------------ | ------------ | ------------ |
| SDKException | 4XX, 5XX     | \*/\*        |