# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestServers < Minitest::Test
    def test_select_global_server_by_id_default
      record_test("servers-select-global-server-by-id-default")
      sdk = SDK.new
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_select_global_server_by_id_valid
      record_test("servers-select-global-server-by-id-valid")

      sdk = SDK.new(server_idx: 0)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_select_global_server_by_id_invalid
      record_test("servers-select-global-server-by-id-invalid")

      error = nil
      begin
        SDK.new(server_idx: 2)
      rescue StandardError => e
        error = e
      end

      refute_nil(error)
      assert_equal("Invalid server index 2", error.message)
    end

    def test_select_global_server_by_id_broken
      record_test("servers-select-global-server-by-id-broken")

      sdk = SDK.new(server_idx: 1)
      refute_nil(sdk)

      error = nil
      begin
        res = sdk.servers.select_global_server
      rescue Faraday::ConnectionFailed => e
        error = e
      end

      refute_nil(error)
      assert_nil(res)
    end
  end
end
