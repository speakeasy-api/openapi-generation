from typing import Union, cast
import httpx
import json

from customhttp.models.security import Security

from customhttp._hooks.types import (
    BeforeRequestContext,
    BeforeRequestHook,
)


class CustomSecurityHook(BeforeRequestHook):
    def before_request(
        self, hook_ctx: BeforeRequestContext, request: httpx.Request
    ) -> Union[httpx.Request, Exception]:
        if hook_ctx.operation_id == "customHttpOnly":
            if hook_ctx.security_source is None:
                return Exception("security source is None")

            sec = hook_ctx.security_source
            if callable(sec):
                sec = sec()

            security = cast(Security, sec)
            custom_http = security.custom_http
            if custom_http is None:
                return Exception("custom_http is not defined")

            request.headers["X-Security-UserID"] = str(custom_http.user_id)
            request.headers["X-Security-Role"] = custom_http.role
            request.headers["X-Security-Passphrase"] = custom_http.passphrase

            if custom_http.access_code is not None:
                request.headers["X-Security-AccessCode"] = str(custom_http.access_code)

            if custom_http.scopes is not None and len(custom_http.scopes) > 0:
                request.headers["X-Security-Scopes"] = json.dumps(custom_http.scopes)

        return request
