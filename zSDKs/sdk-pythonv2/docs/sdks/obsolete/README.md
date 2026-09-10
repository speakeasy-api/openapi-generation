# ~~Obsolete~~

> [!WARNING]
> This SDK is **DEPRECATED**

## Overview

A subSDK in which all operations are deprecated.

### Available Operations

* [~~deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away) instead.

## ~~deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `get_request_body_flattened_away` instead.

### Example Usage

<!-- UsageSnippet language="python" operationID="deprecated1" method="get" path="/deprecated" -->
```python
# Synchronous Example
from speakeasy.new_openapi import SDK


with SDK() as sdk:

    sdk.obsolete.deprecated1()

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

        await sdk.obsolete.deprecated1()

        # Use the SDK ...

asyncio.run(main())
```

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `retries`                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |
| `server_url`                                                        | *Optional[str]*                                                     | :heavy_minus_sign:                                                  | An optional server URL to use.                                      |

### Errors

| Error Type      | Status Code     | Content Type    |
| --------------- | --------------- | --------------- |
| models.SDKError | 4XX, 5XX        | \*/\*           |