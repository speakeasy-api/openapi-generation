# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestServers < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new
    end

    def test_select_global_server_valid
      record_test("servers-select-global-server-valid")

      sdk = SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      sdk = SDK.new(
        server_url: SERVERS[2],
        url_params: {hostname: "localhost", port: HTTPBIN_PORT}
      )
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      sdk = SDK.new(
        server_url: "http://{hostname}:{port}",
        hostname: "localhost",
        port: HTTPBIN_PORT
      )
      refute_nil(sdk)
      res = sdk.servers.select_global_server
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_select_global_server_broken
      record_test("servers-select-global-server-broken")

      error = nil
      begin
        SDK.new(server_idx: 6)
      rescue StandardError => e
        error = e
      end

      refute_nil(error)
      assert_equal("Invalid server index 6", error.message)

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

    def test_select_server_with_id_default
      record_test("servers-select-server-with-id-default")

      # broken server overridden by operation
      sdk = SDK.new(server_url: SERVERS[1])
      refute_nil(sdk)

      res = sdk.servers.select_server_with_id
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_select_server_with_id_valid
      record_test("servers-select-server-with-id-valid")

      # broken server overridden by operation
      sdk = SDK.new(server_url: SERVERS[1])
      refute_nil(sdk)

      res = sdk
        .servers
        .select_server_with_id(
          server_url: ::OpenApiSDK::Servers::SELECT_SERVER_WITH_ID_SERVERS[
            ::OpenApiSDK::Servers::SELECT_SERVER_WITH_ID_SERVERS_VALID
          ]
        )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_select_server_with_id_broken
      record_test("servers-select-server-with-id-broken")

      # default server (valid) overridden by operation
      sdk = SDK.new
      refute_nil(sdk)

      error = nil
      begin
        res = sdk
          .servers
          .select_server_with_id(
            server_url: ::OpenApiSDK::Servers::SELECT_SERVER_WITH_ID_SERVERS[
              ::OpenApiSDK::Servers::SELECT_SERVER_WITH_ID_SERVERS_BROKEN
            ]
          )
      rescue Faraday::ConnectionFailed => e
        error = e
      end

      refute_nil(error)
      assert_nil(res)
    end

    def test_server_with_templates_global
      record_test("servers-server-with-templates-global")

      sdk = SDK.new(server_idx: 2, hostname: "localhost", port: HTTPBIN_PORT)
      refute_nil(sdk)

      res = sdk.servers.server_with_templates_global
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_server_with_templates_global_defaults
      record_test("servers-server-with-templates-global-defaults")

      sdk = SDK.new(server_idx: 2)
      refute_nil(sdk)

      res = sdk.servers.server_with_templates_global
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_server_with_templates_global_enum
      record_test("servers-server-with-templates-global-enum")

      sdk = SDK.new(
        server_idx: 3,
        something: ::OpenApiSDK::Models::ServerVariables::ServerSomething::SOMETHING_ELSE_AGAIN
      )
      refute_nil(sdk)

      res = sdk.servers.server_with_templates_global
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_server_with_templates
      record_test("servers-server-with-templates")

      sdk = SDK.new

      res = sdk.servers.server_with_templates(
        server_url: Utils.template_url(
          ::OpenApiSDK::Servers::SERVER_WITH_TEMPLATES_SERVERS[0],
          {
            hostname: "localhost",
            port: HTTPBIN_PORT
          }
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_server_with_templates_defaults
      record_test("servers-server-with-templates-defaults")

      sdk = SDK.new

      res = sdk.servers.server_with_templates

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_servers_by_id_with_templates
      record_test("servers-server-by-id-with-templates")

      sdk = SDK.new

      res = sdk.servers.servers_by_id_with_templates

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end
  end
end
