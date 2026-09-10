# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestGlobalSecurityFieldsOrdering < Minitest::Test
    def test_global_security_fields_ordering
      record_test("auth-global-security-fields-ordering")

      # When maintainOpenApiOrder: false is set, the first non-nil field (in alphabetical order) is selected.
      # In this case api_key_auth takes precedence over basic_http, which is invalid for this endpoint.
      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "testApiKey",
          basic_http: Models::Shared::SchemeBasicHTTP.new(
            username: "testUser",
            password: "testPass"
          )
        )
      )

      err = assert_raises(Models::Errors::APIError) do
        sdk.auth.global_security_basic_http
      end
      assert_equal(401, err.status_code)
    end
  end

  class TestHoistedSecurityAccessTokenOnly < Minitest::Test
    def test_hoisted_security_access_token_only
      record_test("auth-hoisted-security-access-token-only")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "testApiKey",
          basic_http: Models::Shared::SchemeBasicHTTP.new(
            username: "testUser",
            password: "testPass"
          ),
          access_token: "Bearer ghp_xxxx"
        )
      )

      res = sdk.auth.hoisted_security_access_token_only

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.token)
      assert_equal("Bearer ghp_xxxx", res.token.token)
    end
  end

  class TestHoistedSecurityAccessTokenFirst < Minitest::Test
    def test_hoisted_security_access_token_first
      record_test("auth-hoisted-security-access-token-first")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "testApiKey",
          access_token: "Bearer ghp_xxxx"
        )
      )

      res = sdk.auth.hoisted_security_access_token_first

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.token)
      assert_equal("Bearer ghp_xxxx", res.token.token)
    end
  end

  class TestHoistedSecurityApiKeyFirst < Minitest::Test
    def test_hoisted_security_api_key_first
      record_test("auth-hoisted-security-api-key-first")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "testApiKey",
          access_token: "Bearer ghp_xxxx"
        )
      )

      res = sdk.auth.hoisted_security_api_key_first

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.token)
      assert_equal("testApiKey", res.token.token)
    end
  end

  class TestHoistedSecurityBasicHttpOnly < Minitest::Test
    def test_hoisted_security_basic_http_only
      record_test("auth-hoisted-security-basic-http-only")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "testApiKey",
          basic_http: Models::Shared::SchemeBasicHTTP.new(
            username: "testUser",
            password: "testPass"
          ),
          access_token: "Bearer ghp_xxxx"
        )
      )

      res = sdk.auth.hoisted_security_basic_http_only

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.basic_auth)
      assert(res.basic_auth.authenticated)
      assert_equal("testUser", res.basic_auth.user)
    end
  end

  class TestHoistedSecurityInvalidField < Minitest::Test
    def test_hoisted_security_invalid_field
      record_test("auth-hoisted-security-invalid-field")

      # Provide only basic_http — not valid for access_token_first which expects bearer/apiKey
      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          basic_http: Models::Shared::SchemeBasicHTTP.new(
            username: "user",
            password: "pass"
          )
        )
      )

      err = assert_raises(Models::Errors::APIError) do
        sdk.auth.hoisted_security_access_token_first
      end
      assert_equal(401, err.status_code)
    end
  end
end
