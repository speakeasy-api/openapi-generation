import pytest

from openapi.models import shared
from openapi.models.errors import APIError
from openapi import SDK

from .common_helpers import *
from .helpers import *


def test_function_callbacks_for_o_auth_support_global_security():
    record_test("auth-function-callbacks-oauth-global-security")

    s = SDK(server_url=HTTPBIN_URL, security=lambda: shared.Security(oauth2="Bearer global"))
    assert s is not None

    res = s.auth.global_bearer_auth()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.token is not None
    assert res.token.token == "global"


def test_global_security_fields_ordering():
    record_test("auth-global-security-fields-ordering")

    # When maintainOpenApiOrder: false is set, the first non-nil field (in alphabetical order) is selected.
    # In this case api_key_auth takes precedence over basic_http, which is invalid for this endpoint.
    s = SDK(
        security=shared.Security(
            api_key_auth="testApiKey",
            basic_http=shared.SchemeBasicHTTP(username="testUser", password="testPass"),
        ),
    )

    with pytest.raises(APIError) as exc_info:
        s.auth.global_security_basic_http()
    assert exc_info.value.status_code == 401


def test_hoisted_security_access_token_only():
    record_test("auth-hoisted-security-access-token-only")

    s = SDK(
        security=shared.Security(
            api_key_auth="testApiKey",
            basic_http=shared.SchemeBasicHTTP(username="testUser", password="testPass"),
            access_token="Bearer ghp_xxxx",
        ),
    )

    res = s.auth.hoisted_security_access_token_only()
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.token is not None
    assert res.token.token == "Bearer ghp_xxxx"


def test_hoisted_security_access_token_first():
    record_test("auth-hoisted-security-access-token-first")

    s = SDK(
        security=shared.Security(
            api_key_auth="testApiKey",
            access_token="Bearer ghp_xxxx",
        ),
    )

    res = s.auth.hoisted_security_access_token_first()
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.token is not None
    assert res.token.token == "Bearer ghp_xxxx"


def test_hoisted_security_api_key_first():
    record_test("auth-hoisted-security-api-key-first")

    s = SDK(
        security=shared.Security(
            api_key_auth="testApiKey",
            access_token="Bearer ghp_xxxx",
        ),
    )

    res = s.auth.hoisted_security_api_key_first()
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.token is not None
    assert res.token.token == "testApiKey"


def test_hoisted_security_basic_http_only():
    record_test("auth-hoisted-security-basic-http-only")

    s = SDK(
        security=shared.Security(
            basic_http=shared.SchemeBasicHTTP(username="testUser", password="testPass"),
            api_key_auth="testApiKey",
            access_token="Bearer ghp_xxxx",
        ),
    )

    res = s.auth.hoisted_security_basic_http_only()
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.basic_auth is not None
    assert res.basic_auth.authenticated is True
    assert res.basic_auth.user == "testUser"


def test_hoisted_security_invalid_field():
    record_test("auth-hoisted-security-invalid-field")

    s = SDK(
        security=shared.Security(
            basic_http=shared.SchemeBasicHTTP(username="user", password="pass"),
        ),
    )

    with pytest.raises(APIError) as exc_info:
        s.auth.hoisted_security_access_token_first()
    assert exc_info.value.status_code == 401
