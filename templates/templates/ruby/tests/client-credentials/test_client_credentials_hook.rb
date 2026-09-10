# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "securerandom"

module OpenApiSDK
  class TestClientCredentialsHook < Minitest::Test
    def test_client_credentials_hook_successfully_authenticates
      record_test("hooks-client-credentials-success")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: "speakeasy-sdks",
          client_secret: "supersecret-#{SecureRandom.hex(5)}"
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_client_credentials_hook_successfully_authenticates_global_server
      record_test("hooks-client-credentials-success-global-server")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: "speakeasy-sdks",
          client_secret: "supersecret-#{SecureRandom.hex(5)}"
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request_global_server(request: nil)
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_client_credentials_hook_successfully_authenticates_with_alt_token_url
      record_test("hooks-client-credentials-success-alt-token-url")

      token_url = "/clientcredentials/alt/token"
      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: "speakeasy-sdks",
          client_secret: "supersecret-#{SecureRandom.hex(5)}",
          token_url: token_url,
          scopes: ["alt:one", SDKHooks::ClientCredentialsOAuth2Scope::ALT_TWO.serialize]
        )
      )
      refute_nil(s)

      # 1. Initial token request
      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      # 2. Since the token is already expired, a new one should be requested
      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_client_credentials_hook_no_scopes
      record_test("hooks-client-credentials-no-scopes")

      client_id = "speakeasy-sdks"
      client_secret = "supersecret-#{SecureRandom.hex(5)}"

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: client_id,
          client_secret: client_secret
        )
      )
      refute_nil(s)

      # Expected to fail since the token endpoint requires 'read' and 'write' scopes
      # but the authenticated_request_no_scopes operation does not specify any.
      error = assert_raises(RuntimeError) do
        s.hooks.authenticated_request_no_scopes(request: nil)
      end
      assert_equal("Unexpected status code 400 from token endpoint", error.message)

      # Same check but this time we override the default scopes with an empty list
      s2 = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: client_id,
          client_secret: client_secret,
          scopes: []
        )
      )

      error2 = assert_raises(RuntimeError) do
        s2.hooks.authenticated_request(request: nil)
      end
      assert_equal("Unexpected status code 400 from token endpoint", error2.message)

      # Now use a different token_url that will allow no scopes to be requested
      no_scope_token_url = "/clientcredentials/token?expires_in=90&skip_scopes=true"
      s3 = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: client_id,
          client_secret: client_secret,
          token_url: no_scope_token_url
        )
      )

      res3 = s3.hooks.authenticated_request_no_scopes(request: nil)
      refute_nil(res3)
      refute_nil(res3.http_meta)
      refute_nil(res3.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res3.http_meta.response.status)

      # Since the token is not expired, it should be reused on subsequent call
      res4 = s3.hooks.authenticated_request_no_scopes(request: nil)
      refute_nil(res4)
      refute_nil(res4.http_meta)
      refute_nil(res4.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res4.http_meta.response.status)
    end

    def test_client_credentials_hook_lowercase_bearer_token
      record_test("hooks-client-credentials-lowercase-bearer")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_id: "speakeasy-sdks",
          client_secret: "supersecret-#{SecureRandom.hex(5)}",
          token_url: "/clientcredentials/token?token_type=bearer"
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end
  end
end
