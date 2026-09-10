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

<!-- UsageSnippet language="ruby" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.operation_with_leading_and_trailing_underscores_(qp1: 'renamed')

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `qp1`                                                                                                                                            | *::String*                                                                                                                                       | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |

### Response

**[T.nilable(Operations::V2::Schemas::OperationWithLeadingAndTrailingUnderscoresResponse)](../../models/operations/operationwithleadingandtrailingunderscoresresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## post_file

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="postFile" method="post" path="/file" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: 'example.file',
    content: File.binread('example.file')
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                              | Type                                                                                   | Required                                                                               | Description                                                                            |
| -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `request`                                                                              | [Operations::V2::Schemas::PostFileRequest](../../models/operations/postfilerequest.md) | :heavy_check_mark:                                                                     | The request object to use for the request.                                             |

### Response

**[T.nilable(Operations::V2::Schemas::PostFileResponse)](../../models/operations/postfileresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::Error    | 415, 4XX         | application/json |
| Errors::Error    | 5XX              | application/json |

## get_polymorphism

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_polymorphism

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetPolymorphismResponse)](../../models/operations/getpolymorphismresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_request_body_flattened_away

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  lone_query_param: '<value>'
)
res = s.get_request_body_flattened_away

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter             | Type                  | Required              | Description           |
| --------------------- | --------------------- | --------------------- | --------------------- |
| `lone_query_param`    | *T.nilable(::String)* | :heavy_minus_sign:    | N/A                   |

### Response

**[T.nilable(Operations::V2::Schemas::GetRequestBodyFlattenedAwayResponse)](../../models/operations/getrequestbodyflattenedawayresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_fully_flattened_request

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option6: Components::SecurityOption6.new(
      client_id: '<YOUR_CLIENT_ID_HERE>',
      client_secret: '<YOUR_CLIENT_SECRET_HERE>',
      token_url: '/clientcredentials/token'
    )
  )
)
res = s.get_fully_flattened_request(lang: 'en', request_body: Operations::V2::Schemas::GetFullyFlattenedRequestRequestBody.new(
  name: '<value>'
))

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                                                                      | Type                                                                                                                           | Required                                                                                                                       | Description                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `lang`                                                                                                                         | *::String*                                                                                                                     | :heavy_check_mark:                                                                                                             | N/A                                                                                                                            |
| `request_body`                                                                                                                 | [Operations::V2::Schemas::GetFullyFlattenedRequestRequestBody](../../models/operations/getfullyflattenedrequestrequestbody.md) | :heavy_check_mark:                                                                                                             | N/A                                                                                                                            |
| `max_length`                                                                                                                   | *T.nilable(::Integer)*                                                                                                         | :heavy_minus_sign:                                                                                                             | N/A                                                                                                                            |

### Response

**[T.nilable(Operations::V2::Schemas::GetFullyFlattenedRequestResponse)](../../models/operations/getfullyflattenedrequestresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## create_with_union

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="ruby" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.create_with_union(shape_request: Components::ShapeRequest.new(
  name: '<value>',
  shape: Components::Rectangle.new(
    type: 'rectangle',
    width: 3125.73,
    height: 922.51
  )
))

unless res.object.nil?
  # handle response
end

```

### Parameters

| Parameter                                                       | Type                                                            | Required                                                        | Description                                                     |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `shape_request`                                                 | [Components::ShapeRequest](../../models/shared/shaperequest.md) | :heavy_check_mark:                                              | N/A                                                             |
| `dry_run`                                                       | *T.nilable(T::Boolean)*                                         | :heavy_minus_sign:                                              | If true, validates without creating                             |

### Response

**[T.nilable(Operations::V2::Schemas::CreateWithUnionResponse)](../../models/operations/createwithunionresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## test_endpoint

### Example Usage

<!-- UsageSnippet language="ruby" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.test_endpoint(test_name: '<value>', request_body: Operations::V2::Schemas::TestEndpointRequestBody.new(
  test: '<value>'
))

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                                              | Type                                                                                                   | Required                                                                                               | Description                                                                                            |
| ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| `test_name`                                                                                            | *::String*                                                                                             | :heavy_check_mark:                                                                                     | N/A                                                                                                    |
| `request_body`                                                                                         | [Operations::V2::Schemas::TestEndpointRequestBody](../../models/operations/testendpointrequestbody.md) | :heavy_check_mark:                                                                                     | N/A                                                                                                    |

### Response

**[T.nilable(Operations::V2::Schemas::TestEndpointResponse)](../../models/operations/testendpointresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## create_user

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="ruby" operationID="createUser" method="put" path="/user" example="paired-example" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::BaseUser.new(
  email: 'paired@example.com',
  first_name: 'John'
)
res = s.create_user(request: req)

unless res.user.nil?
  # handle response
end

```
### Example Usage: request-only

<!-- UsageSnippet language="ruby" operationID="createUser" method="put" path="/user" example="request-only" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::BaseUser.new(
  email: 'request-only@example.com'
)
res = s.create_user(request: req)

unless res.user.nil?
  # handle response
end

```
### Example Usage: response-only

<!-- UsageSnippet language="ruby" operationID="createUser" method="put" path="/user" example="response-only" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::BaseUser.new(
  id: '8ffac18c-7d88-4879-b057-e5f45b9ce7de',
  email: 'Virginie47@gmail.com',
  gender: Components::Gender::OTHER
)
res = s.create_user(request: req)

unless res.user.nil?
  # handle response
end

```

### Parameters

| Parameter                                               | Type                                                    | Required                                                | Description                                             |
| ------------------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------- |
| `request`                                               | [Components::BaseUser](../../models/shared/baseuser.md) | :heavy_check_mark:                                      | The request object to use for the request.              |

### Response

**[T.nilable(Operations::V2::Schemas::CreateUserResponse)](../../models/operations/createuserresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_user

Get User

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_user(id: '<id>')

unless res.user.nil?
  # handle response
end

```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *::String*         | :heavy_check_mark: | N/A                |

### Response

**[T.nilable(Operations::V2::Schemas::GetUserResponse)](../../models/operations/getuserresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## update_user

Update User

### Example Usage

<!-- UsageSnippet language="ruby" operationID="updateUser" method="post" path="/user/{id}" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.update_user(id: '<id>', user: Components::User.new(
  id: '8ffac18c-7d88-4879-b057-e5f45b9ce7de',
  email: 'Joanny.Feeney@gmail.com',
  gender: Components::Gender::OTHER
))

unless res.user.nil?
  # handle response
end

```

### Parameters

| Parameter                                       | Type                                            | Required                                        | Description                                     |
| ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- |
| `id`                                            | *::String*                                      | :heavy_check_mark:                              | N/A                                             |
| `user`                                          | [Components::User](../../models/shared/user.md) | :heavy_check_mark:                              | N/A                                             |

### Response

**[T.nilable(Operations::V2::Schemas::UpdateUserResponse)](../../models/operations/updateuserresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## delete_user

Delete User

### Example Usage

<!-- UsageSnippet language="ruby" operationID="deleteUser" method="delete" path="/user/{id}" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.delete_user(id: '<id>')

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *::String*         | :heavy_check_mark: | N/A                |

### Response

**[T.nilable(Operations::V2::Schemas::DeleteUserResponse)](../../models/operations/deleteuserresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## login

Login

### Example Usage

<!-- UsageSnippet language="ruby" operationID="login" method="get" path="/auth/login" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.login

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::LoginResponse)](../../models/operations/loginresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## validate

Validate

### Example Usage

<!-- UsageSnippet language="ruby" operationID="validate" method="get" path="/auth/validate" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.validate

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::ValidateResponse)](../../models/operations/validateresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## chat

### Example Usage

<!-- UsageSnippet language="ruby" operationID="chat" method="post" path="/chat" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Components::ChatModelRequest.new(
  model: 'review-model',
  prompt: 'What is the largest city in the world?',
  stream: false
)
res = s.chat(request: req)

res.chat_stream.each do |event|
  # handle event
  puts event
end


```

### Parameters

| Parameter                                                                                                   | Type                                                                                                        | Required                                                                                                    | Description                                                                                                 |
| ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `request`                                                                                                   | [T.any(Components::ChatModelRequest, Components::ChatAgentRequest)](../../models/operations/chatrequest.md) | :heavy_check_mark:                                                                                          | The request object to use for the request.                                                                  |

### Response

**[T.nilable(Operations::V2::Schemas::ChatResponse)](../../models/operations/chatresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_binary_default_response

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_binary_default_response

unless res.bytes.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetBinaryDefaultResponseResponse)](../../models/operations/getbinarydefaultresponseresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## test_enum_formats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::TestEnumFormatsRequest.new(
  string_array_format: Operations::V2::Schemas::StringArrayFormat::AWAITING_REVIEW_PROCESS,
  string_map_format: Operations::V2::Schemas::StringMapFormat::MODERATE_IMPORTANCE_LEVEL,
  string_partial_map_format: Operations::V2::Schemas::StringPartialMapFormat::INITIAL_DRAFT_VERSION,
  integer_map_format: Operations::V2::Schemas::IntegerMapFormat::SUCCESSFUL_OPERATION_COMPLETE,
  integer_partial_map_format: Operations::V2::Schemas::IntegerPartialMapFormat::PRIMARY_FIRST_OPTION
)
res = s.test_enum_formats(request: req)

unless res.object.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                                            | Type                                                                                                 | Required                                                                                             | Description                                                                                          |
| ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `request`                                                                                            | [Operations::V2::Schemas::TestEnumFormatsRequest](../../models/operations/testenumformatsrequest.md) | :heavy_check_mark:                                                                                   | The request object to use for the request.                                                           |

### Response

**[T.nilable(Operations::V2::Schemas::TestEnumFormatsResponse)](../../models/operations/testenumformatsresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## binary_and_string_upload

### Example Usage

<!-- UsageSnippet language="ruby" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::BinaryAndStringUploadRequest.new(
  binary: File.binread('test.json'),
  string: File.read("test.json")
)
res = s.binary_and_string_upload(request: req)

if res.status_code == 200
  # handle response
end

```

### Parameters

| Parameter                                                                                                        | Type                                                                                                             | Required                                                                                                         | Description                                                                                                      |
| ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `request`                                                                                                        | [Operations::V2::Schemas::BinaryAndStringUploadRequest](../../models/operations/binaryandstringuploadrequest.md) | :heavy_check_mark:                                                                                               | The request object to use for the request.                                                                       |

### Response

**[T.nilable(Operations::V2::Schemas::BinaryAndStringUploadResponse)](../../models/operations/binaryandstringuploadresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_duplicate_export_collision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_duplicate_export_collision

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetDuplicateExportCollisionResponse)](../../models/operations/getduplicateexportcollisionresponse.md)**

### Errors

| Error Type                  | Status Code                 | Content Type                |
| --------------------------- | --------------------------- | --------------------------- |
| Errors::RequestTimeoutError | 408                         | application/json            |
| Errors::APIError            | 4XX, 5XX                    | \*/\*                       |

## get_named_primitive_union

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_named_primitive_union

unless res.some_union.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNamedPrimitiveUnionResponse)](../../models/operations/getnamedprimitiveunionresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_empty_object_error

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_empty_object_error

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetEmptyObjectErrorResponse)](../../models/operations/getemptyobjecterrorresponse.md)**

### Errors

| Error Type                  | Status Code                 | Content Type                |
| --------------------------- | --------------------------- | --------------------------- |
| Errors::FailedResponseError | 500                         | application/json            |
| Errors::APIError            | 4XX, 5XX                    | \*/\*                       |

## url_validation_stress_test

### Example Usage

<!-- UsageSnippet language="ruby" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.url_validation_stress_test

if res.status_code == 200
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::UrlValidationStressTestResponse)](../../models/operations/urlvalidationstresstestresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## parentheses_in_path_allowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="ruby" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.parentheses_in_path_allowed(id: '<id>', template_braces_test: Components::TemplateBracesTest.new(
  field_with_braces_in_description: "A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n",
  field_with_braces_in_title: "A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n",
  field_with_braces_in_example: "A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n"
))

unless res.template_braces_test.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                   | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `id`                                                                        | *::String*                                                                  | :heavy_check_mark:                                                          | N/A                                                                         |
| `template_braces_test`                                                      | [Components::TemplateBracesTest](../../models/shared/templatebracestest.md) | :heavy_check_mark:                                                          | N/A                                                                         |

### Response

**[T.nilable(Operations::V2::Schemas::ParenthesesInPathAllowedResponse)](../../models/operations/parenthesesinpathallowedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_nested_integer_string

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_nested_integer_string

unless res.task_response.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetNestedIntegerStringResponse)](../../models/operations/getnestedintegerstringresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## render_asset

Render a media asset from a text prompt.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="renderAsset" method="post" path="/assets/render" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::RenderAssetRequest.new(
  prompt: '<value>'
)
res = s.render_asset(request: req)

res.asset_stream.each do |event|
  # handle event
  puts event
end


```

### Parameters

| Parameter                                                                                    | Type                                                                                         | Required                                                                                     | Description                                                                                  |
| -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `request`                                                                                    | [Operations::V2::Schemas::RenderAssetRequest](../../models/operations/renderassetrequest.md) | :heavy_check_mark:                                                                           | The request object to use for the request.                                                   |

### Response

**[T.nilable(Operations::V2::Schemas::RenderAssetResponse)](../../models/operations/renderassetresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_asset

Get the current result of an asset job.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getAsset" method="get" path="/assets/{id}" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_asset(id: '<id>', stream: true)

res.asset_status_stream.each do |event|
  # handle event
  puts event
end


```

### Parameters

| Parameter               | Type                    | Required                | Description             |
| ----------------------- | ----------------------- | ----------------------- | ----------------------- |
| `id`                    | *::String*              | :heavy_check_mark:      | N/A                     |
| `stream`                | *T.nilable(T::Boolean)* | :heavy_minus_sign:      | N/A                     |

### Response

**[T.nilable(Operations::V2::Schemas::GetAssetResponse)](../../models/operations/getassetresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## get_error_only_example

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.get_error_only_example

unless res.object.nil?
  # handle response
end

```

### Response

**[T.nilable(Operations::V2::Schemas::GetErrorOnlyExampleResponse)](../../models/operations/geterroronlyexampleresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::Error    | 404              | application/json |
| Errors::APIError | 4XX, 5XX         | \*/\*            |