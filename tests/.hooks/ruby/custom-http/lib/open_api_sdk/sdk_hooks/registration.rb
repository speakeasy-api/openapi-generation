# typed: true
# frozen_string_literal: true

require_relative "./custom_security_hook"

module OpenApiSDK
  module SDKHooks
    class Registration
      extend T::Sig

      sig do
        params(
          hooks: Hooks
        )
          .void
      end
      def self.init_hooks(hooks)
        custom_security_hook = CustomSecurityHook.new
        hooks.register_before_request_hook(custom_security_hook)
      end
    end
  end
end
