# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"

class TestGlobalSecurityBasicHttpSuccess < Minitest::Test
  def test_global_security_basic_http_success
    record_test("auth-basic-http-global-option")

    # Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        basic_http: OpenApiSDK::Models::Shared::BasicHttp.new(
          username: "testUser",
          password: "testPass"
        ),
        access_token: OpenApiSDK::Models::Shared::AccessToken.new(
          access_token: "Bearer ignored"
        )
      )
    )

    res = sdk.auth.global_security_option_basic_http

    refute_nil(res)
    assert_equal(200, res.status_code)
    refute_nil(res.basic_auth)
    assert(res.basic_auth.authenticated)
    assert_equal("testUser", res.basic_auth.user)
  end
end

class TestGlobalSecurityFieldsOrdering < Minitest::Test
  def test_global_security_fields_ordering
    record_test("auth-global-security-option-fields-ordering")

    # Expected to fail since APIKeyAuth takes priority over BasicHTTP in global security definition
    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        api_key_auth: OpenApiSDK::Models::Shared::ApiKeyAuth.new(
          api_key_auth: "Bearer test_api_key"
        ),
        basic_http: OpenApiSDK::Models::Shared::BasicHttp.new(
          username: "testUser",
          password: "testPass"
        )
      )
    )

    err = assert_raises(OpenApiSDK::Models::Errors::APIError) do
      sdk.auth.global_security_option_basic_http
    end
    assert_equal(401, err.status_code)
  end
end

class TestHoistedSecurityAccessTokenFirst < Minitest::Test
  def test_hoisted_security_access_token_first
    record_test("auth-hoisted-security-option-access-token-first")

    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        api_key_auth: OpenApiSDK::Models::Shared::ApiKeyAuth.new(
          api_key_auth: "testApiKey"
        ),
        access_token: OpenApiSDK::Models::Shared::AccessToken.new(
          access_token: "Bearer ghp_xxxx"
        )
      )
    )

    res = sdk.auth.hoisted_security_option_access_token_first

    refute_nil(res)
    assert_equal(200, res.status_code)
    refute_nil(res.token_auth_response)
    assert_equal("Bearer ghp_xxxx", res.token_auth_response.token)
  end
end

class TestHoistedSecurityApiKeyFirst < Minitest::Test
  def test_hoisted_security_api_key_first
    record_test("auth-hoisted-security-option-api-key-first")

    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        api_key_auth: OpenApiSDK::Models::Shared::ApiKeyAuth.new(
          api_key_auth: "testApiKey"
        ),
        access_token: OpenApiSDK::Models::Shared::AccessToken.new(
          access_token: "Bearer ghp_xxxx"
        )
      )
    )

    res = sdk.auth.hoisted_security_option_api_key_first

    refute_nil(res)
    assert_equal(200, res.status_code)
    refute_nil(res.token_auth_response)
    assert_equal("testApiKey", res.token_auth_response.token)
  end
end

class TestHoistedSecurityBasicHttpOnly < Minitest::Test
  def test_hoisted_security_basic_http_only
    record_test("auth-hoisted-security-option-basic-http-only")

    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        api_key_auth: OpenApiSDK::Models::Shared::ApiKeyAuth.new(
          api_key_auth: "testApiKey"
        ),
        basic_http: OpenApiSDK::Models::Shared::BasicHttp.new(
          username: "testUser",
          password: "testPass"
        ),
        access_token: OpenApiSDK::Models::Shared::AccessToken.new(
          access_token: "Bearer ghp_xxxx"
        )
      )
    )

    res = sdk.auth.hoisted_security_option_basic_http_only

    refute_nil(res)
    assert_equal(200, res.status_code)
    refute_nil(res.basic_auth)
    assert(res.basic_auth.authenticated)
    assert_equal("testUser", res.basic_auth.user)
  end
end

class TestHoistedSecurityInvalidOption < Minitest::Test
  def test_hoisted_security_invalid_option
    record_test("auth-hoisted-security-invalid-option")

    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Shared::Security.new(
        basic_http: OpenApiSDK::Models::Shared::BasicHttp.new(
          username: "username",
          password: "password"
        )
      )
    )

    err = assert_raises(OpenApiSDK::Models::Errors::APIError) do
      sdk.auth.hoisted_security_option_access_token_first
    end
    assert_equal(401, err.status_code)
  end
end
