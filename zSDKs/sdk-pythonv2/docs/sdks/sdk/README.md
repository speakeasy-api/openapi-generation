# SDK

## Overview

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

Speakeasy Docs
<https://speakeasy.com/docs>

### Available Operations

* [operation_with_leading_and_trailing_underscores_](#operation_with_leading_and_trailing_underscores_)
* [post_file](#post_file) - Post File
* [get_polymorphism](#get_polymorphism)
* [get_union_errors](#get_union_errors)
* [get_request_body_flattened_away](#get_request_body_flattened_away)
* [get_fully_flattened_request](#get_fully_flattened_request)
* [create_with_union](#create_with_union) - Create with discriminated union request body
* [test_endpoint](#test_endpoint)
* [create_user](#create_user) - Create User
* [get_user](#get_user) - Get User
* [update_user](#update_user) - Update User
* [delete_user](#delete_user) - Delete User
* [login](#login) - Login
* [validate](#validate) - Validate
* [chat](#chat)
* [get_binary_default_response](#get_binary_default_response)
* [test_enum_formats](#test_enum_formats) - Test x-speakeasy-enums in different formats
* [binary_and_string_upload](#binary_and_string_upload)
* [get_error_in_union](#get_error_in_union)
* [get_duplicate_export_collision](#get_duplicate_export_collision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [get_named_primitive_union](#get_named_primitive_union) - Test named primitive union options using title and x-speakeasy-name-override
* [get_empty_object_error](#get_empty_object_error) - Get Empty Object Error
* [url_validation_stress_test](#url_validation_stress_test)
* [parentheses_in_path_allowed](#parentheses_in_path_allowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [get_nested_integer_string](#get_nested_integer_string) - Test nested struct with integer:string tag
* [render_asset](#render_asset) - Render Asset
* [get_asset](#get_asset) - Get Asset
* [get_error_only_example](#get_error_only_example) - Operation with example only on error response

## operation_with_leading_and_trailing_underscores_

### Example Usage

<!-- UsageSnippet language="python" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.operation_with_leading_and_trailing_underscores_(qp1="renamed")

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

        await sdk.operation_with_leading_and_trailing_underscores_(qp1="renamed")

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `qp1`                                                                                                                                            | *str*                                                                                                                                            | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |
| `retries`                                                                                                                                        | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                                                                                 | :heavy_minus_sign:                                                                                                                               | Configuration to override the default retry behavior of the client.                                                                              |                                                                                                                                                  |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## post_file

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="python" operationID="postFile" method="post" path="/file" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.post_file(upload={
        "file_name": "example.file",
        "content": open("example.file", "rb"),
    })

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

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `upload`                                                            | [models.File](../../models/file.md)                                 | :heavy_check_mark:                                                  | A file                                                              |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[httpx.Response](../../models/file.md)**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| models.ErrorsError | 415, 4XX           | application/json   |
| models.ErrorsError | 5XX                | application/json   |

## get_polymorphism

### Example Usage

<!-- UsageSnippet language="python" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_polymorphism()

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

        res = await sdk.get_polymorphism()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetPolymorphismResponse](../../models/getpolymorphismresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_union_errors

### Example Usage

<!-- UsageSnippet language="python" operationID="getUnionErrors" method="get" path="/unionErrors" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_union_errors(page=12)

    while res is not None:
        # Handle items

        res = res.next()
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.get_union_errors(page=12)

        while res is not None:
            # Handle items

            res = res.next()

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         | Example                                                             |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `page`                                                              | *int*                                                               | :heavy_check_mark:                                                  | N/A                                                                 | 12                                                                  |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |                                                                     |

### Response

**[models.GetUnionErrorsResponse](../../models/getunionerrorsresponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| models.ErrorsError  | 404                 | application/json    |
| models.ErrorType1   | 500                 | application/json    |
| models.ErrorType2   | 500                 | application/json    |
| models.TaggedError1 | 4XX                 | application/json    |
| models.TaggedError2 | 4XX                 | application/json    |
| models.SDKError     | 5XX                 | \*/\*               |

## get_request_body_flattened_away

### Example Usage

<!-- UsageSnippet language="python" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK(
    lone_query_param="<value>",
) as sdk:

    sdk.get_request_body_flattened_away()

    # Use the SDK ...
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        lone_query_param="<value>",
    ) as sdk:

        await sdk.get_request_body_flattened_away()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                                                       | Type                                                                                            | Required                                                                                        | Description                                                                                     |
| ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `request`                                                                                       | [models.GetRequestBodyFlattenedAwayRequest](../../models/getrequestbodyflattenedawayrequest.md) | :heavy_check_mark:                                                                              | The request object to use for the request.                                                      |
| `retries`                                                                                       | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                                | :heavy_minus_sign:                                                                              | Configuration to override the default retry behavior of the client.                             |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_fully_flattened_request

### Example Usage

<!-- UsageSnippet language="python" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        option6=speakeasy.new_openapi.SecurityOption6(
            client_id=os.getenv("SPEAKEASY_CLIENT_ID", ""),
            client_secret=os.getenv("SPEAKEASY_CLIENT_SECRET", ""),
            "token_url": "/clientcredentials/token",
        ),
    ),
) as sdk:

    sdk.get_fully_flattened_request(lang="en", name="<value>")

    # Use the SDK ...
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            option6=speakeasy.new_openapi.SecurityOption6(
                client_id=os.getenv("SPEAKEASY_CLIENT_ID", ""),
                client_secret=os.getenv("SPEAKEASY_CLIENT_SECRET", ""),
                "token_url": "/clientcredentials/token",
            ),
        ),
    ) as sdk:

        await sdk.get_fully_flattened_request(lang="en", name="<value>")

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `lang`                                                              | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `name`                                                              | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `max_length`                                                        | *Optional[int]*                                                     | :heavy_minus_sign:                                                  | N/A                                                                 |
| `emoji`                                                             | [Optional[models.Emoji]](../../models/emoji.md)                     | :heavy_minus_sign:                                                  | N/A                                                                 |
| `gif`                                                               | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | N/A                                                                 |
| `reply`                                                             | *Optional[bool]*                                                    | :heavy_minus_sign:                                                  | N/A                                                                 |
| `private`                                                           | *Optional[bool]*                                                    | :heavy_minus_sign:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## create_with_union

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="python" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.create_with_union(name="<value>", shape={
        "type": "rectangle",
        "width": 3125.73,
        "height": 922.51,
    })

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

        res = await sdk.create_with_union(name="<value>", shape={
            "type": "rectangle",
            "width": 3125.73,
            "height": 922.51,
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `name`                                                              | *str*                                                               | :heavy_check_mark:                                                  | Name for the shape                                                  |
| `shape`                                                             | [models.Shape](../../models/shape.md)                               | :heavy_check_mark:                                                  | A discriminated union of shape types                                |
| `dry_run`                                                           | *Optional[bool]*                                                    | :heavy_minus_sign:                                                  | If true, validates without creating                                 |
| `description`                                                       | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | Optional description                                                |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.CreateWithUnionResponse](../../models/createwithunionresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## test_endpoint

### Example Usage

<!-- UsageSnippet language="python" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.test_endpoint(test_name="<value>", test="<value>")

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

        await sdk.test_endpoint(test_name="<value>", test="<value>")

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `test_name`                                                         | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `test`                                                              | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## create_user

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="python" operationID="createUser" method="put" path="/user" example="paired-example" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.create_user(email="paired@example.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", first_name="John", gender="other")

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

        res = await sdk.create_user(email="paired@example.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", first_name="John", gender="other")

        # Handle response
        print(res)

asyncio.run(main())
```
### Example Usage: request-only

<!-- UsageSnippet language="python" operationID="createUser" method="put" path="/user" example="request-only" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.create_user(email="request-only@example.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", gender="other")

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

        res = await sdk.create_user(email="request-only@example.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", gender="other")

        # Handle response
        print(res)

asyncio.run(main())
```
### Example Usage: response-only

<!-- UsageSnippet language="python" operationID="createUser" method="put" path="/user" example="response-only" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.create_user(email="Virginie47@gmail.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", gender="other")

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

        res = await sdk.create_user(email="Virginie47@gmail.com", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", gender="other")

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                                         | Type                                                                              | Required                                                                          | Description                                                                       | Example                                                                           |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `email`                                                                           | *str*                                                                             | :heavy_check_mark:                                                                | N/A                                                                               |                                                                                   |
| `id`                                                                              | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               | 8ffac18c-7d88-4879-b057-e5f45b9ce7de                                              |
| `first_name`                                                                      | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `last_name`                                                                       | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `age`                                                                             | *Optional[float]*                                                                 | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `postal_code`                                                                     | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `gender`                                                                          | [Optional[models.Gender]](../../models/gender.md)                                 | :heavy_minus_sign:                                                                | N/A                                                                               | other                                                                             |
| `associated_ids`                                                                  | List[*str*]                                                                       | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `metadata`                                                                        | [Optional[models.Metadata]](../../models/metadata.md)                             | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `metadata_any`                                                                    | [Optional[models.MetadataAny]](../../models/metadataany.md)                       | :heavy_minus_sign:                                                                | A metadata object with additionalProperties true (any type) - should be flattened |                                                                                   |
| `retries`                                                                         | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                  | :heavy_minus_sign:                                                                | Configuration to override the default retry behavior of the client.               |                                                                                   |

### Response

**[models.User](../../models/user.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_user

Get User

### Example Usage

<!-- UsageSnippet language="python" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_user(id="<id>")

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

        res = await sdk.get_user(id="<id>")

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `id`                                                                | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.User](../../models/user.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## update_user

Update User

### Example Usage

<!-- UsageSnippet language="python" operationID="updateUser" method="post" path="/user/{id}" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.update_user(id_param="<value>", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", email="Joanny.Feeney@gmail.com", gender="other")

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

        res = await sdk.update_user(id_param="<value>", id="8ffac18c-7d88-4879-b057-e5f45b9ce7de", email="Joanny.Feeney@gmail.com", gender="other")

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                                         | Type                                                                              | Required                                                                          | Description                                                                       | Example                                                                           |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `id_param`                                                                        | *str*                                                                             | :heavy_check_mark:                                                                | N/A                                                                               |                                                                                   |
| `id`                                                                              | *str*                                                                             | :heavy_check_mark:                                                                | N/A                                                                               | 8ffac18c-7d88-4879-b057-e5f45b9ce7de                                              |
| `email`                                                                           | *str*                                                                             | :heavy_check_mark:                                                                | N/A                                                                               |                                                                                   |
| `first_name`                                                                      | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `last_name`                                                                       | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `age`                                                                             | *Optional[float]*                                                                 | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `postal_code`                                                                     | *Optional[str]*                                                                   | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `gender`                                                                          | [Optional[models.Gender]](../../models/gender.md)                                 | :heavy_minus_sign:                                                                | N/A                                                                               | other                                                                             |
| `associated_ids`                                                                  | List[*str*]                                                                       | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `metadata`                                                                        | [Optional[models.Metadata]](../../models/metadata.md)                             | :heavy_minus_sign:                                                                | N/A                                                                               |                                                                                   |
| `metadata_any`                                                                    | [Optional[models.MetadataAny]](../../models/metadataany.md)                       | :heavy_minus_sign:                                                                | A metadata object with additionalProperties true (any type) - should be flattened |                                                                                   |
| `retries`                                                                         | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                  | :heavy_minus_sign:                                                                | Configuration to override the default retry behavior of the client.               |                                                                                   |

### Response

**[models.User](../../models/user.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## delete_user

Delete User

### Example Usage

<!-- UsageSnippet language="python" operationID="deleteUser" method="delete" path="/user/{id}" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.delete_user(id="<id>")

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

        await sdk.delete_user(id="<id>")

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `id`                                                                | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## login

Login

### Example Usage

<!-- UsageSnippet language="python" operationID="login" method="get" path="/auth/login" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.login()

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

        res = await sdk.login()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.LoginResponse](../../models/loginresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## validate

Validate

### Example Usage

<!-- UsageSnippet language="python" operationID="validate" method="get" path="/auth/validate" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.validate()

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

        res = await sdk.validate()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.ValidateResponse](../../models/validateresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## chat

### Example Usage

<!-- UsageSnippet language="python" operationID="chat" method="post" path="/chat" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.chat(request={
        "model": "review-model",
        "prompt": "What is the largest city in the world?",
        "stream": False,
    })

    with res as event_stream:
        for event in event_stream:
            # handle event
            print(event, flush=True)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.chat(request={
            "model": "review-model",
            "prompt": "What is the largest city in the world?",
            "stream": False,
        })

        async with res as event_stream:
            async for event in event_stream:
                # handle event
                print(event, flush=True)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `request`                                                           | [models.ChatRequest](../../models/chatrequest.md)                   | :heavy_check_mark:                                                  | The request object to use for the request.                          |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.ChatResponse](../../models/chatresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_binary_default_response

### Example Usage

<!-- UsageSnippet language="python" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_binary_default_response()

    assert res is not None

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

        res = await sdk.get_binary_default_response()

        assert res is not None

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[httpx.Response](../../models/bytes.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## test_enum_formats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="python" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.test_enum_formats(string_array_format="pending_review", string_map_format="medium_priority", string_partial_map_format="draft_mode", integer_map_format=200, integer_partial_map_format=1)

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

        res = await sdk.test_enum_formats(string_array_format="pending_review", string_map_format="medium_priority", string_partial_map_format="draft_mode", integer_map_format=200, integer_partial_map_format=1)

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                                 | Type                                                                      | Required                                                                  | Description                                                               | Example                                                                   |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `string_array_format`                                                     | [models.StringArrayFormat](../../models/stringarrayformat.md)             | :heavy_check_mark:                                                        | String enum with x-speakeasy-enums as array (full coverage)               | pending_review                                                            |
| `string_map_format`                                                       | [models.StringMapFormat](../../models/stringmapformat.md)                 | :heavy_check_mark:                                                        | String enum with x-speakeasy-enums as map (full coverage)                 | medium_priority                                                           |
| `string_partial_map_format`                                               | [models.StringPartialMapFormat](../../models/stringpartialmapformat.md)   | :heavy_check_mark:                                                        | String enum with x-speakeasy-enums as map (partial coverage)              | draft_mode                                                                |
| `integer_map_format`                                                      | [models.IntegerMapFormat](../../models/integermapformat.md)               | :heavy_check_mark:                                                        | Integer enum with x-speakeasy-enums as map (full coverage)                | 200                                                                       |
| `integer_partial_map_format`                                              | [models.IntegerPartialMapFormat](../../models/integerpartialmapformat.md) | :heavy_check_mark:                                                        | Integer enum with x-speakeasy-enums as map (partial coverage)             | 1                                                                         |
| `retries`                                                                 | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)          | :heavy_minus_sign:                                                        | Configuration to override the default retry behavior of the client.       |                                                                           |

### Response

**[models.TestEnumFormatsResponse](../../models/testenumformatsresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## binary_and_string_upload

### Example Usage

<!-- UsageSnippet language="python" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.binary_and_string_upload(binary=open("test.json", "rb").read(), string=open("test.json", "rb").read().decode("utf-8"))

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

        await sdk.binary_and_string_upload(binary=open("test.json", "rb").read(), string=open("test.json", "rb").read().decode("utf-8"))

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         | Example                                                             |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `binary`                                                            | *Optional[bytes]*                                                   | :heavy_minus_sign:                                                  | N/A                                                                 | x-file: test.json                                                   |
| `string`                                                            | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | N/A                                                                 | x-file: test.json                                                   |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |                                                                     |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_error_in_union

### Example Usage

<!-- UsageSnippet language="python" operationID="getErrorInUnion" method="get" path="/errorInUnion" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.get_error_in_union()

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

        await sdk.get_error_in_union()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| models.ErrorsError  | 500                 | application/json    |
| models.TaggedError1 | 500                 | application/json    |
| models.SDKError     | 4XX, 5XX            | \*/\*               |

## get_duplicate_export_collision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="python" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_duplicate_export_collision()

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

        res = await sdk.get_duplicate_export_collision()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetDuplicateExportCollisionResponse](../../models/getduplicateexportcollisionresponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models.RequestTimeoutError | 408                        | application/json           |
| models.SDKError            | 4XX, 5XX                   | \*/\*                      |

## get_named_primitive_union

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="python" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_named_primitive_union()

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

        res = await sdk.get_named_primitive_union()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.SomeUnion](../../models/someunion.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_empty_object_error

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="python" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_empty_object_error()

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

        res = await sdk.get_empty_object_error()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetEmptyObjectErrorResponse](../../models/getemptyobjecterrorresponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models.FailedResponseError | 500                        | application/json           |
| models.SDKError            | 4XX, 5XX                   | \*/\*                      |

## url_validation_stress_test

### Example Usage

<!-- UsageSnippet language="python" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.url_validation_stress_test()

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

        await sdk.url_validation_stress_test()

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

## parentheses_in_path_allowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="python" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.parentheses_in_path_allowed(id="<id>", field_with_braces_in_description="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_default="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_title="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_example="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n")

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

        res = await sdk.parentheses_in_path_allowed(id="<id>", field_with_braces_in_description="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_default="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_title="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n", field_with_braces_in_example="A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n")

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                                                                                                                                                                                                                     | Type                                                                                                                                                                                                                                                          | Required                                                                                                                                                                                                                                                      | Description                                                                                                                                                                                                                                                   | Example                                                                                                                                                                                                                                                       |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`                                                                                                                                                                                                                                                          | *str*                                                                                                                                                                                                                                                         | :heavy_check_mark:                                                                                                                                                                                                                                            | N/A                                                                                                                                                                                                                                                           |                                                                                                                                                                                                                                                               |
| `field_with_braces_in_description`                                                                                                                                                                                                                            | *Optional[str]*                                                                                                                                                                                                                                               | :heavy_minus_sign:                                                                                                                                                                                                                                            | A string with {{ double braces }} and { single braces }<br/>and \{\{ escaped curlies \}\} and `backticks`.<br/>and \`escaped backticks\` and double slashes\\<br/>and 'single quotes' and "double quotes".<br/>and  \'escaped single quotes\' and \"escaped double quotes\".<br/> |                                                                                                                                                                                                                                                               |
| `field_with_braces_in_default`                                                                                                                                                                                                                                | *Optional[str]*                                                                                                                                                                                                                                               | :heavy_minus_sign:                                                                                                                                                                                                                                            | N/A                                                                                                                                                                                                                                                           |                                                                                                                                                                                                                                                               |
| `field_with_braces_in_title`                                                                                                                                                                                                                                  | *Optional[str]*                                                                                                                                                                                                                                               | :heavy_minus_sign:                                                                                                                                                                                                                                            | N/A                                                                                                                                                                                                                                                           |                                                                                                                                                                                                                                                               |
| `field_with_braces_in_example`                                                                                                                                                                                                                                | *Optional[str]*                                                                                                                                                                                                                                               | :heavy_minus_sign:                                                                                                                                                                                                                                            | N/A                                                                                                                                                                                                                                                           | A string with {{ double braces }} and { single braces }<br/>and \{\{ escaped curlies \}\} and `backticks`.<br/>and \`escaped backticks\` and double slashes\\<br/>and 'single quotes' and "double quotes".<br/>and  \'escaped single quotes\' and \"escaped double quotes\".<br/> |
| `retries`                                                                                                                                                                                                                                                     | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                                                                                                                                                                                              | :heavy_minus_sign:                                                                                                                                                                                                                                            | Configuration to override the default retry behavior of the client.                                                                                                                                                                                           |                                                                                                                                                                                                                                                               |

### Response

**[models.TemplateBracesTest](../../models/templatebracestest.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_nested_integer_string

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="python" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_nested_integer_string()

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

        res = await sdk.get_nested_integer_string()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.TaskResponse](../../models/taskresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## render_asset

Render a media asset from a text prompt.

### Example Usage

<!-- UsageSnippet language="python" operationID="renderAsset" method="post" path="/assets/render" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.render_asset(prompt="<value>")

    with res as event_stream:
        for event in event_stream:
            # handle event
            print(event, flush=True)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.render_asset(prompt="<value>")

        async with res as event_stream:
            async for event in event_stream:
                # handle event
                print(event, flush=True)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `prompt`                                                            | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `output_modalities`                                                 | List[[models.OutputModality](../../models/outputmodality.md)]       | :heavy_minus_sign:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.RenderAssetResponse](../../models/renderassetresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_asset

Get the current result of an asset job.

### Example Usage

<!-- UsageSnippet language="python" operationID="getAsset" method="get" path="/assets/{id}" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_asset(id="<id>", stream=True)

    with res as event_stream:
        for event in event_stream:
            # handle event
            print(event, flush=True)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.get_asset(id="<id>", stream=True)

        async with res as event_stream:
            async for event in event_stream:
                # handle event
                print(event, flush=True)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `id`                                                                | *str*                                                               | :heavy_check_mark:                                                  | N/A                                                                 |
| `stream`                                                            | *Optional[bool]*                                                    | :heavy_minus_sign:                                                  | N/A                                                                 |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetAssetResponse](../../models/getassetresponse.md)**

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |

## get_error_only_example

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="python" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.get_error_only_example()

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

        res = await sdk.get_error_only_example()

        # Handle response
        print(res)

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Response

**[models.GetErrorOnlyExampleResponse](../../models/geterroronlyexampleresponse.md)**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| models.ErrorsError | 404                | application/json   |
| models.SDKError    | 4XX, 5XX           | \*/\*              |