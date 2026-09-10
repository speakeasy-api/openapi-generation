import inspect
import json
from types import MappingProxyType

from openapi import SDK, models
from openapi.collections import Collections
from openapi.requestbodies import RequestBodies

from .common_helpers import HTTPBIN_URL, RequestRecorderClient, record_test


def test_collections_containing_null():
    record_test("collections-containing-null")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.collections.collections_containing_null(
        request={
            "required_array": ["foo", None],
            "required_map": {"foo": None, "bar": 123},
            "optional_array": ["foo", None],
            "optional_map": {"foo": None, "bar": 123},
            "array_of_null_union": ["foo", None],
            "map_of_null_union": {"foo": None, "bar": 123},
        }
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.object is not None
    assert res.object.json_ == {
        "requiredArray": ["foo", None],
        "requiredMap": {"foo": None, "bar": 123},
        "optionalArray": ["foo", None],
        "optionalMap": {"foo": None, "bar": 123},
        "arrayOfNullUnion": ["foo", None],
        "mapOfNullUnion": {"foo": None, "bar": 123},
    }


def test_collections_parameter_annotations_advertise_iterables():
    record_test("collections-parameter-annotations-advertise-iterables")

    request_annotation = inspect.signature(
        RequestBodies.form_request_body_string_array
    ).parameters["stuff"].annotation

    assert "Iterable" in str(request_annotation)
    assert (
        "List"
        in str(
            models.operations.CollectionsContainingNullNullishCollections.model_fields[
                "required_array"
            ].annotation
        )
    )


def test_collections_parameters_accept_iterables_direct():
    record_test("collections-parameters-accept-iterables-direct")

    http_client = RequestRecorderClient()
    s = SDK(server_url=HTTPBIN_URL, client=http_client)

    def stuff_generator():
        yield "foo"
        yield "bar"

    res = s.request_bodies.form_request_body_string_array(stuff=stuff_generator())

    assert res is not None
    assert http_client.log
    assert "foo" in http_client.log[0].request_body
    assert "bar" in http_client.log[0].request_body


def test_collections_parameters_accept_iterables():
    record_test("collections-parameters-accept-iterables")

    http_client = RequestRecorderClient()
    s = SDK(server_url=HTTPBIN_URL, client=http_client)

    def required_array_generator():
        yield "foo"
        yield None

    def optional_array_generator():
        yield "bar"
        yield None

    def array_of_null_union_generator():
        yield "baz"
        yield None

    res = s.collections.collections_containing_null(
        request={
            "required_array": required_array_generator(),
            "required_map": {"foo": None, "bar": 123},
            "optional_array": optional_array_generator(),
            "optional_map": {"foo": None, "bar": 123},
            "array_of_null_union": array_of_null_union_generator(),
            "map_of_null_union": {"foo": None, "bar": 123},
        }
    )

    assert res is not None
    assert http_client.log
    assert json.loads(http_client.log[0].request_body) == {
        "requiredArray": ["foo", None],
        "requiredMap": {"foo": None, "bar": 123},
        "optionalArray": ["bar", None],
        "optionalMap": {"foo": None, "bar": 123},
        "arrayOfNullUnion": ["baz", None],
        "mapOfNullUnion": {"foo": None, "bar": 123},
    }


def test_collections_map_input_accepts_non_dict_mapping():
    record_test("collections-map-input-accepts-non-dict-mapping")

    request_annotation = inspect.signature(Collections.map_input).parameters[
        "request"
    ].annotation

    assert "Mapping" in str(request_annotation)

    http_client = RequestRecorderClient()
    s = SDK(server_url=HTTPBIN_URL, client=http_client)

    res = s.collections.map_input(request=MappingProxyType({"a": "b"}))

    assert res is not None
    assert http_client.log
    assert json.loads(http_client.log[0].request_body) == {"a": "b"}
