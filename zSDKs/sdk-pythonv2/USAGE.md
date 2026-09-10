<!-- Start SDK Example Usage [usage] -->
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