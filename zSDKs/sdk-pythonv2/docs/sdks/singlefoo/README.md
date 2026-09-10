# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [get_single_namespace_foo_pet](#get_single_namespace_foo_pet) - Get Single Namespace Foo Pet
* [create_single_namespace_foo_pet](#create_single_namespace_foo_pet) - Create Single Namespace Foo Pet

## get_single_namespace_foo_pet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="python" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.single_foo.get_single_namespace_foo_pet()

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

        res = await sdk.namespace_tests.single_foo.get_single_namespace_foo_pet()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[foo.Pet](../../models/foo/pet.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## create_single_namespace_foo_pet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="python" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.single_foo.create_single_namespace_foo_pet(id="pet-foo-123", name="Fluffy", species="cat")

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

        res = await sdk.namespace_tests.single_foo.create_single_namespace_foo_pet(id="pet-foo-123", name="Fluffy", species="cat")

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         | Example                                                             |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `id`                                                                | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 | pet-foo-123                                                         |
| `name`                                                              | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 | Fluffy                                                              |
| `species`                                                           | *str*                                                               | :heavy_check_mark:                                                  | The species of the pet (e.g., dog, cat)                             | cat                                                                 |
| `models`                                                            | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | A field name that often collides                                    |                                                                     |
| `request`                                                           | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | A field name that often collides                                    |                                                                     |
| `operations`                                                        | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | A field name that often collides                                    |                                                                     |
| `errors`                                                            | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | A field name that often collides                                    |                                                                     |
| `utils`                                                             | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | A field name that often collides                                    |                                                                     |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |                                                                     |

### Response

**[foo.Pet](../../models/foo/pet.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |