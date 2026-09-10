import pytest
from openapi import SDK
from openapi.models import errors
from .common_helpers import record_test, HTTPBIN_URL


def test_api_key_auth_global():
    record_test("auth-api-key-auth-global")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        errors.SDKException, match="API error occurred: Status 401"
    ) as exc_info:
        s.auth.api_key_auth_global()

    assert exc_info.value.status_code == 401
