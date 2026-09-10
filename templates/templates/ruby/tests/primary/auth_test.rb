# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestAuth < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_no_auth
      record_test("auth-no-auth")

      res = @sdk.auth.no_auth

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_basic_auth
      record_test("auth-basic-auth")

      res = @sdk.auth_new.basic_auth_new(
        security: Models::Operations::BasicAuthNewSecurity.new(
          username: "testUser",
          password: "testPass"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_basic_auth_empty
      record_test("auth-basic-auth-empty")

      res = @sdk.auth_new.basic_auth_new(
        security: Models::Operations::BasicAuthNewSecurity.new(
          username: "",
          password: ""
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "",
            password: ""
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_basic_auth_username_only
      record_test("auth-basic-auth-username-only")

      res = @sdk.auth_new.basic_auth_new(
        security: Models::Operations::BasicAuthNewSecurity.new(
          username: "testUser",
          password: ""
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "testUser",
            password: ""
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_basic_auth_password_only
      record_test("auth-basic-auth-password-only")

      res = @sdk.auth_new.basic_auth_new(
        security: Models::Operations::BasicAuthNewSecurity.new(
          username: "",
          password: "testPass"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "",
            password: "testPass"
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_basic_auth_long_password
      # Not recording this test, because it's ruby only
      res = @sdk.auth_new.basic_auth_new(
        security: Models::Operations::BasicAuthNewSecurity.new(
          username: "sdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasda",
          password: "asdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasd"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          basic_auth: Models::Shared::BasicAuth.new(
            username: "sdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasda",
            password: "asdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasdasd"
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_api_key_auth_global
      record_test("auth-api-key-auth-global")
      @sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          api_key_auth: "Bearer test_api_key"
        )
      )
      res = @sdk.auth.api_key_auth_global

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("test_api_key", res.token.token)
    end

    def test_bearer_auth_operation_with_prefix
      record_test("auth-bearer-auth-operation-with-prefix")

      res = @sdk.auth.bearer_auth(
        security: Models::Operations::BearerAuthSecurity.new(
          bearer_auth: "Bearer testToken"
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.token.authenticated, true)
      assert_equal("testToken", res.token.token)
    end

    def test_bearer_auth_operation_without_prefix
      record_test("auth-bearer-auth-operation-without-prefix")

      res = @sdk.auth.bearer_auth(
        security: Models::Operations::BearerAuthSecurity.new(
          bearer_auth: "testToken"
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(res.token.authenticated, true)
      assert_equal("testToken", res.token.token)
    end

    def test_oauth_2_auth
      record_test("auth-oauth2-auth")

      @sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          oauth2: "Bearer testToken"
        )
      )

      res = @sdk.auth_new.oauth2_auth_new(
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_open_id_connect_auth
      record_test("auth-open-id-connect-auth")

      res = @sdk.auth_new.open_id_connect_auth_new(
        security: Models::Operations::OpenIdConnectAuthNewSecurity.new(
          open_id_connect: "Bearer testToken"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_simple_scheme_auth
      record_test("auth-multiple-simple-scheme-auth")

      res = @sdk.auth_new.multiple_simple_scheme_auth(
        security: Models::Operations::MultipleSimpleSchemeAuthSecurity.new(
          api_key_auth_new: "test_api_key",
          oauth2: "Bearer testToken"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            ),
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_scheme_auth
      record_test("auth-multiple-mixed-scheme-auth")

      res = @sdk.auth_new.multiple_mixed_scheme_auth(
        security: Models::Operations::MultipleMixedSchemeAuthSecurity.new(
          api_key_auth_new: "test_api_key",
          basic_auth: Models::Shared::SchemeBasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            )
          ],
          basic_auth: Models::Shared::BasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        )
      )

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_simple_options_auth_first_option
      record_test("auth-multiple-simple-options-auth-first-option")

      res = @sdk.auth_new.multiple_simple_options_auth(
        security: Models::Operations::MultipleSimpleOptionsAuthSecurity.new(
          api_key_auth_new: "test_api_key"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_simple_options_auth_second_option
      record_test("auth-multiple-simple-options-auth-second-option")

      res = @sdk.auth_new.multiple_simple_options_auth(
        security: Models::Operations::MultipleSimpleOptionsAuthSecurity.new(
          oauth2: "Bearer testToken"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_auth_first_option
      record_test("auth-multiple-mixed-options-auth-first-option")

      res = @sdk.auth_new.multiple_mixed_options_auth(
        security: Models::Operations::MultipleMixedOptionsAuthSecurity.new(
          api_key_auth_new: "test_api_key"
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_auth_second_option
      record_test("auth-multiple-mixed-options-auth-second-option")

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
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_with_simple_schemes_auth_first_option
      record_test("auth-multiple-options-with-simple-schemes-auth-first-option")

      res = @sdk.auth_new.multiple_options_with_simple_schemes_auth(
        security: Models::Operations::MultipleOptionsWithSimpleSchemesAuthSecurity.new(
          option1: Models::Operations::MultipleOptionsWithSimpleSchemesAuthSecurityOption1.new(
            api_key_auth_new: "test_api_key",
            oauth2: "Bearer testToken"
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            ),
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_with_simple_schemes_auth_second_option
      record_test("auth-multiple-options-with-simple-schemes-auth-second-option")

      res = @sdk.auth_new.multiple_options_with_simple_schemes_auth(
        security: Models::Operations::MultipleOptionsWithSimpleSchemesAuthSecurity.new(
          option2: Models::Operations::MultipleOptionsWithSimpleSchemesAuthSecurityOption2.new(
            api_key_auth_new: "test_api_key",
            open_id_connect: "Bearer testToken"
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            ),
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_with_mixed_schemes_auth_first_option
      record_test("auth-multiple-options-with-mixed-schemes-auth-first-option")

      res = @sdk.auth_new.multiple_options_with_mixed_schemes_auth(
        security: Models::Operations::MultipleOptionsWithMixedSchemesAuthSecurity.new(
          option1: Models::Operations::MultipleOptionsWithMixedSchemesAuthSecurityOption1.new(
            api_key_auth_new: "test_api_key",
            oauth2: "Bearer testToken"
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            ),
            Models::Shared::HeaderAuth.new(
              header_name: "Authorization",
              expected_value: "Bearer testToken"
            )
          ]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_multiple_mixed_options_with_mixed_schemes_auth_second_option
      record_test("auth-multiple-options-with-mixed-schemes-auth-second-option")

      res = @sdk.auth_new.multiple_options_with_mixed_schemes_auth(
        security: Models::Operations::MultipleOptionsWithMixedSchemesAuthSecurity.new(
          option2: Models::Operations::MultipleOptionsWithMixedSchemesAuthSecurityOption2.new(
            api_key_auth_new: "test_api_key",
            basic_auth: Models::Shared::SchemeBasicAuth.new(
              username: "testUser",
              password: "testPass"
            )
          )
        ),
        request: Models::Shared::AuthServiceRequestBody.new(
          header_auth: [
            Models::Shared::HeaderAuth.new(
              header_name: "x-api-key",
              expected_value: "test_api_key"
            )
          ],
          basic_auth: Models::Shared::BasicAuth.new(
            username: "testUser",
            password: "testPass"
          )
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_function_callbacks_oauth_global_security
      record_test("auth-function-callbacks-oauth-global-security")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security_source: -> { Models::Shared::Security.new(oauth2: "Bearer global") }
      )

      res = sdk.auth.global_bearer_auth

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("global", res.token.token)
    end

    def test_custom_security_option_app_id
      record_test("auth-custom-security-option-app-id")

      sdk = OpenApiSDK::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        security: Models::Shared::Security.new(
          custom_scheme_app_id: Models::Shared::SchemeCustomSchemeAppID.new(
            app_id: "testAppID",
            secret: "testSecret"
          )
        )
      )

      res = sdk.auth_new.custom_scheme_app_id

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end
  end

end
