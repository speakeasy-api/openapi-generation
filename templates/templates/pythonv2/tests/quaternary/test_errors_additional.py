"""Round-trip tests for ``responseSchemaValidation: lenient``."""

import pytest

from openapi import SDK
from openapi.models import errors
from openapi.models.operations import GetMalformedErrorResponseResponseBody

from .common_helpers import record_test


def test_lenient_response_body_validation_mismatches_are_tolerated():
    record_test("errors-response-body-validation-lenient")
    s = SDK()

    missing_required = s.errors.get_malformed_error_response(
        status_code=200, shape="missing-required"
    )
    assert isinstance(missing_required.object, GetMalformedErrorResponseResponseBody)
    assert missing_required.object.count is None

    wrong_type = s.errors.get_malformed_error_response(
        status_code=200, shape="wrong-type"
    )
    assert isinstance(wrong_type.object, GetMalformedErrorResponseResponseBody)
    assert wrong_type.object.count == "not-an-int"


def test_lenient_non_json_success_body_raises_response_validation_error():
    record_test("errors-response-body-validation-lenient-non-json")
    s = SDK()

    # A 200 body that is not parseable JSON has no best-effort typed
    # representation, so even lenient construction surfaces the default
    # ResponseValidationError rather than leaking a raw JSON decode error.
    with pytest.raises(errors.ResponseValidationError) as exc_info:
        s.errors.get_malformed_error_response(status_code=200, shape="html")

    assert exc_info.value.raw_response.status_code == 200
    assert exc_info.value.cause is not None


def test_lenient_known_error_union_variant_stays_typed():
    record_test("errors-response-body-validation-lenient-union-variant")
    s = SDK()

    # The service returns a 5XX discriminated error union whose `tag` selects the
    # mapped variant TaggedError1 (error: string), but the body carries `error`
    # as an object. Lenient keeps the typed variant instead of degrading to the
    # Unknown fallback.
    with pytest.raises(errors.GetMalformedErrorUnionResponseResponseBody) as exc_info:
        s.errors.get_malformed_error_union_response()

    variant = exc_info.value.data
    assert isinstance(variant, errors.TaggedError1Data)
    assert variant.tag == "tag1"
    assert variant.error == {"unexpected": "object"}
