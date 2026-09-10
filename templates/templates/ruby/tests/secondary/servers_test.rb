# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module AlphabeticallyEarly
  class TestServers < Minitest::Test
    def test_select_global_server_valid
      record_test("servers-select-global-server-valid")

      sdk = SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
    end

    def test_select_global_server_broken
      record_test("servers-select-global-server-broken")

      sdk = SDK.new(server_url: "http://broken:12345")
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

    def test_select_global_server_by_name_default
      record_test("servers-select-global-server-by-name-default")
      sdk = SDK.new(server: nil)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
    end

    def test_select_global_server_by_name_valid
      record_test("servers-select-global-server-by-name-valid")

      assert_equal(:default_server, AlphabeticallyEarly::SERVER_DEFAULT_SERVER)
      sdk = SDK.new(server: AlphabeticallyEarly::SERVER_DEFAULT_SERVER)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
    end

    def test_select_global_server_by_name_invalid
      record_test("servers-select-global-server-by-name-invalid")

      error = nil
      begin
        SDK.new(server: :unknown_server)
      rescue StandardError => e
        error = e
      end

      refute_nil(error)
      assert_equal("Invalid server \"unknown_server\"", error.message)
    end

    def test_select_global_server_by_name_broken
      record_test("servers-select-global-server-by-name-broken")

      assert_equal(:broken_server, AlphabeticallyEarly::SERVER_BROKEN_SERVER)
      sdk = SDK.new(server: AlphabeticallyEarly::SERVER_BROKEN_SERVER)
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
