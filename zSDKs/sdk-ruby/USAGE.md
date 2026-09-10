<!-- Start SDK Example Usage [usage] -->
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