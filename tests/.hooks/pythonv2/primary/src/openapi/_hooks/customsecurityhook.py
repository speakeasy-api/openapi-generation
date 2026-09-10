from typing import Union, cast
import httpx

from openapi.models.shared.security import Security

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

            security = cast(Security, sec)
            custom_sec = security.custom_scheme_app_id
            if custom_sec is None:
                return Exception("custom_scheme_app_id is not defined")

            request.headers["X-Security-App-Id"] = custom_sec.app_id
            request.headers["X-Security-Secret"] = custom_sec.secret

        return request
