import pytest

from openapi import SDK
from openapi.models import errors
from openapi.types import BaseModel

from .common_helpers import *
from .test_helpers import *


class _CountModel(BaseModel):
    count: int


_MALFORMED_CASES = [
    (400, "sse", "forced sse"),
    (400, "plain", "internal server timeout"),
    (400, "html", "502 Bad Gateway"),
    (400, "malformed-json", '{"error":'),
    (400, "empty", None),
    (429, "sse", "too_many"),
    (500, "plain", "internal server timeout"),
]


@pytest.mark.parametrize("status,shape,expected_substring", _MALFORMED_CASES)
def test_errors_schema_validation_disabled(status, shape, expected_substring):
    record_test("errors-error-body-validation-disabled")
    s = SDK()
    with pytest.raises(errors.DefaultError) as exc_info:
        s.errors.get_malformed_error_response(status_code=status, shape=shape)
    err = exc_info.value
    assert err.status_code == status
    assert not isinstance(err, errors.ResponseValidationError)
    if expected_substring is not None:
        assert isinstance(err.body, str)
        assert expected_substring in err.body


_MALFORMED_RESPONSE_CASES = [
    ("wrong-type", {"count": "not-an-int"}),
    ("missing-required", {}),
]


@pytest.mark.parametrize("shape,expected_body", _MALFORMED_RESPONSE_CASES)
def test_response_body_validation_disabled(shape, expected_body):
    record_test("errors-response-body-validation-disabled")
    s = SDK()
    # responseSchemaValidation disabled: a schema-violating 200 body is returned
    # as raw decoded JSON instead of raising ResponseValidationError.
    res = s.errors.get_malformed_error_response(status_code=200, shape=shape)
    assert res == expected_body


@pytest.mark.parametrize("shape,expected_body", _MALFORMED_RESPONSE_CASES)
def test_response_body_validation_disabled_raw_response(shape, expected_body):
    record_test("errors-response-body-validation-disabled")
    s = SDK()
    # with_raw_response typed parsing must also honor responseSchemaValidation:
    # both the default `.parse()` and the `.parse(to=...)` override return the
    # raw decoded JSON instead of raising.
    raw = s.errors.with_raw_response.get_malformed_error_response(
        status_code=200, shape=shape
    )
    assert raw.parse() == expected_body
    assert raw.parse(to=_CountModel) == expected_body


# A 4XX/5XX must surface at the request call through the raw/streaming helpers
# — not be handed back as an un-raised APIResponse whose error only fires on a
# later `.parse()`. Mirrors the direct-path behavior (test above raises
# `DefaultError` from `get_malformed_error_response` on an error status).
@pytest.mark.parametrize("status,shape,expected_substring", _MALFORMED_CASES)
def test_errors_raw_response_raises_eagerly(status, shape, expected_substring):
    s = SDK()
    with pytest.raises(errors.DefaultError) as exc_info:
        s.errors.with_raw_response.get_malformed_error_response(
            status_code=status, shape=shape
        )
    err = exc_info.value
    assert err.status_code == status
    if expected_substring is not None:
        assert isinstance(err.body, str)
        assert expected_substring in err.body


@pytest.mark.parametrize("status,shape,expected_substring", _MALFORMED_CASES)
def test_errors_streaming_response_raises_eagerly(status, shape, expected_substring):
    s = SDK()
    with pytest.raises(errors.DefaultError) as exc_info:
        with s.errors.with_streaming_response.get_malformed_error_response(
            status_code=status, shape=shape
        ):
            pass
    assert exc_info.value.status_code == status


# Async eager-raise (raw + streaming) is covered by the primary variant's
# both-mode tests. Secondary enables oAuth2 client-credentials, whose
# `sdk_init` hook requires a sync `client`; `AsyncSDK` only accepts
# `async_client`, so async construction here is blocked by the variant's auth
# wiring — orthogonal to the eager-raise behavior under test.
