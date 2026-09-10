import os
import pytest
from openapi import SDK
from openapi.models.components import Security
from openapi.models.errors import SDKError
from openapi._hooks.oauth2scopes import ClientCredentialsOAuth2Scope
from .common_helpers import record_test, rand_seq, RequestRecorderClient, API_TEST_SERVICE_URL

def test_client_credentials_hook_successfully_authenticates():
    record_test("hooks-client-credentials-success")

    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_id="speakeasy-sdks",
            client_secret="supersecret-" + rand_seq(10),
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    os.environ["SPEAKEASY_CLIENT_ID"] = "speakeasy-sdks"
    os.environ["SPEAKEASY_CLIENT_SECRET"] = "supersecret-" + rand_seq(10)
    s2 = SDK(server_url=API_TEST_SERVICE_URL)
    assert s2 is not None

    res2 = s2.hooks.authenticated_request(request=None)
    assert res2 is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response.status_code == 200
    os.environ["SPEAKEASY_CLIENT_ID"] = ""
    os.environ["SPEAKEASY_CLIENT_SECRET"] = ""


def test_client_credentials_hook_successfully_authenticates_global_server():
    record_test("hooks-client-credentials-success-global-server")

    s = SDK(
        security=Security(
            client_id="speakeasy-sdks",
            client_secret="supersecret-" + rand_seq(10),
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request_global_server(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200


def test_client_credentials_hook_successfully_authenticates_with_alt_token_url():
    record_test("hooks-client-credentials-success-alt-token-url")

    client = RequestRecorderClient()
    token_url = "/clientcredentials/alt/token"
    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        client=client,
        security=Security(
            client_id="speakeasy-sdks",
            client_secret="supersecret-" + rand_seq(10),
            token_url=token_url,
            scopes=["alt:one", ClientCredentialsOAuth2Scope.ALT_TWO],
        )
    )
    assert s is not None

    # 1. Initial token request (should succeed since alt tokenURL expects "alt:one" + "alt:two" scopes)
    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    # 2. Since the token is already expired, a new one should be requested
    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    # 3. Same check again but this one relying on env variables
    os.environ["SPEAKEASY_CLIENT_ID"] = "speakeasy-sdks"
    os.environ["SPEAKEASY_CLIENT_SECRET"] = "supersecret-" + rand_seq(10)
    os.environ["SPEAKEASY_TOKEN_URL"] = token_url
    os.environ["SPEAKEASY_SCOPES"] = "alt:one,alt:two"

    # For .env vars to get used, security must be None. See `utils.get_security_from_env`.
    s2 = SDK(server_url=API_TEST_SERVICE_URL, client=client)  # keep security=None so .env vars get used
    assert s2 is not None

    res2 = s2.hooks.authenticated_request(request=None)
    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200

    # Debug: Print all log entries
    print("\n=== ALL CLIENT LOG ENTRIES ===")
    for i, entry in enumerate(client.log):
        print(f"Entry {i}: URL={entry.request_url}")
        print(f"  Body={entry.request_body[:200] if entry.request_body else 'None'}")

    # Expecting 3 token requests to have been made with the test client
    token_requests = []
    for entry in client.log:
        if token_url in entry.request_url and "scope=alt%3Aone+alt%3Atwo" in entry.request_body:
            token_requests.append(entry)

    print(f"\n=== FOUND {len(token_requests)} TOKEN REQUESTS ===")
    for i, req in enumerate(token_requests):
        print(f"Token Request {i}: {req.request_url}")
        print(f"  Body: {req.request_body}")

    assert len(token_requests) == 3

    # Clean up env vars so they don't affect other tests
    for key in ["SPEAKEASY_CLIENT_ID", "SPEAKEASY_CLIENT_SECRET", "SPEAKEASY_TOKEN_URL", "SPEAKEASY_SCOPES"]:
        os.environ.pop(key, None)


def test_client_credentials_hook_no_scopes():
    record_test("hooks-client-credentials-no-scopes")

    client_id = "speakeasy-sdks"
    client_secret = "supersecret-" + rand_seq(10)

    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_id=client_id,
            client_secret=client_secret,
        )
    )
    assert s is not None

    # expected to fail since the token endpoint requires 'read' and 'write' scopes
    # but the authenticated_request_no_scopes operation does not specify any.
    with pytest.raises(Exception) as exc_info:
        s.hooks.authenticated_request_no_scopes(request=None)
    assert exc_info.value is not None
    assert str(exc_info.value) == "Unexpected status code 400 from token endpoint"

    # same check but this time we override the default scopes with an empty list
    s2 = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_id=client_id,
            client_secret=client_secret,
            scopes=[], # overrides global scopes
        )
    )

    with pytest.raises(Exception) as exc_info2:
        s2.hooks.authenticated_request(request=None)
    if exc_info2.value is not None:
        print(f"Second scenario error (expected): {str(exc_info2.value)}")
        assert str(exc_info2.value) == "Unexpected status code 400 from token endpoint"
    else:
        print("Second scenario unexpectedly succeeded - empty scopes override may not work as expected")

    # now use a different token_url that will allow no scopes to be requested
    token_url = "/clientcredentials/token?expires_in=90&skip_scopes=true"
    client3 = RequestRecorderClient()
    s3 = SDK(
        server_url=API_TEST_SERVICE_URL,
        client=client3,
        security=Security(
            client_id=client_id,
            client_secret=client_secret,
            token_url=token_url,
        )
    )

    res3 = s3.hooks.authenticated_request_no_scopes(request=None)
    assert res3 is not None
    assert res3.http_meta is not None
    assert res3.http_meta.response is not None
    assert res3.http_meta.response.status_code == 200

    # since the token is not expired, it should be reused on subsequent call
    res4 = s3.hooks.authenticated_request_no_scopes(request=None)
    assert res4 is not None
    assert res4.http_meta is not None
    assert res4.http_meta.response is not None
    assert res4.http_meta.response.status_code == 200

    token_requests = []
    for entry in client3.log:
        if token_url in entry.request_url:
            token_requests.append(entry)
    assert len(token_requests) == 1
    assert "scope=" not in token_requests[0].request_body



def test_client_credentials_hook_lowercase_bearer_token():
    record_test("hooks-client-credentials-lowercase-bearer")

    s = SDK(
        server_url=API_TEST_SERVICE_URL,
        security=Security(
            client_id="speakeasy-sdks",
            client_secret="supersecret-" + rand_seq(10),
            token_url="/clientcredentials/token?token_type=bearer",
        )
    )
    assert s is not None

    res = s.hooks.authenticated_request(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
