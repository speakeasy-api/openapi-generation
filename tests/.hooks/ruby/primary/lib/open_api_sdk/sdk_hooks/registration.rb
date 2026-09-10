# typed: true
# frozen_string_literal: true

#
# This file is only ever generated once on the first generation and then is free to be modified.
# Any hooks you wish to add should be registered in the init_hooks method.
#
# Hooks are registered per SDK instance, and are valid for the lifetime of the SDK instance.
#

require_relative "./custom_security_hook"
require_relative "./test_hook"
require_relative "./types"

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
        hooks.register_sdk_init_hook(TestSDKInitHook.new)
        hooks.register_before_request_hook(TestBeforeRequestHook.new)
        hooks.register_after_success_hook(TestAfterSuccessHook.new)
        hooks.register_after_error_hook(TestAfterErroHook.new)

        custom_security_hook = CustomSecurityHook.new
        hooks.register_before_request_hook(custom_security_hook)
      end
    end
  end
end
