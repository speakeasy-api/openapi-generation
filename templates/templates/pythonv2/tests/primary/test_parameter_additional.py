from datetime import date, timedelta
from decimal import Decimal
import json
from uuid import UUID
from openapi import SDK
from openapi.models import shared
from openapi.utils import parse_datetime

from .common_helpers import record_test, HTTPBIN_URL


def test_parameters_header_params_nil():
    # Assuming similar record_test function exists
    record_test("parameters-header-params-nil")

    sdk = SDK(server_url=HTTPBIN_URL)

    res = sdk.parameters.header_params_nil(nullable_header=None)

    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    assert res.http_meta.request is not None
    assert res.http_meta.request.headers.get("Nullable-Header") is None
    assert res.http_meta.request.headers.get("Optional-Header") is None
    assert res.http_meta.request.headers.get("Optional-Nullable-Header") is None


def test_parameters_allow_empty_value_query_params():
    from .helpers import sort_query_parameters

    record_test("parameters-allow-empty-value")

    sdk = SDK(server_url=HTTPBIN_URL)

    res = sdk.parameters.allow_empty_value_query_params(
        arr_param=[],
        arr_param_omit_empty=[],
        nullable_str_param=None,
        num_param=None,
        str_param="",
    )

    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.url, str)

    assert sort_query_parameters(res.res.url) == (
        f"{HTTPBIN_URL}/anything/allowEmptyValue?"
        "arrParam=&nullableStrParam=&numParam=&strParam="
    )

    res2 = sdk.parameters.allow_empty_value_query_params(
        arr_param=["a", "b"],
        arr_param_omit_empty=["x", "y"],
        nullable_str_param="nullable",
        num_param=123,
        str_param="test",
    )

    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.res is not None
    assert isinstance(res2.res.url, str)

    assert sort_query_parameters(res2.res.url) == (
        f"{HTTPBIN_URL}/anything/allowEmptyValue?"
        "arrParam=a&arrParam=b&arrParamOmitEmpty=x&arrParamOmitEmpty=y&"
        "nullableStrParam=nullable&numParam=123&strParam=test"
    )


def test_parameters_path_parameter_json():
    record_test("parameters-path-parameter-json")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.parameters.path_parameter_json(
        json_obj={
            "any": "any",
            "bigint": 8821239038968084,
            "bigint_str": 9223372036854775808,
            "bool_": True,
            "bool_opt": True,
            "date_": date.fromisoformat("2020-01-01"),
            "date_time": parse_datetime("2020-01-01T00:00:00.001Z"),
            "decimal": Decimal("3.141592653589793"),
            "decimal_str": Decimal("3.14159265358979344719667586"),
            "enum": "one",
            "float32": 1.1,
            "float64_str": 1.1,
            "int_": 1,
            "int32": 1,
            "int32_enum": 55,
            "int64_str": 100,
            "int_enum": shared.IntEnum.SECOND,
            "num": 1.1,
            "str_": "test",
            "str_opt": "testOptional",
        }
    )
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.url, str)

    param = res.res.url.split("/")[-1]
    parsed = json.loads(param)
    parsed["date"] = date.fromisoformat(parsed["date"])
    parsed["dateTime"] = parse_datetime(parsed["dateTime"])

    assert parsed == {
        "any": "any",
        "bigint": 8821239038968084,
        "bigintStr": "9223372036854775808",
        "bool": True,
        "boolOpt": True,
        "date": date.fromisoformat("2020-01-01"),
        "dateTime": parse_datetime("2020-01-01T00:00:00.001Z"),
        "decimal": 3.141592653589793,
        "decimalStr": "3.14159265358979344719667586",
        "enum": "one",
        "float32": 1.1,
        "float64Str": "1.1",
        "int": 1,
        "int32": 1,
        "int32Enum": 55,
        "int64Str": "100",
        "intEnum": 2,
        "num": 1.1,
        "str": "test",
        "strOpt": "testOptional",
    }


def test_parameters_path_parameter_formats():
    record_test("parameters-path-parameter-formats")

    s = SDK(server_url=HTTPBIN_URL)

    uuid_value = UUID("12345678-1234-5678-1234-567812345678")
    date_value = date.fromisoformat("2020-01-01")
    date_time_value = parse_datetime("2020-01-01T00:00:00.001Z")
    duration_value = timedelta(hours=1, minutes=30)

    res = s.parameters.path_parameter_formats(
        uuid_param=uuid_value,
        date_param=date_value,
        date_time_param=date_time_value,
        duration_param=duration_value,
        uuid_query=uuid_value,
        date_query=date_value,
        date_time_query=date_time_value,
        duration_query=duration_value,
        x_uuid_header=uuid_value,
        x_date_header=date_value,
        x_date_time_header=date_time_value,
        x_duration_header=duration_value,
    )

    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.url.split("?")[0] == (
        f"{HTTPBIN_URL}/anything/pathParams/formats"
        "/uuid/12345678-1234-5678-1234-567812345678"
        "/date/2020-01-01"
        "/dateTime/2020-01-01T00:00:00.001000Z"
        "/duration/PT1H30M"
    )

    request = res.http_meta.request
    assert request is not None
    assert dict(request.url.params) == {
        "uuidQuery": "12345678-1234-5678-1234-567812345678",
        "dateQuery": "2020-01-01",
        "dateTimeQuery": "2020-01-01T00:00:00.001000Z",
        "durationQuery": "PT1H30M",
    }
    assert request.headers["x-uuid-header"] == "12345678-1234-5678-1234-567812345678"
    assert request.headers["x-date-header"] == "2020-01-01"
    assert request.headers["x-date-time-header"] == "2020-01-01T00:00:00.001000Z"
    assert request.headers["x-duration-header"] == "PT1H30M"


def test_parameters_path_parameter_format_union():
    record_test("parameters-path-parameter-format-union")

    s = SDK(server_url=HTTPBIN_URL)

    res = s.parameters.path_parameter_format_union(id_or_name="widget")

    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.url == f"{HTTPBIN_URL}/anything/pathParams/formatUnion/widget"
