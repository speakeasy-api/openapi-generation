import pytest

from openapi.models.shared import Security, BasicHTTP, APIKeyAuth, AccessToken
from openapi.models.errors import SDKDefaultError
from openapi import SDK

from .common_helpers import *
from .helpers import *


def test_global_security_option_basic_http_success():
    record_test("auth-basic-http-global-option")

    # Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
    with SDK(
        security=Security(
            basic_http=BasicHTTP(username="testUser", password="testPass"),
            access_token=AccessToken(access_token="Bearer ignored"),
        ),
    ) as sdk:
        res = sdk.auth.global_security_option_basic_http()
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.basic_auth is not None
        assert res.basic_auth.authenticated is True
        assert res.basic_auth.user == "testUser"


def test_global_security_option_fields_ordering():
    record_test("auth-global-security-option-fields-ordering")

    # Expected to fail since APIKeyAuth takes priority over BasicHTTP in global security definition
    with SDK(
        security=Security(
            api_key_auth=APIKeyAuth(api_key_auth="Bearer test_api_key"),
            basic_http=BasicHTTP(username="testUser", password="testPass"),
        ),
    ) as sdk:
        with pytest.raises(SDKDefaultError) as exc_info:
            sdk.auth.global_security_option_basic_http()
        assert exc_info.value.status_code == 401


def test_hoisted_security_option_access_token_first():
    record_test("auth-hoisted-security-option-access-token-first")

    with SDK(
        security=Security(
            api_key_auth=APIKeyAuth(api_key_auth="testApiKey"),
            access_token=AccessToken(access_token="Bearer ghp_xxxx"),
        ),
    ) as sdk:
        res = sdk.auth.hoisted_security_option_access_token_first()
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.token_auth_response is not None
        assert res.token_auth_response.token == "Bearer ghp_xxxx"


def test_hoisted_security_option_api_key_first():
    record_test("auth-hoisted-security-option-api-key-first")

    with SDK(
        security=Security(
            api_key_auth=APIKeyAuth(api_key_auth="testApiKey"),
            access_token=AccessToken(access_token="Bearer ghp_xxxx"),
        ),
    ) as sdk:
        res = sdk.auth.hoisted_security_option_api_key_first()
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.token_auth_response is not None
        assert res.token_auth_response.token == "testApiKey"


def test_hoisted_security_option_basic_http_only():
    record_test("auth-hoisted-security-option-basic-http-only")

    with SDK(
        security=Security(
            basic_http=BasicHTTP(username="testUser", password="testPass"),
            api_key_auth=APIKeyAuth(api_key_auth="testApiKey"),
            access_token=AccessToken(access_token="Bearer ghp_xxxx"),
        ),
    ) as sdk:
        res = sdk.auth.hoisted_security_option_basic_http_only()
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.basic_auth is not None
        assert res.basic_auth.authenticated is True
        assert res.basic_auth.user == "testUser"


def test_hoisted_security_invalid_option():
    record_test("auth-hoisted-security-invalid-option")

    with SDK(
        security=Security(
            basic_http=BasicHTTP(username="wrongUser", password="wrongPass"),
        ),
    ) as sdk:
        with pytest.raises(SDKDefaultError) as exc_info:
            sdk.auth.hoisted_security_option_access_token_first()
        assert exc_info.value.status_code == 401
