import random
import string
import os
from openapi import SDK
from openapi.models.components import Security, SchemeClientCredentials
from openapi._hooks.oauth2scopes import ClientCredentialsOAuth2Scope
from .common_helpers import record_test, API_TEST_SERVICE_URL


def test_client_credentials_hook_successfully_authenticates():
    record_test("hooks-client-credentials-basic-success")

    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_credentials=SchemeClientCredentials(
                client_id="speakeasy-sdks",
                client_secret="supersecret-" + rand_seq(10),
                audience="",
            )
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    os.environ["SPEAKEASY_CLIENT_ID"] = "speakeasy-sdks"
    os.environ["SPEAKEASY_CLIENT_SECRET"] = "supersecret-" + rand_seq(10)
    os.environ["SPEAKEASY_AUDIENCE"] = "test"
    s2 = SDK(server_url=API_TEST_SERVICE_URL)
    assert s2 is not None

    res2 = s2.hooks.authenticated_request(request=None)
    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    os.environ["SPEAKEASY_CLIENT_ID"] = ""
    os.environ["SPEAKEASY_CLIENT_SECRET"] = ""
    os.environ.pop("SPEAKEASY_AUDIENCE", None)


def test_client_credentials_hook_successfully_authenticates_global_server():
    record_test("hooks-client-credentials-basic-success-global-server")

    s = SDK(
        security=Security(
            client_credentials=SchemeClientCredentials(
                client_id="speakeasy-sdks",
                client_secret="supersecret-" + rand_seq(10),
                audience="",
            )
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request_global_server(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    res = s.hooks.authenticated_request_global_server(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200


def test_client_credentials_hook_successfully_authenticates_with_alt_token_url():
    record_test("hooks-client-credentials-basic-success-alt-token-url")

    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_credentials=SchemeClientCredentials(
                client_id="speakeasy-sdks",
                client_secret="supersecret-" + rand_seq(10),
                token_url="/clientcredentials/alt/token",
                scopes=["alt:one", ClientCredentialsOAuth2Scope.ALT_TWO],
                audience="",
            )
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    os.environ["SPEAKEASY_CLIENT_ID"] = "speakeasy-sdks"
    os.environ["SPEAKEASY_CLIENT_SECRET"] = "supersecret-" + rand_seq(10)
    os.environ["SPEAKEASY_AUDIENCE"] = "test"
    s2 = SDK(server_url=API_TEST_SERVICE_URL)
    assert s2 is not None

    res2 = s2.hooks.authenticated_request(request=None)
    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    os.environ["SPEAKEASY_CLIENT_ID"] = ""
    os.environ["SPEAKEASY_CLIENT_SECRET"] = ""
    os.environ.pop("SPEAKEASY_AUDIENCE", None)


def rand_seq(n) -> str:
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=n))
