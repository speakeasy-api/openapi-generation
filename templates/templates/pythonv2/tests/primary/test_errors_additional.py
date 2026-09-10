import pytest
import httpx
import logging
from openapi import SDK
from openapi.models import errors, operations
from openapi.types import BaseModel

from .common_helpers import *
from .test_helpers import *


class _CountModel(BaseModel):
    count: int


def test_custom_error_inheritance_and_instanceof():
    """Test that custom errors can be caught as base APIBaseError and isinstance works correctly"""
    record_test("errors-custom-error-inheritance")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # Test TaggedError1 can be caught as specific type and base type
    req1 = operations.TaggedError1RequestBody(tag="tag1", error="Test Error")

    # Test catching as specific custom error type
    with pytest.raises(errors.ErrorUnionDiscriminatedPostResponseBody) as specific_exc:
        s.errors.error_union_discriminated_post(request=req1)

    assert specific_exc.value.data.error == "Test Error"

    # Test catching the same operation as APIBaseError
    from openapi.models.errors.apibaseerror import APIBaseError
    with pytest.raises(APIBaseError) as base_exc:
        s.errors.error_union_discriminated_post(request=req1)

    # Verify base error properties are accessible
    assert base_exc.value.status_code == 400
    assert hasattr(base_exc.value, 'message')
    assert hasattr(base_exc.value, 'body')
    assert hasattr(base_exc.value, 'headers')
    assert hasattr(base_exc.value, 'raw_response')

    # Test isinstance relationship - custom error should be instance of base APIBaseError
    assert isinstance(base_exc.value, APIBaseError)


def test_status_get_error_default_error_codes():
    record_test("errors-status-get-error-default-error-codes")
    logging.basicConfig(level=logging.DEBUG)
    debug_logger = logging.getLogger("sdk")
    s = SDK(server_url=HTTPBIN_URL, debug_logger=debug_logger)
    assert s is not None

    # When `clientServerStatusCodesAsErrors: true` is set (default), 4XX and 5XX ranges
    # are automatically treated as errors.

    # 400 and 500 responses are explicitly defined
    with pytest.raises(
        errors.APIError, match="API error occurred: Status 400"
    ) as exc_info_400:
        s.errors.status_get_error(status_code=400)

    assert exc_info_400.value.status_code == 400
    assert exc_info_400.value.raw_response is not None
    assert exc_info_400.value.raw_response.status_code == 400

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 500"
    ) as exc_info_500:
        s.errors.status_get_error(status_code=500)

    assert exc_info_500.value.status_code == 500
    assert exc_info_500.value.raw_response is not None
    assert exc_info_500.value.raw_response.status_code == 500

    # 404 and 503 responses are undefined but still treated as errors by default
    with pytest.raises(
        errors.APIError, match="API error occurred: Status 404"
    ) as exc_info_404:
        s.errors.status_get_error(status_code=404)

    assert exc_info_404.value.status_code == 404

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info_503:
        s.errors.status_get_error(status_code=503)

    assert exc_info_503.value.status_code == 503


def test_status_get_error_300_non_error():
    record_test("errors-status-get-error300-non-error")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.errors.status_get_error(status_code=300)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 300


def test_status_get_x_speakeasy_errors():
    record_test("errors-status-get-error-x-speakeasy-errors")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # 400 response is explicitly defined and is marked as an error in `x-speakeasy-errors`
    with pytest.raises(
        errors.APIError,
        match='API error occurred: Status 400. Body: {"message":"an error occurred","code":"400","type":"internal"}',
    ) as exc_info_400:
        s.errors.status_get_x_speakeasy_errors(status_code=400)

    assert exc_info_400.value.status_code == 400
    assert exc_info_400.value.raw_response is not None
    assert exc_info_400.value.raw_response.status_code == 400

    # 401 response is undefined but it is marked as an error in `x-speakeasy-errors`
    with pytest.raises(
        errors.APIError,
        match='API error occurred: Status 401. Body: {"message":"an error occurred","code":"401","type":"internal"}',
    ) as exc_info_401:
        s.errors.status_get_x_speakeasy_errors(status_code=401)

    assert exc_info_401.value.status_code == 401
    assert exc_info_401.value.raw_response is not None
    assert exc_info_401.value.raw_response.status_code == 401

    # 402 response is undefined and is not treated as an API error since it's not listed in `x-speakeasy-errors`.
    # Instead we raise a "Unexpected response received" exception.
    with pytest.raises(
        errors.APIError,
            match='Unexpected response received: Status 402. Body: {"message":"an error occurred","code":"402","type":"internal"}',
    ) as exc_info_402:
        s.errors.status_get_x_speakeasy_errors(status_code=402)

    assert exc_info_402.value.status_code == 402
    assert exc_info_402.value.raw_response is not None
    assert exc_info_402.value.raw_response.status_code == 402

    # Both 500 and 501 responses are marked as errors since `5XX` is listed `x-speakeasy-errors`.
    with pytest.raises(errors.Error, match="an error occurred") as exc_info_500:
        s.errors.status_get_x_speakeasy_errors(status_code=500)

    assert exc_info_500.value.data.code == "500"

    with pytest.raises(
        errors.StatusGetXSpeakeasyErrorsResponseBody,
        match='an error occurred',
    ) as exc_info_501:
        s.errors.status_get_x_speakeasy_errors(status_code=501)

    assert exc_info_501.value.data.code == "501"
    assert exc_info_501.value.data.http_meta is not None
    assert exc_info_501.value.data.http_meta.response is not None
    assert exc_info_501.value.data.http_meta.response.status_code == 501


def test_status_get_success_x_speakeasy_errors_empty_status_code_list():
    record_test("errors-status-get-error-x-speakeasy-errors-none")

    s = SDK()
    assert s is not None

    # `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
    # with a dummy list, meaning all responses are treated as non-errors.

    res = s.errors.status_get_non_error(status_code=200)
    assert res is not None
    assert res.http_meta.response.status_code == 200

    res = s.errors.status_get_non_error(status_code=400)
    assert res is not None
    assert res.http_meta.response.status_code == 400

    res = s.errors.status_get_non_error(status_code=500)
    assert res is not None
    assert res.http_meta.response.status_code == 500


def test_status_get_success_x_speakeasy_errors_unspecified_responses():
    record_test("errors-status-get-error-x-speakeasy-errors-default")

    s = SDK()
    assert s is not None

    # `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
    # by marking all *unspecified* responses as errors.

    # 200 and 400 responses are explicitly defined, so they are treated as non-errors.
    res = s.errors.status_get_default_error(status_code=200)
    assert res is not None
    assert res.http_meta.response.status_code == 200

    res = s.errors.status_get_default_error(status_code=400)
    assert res is not None
    assert res.http_meta.response.status_code == 400

    # 404 and 500 responses are undefined, so they are treated as errors.
    with pytest.raises(
        errors.APIError, match="API error occurred: Status 404"
    ) as exc_info_404:
        s.errors.status_get_default_error(status_code=404)

    assert exc_info_404.value.status_code == 404

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 500"
    ) as exc_info_500:
        s.errors.status_get_default_error(status_code=500)

    assert exc_info_500.value.status_code == 500

    # To make sure the catch-all "default" code gets properly applied in `templateErrorStatusCodesCheck`,
    # an AfterError hook was added to Hooks/TestHook.py to recover from 418 error.
    res = s.errors.status_get_default_error(status_code=418)
    assert res is not None
    assert res.http_meta.response.status_code == 200


def test_connection_error_get():
    record_test("errors-connection-error")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        httpx.ConnectError,
        match=r"(Name or service not known|nodename nor servname provided, or not known)",
    ):
        s.errors.connection_error_get()


def test_union_of_errors_post():
    record_test("errors-union-of-errors")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    req1 = operations.ErrorType1RequestBody(error="Error1")
    with pytest.raises(
        errors.ErrorUnionPostResponseBody,
        match='{"error":"Error1"}',
    ) as error1:
        s.errors.error_union_post(request=req1)
    assert error1.value.data.error == "Error1"
    assert error1.value.data.http_meta is not None
    assert error1.value.data.http_meta.response is not None
    assert error1.value.data.http_meta.response.status_code == 500

    req2 = operations.ErrorType2RequestBody(
        error_type2_message=operations.ErrorType2Message(message="Error2")
    )
    with pytest.raises(
        errors.ErrorUnionPostResponseBody,
        match='{"error":{"message":"Error2"}}',
    ) as error2:
        s.errors.error_union_post(request=req2)
    assert isinstance(error2.value.data.error, errors.SchemasError)
    assert error2.value.data.error.message == "Error2"
    assert error2.value.data.http_meta is not None
    assert error2.value.data.http_meta.response is not None
    assert error2.value.data.http_meta.response.status_code == 500


def test_discriminated_union_of_errors_post():
    record_test("errors-union-of-errors-discriminated")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    req1 = operations.TaggedError1RequestBody(tag="tag1", error="Error1")
    with pytest.raises(
        errors.ErrorUnionDiscriminatedPostResponseBody,
        match='{"error":"Error1","tag":"tag1"}',
    ) as error1:
        s.errors.error_union_discriminated_post(request=req1)
    assert error1.value.data.error == "Error1"
    assert error1.value.data.http_meta is not None
    assert error1.value.data.http_meta.response is not None
    assert error1.value.data.http_meta.response.status_code == 400

    req2 = operations.TaggedError2RequestBody(
        tag="tag2",
        tagged_error2_message=operations.TaggedError2Message(message="Error2"),
    )
    with pytest.raises(
        errors.ErrorUnionDiscriminatedPostResponseBody,
        match='{"error":{"message":"Error2"},"tag":"tag2"}',
    ) as error2:
        s.errors.error_union_discriminated_post(request=req2)
    assert isinstance(error2.value.data.error, errors.SchemasTaggedError2Error)
    assert error2.value.data.error.message == "Error2"
    assert error2.value.data.http_meta is not None
    assert error2.value.data.http_meta.response is not None
    assert error2.value.data.http_meta.response.status_code == 400


def test_error_hashability():
    """Test that error classes are hashable and can be used in sets/dicts"""

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # Test that we can catch errors and use them in sets/dicts
    req1 = operations.TaggedError1RequestBody(tag="tag1", error="Test Error 1")
    req2 = operations.TaggedError1RequestBody(tag="tag1", error="Test Error 2")

    errors_caught = []

    # Catch first error
    try:
        s.errors.error_union_discriminated_post(request=req1)
    except errors.ErrorUnionDiscriminatedPostResponseBody as e:
        errors_caught.append(e)

    # Catch second error
    try:
        s.errors.error_union_discriminated_post(request=req2)
    except errors.ErrorUnionDiscriminatedPostResponseBody as e:
        errors_caught.append(e)

    assert len(errors_caught) == 2

    # Test that errors can be used in sets (requires hashability)
    error_set = set(errors_caught)
    assert len(error_set) == 2, "Both errors should be in the set since they have different data"

    # Test that errors can be used as dictionary keys
    error_dict = {
        errors_caught[0]: "First error",
        errors_caught[1]: "Second error"
    }
    assert len(error_dict) == 2, "Both errors should be usable as dict keys"

    # Test hash consistency - same error should have same hash
    try:
        s.errors.error_union_discriminated_post(request=req1)
    except errors.ErrorUnionDiscriminatedPostResponseBody as e:
        same_error = e

    # The hash should be consistent for the same error data
    hash1 = hash(errors_caught[0])
    hash2 = hash(same_error)
    assert hash1 == hash2, "Same error data should produce same hash"

    # Test with base error classes too
    from openapi.models.errors.apibaseerror import APIBaseError
    try:
        s.errors.error_union_discriminated_post(request=req1)
    except APIBaseError as base_error:
        # Should be hashable
        base_error_set = {base_error}
        assert len(base_error_set) == 1

        base_error_dict = {base_error: "base error"}
        assert len(base_error_dict) == 1


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
def test_errors_schema_validation_enabled(status, shape, expected_substring):
    record_test("errors-error-body-validation-enabled")
    s = SDK()
    with pytest.raises(errors.ResponseValidationError) as exc_info:
        s.errors.get_malformed_error_response(status_code=status, shape=shape)
    err = exc_info.value
    assert err.status_code == status
    assert err.raw_response.status_code == status
    assert err.message.startswith("Response validation failed:")
    assert err.cause is not None
    if expected_substring is not None:
        assert isinstance(err.body, str)
        assert expected_substring in err.body


@pytest.mark.parametrize("shape", ["wrong-type", "missing-required"])
def test_response_body_validation_enabled(shape):
    record_test("errors-response-body-validation-enabled")
    s = SDK()
    # responseSchemaValidation enabled (default): a schema-violating 200 body
    # raises ResponseValidationError instead of being returned.
    with pytest.raises(errors.ResponseValidationError) as exc_info:
        s.errors.get_malformed_error_response(status_code=200, shape=shape)
    assert exc_info.value.raw_response.status_code == 200


@pytest.mark.parametrize("shape", ["wrong-type", "missing-required"])
def test_response_body_validation_enabled_raw_response(shape):
    record_test("errors-response-body-validation-enabled")
    s = SDK()
    # with_raw_response typed parsing honors responseSchemaValidation (enabled by
    # default): the `.parse(to=...)` override raises ResponseValidationError.
    raw = s.errors.with_raw_response.get_malformed_error_response(
        status_code=200, shape=shape
    )
    with pytest.raises(errors.ResponseValidationError):
        raw.parse(to=_CountModel)


# A 4XX/5XX must surface at the request call through the raw/streaming helpers,
# matching the direct path (`test_status_get_error_default_error_codes` raises
# `errors.APIError` on the call). Pre-fix the raw/streaming helpers handed back
# an un-raised APIResponse whose error only fired on a later `.parse()`.
_EAGER_ERROR_STATUSES = [400, 404, 500, 503]


@pytest.mark.parametrize("status", _EAGER_ERROR_STATUSES)
def test_status_get_error_raw_response_raises_eagerly(status):
    s = SDK(server_url=HTTPBIN_URL)
    with pytest.raises(errors.APIError) as exc_info:
        s.errors.with_raw_response.status_get_error(status_code=status)
    assert exc_info.value.status_code == status


@pytest.mark.parametrize("status", _EAGER_ERROR_STATUSES)
def test_status_get_error_streaming_response_raises_eagerly(status):
    s = SDK(server_url=HTTPBIN_URL)
    with pytest.raises(errors.APIError) as exc_info:
        with s.errors.with_streaming_response.status_get_error(status_code=status):
            pass
    assert exc_info.value.status_code == status


@pytest.mark.asyncio()
@pytest.mark.parametrize("status", _EAGER_ERROR_STATUSES)
async def test_status_get_error_raw_response_raises_eagerly_async(status):
    async with SDK(server_url=HTTPBIN_URL) as s:
        with pytest.raises(errors.APIError) as exc_info:
            await s.errors.with_raw_response.status_get_error_async(
                status_code=status
            )
        assert exc_info.value.status_code == status


@pytest.mark.asyncio()
@pytest.mark.parametrize("status", _EAGER_ERROR_STATUSES)
async def test_status_get_error_streaming_response_raises_eagerly_async(status):
    async with SDK(server_url=HTTPBIN_URL) as s:
        with pytest.raises(errors.APIError) as exc_info:
            async with s.errors.with_streaming_response.status_get_error_async(
                status_code=status
            ):
                pass
        assert exc_info.value.status_code == status
