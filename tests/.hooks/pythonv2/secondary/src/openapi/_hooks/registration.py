from .customsecurityhook import CustomSecurityHook
from .types import Hooks


def init_hooks(hooks: Hooks):
    custom_security_hook = CustomSecurityHook()
    hooks.register_before_request_hook(custom_security_hook)
