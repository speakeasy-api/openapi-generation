import httpx
import platform
from typing import Any, Optional, Tuple, Union

from openapi.httpclient import HttpClient
from openapi.sdkconfiguration import SDKConfiguration
from .types import (
    BeforeRequestContext,
    BeforeRequestHook,
    AfterSuccessContext,
    AfterSuccessHook,
    AfterErrorContext,
    AfterErrorHook,
    AfterParseErrorContext,
    AfterParseErrorHook,
    SDKInitHook,
)
from urllib.parse import urlparse, parse_qs, urlencode, urlunparse


class TestClient(HttpClient):
    client: HttpClient

    def __init__(self, client: HttpClient):
        self.client = client

    def send(
        self,
        request: httpx.Request,
        *,
        stream: bool = False,
        auth: Union[
            httpx._types.AuthTypes, httpx._client.UseClientDefault, None
        ] = httpx.USE_CLIENT_DEFAULT,
        follow_redirects: Union[
            bool, httpx._client.UseClientDefault
        ] = httpx.USE_CLIENT_DEFAULT,
    ) -> httpx.Response:
        request.headers["Client-Level-Header"] = "added by client"

        return self.client.send(
            request, stream=stream, auth=auth, follow_redirects=follow_redirects
        )

    def build_request(
        self,
        method: str,
        url: httpx._types.URLTypes,
        *,
        content: Optional[httpx._types.RequestContent] = None,
        data: Optional[httpx._types.RequestData] = None,
        files: Optional[httpx._types.RequestFiles] = None,
        json: Optional[Any] = None,
        params: Optional[httpx._types.QueryParamTypes] = None,
        headers: Optional[httpx._types.HeaderTypes] = None,
        cookies: Optional[httpx._types.CookieTypes] = None,
        timeout: Union[
            httpx._types.TimeoutTypes, httpx._client.UseClientDefault
        ] = httpx.USE_CLIENT_DEFAULT,
        extensions: Optional[httpx._types.RequestExtensions] = None,
    ) -> httpx.Request:
        return self.client.build_request(
            method,
            url,
            content=content,
            data=data,
            files=files,
            json=json,
            params=params,
            headers=headers,
            cookies=cookies,
            timeout=timeout,
            extensions=extensions,
        )

    def close(self) -> None:
        self.client.close()


class TestHook(SDKInitHook, BeforeRequestHook, AfterSuccessHook, AfterErrorHook, AfterParseErrorHook):
    init_sdk_version: str = ""

    def sdk_init(self, config: SDKConfiguration) -> SDKConfiguration:
        if config.client is None:
            raise Exception("Client is required")

        self.init_sdk_version = config.sdk_version
        config.client = TestClient(config.client)
        return config

    def before_request(
        self, hook_ctx: BeforeRequestContext, request: httpx.Request
    ) -> Union[httpx.Request, Exception]:
        request.headers["Idempotency-Key"] = "some-key"

        if hook_ctx.operation_id == "testHooks":
            # Parse the original URL
            parsed_url = urlparse(str(request.url))
            query_params = parse_qs(parsed_url.query)

            # Update the query parameter
            query_params["someParam"] = ["overriddenParam"]

            # Reconstruct the URL with the updated query parameter
            updated_query_string = urlencode(query_params, doseq=True)
            updated_url_parts = parsed_url._replace(query=updated_query_string)

            request.url = httpx.URL(urlunparse(updated_url_parts))
        if hook_ctx.operation_id == "authorizationHeaderModification":
            request.headers["Authorization"] = (
                request.headers["Authorization"] + " modified"
            )
        if hook_ctx.operation_id == "testHooksBeforeCreateRequestPaths":
            request.headers["old-pathname"] = request.url.path
        if hook_ctx.operation_id == "hooksCustomUserAgent":
            request.headers["User-Agent"] = (
                f"acme-corp/{hook_ctx.config.sdk_version} acme-corp/{platform.python_version()}"
            )
            request.headers["X-Test-Gen-Version"] = hook_ctx.config.gen_version
            request.headers["X-Test-Doc-Version"] = hook_ctx.config.openapi_doc_version
            request.headers["X-Test-Init-Sdk-Version"] = self.init_sdk_version

        return request

    def after_success(
        self, hook_ctx: AfterSuccessContext, response: httpx.Response
    ) -> Union[httpx.Response, Exception]:
        if hook_ctx.operation_id == "testHooksAfterResponse":
            return Exception("validation failed")

        if hook_ctx.operation_id == "hooksOperationMetadata":
            response.headers["x-test-tags"] = ",".join(hook_ctx.tags or [])
            extensions = hook_ctx.extensions or {}
            response.headers["x-test-extension"] = str(
                extensions.get("x-test-operation-metadata", "none")
            )
            response.headers["x-test-template-delimiters"] = str(
                extensions.get("x-test-template-delimiters", "none")
            )
            speakeasy_keys = [k for k in extensions if k.startswith("x-speakeasy-")]
            response.headers["x-test-speakeasy-keys"] = ",".join(speakeasy_keys) or "none"

        return response

    def after_parse_error(
        self,
        hook_ctx: AfterParseErrorContext,
        response: httpx.Response,
        error: Exception,
    ) -> Exception:
        """Echo `hook_ctx.response.mode` into the substitute exception so the
        test can lock that the parse-error hook fires for both `parsed` and
        `raw` modes."""
        if response.request.headers.get(
            "x-test-after-parse-error"
        ) and hook_ctx.operation_id in (
            "statusGetDefaultError",
            "getMalformedErrorResponse",
        ):
            return RuntimeError(
                f"parse_error fired mode={hook_ctx.response.mode}"
            )
        return error

    def after_error(
        self,
        hook_ctx: AfterErrorContext,
        response: Optional[httpx.Response],
        error: Optional[Exception],
    ) -> Union[Tuple[Optional[httpx.Response], Optional[Exception]], Exception]:
        if (
            response is not None
            and response.request.headers.get("x-test-after-error")
            and hook_ctx.operation_id == "getMalformedErrorResponse"
        ):
            return (None, RuntimeError(f"error fired mode={hook_ctx.response.mode}"))

        if hook_ctx.operation_id == "testHooksError":
            if response is not None and response.status_code != 400:
                return (None, Exception("expected status code 400"))
            return (None, Exception("special test error case"))

        if hook_ctx.operation_id == "statusGetDefaultError":
            if response is not None and response.status_code == 418:
                # Echo the propagated response mode back as a header so the
                # test can assert it (locks gen.yaml `rawResponseHelpers: true`
                # threading response metadata through `HookContext`).
                return (
                    httpx.Response(
                        status_code=200,
                        headers={
                            "x-test-response-mode": hook_ctx.response.mode,
                        },
                    ),
                    None,
                )

        return (response, error)
