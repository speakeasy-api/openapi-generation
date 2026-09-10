from openapi import AsyncSDK

from .common_helpers import *
import httpx
import pytest


@pytest.mark.asyncio()
async def test_globals_query_parameter_get_uses_global():
    record_test("globals-query-parameter-get-uses-global")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_query_param="test") as s:
        assert s is not None

        res = await s.globals.globals_query_parameter_get()
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.args.global_query_param == "test"


@pytest.mark.asyncio()
async def test_globals_query_parameter_get_uses_local():
    record_test("globals-query-parameter-get-uses-local")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_query_param="test") as s:
        assert s is not None

        res = await s.globals.globals_query_parameter_get(global_query_param="local")
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.args.global_query_param == "local"


@pytest.mark.asyncio()
async def test_global_path_parameter_get_uses_global():
    record_test("globals-path-parameter-get-uses-global")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_path_param=1) as s:
        assert s is not None

        res = await s.globals.global_path_parameter_get()
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.url == f"{HTTPBIN_URL}/anything/globals/pathParameter/1"


@pytest.mark.asyncio()
async def test_global_path_parameter_get_uses_local():
    record_test("globals-path-parameter-get-uses-local")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_path_param=1) as s:
        assert s is not None

        res = await s.globals.global_path_parameter_get(global_path_param=2)
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.url == f"{HTTPBIN_URL}/anything/globals/pathParameter/2"


@pytest.mark.asyncio()
async def test_global_header_get_uses_global():
    record_test("globals-header-get-uses-global")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_header_param=True) as s:
        assert s is not None

        res = await s.globals.globals_header_get()
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.headers is not None
        assert res.res.headers["Globalheaderparam"] == "true"


@pytest.mark.asyncio()
async def test_global_header_get_uses_local():
    record_test("globals-header-get-uses-local")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_header_param=True) as s:
        assert s is not None

        res = await s.globals.globals_header_get(global_header_param=False)
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.headers is not None
        assert res.res.headers["Globalheaderparam"] == "false"


@pytest.mark.asyncio()
async def test_global_header_keeps_custom_client_headers():
    record_test("globals-header-keeps-custom-client-headers")

    http_client = httpx.AsyncClient(headers={"x-custom-header": "someValue"})

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token", global_header_param=True, async_client=http_client) as s:
        assert s is not None

        res = await s.globals.globals_header_get()
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.headers is not None
        assert res.res.headers["Globalheaderparam"] == "true"
        assert res.res.headers["X-Custom-Header"] == "someValue"


@pytest.mark.asyncio()
async def test_globals_hidden_post():
    record_test("globals-hidden-post")

    async with AsyncSDK(
        server_url=HTTPBIN_URL,
        api_key_auth="token",
        global_hidden_query_param="hello",
        global_hidden_header_param="world",
        global_hidden_path_param="test",
    ) as s:
        assert s is not None

        res = await s.globals.globals_hidden_post(
            test="friend",
            other=37,
        )
        assert res is not None
        assert res.status_code == 200
        assert res.res is not None
        assert res.res.args.global_hidden_query_param == "hello"
        assert res.res.json_.test == "friend"
        assert res.res.json_.other == 37
        assert res.res.headers["Globalhiddenheaderparam"] == "world"
        assert (
            res.res.url
            == f"{HTTPBIN_URL}/anything/globals/hidden/test?globalHiddenQueryParam=hello"
        )
