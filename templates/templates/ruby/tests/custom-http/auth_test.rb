# frozen_string_literal: true
# typed: true

require_relative "../lib/customhttp"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"

module CustomHttp
  class TestAuth < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.api_test_service_url)
    end

    def test_custom_http_scheme_only
      record_test("auth-custom-security-scheme-only")

      test_scopes = ["read:products", "write:products"]

      @sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.api_test_service_url,
        custom_http: OpenApiSDK::Models::Components::SchemeCustomHTTPSecurity.new(
          user_id: 54_321,
          role: OpenApiSDK::Models::Components::Role::MANAGER,
          passphrase: "secure-passphrase-123",
          access_code: 104,
          scopes: test_scopes
        )
      )

      res = @sdk.auth.custom_http_only

      refute_nil(res)
      refute_nil(res.object)
      assert_equal("access_granted", res.object.grant)
      assert_equal(test_scopes, res.object.scopes)
    end
  end
end
