# typed: true
# frozen_string_literal: true

require_relative "./types"

module OpenApiSDK
  module SDKHooks
    class CustomSecurityHook < AbstractSDKHook
      extend T::Sig

      sig do
        override
          .params(
            hook_ctx: BeforeRequestHookContext,
            request: Faraday::Request
          )
          .returns(Faraday::Request)
      end
      def before_request(hook_ctx:, request:)
        request.headers["Idempotency-Key"] = "some-key"

        case hook_ctx.operation_id
        when "customSchemeAppId"

          raise ArgumentError, "Security source is null" if hook_ctx.security_source.nil?

          security = T.must(hook_ctx.security_source).call
          raise ArgumentError, "Security source is not of type Security" unless security.is_a?(Models::Shared::Security)

          raise "CustomSchemeAppID security is not defined" if security.custom_scheme_app_id.nil?

          request.headers["X-Security-App-Id"] = security.custom_scheme_app_id.app_id
          request.headers["X-Security-Secret"] = security.custom_scheme_app_id.secret
        end

        request
      end

    end
  end
end
