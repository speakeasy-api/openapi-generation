from openapi import SDK
from openapi._version import __response_mode_header__
from openapi.models.operations import *
from openapi.models import errors
from openapi.utils.response_helpers import APIResponse, StreamedAPIResponse, consume_response_mode

from .common_helpers import *


def test_headers_override_request_headers():
    record_test("headers-override-request-headers")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    res = s.methods.method_get(
        http_headers={
            "x-inject-header-1": "foo",
            "x-inject-header-2": "bar",
        }
    )

    assert res.http_meta is not None

    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.object is not None
    assert res.object == MethodGetResponseBody(
        status="OK",
    )

    assert res.http_meta.request is not None
    assert res.http_meta.request.headers.get("x-inject-header-1") == "foo"
    assert res.http_meta.request.headers.get("x-inject-header-2") == "bar"


def test_raw_response_helpers_strip_internal_header():
    record_test("raw-response-helpers-strip-internal-header")

    s = SDK(server_url=HTTPBIN_URL)

    raw = s.methods.with_raw_response.method_get()
    assert isinstance(raw, APIResponse)
    assert raw.status_code == 200
    assert raw.http_request.headers.get("x-openapi-response-mode") is None

    with raw as raw_context:
        assert raw_context.parse().object == MethodGetResponseBody(
            status="OK",
        )

    parsed = raw.parse()
    assert parsed.object == MethodGetResponseBody(
        status="OK",
    )

    mode, remaining_headers = consume_response_mode(
        {
            __response_mode_header__.upper(): "raw",
            "x-public-header": "value",
        }
    )
    assert mode == "raw"
    assert remaining_headers == {"x-public-header": "value"}

    with s.methods.with_streaming_response.method_get() as streamed:
        assert isinstance(streamed, StreamedAPIResponse)
        assert streamed.status_code == 200
        assert streamed.http_request.headers.get("x-openapi-response-mode") is None
        assert streamed.parse().object == MethodGetResponseBody(
            status="OK",
        )


def test_headers_empty_response_body_with_headers():
    record_test("headers-empty-response-body-with-headers")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    res = s.response_headers.response_body_empty_with_headers(
        x_number_header=1.1,
        x_string_header="hello"
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.http_meta.response.headers is not None
    assert res.http_meta.response.headers["x-string-header"] == "hello"
    assert res.http_meta.response.headers["x-number-header"] == "1.1"

    # In Python, the headers dictionary is not templated when http_meta exists
    assert not hasattr(res, 'headers')


def test_headers_response_without_headers():
    record_test("headers-response-headers-none")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    # The 200 response in `errorResponseHeaders` does not include any headers.
    # but other responses in the same operation do.
    res200 = s.response_headers.error_response_headers(status_code=200, include_header=True)

    assert res200.http_meta is not None
    assert res200.http_meta.response is not None
    assert res200.http_meta.response.status_code == 200
    assert res200.auth_token is not None
    assert res200.auth_token.token == "test-token"


def test_headers_error_responses_without_headers():
    record_test("headers-error-response-headers-none")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    # 1. Basic error response (no body nor custom headers) whose
    # operation includes another response that has headers.
    try:
        s.response_headers.response_headers(status_code=400, include_required_header=True)
        assert False, "Expected error"
    except errors.APIError as err:
        assert err.status_code == 400
        assert err.body == ""

    # 2. Error response with a body but no custom headers whose
    # operation includes another response that has headers.
    try:
        s.response_headers.response_headers(status_code=500, include_required_header=True)
        assert False, "Expected error"
    except errors.APIError as err:
        assert err.status_code == 500
        assert err.body == '"Internal server error."\n'


def test_headers_response_headers_optional():
    record_test("headers-response-headers-optional")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    # 1. Success response with required header included
    res1 = s.response_headers.response_headers(
        status_code=200,
        include_required_header=True
    )
    assert res1.http_meta is not None
    assert res1.http_meta.response is not None
    assert res1.http_meta.response.status_code == 200
    assert res1.auth_token is not None
    assert res1.auth_token.token == "test-token"
    assert res1.http_meta.response.headers is not None
    assert res1.http_meta.response.headers["x-required-header"] == "required"
    assert "x-optional-header" not in res1.http_meta.response.headers

    # 2. Success response with required header omitted - SDK should not throw
    res2 = s.response_headers.response_headers(
        status_code=200,
        include_required_header=False
    )
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.auth_token is not None
    assert res2.auth_token.token == "test-token"
    assert "x-required-header" not in res2.http_meta.response.headers
    assert "x-optional-header" not in res2.http_meta.response.headers


def test_headers_error_response_headers_optional():
    record_test("headers-error-response-headers-optional")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    # 1. Error response with Retry-After header included
    try:
        s.response_headers.error_response_headers(
            status_code=429,
            include_header=True
        )
        assert False, "Expected error"
    except errors.APIError as err1:
        assert err1.status_code == 429
        assert err1.headers.get("retry-after") == "60"
        assert err1.body == '"Too many attempts. Please try again later."\n'

    # 2. Error response with Retry-After header omitted - SDK should not throw
    try:
        s.response_headers.error_response_headers(
            status_code=429,
            include_header=False
        )
        assert False, "Expected error"
    except errors.APIError as err2:
        assert err2.status_code == 429
        assert err2.headers.get("retry-after") is None
        assert err2.body == '"Too many attempts. Please try again later."\n'
