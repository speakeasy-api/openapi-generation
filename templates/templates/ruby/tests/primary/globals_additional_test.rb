# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestGlobals < Minitest::Test

    def test_globals_query_parameter_get_uses_global
      record_test("globals-query-parameter-get-uses-global")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_query_param: "test")
      refute_nil(@sdk)

      res = @sdk.globals.globals_query_parameter_get
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.args.global_query_param, "test")
    end

    def test_globals_query_parameter_get_uses_local
      record_test("globals-query-parameter-get-uses-local")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_query_param: "test")
      refute_nil(@sdk)

      res = @sdk.globals.globals_query_parameter_get(global_query_param: "local")

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.args.global_query_param, "local")
    end

    def test_global_path_parameter_get_uses_global
      record_test("globals-path-parameter-get-uses-global")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_path_param: 1)
      refute_nil(@sdk)

      res = @sdk.globals.global_path_parameter_get
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.url, "#{CommonHelpers.httpbin_url}/anything/globals/pathParameter/1")
    end

    def test_global_path_parameter_get_uses_local
      record_test("globals-path-parameter-get-uses-local")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_path_param: 1)
      refute_nil(@sdk)

      res = @sdk.globals.global_path_parameter_get(global_path_param: 2)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.url, "#{CommonHelpers.httpbin_url}/anything/globals/pathParameter/2")
    end

    def test_global_header_get_uses_global
      record_test("globals-header-get-uses-global")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_header_param: true)
      refute_nil(@sdk)

      res = @sdk.globals.globals_header_get
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.headers["Globalheaderparam"], "true")
    end

    def test_global_header_get_uses_local
      record_test("globals-header-get-uses-local")

      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, global_header_param: true)
      refute_nil(@sdk)

      res = @sdk.globals.globals_header_get(global_header_param: false)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.res.headers["Globalheaderparam"], "false")
    end
  end
end
