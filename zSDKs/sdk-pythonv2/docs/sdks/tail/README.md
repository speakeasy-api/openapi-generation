# Group.SubGroup.Empty.Tail

## Overview

### Available Operations

* [nested_group_op](#nested_group_op) - An operation at the group's deepest level

## nested_group_op

Notice that 'group.flattened' has no operations.


### Example Usage

<!-- UsageSnippet language="python" operationID="nestedGroupOp" method="get" path="/group/nested" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.group.sub_group.empty.tail.nested_group_op()

    # Use the SDK ...
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        await sdk.group.sub_group.empty.tail.nested_group_op()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |