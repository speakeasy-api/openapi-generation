# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestServers < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new
    end

    def test_select_global_server_by_name_default
      record_test("servers-select-global-server-by-name-default")

      res = @sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_select_global_server_by_name_invalid
      record_test("servers-select-global-server-by-name-invalid")

      error = nil
      begin
        SDK.new(server: :unknown)
      rescue StandardError => e
        error = e
      end

      refute_nil(error)
      assert_equal("Invalid server \"unknown\"", error.message)
    end

    def test_select_global_server_by_name_with_templates_defaults
      record_test("servers-select-global-server-by-name-with-templates-defaults")

      assert_equal(:templated, OpenApiSDK::SERVER_TEMPLATED)
      sdk = SDK.new(server: OpenApiSDK::SERVER_TEMPLATED)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_select_global_server_by_name_with_templates_valid
      record_test("servers-select-global-server-by-name-with-templates-valid")

      assert_equal(:templated, OpenApiSDK::SERVER_TEMPLATED)
      sdk = SDK.new(server: :templated, hostname: "127.0.0.1", port: HTTPBIN_PORT)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_select_global_server_by_name_with_templates_broken
      record_test("servers-select-global-server-by-name-with-templates-broken")

      sdk = SDK.new(server: OpenApiSDK::SERVER_TEMPLATED, hostname: "broken", port: "12345")
      refute_nil(sdk)
      error = nil
      begin
        sdk.servers.select_global_server
      rescue Faraday::ConnectionFailed => e
        error = e
      end

      refute_nil(error)
      assert_includes(error.message, "broken:12345")
    end
  end
end
