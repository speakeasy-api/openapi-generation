from customhttp import SDK
from customhttp.models.security import SchemeCustomHTTPSecurity
from .common_helpers import record_test, API_TEST_SERVICE_URL
import pytest


@pytest.mark.asyncio()
async def test_custom_http_scheme_only():
    record_test("auth-custom-security-scheme-only")

    test_scopes = ["read:products", "write:products"]

    async with SDK(
        server_url=API_TEST_SERVICE_URL,
        custom_http=SchemeCustomHTTPSecurity(
            user_id=54321,
            role="manager",
            passphrase="secure-passphrase-123",
            access_code=104,
            scopes=test_scopes,
        )
    ) as s:
        assert s is not None

        res = await s.auth.custom_http_only_async()
        assert res is not None
        assert res.grant == "access_granted"
        assert res.scopes == test_scopes
