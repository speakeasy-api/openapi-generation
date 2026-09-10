from .asynctypes import AsyncHooks
from .asynctesthook import AsyncTestHook
from .testhook import TestHook
from .customsecurityhook import CustomSecurityHook
from .adapters import (
    SyncToAsyncBeforeRequestAdapter,
    SyncToAsyncAfterSuccessAdapter,
    SyncToAsyncAfterErrorAdapter,
)


def init_async_hooks(hooks: AsyncHooks):
    """Initialize async test hooks for integration testing."""
    async_test_hook = AsyncTestHook()

    hooks.register_before_request_hook(async_test_hook)
    hooks.register_after_success_hook(async_test_hook)
    hooks.register_after_error_hook(async_test_hook)
    hooks.register_after_parse_error_hook(async_test_hook)

    # Register sync hooks adapted to async for compatibility
    # This allows sync hooks to work in async contexts via thread pool
    test_hook = TestHook()
    custom_security_hook = CustomSecurityHook()

    hooks.register_before_request_hook(SyncToAsyncBeforeRequestAdapter(test_hook))
    hooks.register_before_request_hook(SyncToAsyncBeforeRequestAdapter(custom_security_hook))
    hooks.register_after_success_hook(SyncToAsyncAfterSuccessAdapter(test_hook))
    hooks.register_after_error_hook(SyncToAsyncAfterErrorAdapter(test_hook))
