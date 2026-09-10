# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestEnumsAdditional < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_open_enums_round_trip
      record_test("open-enums-round-trip")

      # Send request with unknown enum values: 'purple' for Color, 'tick' for Icon, 2160 for HeroWidth
      res = @sdk.enums.enums_post_open_enum_unrecognized(
        request: Models::Shared::ThemeRequestOpaque.new(
          color: "purple",
          icon: "tick",
          hero_width: 2160
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      theme = res.theme_response&.json
      refute_nil(theme)

      # 'purple' is unknown for Color (open enum) - should deserialize without raising
      assert_equal("purple", theme.color.serialize)
      # 'tick' is a known value for Icon - should return the constant
      assert_equal("tick", theme.icon.serialize)
      # 2160 is unknown for HeroWidth (open enum) - should deserialize without raising
      assert_equal(2160, theme.hero_width.serialize)

      # Round-trip: send the deserialized values back
      res2 = @sdk.enums.enums_post_open_enum_unrecognized(
        request: Models::Shared::ThemeRequestOpaque.new(
          color: theme.color.serialize.to_s,
          icon: theme.icon.serialize.to_s,
          hero_width: theme.hero_width.serialize.to_i
        )
      )
      refute_nil(res2)
      assert_equal(Rack::Utils.status_code(:ok), res2.http_meta.response.status)

      theme2 = res2.theme_response&.json
      refute_nil(theme2)
      assert_equal("purple", theme2.color.serialize)
      assert_equal("tick", theme2.icon.serialize)
      assert_equal(2160, theme2.hero_width.serialize)
    end

    def test_open_enums_round_trip_string_union
      record_test("open-enums-round-trip-string-union")

      res = @sdk.enums.enums_post_open_enum_unrecognized(
        request: Models::Shared::ThemeRequestOpaque.new(
          color: "purple",
          icon: "tick",
          hero_width: 2160
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      theme = res.theme_response&.json
      refute_nil(theme)

      assert_equal("purple", theme.color.serialize)
      assert_equal("tick", theme.icon.serialize)
      assert_equal(2160, theme.hero_width.serialize)

      # Round-trip: pass the deserialized response values back as a new request
      res2 = @sdk.enums.enums_post_open_enum_unrecognized(
        request: Models::Shared::ThemeRequestOpaque.new(
          color: theme.color.serialize.to_s,
          icon: theme.icon.serialize.to_s,
          hero_width: theme.hero_width.serialize.to_i
        )
      )
      refute_nil(res2)
      assert_equal(Rack::Utils.status_code(:ok), res2.http_meta.response.status)

      theme2 = res2.theme_response&.json
      refute_nil(theme2)
      assert_equal("purple", theme2.color.serialize)
      assert_equal("tick", theme2.icon.serialize)
      assert_equal(2160, theme2.hero_width.serialize)
    end

    def test_open_enums_response_usage
      record_test("open-enums-response-usage")

      # Step 1: request-only enum should remain closed (known value)
      res1 = @sdk.open_enums.open_enums_response_usage_request_only(
        request: Models::Shared::ObjectWithEnumInRequestOnly.new(
          action: Models::Shared::EnumUsedInRequestOnly::CREATE,
          data: "test data"
        )
      )
      refute_nil(res1)
      assert_equal(Rack::Utils.status_code(:ok), res1.http_meta.response.status)

      # Step 2: explicitly open enum (known value)
      res2 = @sdk.open_enums.open_enums_response_usage_explicitly_open(
        request: Models::Shared::ObjectWithEnumExplicitlyOpen.new(
          mode: Models::Shared::EnumUsedInRequestExplicitlyOpen::ALPHA,
          value: "test value"
        )
      )
      refute_nil(res2)
      assert_equal(Rack::Utils.status_code(:ok), res2.http_meta.response.status)

      # Step 3: enum used in both request and response (known value)
      res3 = @sdk.open_enums.open_enums_response_usage_both_request_and_response(
        request: Models::Shared::ObjectWithEnumInBoth.new(
          state: Models::Shared::EnumUsedInBothRequestAndResponse::ACTIVE,
          description: "known value test"
        )
      )
      refute_nil(res3)
      assert_equal(Rack::Utils.status_code(:ok), res3.http_meta.response.status)
    end

    def test_parameter_open_enum
      record_test("parameters-open-enum")

      res = @sdk.parameters.parameter_open_enum(
        param_p: Models::Shared::OpenEnum::ONE_HUNDRED_AND_ONE,
        param_q: Models::Shared::OpenEnum::FOUR_HUNDRED_AND_FOUR,
        param_h: Models::Shared::OpenEnum::ONE_HUNDRED_AND_ONE
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal("#{CommonHelpers.httpbin_url}/anything/openEnum/101/suffix?param-q=404", res.res.url)
      assert_equal("101", res.res.headers["Param-H"])
    end
  end
end
