import os
import httpx
from openapi import SDK, AsyncSDK
from openapi.models.errors import SDKError
from .common_helpers import *
import pytest


@pytest.mark.asyncio()
async def test_global_security_flattening():
    record_test("auth-global-security-flattening")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="Bearer testToken") as s:
        assert s is not None

        res = await s.auth.api_key_auth_global()
        assert res is not None
        assert res.status_code == 200


@pytest.mark.asyncio()
async def test_global_security_flattening_callback():
    record_test("auth-global-security-flattening-callback")

    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth=lambda: "Bearer testToken") as s:
        assert s is not None

        res = await s.auth.api_key_auth_global()
        assert res is not None
        assert res.status_code == 200


def test_global_security_flattening_env_var_fallback_sync():
    record_test("auth-global-security-flattening-env-var-fallback")

    os.environ["SPEAKEASY_API_KEY_AUTH"] = "Bearer testToken"
    try:
        with SDK() as s:
            assert s is not None
            assert s.sdk_configuration.security is None

            res = s.auth.api_key_auth_global()
            assert res is not None
            assert res.status_code == 200
    finally:
        os.environ.pop("SPEAKEASY_API_KEY_AUTH", None)


@pytest.mark.asyncio()
async def test_global_security_flattening_env_var_fallback_async():
    record_test("auth-global-security-flattening-env-var-fallback")

    os.environ["SPEAKEASY_API_KEY_AUTH"] = "Bearer testToken"
    try:
        async with AsyncSDK() as s:
            assert s is not None
            assert s.sdk_configuration.security is None

            res = await s.auth.api_key_auth_global()
            assert res is not None
            assert res.status_code == 200
    finally:
        os.environ.pop("SPEAKEASY_API_KEY_AUTH", None)


def test_global_security_flattening_empty_string_ignores_env_var():
    record_test("auth-global-security-flattening-empty")

    captured = {}

    def capture(request):
        captured["headers"] = request.headers
        return httpx.Response(401, request=request)

    os.environ["SPEAKEASY_API_KEY_AUTH"] = "Bearer testToken"
    try:
        client = httpx.Client(transport=httpx.MockTransport(capture))
        with SDK(server_url=HTTPBIN_URL, api_key_auth="", client=client) as s:
            with pytest.raises(SDKError):
                s.auth.api_key_auth_global()
    finally:
        os.environ.pop("SPEAKEASY_API_KEY_AUTH", None)

    assert "authorization" not in captured["headers"]
