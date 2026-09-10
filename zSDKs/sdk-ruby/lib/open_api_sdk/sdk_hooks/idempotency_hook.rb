# typed: true
# frozen_string_literal: true

require_relative './types'
require 'securerandom'
require 'sorbet-runtime'

module OpenApiSDK
  module SDKHooks
    class IdempotencyHook < AbstractSDKHook
      extend T::Sig

      sig do
        override.params(
          hook_ctx: BeforeRequestHookContext,
          request: Faraday::Request
        ).returns(Faraday::Request)
      end
      def before_request(hook_ctx:, request:)
        request.headers['Idempotency-Key'] = SecureRandom.uuid
        request
      end
    end
  end
end
