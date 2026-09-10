# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestFlattening < Minitest::Test
    def test_component_body_and_param_no_conflict
      record_test("flattening-component-body-and-param-no-conflict")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      obj = create_simple_object

      res = @sdk.flattening.component_body_and_param_no_conflict(simple_object: obj, param_str: "param test")
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal(res.res.args["paramStr"], "param test")
      assert_equal(res.res.json, obj)
    end

    def test_component_body_and_param_conflict
      record_test("flattening-component-body-and-param-conflict")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      obj = create_simple_object

      res = @sdk.flattening.component_body_and_param_conflict(simple_object: obj, str_: "param test")
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal(res.res.args["str"], "param test")
      assert_equal(res.res.json, obj)
    end

    def test_inline_body_and_param_conflict
      record_test("flattening-inline-body-and-param-conflict")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      res = @sdk.flattening.inline_body_and_param_conflict(
        request_body: Models::Operations::InlineBodyAndParamConflictRequestBody.new(
          str_: "body test"
        ),
        str_: "param test"
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal(res.res.args["str"], "param test")
      assert_equal(res.res.json.str_, "body test")
    end

    def test_inline_body_and_param_no_conflict
      record_test("flattening-inline-body-and-param-no-conflict")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      res = @sdk.flattening.inline_body_and_param_no_conflict(
        request_body: Models::Operations::InlineBodyAndParamNoConflictRequestBody.new(
          body_str: "body test"
        ),
        param_str: "param test"
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal(res.res.args["paramStr"], "param test")
      assert_equal(res.res.json.body_str, "body test")
    end

    def test_conflicting_params
      record_test("flattening-conflicting-params")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      res = @sdk.flattening.conflicting_params(str_path_parameter: "pathParam", str_query_parameter: "queryParam")
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_match(%r{/pathParam?}, res.res.url)
      assert_equal(res.res.args["str"], "queryParam")
    end

    def test_required_body_all_optional
      record_test("flattening-required-body-all-optional")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(@sdk)

      res = @sdk.flattening.required_body_all_optional(
        request: Models::Shared::ObjWithOptionalProperties.new(
          opt_str: "body test",
          opt_int: 1
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_equal(res.res.json.opt_str, "body test")
      assert_equal(res.res.json.opt_int, 1)

      assert_raises(ArgumentError, "wrong number of arguments (given 0, expected 1)") do
        @sdk.flattening.required_body_all_optional
      end
    end

  end
end
