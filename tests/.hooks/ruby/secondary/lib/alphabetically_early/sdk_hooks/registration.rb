# typed: true
# frozen_string_literal: true

#
# This file is only ever generated once on the first generation and then is free to be modified.
# Any hooks you wish to add should be registered in the init_hooks method.
#
# Hooks are registered per SDK instance, and are valid for the lifetime of the SDK instance.
#

require_relative "./custom_security_hook"
require_relative "./types"

module AlphabeticallyEarly
  module SDKHooks
    class Registration

      def self.init_hooks(hooks)
        custom_security_hook = CustomSecurityHook.new
        hooks.register_before_request_hook(custom_security_hook)
      end
    end
  end
end
