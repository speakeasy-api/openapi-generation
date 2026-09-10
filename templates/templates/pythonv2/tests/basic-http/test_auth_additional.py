import pytest

from openapi.models import *
from openapi.errors import SDKError
from openapi import SDK

from .common_helpers import *


def test_basic_auth_optional():
    record_test("auth-basic-auth-operation-optional")

    with SDK(
        security=Security(
            username="wrongUser",
            password="wrongPass",
        ),
    ) as sdk:
        res = sdk.auth.basic_auth_optional(
            security=BasicAuthOptionalSecurity(
                username="testUser",
                password="testPass",
            ),
        )
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.basic_auth_response is not None
        assert res.basic_auth_response.authenticated is True
        assert res.basic_auth_response.user == "testUser"

    with SDK(
        security=Security(
            username="testUser",
            password="testPass",
        ),
    ) as sdk:
        with pytest.raises(SDKError):
            sdk.auth.basic_auth_optional(
                security=BasicAuthOptionalSecurity(),
            )
