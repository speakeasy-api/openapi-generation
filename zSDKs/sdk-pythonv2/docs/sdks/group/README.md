# Group

## Overview

### Available Operations

* [root_group_op](#root_group_op) - An operation at the group's root level

## root_group_op

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="python" operationID="rootGroupOp" method="get" path="/group/root" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.group.root_group_op()

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

        await sdk.group.root_group_op()

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