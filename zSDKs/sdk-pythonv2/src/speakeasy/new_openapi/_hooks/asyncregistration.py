from .asynctypes import AsyncHooks


# This file is only ever generated once on the first generation and then is free to be modified.
# Any async hooks you wish to add should be registered in the init_async_hooks function.
# Feel free to define them in this file or in separate files in the hooks folder.


def init_async_hooks(hooks: AsyncHooks):
    # pylint: disable=unused-argument
    """Add async hooks by calling hooks.register_{sdk_init/before_request/after_success/after_error}_hook
    with an instance of a hook that implements that specific AsyncHook interface.
    Async hooks are registered per SDK instance, and are valid for the lifetime of the SDK instance.

    Async hooks are invoked in async contexts and allow you to use async/await for non-blocking I/O operations.

    Example:
        from .asynctypes import AsyncBeforeRequestHook, BeforeRequestContext
        import httpx

        class MyAsyncHook(AsyncBeforeRequestHook):
            async def before_request(self, hook_ctx: BeforeRequestContext, request: httpx.Request):
                # Perform async operations
                token = await fetch_token_from_external_service()

                # Modify request
                headers = dict(request.headers)
                headers["Authorization"] = f"Bearer {token}"
                return httpx.Request(request.method, request.url, headers=headers, content=request.content)

        hooks.register_before_request_hook(MyAsyncHook())
    """
