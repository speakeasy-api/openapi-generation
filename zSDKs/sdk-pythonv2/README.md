# openapi

Developer-friendly & type-safe Python SDK specifically catered to leverage *openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=python)
[![License: MIT](https://img.shields.io/badge/LICENSE_//_MIT-3b5bdb?style=for-the-badge&labelColor=eff6ff)](https://opensource.org/licenses/MIT)


<br /><br />
> [!IMPORTANT]
> This SDK is not yet ready for production use. Delete this section before > publishing to a package manager.

<!-- Start Summary [summary] -->
## Summary

SDK Review: A test document for reviewing the SDK.

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

For more information about the API: [Speakeasy Docs](https://speakeasy.com/docs)
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [openapi](#openapi)
  * [SDK Installation](#sdk-installation)
  * [IDE Support](#ide-support)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Global Parameters](#global-parameters)
  * [Server-sent event streaming](#server-sent-event-streaming)
  * [Pagination](#pagination)
  * [File uploads](#file-uploads)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
  * [Custom HTTP Client](#custom-http-client)
  * [Resource Management](#resource-management)
  * [Debugging](#debugging)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

> [!TIP]
> To finish publishing your SDK to PyPI you must [run your first generation action](https://www.speakeasy.com/docs/github-setup#step-by-step-guide).


> [!NOTE]
> **Python version upgrade policy**
>
> Once a Python version reaches its [official end of life date](https://devguide.python.org/versions/), a 3-month grace period is provided for users to upgrade. Following this grace period, the minimum python version supported in the SDK will be updated.

The SDK can be installed with *uv*, *pip*, or *poetry* package managers.

### uv

*uv* is a fast Python package installer and resolver, designed as a drop-in replacement for pip and pip-tools. It's recommended for its speed and modern Python tooling capabilities.

```bash
uv add git+https://github.com/speakeasy-sdks/test-sdk.git
```

### PIP

*PIP* is the default package installer for Python, enabling easy installation and management of packages from PyPI via the command line.

```bash
pip install git+https://github.com/speakeasy-sdks/test-sdk.git
```

### Poetry

*Poetry* is a modern tool that simplifies dependency management and package publishing by using a single `pyproject.toml` file to handle project metadata and dependencies.

```bash
poetry add git+https://github.com/speakeasy-sdks/test-sdk.git
```

### Shell and script usage with `uv`

You can use this SDK in a Python shell with [uv](https://docs.astral.sh/uv/) and the `uvx` command that comes with it like so:

```shell
uvx --from openapi python
```

It's also possible to write a standalone Python script without needing to set up a whole project like so:

```python
#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "openapi",
# ]
# ///

from speakeasy.new_openapi import SDK

sdk = SDK(
  # SDK arguments
)

# Rest of script here...
```

Once that is saved to a file, you can run it with `uv run script.py` where
`script.py` can be replaced with the actual file name.
<!-- End SDK Installation [installation] -->

<!-- Start IDE Support [idesupport] -->
## IDE Support

### PyCharm

Generally, the SDK will work well with most IDEs out of the box. However, when using PyCharm, you can enjoy much better integration with Pydantic by installing an additional plugin.

- [PyCharm Pydantic Plugin](https://docs.pydantic.dev/latest/integrations/pycharm/)
<!-- End IDE Support [idesupport] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

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

### Example 2

```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.tag1.post_file_with_encoding(file={
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

        res = await sdk.tag1.post_file_with_encoding(file={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Example 3

```python
# Synchronous Example
from datetime import date
from decimal import Decimal
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK
from speakeasy.new_openapi.utils import parse_datetime


with SDK(
    deprecated_query_param1="some example query param",
    deprecated_query_param2="some example query param",
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

    res = sdk.test_group.tag2.post_test(obj=speakeasy.new_openapi.ExhaustiveObject(
        str_="example",
        bool_=True,
        integer=999999,
        int32=1,
        num=1.1,
        float32=8499.3,
        date_=date.fromisoformat("2020-01-01"),
        date_time=parse_datetime("2020-01-01T00:00:00Z"),
        anything="<value>",
        bool_opt=True,
        int_opt_null=999999,
        num_opt_null=1.1,
        int_enum=3,
        int32_enum=69,
        bigint=593288,
        decimal_str=Decimal("7028.3"),
        obj=speakeasy.new_openapi.SimpleObject(
            str_="example",
        ),
        map={

        },
        arr=[],
        any="<value>",
        nullable_int_enum=3,
        nullable_string_enum="Second",
        color="green",
        icon="tick",
        hero_width=480,
    ), type_="type1")

    assert res is not None

    # Handle response
    print(res)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from datetime import date
from decimal import Decimal
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK
from speakeasy.new_openapi.utils import parse_datetime

async def main():

    async with AsyncSDK(
        deprecated_query_param1="some example query param",
        deprecated_query_param2="some example query param",
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.test_group.tag2.post_test(obj=speakeasy.new_openapi.ExhaustiveObject(
            str_="example",
            bool_=True,
            integer=999999,
            int32=1,
            num=1.1,
            float32=8499.3,
            date_=date.fromisoformat("2020-01-01"),
            date_time=parse_datetime("2020-01-01T00:00:00Z"),
            anything="<value>",
            bool_opt=True,
            int_opt_null=999999,
            num_opt_null=1.1,
            int_enum=3,
            int32_enum=69,
            bigint=593288,
            decimal_str=Decimal("7028.3"),
            obj=speakeasy.new_openapi.SimpleObject(
                str_="example",
            ),
            map={

            },
            arr=[],
            any="<value>",
            nullable_int_enum=3,
            nullable_string_enum="Second",
            color="green",
            icon="tick",
            hero_width=480,
        ), type_="type1")

        assert res is not None

        # Handle response
        print(res)

asyncio.run(main())
```

### A custom readme heading

A custom usage description

```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    query_param1="some example query param",
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

    res = sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

    while res is not None:
        # Handle items

        res = res.next()
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
        query_param1="some example query param",
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

        while res is not None:
            # Handle items

            res = res.next()

asyncio.run(main())
```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives by setting the `security` optional parameter when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     | Environment Variable                          |
| ------------------------- | ---- | ---------- | --------------------------------------------- |
| `username`<br/>`password` | http | HTTP Basic | `SPEAKEASY_USERNAME`<br/>`SPEAKEASY_PASSWORD` |

```python
# Synchronous Example
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        user_pass_auth=speakeasy.new_openapi.UserPassAuth(
            username="<USERNAME>",
            password="<PASSWORD>",
        ),
    ),
) as sdk:

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
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            user_pass_auth=speakeasy.new_openapi.UserPassAuth(
                username="<USERNAME>",
                password="<PASSWORD>",
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name          | Type   | Scheme      | Environment Variable    |
| ------------- | ------ | ----------- | ----------------------- |
| `bearer_auth` | http   | HTTP Bearer | `SPEAKEASY_BEARER_AUTH` |
| `my_api_key`  | apiKey | API key     | `SPEAKEASY_MY_API_KEY`  |

```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        option2=speakeasy.new_openapi.SecurityOption2(
            bearer_auth="<YOUR_JWT>",
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

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
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            option2=speakeasy.new_openapi.SecurityOption2(
                bearer_auth="<YOUR_JWT>",
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       | Environment Variable |
| -------- | ------ | ------------ | -------------------- |
| `oauth2` | oauth2 | OAuth2 token | `SPEAKEASY_OAUTH2`   |

```python
# Synchronous Example
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        option3=speakeasy.new_openapi.SecurityOption3(
            oauth2="Bearer <YOUR_OAUTH2_TOKEN>",
        ),
    ),
) as sdk:

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
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            option3=speakeasy.new_openapi.SecurityOption3(
                oauth2="Bearer <YOUR_OAUTH2_TOKEN>",
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                  | Type | Scheme      | Environment Variable                      |
| --------------------- | ---- | ----------- | ----------------------------------------- |
| `app_id`<br/>`secret` | http | Custom HTTP | `SPEAKEASY_APP_ID`<br/>`SPEAKEASY_SECRET` |

```python
# Synchronous Example
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        option4=speakeasy.new_openapi.SecurityOption4(
            app_id="app-speakeasy-123",
            secret="MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI",
        ),
    ),
) as sdk:

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
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            option4=speakeasy.new_openapi.SecurityOption4(
                app_id="app-speakeasy-123",
                secret="MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI",
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name          | Type   | Scheme       | Environment Variable    |
| ------------- | ------ | ------------ | ----------------------- |
| `mobile_auth` | oauth2 | OAuth2 token | `SPEAKEASY_MOBILE_AUTH` |

```python
# Synchronous Example
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        option5=speakeasy.new_openapi.SecurityOption5(
            mobile_auth="Bearer <YOUR_OAUTH2_TOKEN>",
        ),
    ),
) as sdk:

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
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            option5=speakeasy.new_openapi.SecurityOption5(
                mobile_auth="Bearer <YOUR_OAUTH2_TOKEN>",
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                            | Type   | Scheme                         | Environment Variable                                                          |
| ----------------------------------------------- | ------ | ------------------------------ | ----------------------------------------------------------------------------- |
| `client_id`<br/>`client_secret`<br/>`token_url` | oauth2 | OAuth2 Client Credentials Flow | `SPEAKEASY_CLIENT_ID`<br/>`SPEAKEASY_CLIENT_SECRET`<br/>`SPEAKEASY_TOKEN_URL` |

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

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name         | Type   | Scheme  | Environment Variable   |
| ------------ | ------ | ------- | ---------------------- |
| `my_api_key` | apiKey | API key | `SPEAKEASY_MY_API_KEY` |

```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

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
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import AsyncSDK

async def main():

    async with AsyncSDK(
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.tag1.auth(security=speakeasy.new_openapi.AuthSecurity(
        access_token=os.getenv("SPEAKEASY_ACCESS_TOKEN", ""),
    ))

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

    async with AsyncSDK() as sdk:

        await sdk.tag1.auth(security=speakeasy.new_openapi.AuthSecurity(
            access_token=os.getenv("SPEAKEASY_ACCESS_TOKEN", ""),
        ))

        # Use the SDK ...

asyncio.run(main())
```
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### [SDK](docs/sdks/sdk/README.md)

* [operation_with_leading_and_trailing_underscores_](docs/sdks/sdk/README.md#operation_with_leading_and_trailing_underscores_)
* [post_file](docs/sdks/sdk/README.md#post_file) - Post File
* [get_polymorphism](docs/sdks/sdk/README.md#get_polymorphism)
* [get_union_errors](docs/sdks/sdk/README.md#get_union_errors)
* [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away)
* [get_fully_flattened_request](docs/sdks/sdk/README.md#get_fully_flattened_request)
* [create_with_union](docs/sdks/sdk/README.md#create_with_union) - Create with discriminated union request body
* [test_endpoint](docs/sdks/sdk/README.md#test_endpoint)
* [create_user](docs/sdks/sdk/README.md#create_user) - Create User
* [get_user](docs/sdks/sdk/README.md#get_user) - Get User
* [update_user](docs/sdks/sdk/README.md#update_user) - Update User
* [delete_user](docs/sdks/sdk/README.md#delete_user) - Delete User
* [login](docs/sdks/sdk/README.md#login) - Login
* [validate](docs/sdks/sdk/README.md#validate) - Validate
* [chat](docs/sdks/sdk/README.md#chat)
* [get_binary_default_response](docs/sdks/sdk/README.md#get_binary_default_response)
* [test_enum_formats](docs/sdks/sdk/README.md#test_enum_formats) - Test x-speakeasy-enums in different formats
* [binary_and_string_upload](docs/sdks/sdk/README.md#binary_and_string_upload)
* [get_error_in_union](docs/sdks/sdk/README.md#get_error_in_union)
* [get_duplicate_export_collision](docs/sdks/sdk/README.md#get_duplicate_export_collision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [get_named_primitive_union](docs/sdks/sdk/README.md#get_named_primitive_union) - Test named primitive union options using title and x-speakeasy-name-override
* [get_empty_object_error](docs/sdks/sdk/README.md#get_empty_object_error) - Get Empty Object Error
* [url_validation_stress_test](docs/sdks/sdk/README.md#url_validation_stress_test)
* [parentheses_in_path_allowed](docs/sdks/sdk/README.md#parentheses_in_path_allowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [get_nested_integer_string](docs/sdks/sdk/README.md#get_nested_integer_string) - Test nested struct with integer:string tag
* [render_asset](docs/sdks/sdk/README.md#render_asset) - Render Asset
* [get_asset](docs/sdks/sdk/README.md#get_asset) - Get Asset
* [get_error_only_example](docs/sdks/sdk/README.md#get_error_only_example) - Operation with example only on error response

### [Group](docs/sdks/group/README.md)

* [root_group_op](docs/sdks/group/README.md#root_group_op) - An operation at the group's root level

#### [Group.SubGroup](docs/sdks/subgroup/README.md)

* [sub_group_op](docs/sdks/subgroup/README.md#sub_group_op) - An operation at the group's top level

##### [Group.SubGroup.Empty.Tail](docs/sdks/tail/README.md)

* [nested_group_op](docs/sdks/tail/README.md#nested_group_op) - An operation at the group's deepest level

### [NamespaceTests.Conflicts](docs/sdks/conflicts/README.md)

* [get_namespace_conflict](docs/sdks/conflicts/README.md#get_namespace_conflict) - Get Namespace Conflict Test
* [put_namespace_conflict](docs/sdks/conflicts/README.md#put_namespace_conflict) - Put Property Name Conflicts Behind
* [create_namespace_conflict](docs/sdks/conflicts/README.md#create_namespace_conflict) - Create Namespace Conflict Test
* [get_triple_namespace_conflict](docs/sdks/conflicts/README.md#get_triple_namespace_conflict) - Get Triple Namespace Conflict Test
* [get_pet_owners](docs/sdks/conflicts/README.md#get_pet_owners) - Get Pet Owners

### [NamespaceTests.SingleBar](docs/sdks/singlebar/README.md)

* [get_single_namespace_bar_pet](docs/sdks/singlebar/README.md#get_single_namespace_bar_pet) - Get Single Namespace Bar Pet

### [NamespaceTests.SingleFoo](docs/sdks/singlefoo/README.md)

* [get_single_namespace_foo_pet](docs/sdks/singlefoo/README.md#get_single_namespace_foo_pet) - Get Single Namespace Foo Pet
* [create_single_namespace_foo_pet](docs/sdks/singlefoo/README.md#create_single_namespace_foo_pet) - Create Single Namespace Foo Pet

### [NamespaceTests.Types](docs/sdks/types/README.md)

* [get_namespace_types](docs/sdks/types/README.md#get_namespace_types) - Get Namespace Types Test
* [get_namespace_animal](docs/sdks/types/README.md#get_namespace_animal) - Get Namespace Animal (Discriminated Union)
* [get_namespace_vehicle](docs/sdks/types/README.md#get_namespace_vehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [get_namespace_organization](docs/sdks/types/README.md#get_namespace_organization) - Get Namespace Organization (Nested Inline Schemas)

### [~~Obsolete~~](docs/sdks/obsolete/README.md)

* [~~deprecated1~~](docs/sdks/obsolete/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away) instead.

### [Tag1](docs/sdks/tag1/README.md)

* [~~deprecated1~~](docs/sdks/tag1/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away) instead.
* [auth](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [list_test1](docs/sdks/tag1/README.md#list_test1) - Get Test1
* [post_file_with_encoding](docs/sdks/tag1/README.md#post_file_with_encoding) - Post File With Encoding

### [TestGroup.Tag2](docs/sdks/tag2/README.md)

* [post_test](docs/sdks/tag2/README.md#post_test) - Post Test2

### [TestGroup.Tag3](docs/sdks/tag3/README.md)

* [post_test](docs/sdks/tag3/README.md#post_test) - Post Test2

</details>
<!-- End Available Resources and Operations [operations] -->

<!-- Start Global Parameters [global-parameters] -->
## Global Parameters

Certain parameters are configured globally. These parameters may be set on the SDK client instance itself during initialization. When configured as an option during SDK initialization, These global values will be used as defaults on the operations that use them. When such operations are called, there is a place in each to override the global value, if needed.

For example, you can set `queryParam1` to `"some example query param"` at SDK initialization and then you do not have to pass the same value on calls to operations like `get_request_body_flattened_away`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.
Global parameters can also be set via environment variable.

| Name                    | Type | Description                                                                        | Environment                       |
| ----------------------- | ---- | ---------------------------------------------------------------------------------- | --------------------------------- |
| query_param1            | str  | A long winded, multi-line description<br/>for the query parameter number one.<br/> | SPEAKEASY_QUERY_PARAM1            |
| deprecated_query_param1 | str  | A deprecated description                                                           | SPEAKEASY_DEPRECATED_QUERY_PARAM1 |
| deprecated_query_param2 | str  | The deprecated_query_param2 parameter.                                             | SPEAKEASY_DEPRECATED_QUERY_PARAM2 |
| lone_query_param        | str  | The lone_query_param parameter.                                                    | SPEAKEASY_LONE_QUERY_PARAM        |

### Example

```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK(
    lone_query_param="<value>",
    query_param1="some example query param",
    deprecated_query_param1="some example query param",
    deprecated_query_param2="some example query param",
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
        query_param1="some example query param",
        deprecated_query_param1="some example query param",
        deprecated_query_param2="some example query param",
    ) as sdk:

        await sdk.get_request_body_flattened_away()

        # Use the SDK ...

asyncio.run(main())
```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Server-sent event streaming [eventstream] -->
## Server-sent event streaming

[Server-sent events][mdn-sse] are used to stream content from certain
operations. These operations will expose the stream as [Generator][generator] that
can be consumed using a simple `for` loop. The loop will
terminate when the server no longer has any events to send and closes the
underlying connection.  

The stream is also a [Context Manager][context-manager] and can be used with the `with` statement and will close the
underlying connection when the context is exited.

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

[mdn-sse]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
[generator]: https://book.pythontips.com/en/latest/generators.html
[context-manager]: https://book.pythontips.com/en/latest/context_managers.html
<!-- End Server-sent event streaming [eventstream] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you make your SDK calls as usual, but the
returned response object will have a `Next` method that can be called to pull down the next group of results. If the
return value of `Next` is `None`, then there are no more pages to be fetched.

Here's an example of one such pagination call:
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    query_param1="some example query param",
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

    res = sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

    while res is not None:
        # Handle items

        res = res.next()
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
        query_param1="some example query param",
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param")

        while res is not None:
            # Handle items

            res = res.next()

asyncio.run(main())
```
<!-- End Pagination [pagination] -->

<!-- Start File uploads [file-upload] -->
## File uploads

Certain SDK methods accept file objects as part of a request body or multi-part request. It is possible and typically recommended to upload files as a stream rather than reading the entire contents into memory. This avoids excessive memory consumption and potentially crashing with out-of-memory errors when working with very large files. The following example demonstrates how to attach a file stream to a request.

> [!TIP]
>
> For endpoints that handle file uploads bytes arrays can also be used. However, using streams is recommended for large files.
>

```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    res = sdk.tag1.post_file_with_encoding(file={
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

        res = await sdk.tag1.post_file_with_encoding(file={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```
<!-- End File uploads [file-upload] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide a `RetryConfig` object to the call:
```python
# Synchronous Example
from speakeasy.new_openapi import SDK
from speakeasy.new_openapi.utils import BackoffStrategy, RetryConfig


with SDK() as sdk:

    res = sdk.post_file(upload={
        "file_name": "example.file",
        "content": open("example.file", "rb"),
    },
        RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False))

    # Handle response
    print(res)
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK
from speakeasy.new_openapi.utils import BackoffStrategy, RetryConfig

async def main():

    async with AsyncSDK() as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        },
            RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False))

        # Handle response
        print(res)

asyncio.run(main())
```

If you'd like to override the default retry strategy for all operations that support retries, you can use the `retry_config` optional parameter when initializing the SDK:
```python
# Synchronous Example
from speakeasy.new_openapi import SDK
from speakeasy.new_openapi.utils import BackoffStrategy, RetryConfig


with SDK(
    retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False),
) as sdk:

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
from speakeasy.new_openapi.utils import BackoffStrategy, RetryConfig

async def main():

    async with AsyncSDK(
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False),
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

[`SDKBaseError`](./src/speakeasy/new_openapi/models/sdkbaseerror.py) is the base class for all HTTP error responses. It has the following properties:

| Property           | Type             | Description                                                                             |
| ------------------ | ---------------- | --------------------------------------------------------------------------------------- |
| `err.message`      | `str`            | Error message                                                                           |
| `err.status_code`  | `int`            | HTTP response status code eg `404`                                                      |
| `err.headers`      | `httpx.Headers`  | HTTP response headers                                                                   |
| `err.body`         | `str`            | HTTP body. Can be empty string if no body is returned.                                  |
| `err.raw_response` | `httpx.Response` | Raw HTTP response                                                                       |
| `err.data`         |                  | Optional. Some errors may contain structured data. [See Error Classes](#error-classes). |

### Example
```python
# Synchronous Example
from speakeasy.new_openapi import SDK, models


with SDK() as sdk:
    res = None
    try:

        res = sdk.get_union_errors(page=12)

        while res is not None:
            # Handle items

            res = res.next()


    except models.SDKBaseError as e:
        # The base class for HTTP error responses
        print(e.message)
        print(e.status_code)
        print(e.body)
        print(e.headers)
        print(e.raw_response)

        # Depending on the method different errors may be thrown
        if isinstance(e, models.ErrorsError):
            print(e.data.error)  # str
            print(e.data.code)  # int
```

</br>

An Async SDK client can also be used to make asynchronous requests by importing it and asyncio.

```python
# Asynchronous Example
import asyncio
from speakeasy.new_openapi import AsyncSDK, models

async def main():

    async with AsyncSDK() as sdk:
        res = None
        try:

            res = await sdk.get_union_errors(page=12)

            while res is not None:
                # Handle items

                res = res.next()


            except models.SDKBaseError as e:
                # The base class for HTTP error responses
                print(e.message)
                print(e.status_code)
                print(e.body)
                print(e.headers)
                print(e.raw_response)

                # Depending on the method different errors may be thrown
                if isinstance(e, models.ErrorsError):
                    print(e.data.error)  # str
                    print(e.data.code)  # int

asyncio.run(main())
```

### Error Classes
**Primary error:**
* [`SDKBaseError`](./src/speakeasy/new_openapi/models/sdkbaseerror.py): The base class for HTTP error responses.

<details><summary>Less common errors (14)</summary>

<br />

**Network errors:**
* [`httpx.RequestError`](https://www.python-httpx.org/exceptions/#httpx.RequestError): Base class for request errors.
    * [`httpx.ConnectError`](https://www.python-httpx.org/exceptions/#httpx.ConnectError): HTTP client was unable to make a request to a server.
    * [`httpx.TimeoutException`](https://www.python-httpx.org/exceptions/#httpx.TimeoutException): HTTP request timed out.


**Inherit from [`SDKBaseError`](./src/speakeasy/new_openapi/models/sdkbaseerror.py)**:
* [`ErrorsError`](./src/speakeasy/new_openapi/models/errorserror.py): A not-so-long multi-line error model description. Applicable to 7 of 50 methods.*
* [`BadRequestResponseError`](./src/speakeasy/new_openapi/models/badrequestresponseerror.py): Bad Request. Status code `400`. Applicable to 2 of 50 methods.*
* [`TaggedError1`](./src/speakeasy/new_openapi/models/taggederror1.py): Applicable to 2 of 50 methods.*
* [`RequestTimeoutError`](./src/speakeasy/new_openapi/models/requesttimeouterror.py): A spec-defined error that collides with the built-in RequestTimeoutError in httpclienterrors.ts. Status code `408`. Applicable to 1 of 50 methods.*
* [`TaggedError2`](./src/speakeasy/new_openapi/models/taggederror2.py): Something went wrong. Status code `4XX`. Applicable to 1 of 50 methods.*
* [`ErrorType1`](./src/speakeasy/new_openapi/models/errortype1.py): An error of type one. Status code `500`. Applicable to 1 of 50 methods.*
* [`ErrorType2`](./src/speakeasy/new_openapi/models/errortype2.py): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
* [`FailedResponseError`](./src/speakeasy/new_openapi/models/failedresponseerror.py): An error response with an empty object schema. Status code `500`. Applicable to 1 of 50 methods.*
* [`Test2ResponseError`](./src/speakeasy/new_openapi/models/test2responseerror.py): Internal Server Error. Status code `500`. Applicable to 1 of 50 methods.*
* [`ResponseValidationError`](./src/speakeasy/new_openapi/models/responsevalidationerror.py): Type mismatch between the response data and the expected Pydantic model. Provides access to the Pydantic validation error via the `cause` attribute.

</details>

\* Check [the method documentation](#available-resources-and-operations) to see if the error is applicable.
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally by passing a server index to the `server_idx: int` optional parameter when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values through the additional parameters made available in the SDK constructor:

| Variable    | Parameter                 | Supported Values                      | Default       | Description                              |
| ----------- | ------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain: str`          | str                                   | `"api"`       |                                          |
| `version`   | `version: str`            | str                                   | `"1"`         |                                          |
| `HostName`  | `host_name: str`          | str                                   | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port: models.ServerPORT` | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK(
    server_idx=2,
    host_name="localhost",
    port="443",
) as sdk:

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

    async with AsyncSDK(
        server_idx=2,
        host_name="localhost",
        port="443",
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Override Server URL Per-Client

The default server can also be overridden globally by passing a URL to the `server_url: str` optional parameter when initializing the SDK client instance. For example:
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK(
    server_url="http://localhost:8080",
) as sdk:

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

    async with AsyncSDK(
        server_url="http://localhost:8080",
    ) as sdk:

        res = await sdk.post_file(upload={
            "file_name": "example.file",
            "content": open("example.file", "rb"),
        })

        # Handle response
        print(res)

asyncio.run(main())
```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```python
# Synchronous Example
import os
import speakeasy.new_openapi
from speakeasy.new_openapi import SDK


with SDK(
    query_param1="some example query param",
    security=speakeasy.new_openapi.Security(
        my_api_key=speakeasy.new_openapi.MyAPIKey(
            my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
        ),
    ),
) as sdk:

    res = sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param", server_url="http://localhost:35123")

    while res is not None:
        # Handle items

        res = res.next()
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
        query_param1="some example query param",
        security=speakeasy.new_openapi.Security(
            my_api_key=speakeasy.new_openapi.MyAPIKey(
                my_api_key=os.getenv("SPEAKEASY_MY_API_KEY", ""),
            ),
        ),
    ) as sdk:

        res = await sdk.tag1.list_test1(query_param2=1, page=100, header_param1="some example header param", server_url="http://localhost:35123")

        while res is not None:
            # Handle items

            res = res.next()

asyncio.run(main())
```
<!-- End Server Selection [server] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The Python SDK makes API calls using the [httpx](https://www.python-httpx.org/) HTTP library.  In order to provide a convenient way to configure timeouts, cookies, proxies, custom headers, and other low-level configuration, you can initialize the SDK client with your own HTTP client instance.
Depending on whether you are using the sync or async version of the SDK, you can pass an instance of `HttpClient` or `AsyncHttpClient` respectively, which are Protocol's ensuring that the client has the necessary methods to make API calls.
This allows you to wrap the client with your own custom logic, such as adding custom headers, logging, or error handling, or you can just pass an instance of `httpx.Client` or `httpx.AsyncClient` directly.

For example, you could specify a header for every request that this sdk makes as follows:
```python
from speakeasy.new_openapi import SDK
import httpx

http_client = httpx.Client(headers={"x-custom-header": "someValue"})
s = SDK(client=http_client)
```

or you could wrap the client with your own custom logic:
```python
from speakeasy.new_openapi import SDK
from speakeasy.new_openapi.httpclient import AsyncHttpClient
import httpx

class CustomClient(AsyncHttpClient):
    client: AsyncHttpClient

    def __init__(self, client: AsyncHttpClient):
        self.client = client

    async def send(
        self,
        request: httpx.Request,
        *,
        stream: bool = False,
        auth: Union[
            httpx._types.AuthTypes, httpx._client.UseClientDefault, None
        ] = httpx.USE_CLIENT_DEFAULT,
        follow_redirects: Union[
            bool, httpx._client.UseClientDefault
        ] = httpx.USE_CLIENT_DEFAULT,
    ) -> httpx.Response:
        request.headers["Client-Level-Header"] = "added by client"

        return await self.client.send(
            request, stream=stream, auth=auth, follow_redirects=follow_redirects
        )

    def build_request(
        self,
        method: str,
        url: httpx._types.URLTypes,
        *,
        content: Optional[httpx._types.RequestContent] = None,
        data: Optional[httpx._types.RequestData] = None,
        files: Optional[httpx._types.RequestFiles] = None,
        json: Optional[Any] = None,
        params: Optional[httpx._types.QueryParamTypes] = None,
        headers: Optional[httpx._types.HeaderTypes] = None,
        cookies: Optional[httpx._types.CookieTypes] = None,
        timeout: Union[
            httpx._types.TimeoutTypes, httpx._client.UseClientDefault
        ] = httpx.USE_CLIENT_DEFAULT,
        extensions: Optional[httpx._types.RequestExtensions] = None,
    ) -> httpx.Request:
        return self.client.build_request(
            method,
            url,
            content=content,
            data=data,
            files=files,
            json=json,
            params=params,
            headers=headers,
            cookies=cookies,
            timeout=timeout,
            extensions=extensions,
        )

s = SDK(async_client=CustomClient(httpx.AsyncClient()))
```
### httpx2 (Pydantic's httpx fork)

[httpx2](https://httpx2.pydantic.dev/) is Pydantic's maintained fork of `httpx`. To run this SDK on httpx2, call `alias_httpx()` at your program's entry point, before importing the SDK, so every `import httpx` — including the ones inside the SDK — resolves to `httpx2`:
```python
import httpx2

httpx2.alias_httpx()

from speakeasy.new_openapi import SDK

s = SDK()
```

An SDK can also be generated against httpx2 directly, so it depends on the fork instead of `httpx`, by setting `python.httpClientLibrary: httpx2` in `gen.yaml`.
<!-- End Custom HTTP Client [http-client] -->

<!-- Start Resource Management [resource-management] -->
## Resource Management

The `SDK` and `AsyncSDK` classes implement the context manager protocol and register finalizer functions to close the underlying HTTPX clients they use under the hood. This will close HTTP connections, release memory and free up other resources held by the SDKs. In short-lived Python programs and notebooks that make a few SDK method calls, resource management may not be a concern. However, in longer-lived programs, it is beneficial to create SDK instances via [context managers][context-manager] and reuse them across the application.

[context-manager]: https://docs.python.org/3/reference/datamodel.html#context-managers

```python
from speakeasy.new_openapi import AsyncSDK, SDK
def main():

    with SDK() as sdk:
        # Rest of application here...


# Or when using async:
async def amain():

    async with AsyncSDK() as sdk:
        # Rest of application here...
```
<!-- End Resource Management [resource-management] -->

<!-- Start Debugging [debug] -->
## Debugging

You can setup your SDK to emit debug logs for SDK requests and responses.

You can pass your own logger class directly into your SDK.
```python
from speakeasy.new_openapi import SDK
import logging

logging.basicConfig(level=logging.DEBUG)
s = SDK(debug_logger=logging.getLogger("speakeasy.new_openapi"))
```

You can also enable a default debug logger by setting an environment variable `SPEAKEASY_DEBUG` to true.
<!-- End Debugging [debug] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=python)
