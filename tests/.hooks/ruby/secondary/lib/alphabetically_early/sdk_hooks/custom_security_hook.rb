# typed: true
# frozen_string_literal: true

require_relative "./types"

module AlphabeticallyEarly
  module SDKHooks
    class CustomSecurityHook < AbstractSDKHook

      def before_request(hook_ctx:, request:)
        request.headers["Idempotency-Key"] = "some-key"

        case hook_ctx.operation_id
        when "customSchemeAppId"

          raise ArgumentError, "Security source is null" if hook_ctx.security_source.nil?

          custom_security = hook_ctx.security_source.call
          unless custom_security.is_a?(Models::Operations::CustomSchemeAppIdSecurity)
            raise ArgumentError, "Security source is not of type CustomSchemeAppId"
          end

          request.headers["X-Security-App-Id"] = custom_security.app_id
          request.headers["X-Security-Secret"] = custom_security.secret
        end

        request
      end

    end
  end
end
