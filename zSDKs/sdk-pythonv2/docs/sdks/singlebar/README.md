# NamespaceTests.SingleBar

## Overview

### Available Operations

* [get_single_namespace_bar_pet](#get_single_namespace_bar_pet) - Get Single Namespace Bar Pet

## get_single_namespace_bar_pet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="python" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.single_bar.get_single_namespace_bar_pet()

    # Handle response
    print(res)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.namespace_tests.single_bar.get_single_namespace_bar_pet()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[bar.Pet](../../models/bar/pet.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |