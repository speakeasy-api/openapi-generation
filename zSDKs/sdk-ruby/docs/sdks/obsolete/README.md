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

<!-- UsageSnippet language="ruby" operationID="deprecated1" method="get" path="/deprecated" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.obsolete.deprecated1

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                      | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `server_url`                   | *String*                       | :heavy_minus_sign:             | An optional server URL to use. |

### Response

**[T.nilable(Operations::V2::Schemas::Deprecated1Response)](../../models/operations/deprecated1response.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |