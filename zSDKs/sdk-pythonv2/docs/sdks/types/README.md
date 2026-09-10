# NamespaceTests.Types

## Overview

### Available Operations

* [get_namespace_types](#get_namespace_types) - Get Namespace Types Test
* [get_namespace_animal](#get_namespace_animal) - Get Namespace Animal (Discriminated Union)
* [get_namespace_vehicle](#get_namespace_vehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [get_namespace_organization](#get_namespace_organization) - Get Namespace Organization (Nested Inline Schemas)

## get_namespace_types

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.types.get_namespace_types()

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

        res = await sdk.namespace_tests.types.get_namespace_types()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.NamespaceTypesTest](../../models/namespacetypestest.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_namespace_animal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.types.get_namespace_animal()

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

        res = await sdk.namespace_tests.types.get_namespace_animal()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[foo.Animal](../../models/foo/animal.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_namespace_vehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.types.get_namespace_vehicle()

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

        res = await sdk.namespace_tests.types.get_namespace_vehicle()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[bar.Vehicle](../../models/bar/vehicle.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_namespace_organization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.namespace_tests.types.get_namespace_organization()

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

        res = await sdk.namespace_tests.types.get_namespace_organization()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[foo.Organization](../../models/foo/organization.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |