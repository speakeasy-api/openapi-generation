from openapi import SDK
from .common_helpers import record_test, HTTPBIN_URL


def test_follow_redirects():
    record_test("redirects-are-followed")

    sdk = SDK(server_url=HTTPBIN_URL)

    res = sdk.redirects.redirects_are_followed()
    assert res.name is not None
    assert res.name == "John Doe"
