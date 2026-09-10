from openapi import AsyncSDK
from .common_helpers import record_test, HTTPBIN_PORT
import pytest


@pytest.mark.asyncio()
async def test_select_global_server_by_name_with_templates_defaults():
    record_test("servers-select-global-server-by-name-with-templates-defaults")

    async with AsyncSDK(api_key_auth="token", server="TEMPLATED") as s:
        assert s is not None

        res = await s.servers.select_global_server()

        assert res is not None
        assert res.status_code == 200


@pytest.mark.asyncio()
async def test_select_global_server_by_name_with_templates_valid():
    record_test("servers-select-global-server-by-name-with-templates-valid")

    async with AsyncSDK(
        api_key_auth="token", server="TEMPLATED", hostname="localhost", port=HTTPBIN_PORT
    ) as s:
        assert s is not None

        res = await s.servers.select_global_server()

        assert res is not None
        assert res.status_code == 200


@pytest.mark.asyncio()
async def test_select_global_server_by_name_with_templates_broken():
    record_test("servers-select-global-server-by-name-with-templates-broken")

    async with AsyncSDK(api_key_auth="token", server="TEMPLATED", hostname="broken", port="12345") as s:
        assert s is not None

        error = None

        try:
            await s.servers.select_global_server()
        except Exception as err:
            error = err

        assert error is not None
