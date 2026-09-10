import asyncio
import httpx
from typing import Any, Optional, Tuple, Union

from openapi.httpclient import AsyncHttpClient
from openapi.sdkconfiguration import SDKConfiguration
from .asynctypes import (
    AsyncBeforeRequestHook,
    AsyncAfterSuccessHook,
    AsyncAfterErrorHook,
    AsyncAfterParseErrorHook,
)
from .types import (
    SDKInitHook,
    BeforeRequestContext,
    AfterSuccessContext,
    AfterErrorContext,
    AfterParseErrorContext,
)
from urllib.parse import urlparse, parse_qs, urlencode, urlunparse


class AsyncTestClient(AsyncHttpClient):
    client: AsyncHttpClient

    def __init__(self, client: AsyncHttpClient):
        self.client = client

    async def send(
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
        # Simulate async I/O operation
        await asyncio.sleep(0.001)

        request.headers["Async-Client-Level-Header"] = "added by async client"

        return await self.client.send(
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

    async def aclose(self) -> None:
        await self.client.aclose()


class AsyncTestSDKInitHook(SDKInitHook):
    """Sync SDK init hook that wraps the async client for testing.

    Client wrapping doesn't require async operations, so we use a sync hook.
    """
    def sdk_init(self, config: SDKConfiguration) -> SDKConfiguration:
        if config.async_client is None:
            raise Exception("Async client is required")

        config.async_client = AsyncTestClient(config.async_client)
        return config


class AsyncTestHook(
    AsyncBeforeRequestHook,
    AsyncAfterSuccessHook,
    AsyncAfterErrorHook,
    AsyncAfterParseErrorHook,
):
    async def before_request(
        self, hook_ctx: BeforeRequestContext, request: httpx.Request
    ) -> Union[httpx.Request, Exception]:
        # Simulate async I/O (e.g., fetching from cache, external API call)
        await asyncio.sleep(0.001)

        request.headers["Async-Idempotency-Key"] = "async-key"

        if hook_ctx.operation_id == "testAsyncHooks":
            # Parse the original URL
            parsed_url = urlparse(str(request.url))
            query_params = parse_qs(parsed_url.query)

            # Update the query parameter
            query_params["asyncParam"] = ["asyncOverriddenParam"]

            # Reconstruct the URL with the updated query parameter
            updated_query_string = urlencode(query_params, doseq=True)
            updated_url_parts = parsed_url._replace(query=updated_query_string)

            request.url = httpx.URL(urlunparse(updated_url_parts))

        if hook_ctx.operation_id == "testAsyncHooksBeforeCreateRequestPaths":
            request.headers["async-old-pathname"] = request.url.path

        return request

    async def after_success(
        self, hook_ctx: AfterSuccessContext, response: httpx.Response
    ) -> Union[httpx.Response, Exception]:
        # Simulate async I/O (e.g., logging to external service)
        await asyncio.sleep(0.001)

        if hook_ctx.operation_id == "testAsyncHooksAfterResponse":
            return Exception("async validation failed")

        return response

    async def after_error(
        self,
        hook_ctx: AfterErrorContext,
        response: Optional[httpx.Response],
        error: Optional[Exception],
    ) -> Union[Tuple[Optional[httpx.Response], Optional[Exception]], Exception]:
        # Simulate async I/O (e.g., reporting error to monitoring service)
        await asyncio.sleep(0.001)

        if (
            response is not None
            and response.request.headers.get("x-test-after-error")
            and hook_ctx.operation_id == "getMalformedErrorResponse"
        ):
            return (None, RuntimeError(f"error fired mode={hook_ctx.response.mode}"))

        if hook_ctx.operation_id == "testAsyncHooksError":
            if response is not None and response.status_code != 400:
                return (None, Exception("expected status code 400"))
            return (None, Exception("async special test error case"))

        return (response, error)

    async def after_parse_error(
        self,
        hook_ctx: AfterParseErrorContext,
        response: httpx.Response,
        error: Exception,
    ) -> Exception:
        await asyncio.sleep(0.001)

        if response.request.headers.get(
            "x-test-after-parse-error"
        ) and hook_ctx.operation_id in (
            "statusGetDefaultError",
            "getMalformedErrorResponse",
        ):
            return RuntimeError(
                f"async parse_error fired mode={hook_ctx.response.mode} "
                f"execution={hook_ctx.response.execution}"
            )

        return error
