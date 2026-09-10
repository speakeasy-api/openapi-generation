# NamespaceTests.Conflicts

## Overview

### Available Operations

* [get_namespace_conflict](#get_namespace_conflict) - Get Namespace Conflict Test
* [put_namespace_conflict](#put_namespace_conflict) - Put Property Name Conflicts Behind
* [create_namespace_conflict](#create_namespace_conflict) - Create Namespace Conflict Test
* [get_triple_namespace_conflict](#get_triple_namespace_conflict) - Get Triple Namespace Conflict Test
* [get_pet_owners](#get_pet_owners) - Get Pet Owners

## get_namespace_conflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references two different Pet types from different namespaces.
The SDK should properly import and alias both Pet types.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.conflicts.get_namespace_conflict()

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

        res = await sdk.namespace_tests.conflicts.get_namespace_conflict()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.NamespaceConflictTest](../../models/namespaceconflicttest.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## put_namespace_conflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="python" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.namespace_tests.conflicts.put_namespace_conflict()

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

        await sdk.namespace_tests.conflicts.put_namespace_conflict()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                                   | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `request`                                                                   | [models.ObjWithRenamedProperties](../../models/objwithrenamedproperties.md) | :heavy_check_mark:                                                          | The request object to use for the request.                                  |
| `retries`                                                                   | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)            | :heavy_minus_sign:                                                          | Configuration to override the default retry behavior of the client.         |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## create_namespace_conflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="python" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.conflicts.create_namespace_conflict(id="pet-foo-123", name="Fluffy", species="cat")

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

        res = await sdk.namespace_tests.conflicts.create_namespace_conflict(id="pet-foo-123", name="Fluffy", species="cat")

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

**[bar.Pet](../../models/bar/pet.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_triple_namespace_conflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="python" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.conflicts.get_triple_namespace_conflict()

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

        res = await sdk.namespace_tests.conflicts.get_triple_namespace_conflict()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.TripleNamespaceConflictTest](../../models/triplenamespaceconflicttest.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_pet_owners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="python" operationID="getPetOwners" method="get" path="/petOwners" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.conflicts.get_pet_owners()

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

        res = await sdk.namespace_tests.conflicts.get_pet_owners()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetPetOwnersResponse](../../models/getpetownersresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |