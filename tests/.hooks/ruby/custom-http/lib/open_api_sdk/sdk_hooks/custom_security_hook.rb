# typed: true
# frozen_string_literal: true

require "json"

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
        case hook_ctx.operation_id
        when "customHttpOnly"

          raise ArgumentError, "Security source is null" if hook_ctx.security_source.nil?

          security = T.must(hook_ctx.security_source).call
          unless security.is_a?(Models::Components::Security)
            raise ArgumentError, "Security source is not of type Security"
          end

          raise "CustomHttp security is not defined" if security.custom_http.nil?

          custom_http = security.custom_http

          request.headers["X-Security-UserID"] = custom_http.user_id.to_s
          request.headers["X-Security-Role"] = custom_http.role.serialize
          request.headers["X-Security-Passphrase"] = custom_http.passphrase

          unless custom_http.access_code.nil?
            request.headers["X-Security-AccessCode"] = custom_http.access_code.to_s
          end

          unless custom_http.scopes.nil? || custom_http.scopes.empty?
            request.headers["X-Security-Scopes"] = custom_http.scopes.to_json
          end
        end

        request
      end

    end
  end
end
