# Tag1

## Overview

The first tag.

### Available Operations

* [~~deprecated1~~](#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [get_request_body_flattened_away](docs/sdks/sdk/README.md#get_request_body_flattened_away) instead.
* [auth](#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [list_test1](#list_test1) - Get Test1
* [post_file_with_encoding](#post_file_with_encoding) - Post File With Encoding

## ~~deprecated1~~

Deprecated Operation

> :warning: **DEPRECATED**: This endpoint is deprecated.. Use `get_request_body_flattened_away` instead.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="deprecated1" method="get" path="/deprecated" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.tag1.deprecated1

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

## auth

This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK


### Example Usage

<!-- UsageSnippet language="ruby" operationID="auth" method="get" path="/auth" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.tag1.auth(security: Operations::V2::Schemas::AuthSecurity.new(
  access_token: '<YOUR_ACCESS_TOKEN_HERE>'
))

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                            | Type                                                                                 | Required                                                                             | Description                                                                          |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `security`                                                                           | [Operations::V2::Schemas::AuthSecurity](../../operations/v2/schemas/authsecurity.md) | :heavy_check_mark:                                                                   | The security requirements to use for the request.                                    |

### Response

**[T.nilable(Operations::V2::Schemas::AuthResponse)](../../models/operations/authresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## list_test1

This is a {{test}} endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="listTest1" method="get" path="/test1/{page}" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  query_param1: 'some example query param',
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: '<YOUR_API_KEY_HERE>'
    )
  )
)
res = s.tag1.list_test1(page: 100, query_param2: Operations::V2::Schemas::QueryParam2::ONE, header_param1: 'some example header param')

unless res.object.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                                                                                                                                                                                   | Type                                                                                                                                                                                                                                        | Required                                                                                                                                                                                                                                    | Description                                                                                                                                                                                                                                 | Example                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `page`                                                                                                                                                                                                                                      | *::Integer*                                                                                                                                                                                                                                 | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | 100                                                                                                                                                                                                                                         |
| `query_param2`                                                                                                                                                                                                                              | [Operations::V2::Schemas::QueryParam2](../../models/operations/queryparam2.md)                                                                                                                                                              | :heavy_check_mark:                                                                                                                                                                                                                          | An [enum](https://enum.com) "query parameter"<br/>that is not easily described in a single line.<br/><br/>**Available Values:**<br/>\| Value \| Description \|<br/>\|-------\|-------------\|<br/>\| 0     \| No data     \|<br/>\| 1     \| Partial     \|<br/>\| 2     \| Complete    \| | 1                                                                                                                                                                                                                                           |
| `header_param1`                                                                                                                                                                                                                             | *::String*                                                                                                                                                                                                                                  | :heavy_check_mark:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example header param                                                                                                                                                                                                                   |
| `query_param1`                                                                                                                                                                                                                              | *T.nilable(::String)*                                                                                                                                                                                                                       | :heavy_minus_sign:                                                                                                                                                                                                                          | N/A                                                                                                                                                                                                                                         | some example query param                                                                                                                                                                                                                    |
| `server_url`                                                                                                                                                                                                                                | *String*                                                                                                                                                                                                                                    | :heavy_minus_sign:                                                                                                                                                                                                                          | An optional server URL to use.                                                                                                                                                                                                              | http://localhost:8080                                                                                                                                                                                                                       |

### Response

**[T.nilable(Operations::V2::Schemas::ListTest1Response)](../../models/operations/listtest1response.md)**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| Errors::BadRequestResponseError | 400                             | application/json                |
| Errors::Error                   | 500                             | application/json                |
| Errors::APIError                | 4XX, 5XX                        | \*/\*                           |

## post_file_with_encoding

This endpoint tests the encoding field with multipart/form-data content type.
According to OpenAPI 3.0.3 spec, the encoding field is valid for both
application/x-www-form-urlencoded and multipart/* media types.

This test includes multiple content types for the file field to verify
handling of comma-separated content types in encoding.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="postFileWithEncoding" method="post" path="/fileWithEncoding" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::PostFileWithEncodingRequest.new(
  file: Operations::V2::Schemas::File.new(
    file_name: 'example.file',
    content: File.binread('example.file')
  )
)
res = s.tag1.post_file_with_encoding(request: req)

unless res.object.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                                                      | Type                                                                                                           | Required                                                                                                       | Description                                                                                                    |
| -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `request`                                                                                                      | [Operations::V2::Schemas::PostFileWithEncodingRequest](../../models/operations/postfilewithencodingrequest.md) | :heavy_check_mark:                                                                                             | The request object to use for the request.                                                                     |

### Response

**[T.nilable(Operations::V2::Schemas::PostFileWithEncodingResponse)](../../models/operations/postfilewithencodingresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::Error    | 415              | application/json |
| Errors::APIError | 4XX, 5XX         | \*/\*            |