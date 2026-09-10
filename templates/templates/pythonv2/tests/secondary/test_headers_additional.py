from openapi import SDK

from .common_helpers import *


def test_headers_response_body_with_headers_flat():
    record_test("headers-response-body-with-headers-flat")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    res = s.response_headers.response_headers(
        status_code=200,
        include_required_header=True
    )
    assert res is not None
    assert res.result is not None
    assert res.result.token == "test-token"
    assert res.headers is not None
    assert res.headers["x-required-header"] == ["required"]


def test_headers_empty_response_body_with_headers_flat():
    record_test("headers-empty-response-body-with-headers-flat")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    res = s.response_headers.response_body_empty_with_headers(
        x_number_header=1.1,
        x_string_header="hello"
    )

    assert res is not None
    assert res.headers is not None
    assert res.headers["x-string-header"] == ["hello"]
    assert res.headers["x-number-header"] == ["1.1"]
