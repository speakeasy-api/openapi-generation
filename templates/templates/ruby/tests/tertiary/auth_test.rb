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

    def test_no_auth
      record_test("auth-no-auth")

      res = @sdk.auth.no_auth

      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.status_code)
    end

    def test_api_key_auth_global
      record_test("auth-api-key-auth-global")

      err = assert_raises(Models::Errors::APIError) do
        @sdk.auth.api_key_auth_global
      end

      assert_equal(Rack::Utils.status_code(:unauthorized), err.status_code)
    end

  end
end
