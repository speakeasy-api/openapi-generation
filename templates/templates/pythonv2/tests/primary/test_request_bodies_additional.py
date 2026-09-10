import io
import os
import pytest
from datetime import timedelta
from decimal import Decimal
from uuid import UUID

from openapi import SDK
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.types import BaseModel
from openapi.utils import marshal_json, unmarshal_json
from pydantic import ValidationError

from .common_helpers import *
from .test_helpers import *


def test_request_body_put_multipart_file():
    record_test("request-bodies-put-multipart-file")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    f = open(os.path.dirname(__file__) + "/testUpload.json", mode="rb")
    data = f.read()

    res = s.request_bodies.request_body_put_multipart_file(
        file=RequestBodyPutMultipartFileFile(content=data, file_name="testUpload.json")
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.files == {"file": data.decode("utf-8")}


def test_request_body_put_multipart_file_ref():
    record_test("request-bodies-put-multipart-file-ref")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    f = open(os.path.dirname(__file__) + "/testUpload.json", mode="rb")
    data = f.read()

    res = s.request_bodies.request_body_put_multipart_file_ref(
        file=BinaryString(content=data, file_name="testUpload.json")
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.files == {"file": data.decode("utf-8")}


def test_request_body_put_multipart_file_bytesio():
    """Multipart content accepts an in-memory binary stream (io.BytesIO).

    Guards the request-stream union widening to io.IOBase: io.BytesIO is not an
    instance of typing.IO nor io.BufferedReader, so the previous
    Union[bytes, IO[bytes], io.BufferedReader] rejected it at pydantic
    validation.
    """
    record_test("request-bodies-put-multipart-file-bytesio")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with open(os.path.dirname(__file__) + "/testUpload.json", mode="rb") as f:
        data = f.read()

    res = s.request_bodies.request_body_put_multipart_file(
        file=RequestBodyPutMultipartFileFile(
            content=io.BytesIO(data), file_name="testUpload.json"
        )
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.files == {"file": data.decode("utf-8")}


def test_request_body_put_multipart_file_stringio():
    """Multipart content accepts an in-memory text stream (io.StringIO).

    httpx rejects io.StringIO for multipart at the transport layer, so the
    serializer converts text streams to bytes before handoff.
    """
    record_test("request-bodies-put-multipart-file-stringio")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with open(os.path.dirname(__file__) + "/testUpload.json", mode="rb") as f:
        text = f.read().decode("utf-8")

    res = s.request_bodies.request_body_put_multipart_file(
        file=RequestBodyPutMultipartFileFile(
            content=io.StringIO(text), file_name="testUpload.json"
        )
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.files == {"file": text}


def test_request_body_put_multipart_file_handle():
    """Multipart content accepts an open file handle from open(...)."""
    record_test("request-bodies-put-multipart-file-handle")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    path = os.path.dirname(__file__) + "/testUpload.json"
    with open(path, mode="rb") as f:
        data = f.read()

    with open(path, mode="rb") as handle:
        res = s.request_bodies.request_body_put_multipart_file(
            file=RequestBodyPutMultipartFileFile(
                content=handle, file_name="testUpload.json"
            )
        )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.files == {"file": data.decode("utf-8")}


def test_request_body_put_bytes():
    record_test("request-bodies-put-bytes")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    f = open(os.path.dirname(__file__) + "/testUpload.json", mode="rb")
    data = f.read()
    res = s.request_bodies.request_body_put_bytes(request=data)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.data == data.decode("utf-8")


def test_request_no_body_no_content_type():
    record_test("request-bodies-no-body-no-content-type")

    s = SDK(server_url=HTTPBIN_URL)

    assert s is not None

    res = s.methods.method_get(server_url=API_TEST_SERVICE_URL)

    assert res.http_meta is not None

    assert res.http_meta.request is not None
    assert res.http_meta.request.content == b""
    assert res.http_meta.request.headers.get("content-type") is None

    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.object is not None
    assert res.object == MethodGetResponseBody(
        status="OK",
    )


def test_request_bodies_wildcard_no_content_type():
    # Assuming similar record_test function exists
    record_test("request-bodies-wildcard-no-content-type")

    sdk = SDK(server_url=HTTPBIN_URL)

    res = sdk.request_bodies.request_body_post_wildcard()

    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    assert res.http_meta.request is not None
    assert res.http_meta.request.headers.get("Content-Type") is None


@pytest.mark.parametrize(
    "test_id,obj,expected_dump",
    [
        ("required", ObjectWithOptionalFalseNullableFalseField(id=0, medical_record=MedicalRecord()), {"id": 0, "medicalRecord": {}}),
        ("nullable", ObjectWithOptionalFalseNullableTrueField(id=1, medical_record=None), {"id": 1, "medicalRecord": None}),
        ("optionalUnset", ObjectWithOptionalTrueNullableFalseField(id=2), {"id": 2}),
        ("optionalNullableUnset", ObjectWithOptionalTrueNullableTrueField(id=3), {"id": 3}),
        ("optionalSet", ObjectWithOptionalTrueNullableFalseField(id=4, medical_record=ObjectWithOptionalTrueNullableFalseFieldMedicalRecord()), {"id": 4, "medicalRecord": {}}),
        ("optionalNullableSet", ObjectWithOptionalTrueNullableTrueField(id=5, medical_record=None), {"id": 5, "medicalRecord": None}),
    ]
)
def test_request_bodies_serialize_model(test_id, obj, expected_dump):
    """ Tests custom serialization logic for handling optional and nullable fields"""
    assert obj.model_dump(by_alias=True, mode='json') == {**expected_dump, "type": "mammal"}


def test_request_bodies_optional_complex_number_types_unpopulated():
    """
    Verifies that calling the SDK method with no optional numeric fields
    sends {} and receives an empty object back. The old broken None guard
    (is_union(type(d))) always returned False for NoneType, causing
    PydanticSerializationError on model_dump(). After the fix, optional
    decimal/float/bigint fields set to None serialize correctly.

    Written as an additional test because the arazzo test generator
    populates schema example values even when payload is {}, which
    prevents testing the unpopulated case via generated tests.
    """
    record_test("request-bodies-complex-number-types-optional-unpopulated")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.request_bodies.request_body_post_complex_number_types_optional()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.object is not None
    assert res.object == RequestBodyPostComplexNumberTypesOptionalResponseBody(
        json_=OptionalComplexNumberTypes(),
    )


def test_request_bodies_complex_number_types_validation_error():
    """
    Verifies that passing None for required complex number type fields
    still raises a ValidationError. The serializer fix (if d is None: return None)
    must not suppress validation for required fields.
    """
    record_test("request-bodies-complex-number-types-required-missing")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(ValidationError, match="6 validation errors for ComplexNumberTypes") as exc_info:
        s.request_bodies.request_body_post_complex_number_types(
            path_big_int=8821239038968084,
            path_big_int_str=9223372036854775808,
            path_decimal=Decimal("3.141592653589793"),
            path_decimal_str=Decimal("3.14159265358979344719667586"),
            path_float64_str=1.1,
            path_int64_str=100,
            query_big_int=8821239038968084,
            query_big_int_str=9223372036854775808,
            query_decimal=Decimal("3.141592653589793"),
            query_decimal_str=Decimal("3.14159265358979344719667586"),
            query_float64_str=1.1,
            query_int64_str=100,
            bigint=None,
            bigint_str=None,
            decimal=None,
            decimal_str=None,
            float64_str=None,
            int64_str=None,
        )

    errs = exc_info.value.errors()
    assert {e["loc"][0] for e in errs} == {"bigint", "bigint_str", "decimal", "decimal_str", "float64_str", "int64_str"}


_BASE64_NON_UTF8_BYTES = b"\xff\xfe\x00\x01binary\xc3\x28"


def _b64(data: bytes) -> str:
    import base64 as _b
    return _b.b64encode(data).decode("ascii")


def test_request_bodies_base64_input_mode_file_split_components():
    """Split request / response component schemas under the file mode.

    Exercises the recommended spec-author pattern: the request schema applies
    x-speakeasy-base64-input-mode: file on opted-in properties, while the
    response uses a separate component schema. All three accepted input shapes
    (pre-encoded str, BytesIO binary stream, filesystem path) are sent in a
    single round-trip; the wire equality with locally computed base64 proves
    the SDK encoded each one correctly. The isinstance checks on the response
    fields are the parity guard for handleString's params.IsRequest gating —
    response-side opted-in fields must stay typed str even though the schema
    carries the extension.
    """
    import io
    import pathlib
    import tempfile
    record_test("request-bodies-base64-input-mode-file-split-components")

    with tempfile.NamedTemporaryFile(delete=False) as fh:
        fh.write(_BASE64_NON_UTF8_BYTES)
        path = pathlib.Path(fh.name)
    expected_path = _b64(_BASE64_NON_UTF8_BYTES)
    expected_stream = _b64(_BASE64_NON_UTF8_BYTES)

    s = SDK()
    res = s.request_bodies.post_base64_input_mode(
        data_byte=path,
        data_content_encoding=io.BytesIO(_BASE64_NON_UTF8_BYTES),
        data_plain="plain-value",
    )

    assert res.res is not None and res.res.json_ is not None
    # Response schema applies x-speakeasy-base64-input-mode on
    # data_content_encoding. applyBase64InputMode is gated to params.IsRequest
    # so the response-side TypeDef does not pick up the union — opted-in
    # response fields remain typed str.
    assert isinstance(res.res.json_.data_byte, str)
    assert isinstance(res.res.json_.data_content_encoding, str)
    assert res.res.json_.data_byte == expected_path
    assert res.res.json_.data_content_encoding == expected_stream
    assert res.res.json_.data_plain == "plain-value"


def test_request_bodies_base64_input_mode_file_shared_component():
    """Single component used for both request body and response.

    Behavior contract this test pins down:

    - Request side: opted-in field accepts BytesIO; the BeforeValidator on
      Base64EncodedString converts it to a base64 str at validation time.
    - Wire: server (httpbin) echoes the JSON request body inside its envelope.
    - Response side: opted-in field is typed Base64EncodedString (= str), the
      same as in the split-schema variant. Sharing a component across both
      directions does not widen the read-site type; callers do not need to
      narrow with isinstance before string operations.
    - Re-serialization: model_dump() on a response instance round-trips the
      base64 str unchanged (BeforeValidator is a no-op on str inputs).
    """
    import io
    record_test("request-bodies-base64-input-mode-file-shared-component")

    s = SDK()
    plain_value = "plain-shared"
    raw_bytes_for_byte_field = b"plain-byte-payload"
    pre_encoded_byte = _b64(raw_bytes_for_byte_field)
    expected_content_encoding = _b64(_BASE64_NON_UTF8_BYTES)

    res = s.request_bodies.post_base64_input_mode_shared(
        data_plain=plain_value,
        data_byte=pre_encoded_byte,
        data_content_encoding=io.BytesIO(_BASE64_NON_UTF8_BYTES),
    )

    assert res.res is not None and res.res.json_ is not None

    # Runtime: every opted-in response field is plain str (the BeforeValidator
    # passed the wire str through unchanged because str is not PathLike/IOBase).
    assert isinstance(res.res.json_.data_plain, str)
    assert isinstance(res.res.json_.data_byte, str)
    assert isinstance(res.res.json_.data_content_encoding, str)

    # Wire contents round-trip exactly: the BytesIO input was base64-encoded at
    # request validation, echoed by httpbin, and parsed back into a plain str.
    assert res.res.json_.data_plain == plain_value
    assert res.res.json_.data_byte == pre_encoded_byte
    assert res.res.json_.data_content_encoding == expected_content_encoding

    # Re-serializing the response must keep the wire shape intact.
    dumped = res.res.json_.model_dump(by_alias=True)
    assert dumped["dataPlain"] == plain_value
    assert dumped["dataByte"] == pre_encoded_byte
    assert dumped["dataContentEncoding"] == expected_content_encoding


def test_request_bodies_base64_file_input_read_is_idempotent():

    import io
    from openapi.types.base64fileinput import encode_base64_file_input

    record_test("request-bodies-base64-file-input-idempotent")

    expected = _b64(_BASE64_NON_UTF8_BYTES)

    # Repeated validation of the same stream yields the same encoding.
    stream = io.BytesIO(_BASE64_NON_UTF8_BYTES)
    assert encode_base64_file_input(stream) == expected
    assert encode_base64_file_input(stream) == expected

    # A stream positioned mid-way encodes from its current position and is
    # restored there, so a retry sees identical content.
    prefixed = io.BytesIO(b"prefix-" + _BASE64_NON_UTF8_BYTES)
    prefixed.seek(7)
    assert encode_base64_file_input(prefixed) == expected
    assert prefixed.tell() == 7
    assert encode_base64_file_input(prefixed) == expected


# Coverage for the uuidFormat config flag. With uuidFormat enabled,
# `format: uuid` string schemas generate as uuid.UUID instead of str. Exercises
# the `customer_id` (format: uuid) field on optionalRequestBodyPost through the
# SDK's real wire serializers.


def test_request_bodies_uuid_format():
    """customer_id is a native uuid.UUID and survives marshal -> unmarshal."""
    record_test("request-bodies-uuid-format")

    uid = UUID("12345678-1234-5678-1234-567812345678")
    body = OptionalRequestBodyPostRequestBody(customer_id=uid, include_archived=True)

    assert isinstance(body.customer_id, UUID)
    assert body.customer_id == uid

    raw = marshal_json(body, OptionalRequestBodyPostRequestBody)

    # On the wire the UUID is the canonical lowercase string form.
    assert '"customer_id":"12345678-1234-5678-1234-567812345678"' in raw

    restored = unmarshal_json(raw, OptionalRequestBodyPostRequestBody)

    assert isinstance(restored.customer_id, UUID)
    assert restored.customer_id == uid
    assert restored.include_archived is True


# Coverage for the durationFormat config flag. With durationFormat
# enabled, `format: duration` string schemas generate as datetime.timedelta
# instead of str. Exercises the `session_timeout` (format: duration) field on
# optionalRequestBodyPost, including the ISO 8601 duration string form on the wire.


def test_request_bodies_duration_format():
    """session_timeout is a native timedelta and survives marshal -> unmarshal."""
    record_test("request-bodies-duration-format")

    td = timedelta(hours=1, minutes=30)
    body = OptionalRequestBodyPostRequestBody(session_timeout=td, include_archived=True)

    assert isinstance(body.session_timeout, timedelta)
    assert body.session_timeout == td

    raw = marshal_json(body, OptionalRequestBodyPostRequestBody)

    # On the wire the duration is the ISO 8601 duration string form.
    assert '"session_timeout":"PT1H30M"' in raw

    restored = unmarshal_json(raw, OptionalRequestBodyPostRequestBody)

    assert isinstance(restored.session_timeout, timedelta)
    assert restored.session_timeout == td
    assert restored.include_archived is True
