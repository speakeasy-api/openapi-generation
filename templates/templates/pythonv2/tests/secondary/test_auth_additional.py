import os
from openapi.models import operations, shared
from openapi import SDK
from .common_helpers import record_test, rand_seq, RequestRecorderClient, HTTPBIN_URL


def test_no_auth():
    record_test("auth-hoisted-no-auth-retained")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    s.auth.no_auth()


def test_basic_auth():
    record_test("auth-hoisted-basic-auth")

    s = SDK(
        server_url=HTTPBIN_URL,
        security=shared.Security(
            username="testUser",
            password="testPass",
        )
    )
    assert s is not None

    res = s.auth.basic_auth(user="testUser", passwd="testPass")
    assert res is not None
    assert res.authenticated is True

    os.environ["SPEAKEASY_USERNAME"] = "testUser"
    os.environ["SPEAKEASY_PASSWORD"] = "testPass"
    s2 = SDK(
        server_url=HTTPBIN_URL,
        security=shared.Security(
            username="testUser",
            password="testPass",
        )
    )
    assert s2 is not None

    res2 = s2.auth.basic_auth(user="testUser", passwd="testPass")
    assert res2 is not None
    assert res2.authenticated is True
    os.environ["SPEAKEASY_USERNAME"] = ""
    os.environ["SPEAKEASY_PASSWORD"] = ""


def test_multiple_mixed_options_auth():
    record_test("auth-hoisted-operation-auth-retained")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    s.auth_new.multiple_mixed_options_auth(
        request=shared.AuthServiceRequestBody(
            basic_auth=shared.AuthServiceRequestBodyBasicAuth(
                username="testUser",
                password="testPass",
            )
        ),
        security=operations.MultipleMixedOptionsAuthSecurity(
            basic_auth=shared.SchemeBasicAuth(
                username="testUser",
                password="testPass",
            )
        ),
    )


def test_authenticated_request_operation_level_oauth2():
    record_test("auth-operation-level-oauth2")

    recorder = RequestRecorderClient()
    s = SDK(server_url=HTTPBIN_URL, client=recorder)
    assert s is not None

    client_id = "speakeasy-sdks"
    client_secret = "supersecret-" + rand_seq(10)

    # A token should be requested with 'read', 'write' and 'erase' scopes
    s.hooks.authenticated_request(
        security=operations.AuthenticatedRequestSecurity(
            client_id=client_id,
            client_secret=client_secret,
            audience="",
        )
    )

    # This operation requires 'read' and 'write' scopes.
    # The same token should be reused since [read, write, erase] is a superset of [read, write].
    s.hooks.authenticated_request_unflattened(
        security=operations.AuthenticatedRequestUnflattenedSecurity(
            client_credentials=shared.SchemeClientCredentials(
                client_id="speakeasy-sdks",
                client_secret=client_secret,
                audience="",
            )
        )
    )

    # Verify that only a single token was requested
    token_requested = False
    for entry in recorder.log:
        if "/clientcredentials/token" in entry.request_url:
            if token_requested:
                assert False, f"Expected only a single token request"

            assert "scope=read+write+erase" in entry.request_body
            token_requested = True

