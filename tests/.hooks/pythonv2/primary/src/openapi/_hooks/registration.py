from .customsecurityhook import CustomSecurityHook
from .types import Hooks
from .testhook import TestHook
from .asynctesthook import AsyncTestSDKInitHook


def init_hooks(hooks: Hooks):
    test_hook = TestHook()
    custom_security_hook = CustomSecurityHook()
    async_init_hook = AsyncTestSDKInitHook()

    hooks.register_sdk_init_hook(test_hook)
    hooks.register_sdk_init_hook(async_init_hook)

    hooks.register_before_request_hook(test_hook)
    hooks.register_before_request_hook(custom_security_hook)

    hooks.register_after_success_hook(test_hook)

    hooks.register_after_error_hook(test_hook)

    hooks.register_after_parse_error_hook(test_hook)
