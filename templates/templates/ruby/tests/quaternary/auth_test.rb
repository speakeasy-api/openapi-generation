# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestAuth < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_global_security_flattening
      record_test("auth-global-security-flattening")
      @sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        api_key_auth: "Bearer testToken"
      )

      res = @sdk.auth.api_key_auth_global

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
      assert_equal(res.token.authenticated, true)
      assert_equal("testToken", res.token.token)
    end

    def test_global_security_flattening_callback
      record_test("auth-global-security-flattening-callback")
      @sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security_source: -> { Models::Shared::Security.new(api_key_auth: "Bearer testToken") }
      )

      res = @sdk.auth.api_key_auth_global
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
      assert_equal(res.token.authenticated, true)
      assert_equal("testToken", res.token.token)
    end
  end
end
