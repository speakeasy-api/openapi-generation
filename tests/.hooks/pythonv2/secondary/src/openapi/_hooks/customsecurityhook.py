from typing import Union, cast
import httpx

from openapi.models.operations.customschemeappid import CustomSchemeAppIDSecurity

from .types import (
    BeforeRequestContext,
    BeforeRequestHook,
)


class CustomSecurityHook(BeforeRequestHook):
    def before_request(
        self, hook_ctx: BeforeRequestContext, request: httpx.Request
    ) -> Union[httpx.Request, Exception]:
        if hook_ctx.operation_id == "customSchemeAppId":
            if hook_ctx.security_source is None:
                return Exception("security source is None")

            sec = hook_ctx.security_source
            if callable(sec):
                sec = sec()

            custom_sec = cast(CustomSchemeAppIDSecurity, sec)
            request.headers["X-Security-App-Id"] = custom_sec.app_id
            request.headers["X-Security-Secret"] = custom_sec.secret

        return request
