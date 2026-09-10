# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module AlphabeticallyEarly
  class TestAuth < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_no_auth
      record_test("auth-hoisted-no-auth-retained")

      res = @sdk.auth.no_auth

      assert_nil(res)
    end

    def test_basic_auth
      record_test("auth-hoisted-basic-auth")

      @sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          username: "testUser",
          password: "testPass"
        )
      )

      res = @sdk.auth.basic_auth(
        passwd: "testPass",
        user: "testUser"
      )

      refute_nil(res)
      assert_equal(res.authenticated, true)
    end

    def test_multiple_mixed_options_auth
      record_test("auth-hoisted-operation-auth-retained")

      res = @sdk.auth_new.multiple_mixed_options_auth(
        security: Models::Operations::MultipleMixedOptionsAuthSecurity.new(
          basic_auth: Models::Shared::SchemeBasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        )
      )
      assert_nil(res)
    end

    def test_custom_security_scheme_app_id
      record_test("auth-custom-security-scheme-app-id")

      res = @sdk.auth_new.custom_scheme_app_id(
        security: Models::Operations::CustomSchemeAppIdSecurity.new(
          app_id: "testAppID",
          secret: "testSecret"
        )
      )
      assert_nil(res)
    end
  end
end
