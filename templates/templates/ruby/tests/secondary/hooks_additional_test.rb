# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module AlphabeticallyEarly
  class TestHooksAdditional < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_hooks_available_oauth2_scopes
      # Duplicate of primary's 'hooks-available-oauth2-scopes'.
      # Included to ensure the `serialize` and `deserialize` methods work
      # as expected when using ::Crystalline::Enum (typingStrategy: none)
      # see `lib/crystalline/types.rb` for more details

      assert_equal("read", SDKHooks::ClientCredentialsOAuth2Scope::READ.serialize)
      assert_equal("models:read", SDKHooks::OAuth2Scope::MODELS_READ.serialize)

      assert_equal(
        SDKHooks::ClientCredentialsOAuth2Scope::WRITE,
        SDKHooks::ClientCredentialsOAuth2Scope.deserialize("write")
      )
      assert_equal(SDKHooks::OAuth2Scope::STORE_MANAGE, SDKHooks::OAuth2Scope.deserialize("store_manage"))

      error = assert_raises(RuntimeError) do
        SDKHooks::OAuth2Scope.deserialize("unknown")
      end

      assert_match(/unknown/, error.message)
    end
  end
end
