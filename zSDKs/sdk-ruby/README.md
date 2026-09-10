# openapi

Developer-friendly & type-safe Ruby SDK specifically catered to leverage *openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=ruby)
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
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Global Parameters](#global-parameters)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

The SDK can be installed using [RubyGems](https://rubygems.org/):

```bash
gem install specific_install
gem specific_install https://github.com/speakeasy-sdks/test-sdk 
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

### Example 2

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

req = Operations::V2::Schemas::PostFileWithEncodingRequest.new(
  file: Operations::V2::Schemas::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.tag1.post_file_with_encoding(request: req)

unless res.object.nil?
  # handle response
end

```

### Example 3

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  deprecated_query_param1: "some example query param",
  deprecated_query_param2: "some example query param",
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)
res = s.test_group.tag2.post_test(
  test2_request: Components::Test2Request.new(
    obj: Components::ExhaustiveObject.new(
      str_: "example",
      bool: true,
      integer: 999_999,
      int32: 1,
      num: 1.1,
      float32: 8499.3,
      date: Date.parse("2020-01-01"),
      date_time: DateTime.iso8601("2020-01-01T00:00:00Z"),
      anything: "<value>",
      bool_opt: true,
      int_opt_null: 999_999,
      num_opt_null: 1.1,
      int_enum: Components::IntEnum::THIRD,
      int32_enum: Components::Int32Enum::SIXTY_NINE,
      bigint: 702_830,
      decimal_str: "<value>",
      obj: Components::SimpleObject.new(
        str_: "example"
      ),
      map: {
        "key" => Components::SimpleObject.new(
          str_: "example"
        )
      },
      arr: [
        Components::SimpleObject.new(
          str_: "example"
        )
      ],
      any: "<value>",
      nullable_int_enum: Components::NullableIntEnum::THIRD,
      nullable_string_enum: Components::NullableStringEnum::SECOND,
      color: Components::Color::GREEN,
      icon: Components::Icon::TICK,
      hero_width: Components::HeroWidth::FOUR_HUNDRED_AND_EIGHTY
    ),
    type: Components::Type::SUPER_TYPE1
  )
)

unless res.body.nil?
  # handle response
end

```

### A custom readme heading

A custom usage description

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  query_param1: "some example query param",
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)
res = s
  .tag1
  .list_test1(
    page: 100,
    query_param2: Operations::V2::Schemas::QueryParam2::ONE,
    header_param1: "some example header param"
  )

unless res.object.nil?
  # handle response
end

```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives by setting the `security` optional parameter when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     |
| ------------------------- | ---- | ---------- |
| `username`<br/>`password` | http | HTTP Basic |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    user_pass_auth: Components::UserPassAuth.new(
      username: "<USERNAME>",
      password: "<PASSWORD>"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name          | Type   | Scheme      |
| ------------- | ------ | ----------- |
| `bearer_auth` | http   | HTTP Bearer |
| `my_api_key`  | apiKey | API key     |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option2: Components::SecurityOption2.new(
      bearer_auth: "<YOUR_JWT>",
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       |
| -------- | ------ | ------------ |
| `oauth2` | oauth2 | OAuth2 token |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option3: Components::SecurityOption3.new(
      oauth2: "Bearer <YOUR_OAUTH2_TOKEN>"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                  | Type | Scheme      |
| --------------------- | ---- | ----------- |
| `app_id`<br/>`secret` | http | Custom HTTP |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option4: Components::SecurityOption4.new(
      app_id: "app-speakeasy-123",
      secret: "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### Option5

The `Option5` alternative relies on the following scheme:

| Name          | Type   | Scheme       |
| ------------- | ------ | ------------ |
| `mobile_auth` | oauth2 | OAuth2 token |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option5: Components::SecurityOption5.new(
      mobile_auth: "Bearer <YOUR_OAUTH2_TOKEN>"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                            | Type   | Scheme                         |
| ----------------------------------------------- | ------ | ------------------------------ |
| `client_id`<br/>`client_secret`<br/>`token_url` | oauth2 | OAuth2 Client Credentials Flow |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    option6: Components::SecurityOption6.new(
      client_id: "<YOUR_CLIENT_ID_HERE>",
      client_secret: "<YOUR_CLIENT_SECRET_HERE>",
      token_url: "/clientcredentials/token"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name         | Type   | Scheme  |
| ------------ | ------ | ------- |
| `my_api_key` | apiKey | API key |

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.tag1.auth(
  security: Operations::V2::Schemas::AuthSecurity.new(
    access_token: "<YOUR_ACCESS_TOKEN_HERE>"
  )
)

if res.status_code == 200
  # handle response
end

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

For example, you can set `queryParam1` to `'some example query param'` at SDK initialization and then you do not have to pass the same value on calls to operations like `get_request_body_flattened_away`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.

| Name                    | Type     | Description                                                                        |
| ----------------------- | -------- | ---------------------------------------------------------------------------------- |
| query_param1            | ::String | A long winded, multi-line description<br/>for the query parameter number one.<br/> |
| deprecated_query_param1 | ::String | A deprecated description                                                           |
| deprecated_query_param2 | ::String | The deprecated_query_param2 parameter.                                             |
| lone_query_param        | ::String | The lone_query_param parameter.                                                    |

### Example

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  lone_query_param: "<value>",
  query_param1: "some example query param",
  deprecated_query_param1: "some example query param",
  deprecated_query_param2: "some example query param"
)
res = s.get_request_body_flattened_away

if res.status_code == 200
  # handle response
end

```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide a `RetryConfig` object to the call:
```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  deprecated_query_param1: "some example query param",
  deprecated_query_param2: "some example query param",
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)
res = s.test_group.tag2.post_test(
  test2_request: Components::Test2Request.new(
    obj: Components::ExhaustiveObject.new(
      str_: "example",
      bool: true,
      integer: 999_999,
      int32: 1,
      num: 1.1,
      float32: 8499.3,
      date: Date.parse("2020-01-01"),
      date_time: DateTime.iso8601("2020-01-01T00:00:00Z"),
      anything: "<value>",
      bool_opt: true,
      int_opt_null: 999_999,
      num_opt_null: 1.1,
      int_enum: Components::IntEnum::THIRD,
      int32_enum: Components::Int32Enum::SIXTY_NINE,
      bigint: 702_830,
      decimal_str: "<value>",
      obj: Components::SimpleObject.new(
        str_: "example"
      ),
      map: {
        "key" => Components::SimpleObject.new(
          str_: "example"
        )
      },
      arr: [
        Components::SimpleObject.new(
          str_: "example"
        )
      ],
      any: "<value>",
      nullable_int_enum: Components::NullableIntEnum::THIRD,
      nullable_string_enum: Components::NullableStringEnum::SECOND,
      color: Components::Color::GREEN,
      icon: Components::Icon::TICK,
      hero_width: Components::HeroWidth::FOUR_HUNDRED_AND_EIGHTY
    ),
    type: Components::Type::SUPER_TYPE1
  )
)

unless res.body.nil?
  # handle response
end

```

If you'd like to override the default retry strategy for all operations that support retries, you can use the `retry_config` optional parameter when initializing the SDK:
```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  retry_config: Utils::RetryConfig.new(
    backoff: Utils::BackoffStrategy.new(
      exponent: 1.1,
      initial_interval: 1,
      max_elapsed_time: 100,
      max_interval: 50
    ),
    retry_connection_errors: false,
    strategy: "backoff"
  ),
  deprecated_query_param1: "some example query param",
  deprecated_query_param2: "some example query param",
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)
res = s.test_group.tag2.post_test(
  test2_request: Components::Test2Request.new(
    obj: Components::ExhaustiveObject.new(
      str_: "example",
      bool: true,
      integer: 999_999,
      int32: 1,
      num: 1.1,
      float32: 8499.3,
      date: Date.parse("2020-01-01"),
      date_time: DateTime.iso8601("2020-01-01T00:00:00Z"),
      anything: "<value>",
      bool_opt: true,
      int_opt_null: 999_999,
      num_opt_null: 1.1,
      int_enum: Components::IntEnum::THIRD,
      int32_enum: Components::Int32Enum::SIXTY_NINE,
      bigint: 702_830,
      decimal_str: "<value>",
      obj: Components::SimpleObject.new(
        str_: "example"
      ),
      map: {
        "key" => Components::SimpleObject.new(
          str_: "example"
        )
      },
      arr: [
        Components::SimpleObject.new(
          str_: "example"
        )
      ],
      any: "<value>",
      nullable_int_enum: Components::NullableIntEnum::THIRD,
      nullable_string_enum: Components::NullableStringEnum::SECOND,
      color: Components::Color::GREEN,
      icon: Components::Icon::TICK,
      hero_width: Components::HeroWidth::FOUR_HUNDRED_AND_EIGHTY
    ),
    type: Components::Type::SUPER_TYPE1
  )
)

unless res.body.nil?
  # handle response
end

```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or raise an error.

By default an API error will raise a `Errors::APIError`, which has the following properties:

| Property       | Type                                    | Description           |
|----------------|-----------------------------------------|-----------------------|
| `message`     | *string*                                 | The error message     |
| `status_code`  | *int*                                   | The HTTP status code  |
| `raw_response` | *Faraday::Response*                     | The raw HTTP response |
| `body`        | *string*                                 | The response content  |

When custom error responses are specified for an operation, the SDK may also throw their associated exception. You can refer to respective *Errors* tables in SDK docs for more details on possible exception types for each operation. For example, the `post_file` method throws the following exceptions:

| Error Type    | Status Code | Content Type     |
| ------------- | ----------- | ---------------- |
| Errors::Error | 415, 4XX    | application/json |
| Errors::Error | 5XX         | application/json |

### Example

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new

begin
  req = Operations::V2::Schemas::PostFileRequest.new(
    upload: Components::File.new(
      file_name: "example.file",
      content: File.binread("example.file")
    )
  )
  res = s.post_file(request: req)

  unless res.file.nil?
    # handle response
  end

rescue Errors::Error => e
  # handle e.container data
  raise e
rescue Errors::Error => e
  # handle e.container data
  raise e
rescue Errors::APIError => e
  # handle default exception
  raise e
end

```
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally by passing a server index to the `server_idx (Integer)` optional parameter when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values through the additional parameters made available in the SDK constructor:

| Variable    | Parameter                                                      | Supported Values                      | Default       | Description                              |
| ----------- | -------------------------------------------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `subdomain (::String)`                                         | ::String                              | `"api"`       |                                          |
| `version`   | `version (::String)`                                           | ::String                              | `"1"`         |                                          |
| `HostName`  | `host_name (::String)`                                         | ::String                              | `"localhost"` | The hostname of the server.              |
| `PORT`      | `port (::OpenApiSDK::Components::ServerVariables::ServerPORT)` | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  server_idx: 2,
  host_name: "localhost",
  port: "443"
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

### Override Server URL Per-Client

The default server can also be overridden globally by passing a URL to the `server_url (String)` optional parameter when initializing the SDK client instance. For example:
```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  server_url: "http://localhost:8080"
)

req = Operations::V2::Schemas::PostFileRequest.new(
  upload: Components::File.new(
    file_name: "example.file",
    content: File.binread("example.file")
  )
)
res = s.post_file(request: req)

unless res.file.nil?
  # handle response
end

```

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```ruby
require "openapi"

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  query_param1: "some example query param",
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: "<YOUR_API_KEY_HERE>"
    )
  )
)
res = s
  .tag1
  .list_test1(
    server_url: "http://localhost:35123",
    page: 100,
    query_param2: Operations::V2::Schemas::QueryParam2::ONE,
    header_param1: "some example header param"
  )

unless res.object.nil?
  # handle response
end

```
<!-- End Server Selection [server] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi&utm_campaign=ruby)
