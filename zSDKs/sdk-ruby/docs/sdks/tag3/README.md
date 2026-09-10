# TestGroup.Tag3

## Overview

### Available Operations

* [post_test](#post_test) - Post Test2

## post_test

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="ruby" operationID="postTest2" method="post" path="/test2" -->
```ruby
require 'openapi'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  deprecated_query_param1: 'some example query param',
  deprecated_query_param2: 'some example query param',
  security: Components::Security.new(
    my_api_key: Components::MyApiKey.new(
      my_api_key: '<YOUR_API_KEY_HERE>'
    )
  )
)
res = s.test_group.tag3.post_test(test2_request: Components::Test2Request.new(
  obj: Components::ExhaustiveObject.new(
    str_: 'example',
    bool: true,
    integer: 999_999,
    int32: 1,
    num: 1.1,
    float32: 8499.3,
    date: Date.parse('2020-01-01'),
    date_time: DateTime.iso8601('2020-01-01T00:00:00Z'),
    anything: '<value>',
    bool_opt: true,
    int_opt_null: 999_999,
    num_opt_null: 1.1,
    int_enum: Components::IntEnum::THIRD,
    int32_enum: Components::Int32Enum::SIXTY_NINE,
    bigint: 702_830,
    decimal_str: '<value>',
    obj: Components::SimpleObject.new(
      str_: 'example'
    ),
    map: {
      'key' => Components::SimpleObject.new(
        str_: 'example'
      ),
    },
    arr: [
      Components::SimpleObject.new(
        str_: 'example'
      ),
    ],
    any: '<value>',
    nullable_int_enum: Components::NullableIntEnum::THIRD,
    nullable_string_enum: Components::NullableStringEnum::SECOND,
    color: Components::Color::GREEN,
    icon: Components::Icon::TICK,
    hero_width: Components::HeroWidth::FOUR_HUNDRED_AND_EIGHTY
  ),
  type: Components::Type::SUPER_TYPE1
))

unless res.body.nil?
  # handle response
end

```

### Parameters

| Parameter                                                                                                               | Type                                                                                                                    | Required                                                                                                                | Description                                                                                                             | Example                                                                                                                 |
| ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `test2_request`                                                                                                         | [Components::Test2Request](../../models/shared/test2request.md)                                                         | :heavy_check_mark:                                                                                                      | N/A                                                                                                                     |                                                                                                                         |
| `deprecated_query_param1`                                                                                               | *T.nilable(::String)*                                                                                                   | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `deprecated_query_param2`                                                                                               | *T.nilable(::String)*                                                                                                   | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `server_url`                                                                                                            | *String*                                                                                                                | :heavy_minus_sign:                                                                                                      | An optional server URL to use.                                                                                          | http://localhost:8080                                                                                                   |

### Response

**[T.nilable(Operations::V2::Schemas::PostTest2Response)](../../models/operations/posttest2response.md)**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| Errors::BadRequestResponseError | 400                             | application/json                |
| Errors::Error                   | 404                             | application/json                |
| Errors::Test2ResponseError      | 500                             | application/json                |
| Errors::APIError                | 4XX, 5XX                        | \*/\*                           |