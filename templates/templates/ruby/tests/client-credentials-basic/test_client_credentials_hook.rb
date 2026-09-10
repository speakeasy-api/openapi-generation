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
  class TestClientCredentialsBasicHook < Minitest::Test
    def test_client_credentials_basic_hook_successfully_authenticates
      record_test("hooks-client-credentials-basic-success")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_credentials: Models::Shared::SchemeClientCredentials.new(
            client_id: "speakeasy-sdks",
            client_secret: "supersecret-#{SecureRandom.hex(5)}",
            audience: ""
          )
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_client_credentials_basic_hook_successfully_authenticates_global_server
      record_test("hooks-client-credentials-basic-success-global-server")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_credentials: Models::Shared::SchemeClientCredentials.new(
            client_id: "speakeasy-sdks",
            client_secret: "supersecret-#{SecureRandom.hex(5)}",
            audience: ""
          )
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request_global_server(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      res = s.hooks.authenticated_request_global_server(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end

    def test_client_credentials_basic_hook_successfully_authenticates_with_alt_token_url
      record_test("hooks-client-credentials-basic-success-alt-token-url")

      s = OpenApiSDK::SDK.new(
        security: Models::Shared::Security.new(
          client_credentials: Models::Shared::SchemeClientCredentials.new(
            client_id: "speakeasy-sdks",
            client_secret: "supersecret-#{SecureRandom.hex(5)}",
            token_url: "/clientcredentials/alt/token",
            scopes: ["alt:one", SDKHooks::ClientCredentialsOAuth2Scope::ALT_TWO.serialize],
            audience: ""
          )
        )
      )
      refute_nil(s)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)

      res = s.hooks.authenticated_request(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end
  end
end
