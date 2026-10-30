import json
import warnings

import pydantic
import pytest

from openapi import SDK, AsyncSDK, utils
from openapi.models import shared
from openapi.types import UnrecognizedInt, UnrecognizedStr
from openapi.utils import ALLOW_UNKNOWN_UNION_VARIANTS
from .common_helpers import HTTPBIN_URL, record_test
from .unknown_fields import drop_unexpected_fields
from .unknown_fields_cases import UNKNOWN_FIELD_CASES


@pytest.mark.parametrize("child", [
    {"displayName": "child", "future": True},
    {"displayName": "child", "display_name": "unknown"},
    {"displayName": "child", "child": {"displayName": "grandchild", "future": {"nested": [None, 0]}}},
    {"displayName": "child", "nullable": None, "future": None},
])
def test_flattened_request_keeps_nested_unknown_fields(child):
    record_test("unknown-fields-json-round-trip")
    with SDK(server_url=HTTPBIN_URL) as sdk:
        response = sdk.echo_unknown_fields(label="root", child=child)
    payload = {"label": "root", "child": child}
    assert response.raw_response.json()["json"] == payload
    assert response.object.json_.to_dict() == payload
    assert json.loads(response.object.json_.to_json()) == payload


@pytest.mark.asyncio()
async def test_flattened_request_keeps_nested_unknown_fields_async():
    child = {"displayName": "child", "child": {"displayName": "grandchild", "future": {"nested": [None, 0]}}}
    async with AsyncSDK(server_url=HTTPBIN_URL) as sdk:
        response = await sdk.echo_unknown_fields(label="root", child=child)
    payload = {"label": "root", "child": child}
    assert response.raw_response.json()["json"] == payload
    assert response.object.json_.to_dict() == payload
    assert json.loads(response.object.json_.to_json()) == payload


def test_flattened_request_keeps_unknown_fields_on_nested_models():
    child = shared.FieldNode(display_name="child", future=None)
    with SDK(server_url=HTTPBIN_URL) as sdk:
        response = sdk.echo_flattened_unknown_fields(label="root", child=child)
    assert response.raw_response.json()["json"] == {"label": "root", "child": {"displayName": "child", "future": None}}


def test_flattened_request_has_no_argument_for_top_level_unknown_fields():
    with SDK(server_url=HTTPBIN_URL) as sdk:
        with pytest.raises(TypeError):
            sdk.echo_flattened_unknown_fields(label="root", future=True)  # type: ignore[call-arg]


@pytest.mark.parametrize("fields", [fields for fields, _ in UNKNOWN_FIELD_CASES])
def test_to_dict_round_trips_unknown_fields(fields):
    payload = {"label": "root", **fields}
    model = shared.FieldEnvelopeInput.model_validate(payload)
    assert model.to_dict(mode="json") == payload
    assert json.loads(model.to_json()) == payload
    assert shared.FieldEnvelopeInput.model_validate(model.to_dict()).to_dict(mode="json") == payload


@pytest.mark.parametrize("options, expected", [
    ({}, {"displayName": "child", "nullable": None, "future": None}),
    ({"use_api_names": False}, {"display_name": "child", "nullable": None, "future": None}),
    ({"exclude_unset": False}, {"displayName": "child", "nullable": None, "future": None}),
    ({"exclude_defaults": True}, {"displayName": "child", "nullable": None, "future": None}),
    ({"exclude_none": True}, {"displayName": "child"}),
])
def test_model_serialization_methods(options, expected):
    model = shared.FieldNode.model_validate({"displayName": "child", "nullable": None, "future": None})
    assert model.to_dict(**options) == expected
    assert model.to_dict(mode="json", **options) == expected
    assert json.loads(model.to_json(**options)) == expected


@pytest.mark.parametrize("options", [{"exclude_none": True}, {"exclude_defaults": True}])
def test_excluded_field_keeps_unknown_field_named_like_it(options):
    model = shared.FieldEnvelopeInput.model_validate({"label": "root", "plainChoice": None, "plain_choice": "unknown"})
    assert model.to_dict(**options) == {"label": "root", "plain_choice": "unknown"}
    assert model.model_dump(by_alias=True, **options) == {"label": "root", "plain_choice": "unknown"}


def test_drop_unexpected_fields_ignores_only_undeclared_fields():
    expected = shared.TypedFields(label="typed")
    assert drop_unexpected_fields(shared.TypedFields.model_validate({"label": "typed", "future": "value"}), expected) != expected
    assert drop_unexpected_fields(shared.FieldNode.model_validate({"displayName": "child", "future": True}), shared.FieldNode(display_name="child")) == shared.FieldNode(display_name="child")


def test_drop_unexpected_fields_in_typed_additional_properties():
    expected = shared.TypedNodes.model_validate({"label": "nodes", "one": {"displayName": "child"}})
    nested_unknown = shared.TypedNodes.model_validate({"label": "nodes", "one": {"displayName": "child", "future": True}})
    extra_entry = shared.TypedNodes.model_validate({"label": "nodes", "one": {"displayName": "child"}, "two": {"displayName": "other"}})
    changed_value = shared.TypedNodes.model_validate({"label": "nodes", "one": {"displayName": "changed"}})
    assert drop_unexpected_fields(nested_unknown, expected) == expected
    assert drop_unexpected_fields(extra_entry, expected) != expected
    assert drop_unexpected_fields(changed_value, expected) != expected


def test_model_serialization_methods_field_names():
    model = shared.FieldEnvelopeInput.model_validate({"label": "root", "child": {"displayName": "child", "child": {"displayName": "grandchild"}}, "futureKind": "third", "plainChoice": {"text": "t"}, "future": True})
    api_names = {"label": "root", "child": {"displayName": "child", "child": {"displayName": "grandchild"}}, "futureKind": "third", "plainChoice": {"text": "t"}, "future": True}
    field_names = {"label": "root", "child": {"display_name": "child", "child": {"display_name": "grandchild"}}, "future_kind": "third", "plain_choice": {"text": "t"}, "future": True}
    assert model.model_dump(exclude_unset=True) == api_names
    assert model.to_dict() == api_names
    assert model.to_dict(use_api_names=False) == field_names
    assert json.loads(model.to_json(use_api_names=False)) == field_names
    assert shared.FieldEnvelopeInput.model_validate(model.to_dict(use_api_names=False)).to_dict() == api_names


@pytest.mark.parametrize("typ, value", [(UnrecognizedStr, "future"), (UnrecognizedInt, 3)])
def test_unrecognized_enum_serialization(typ, value):
    adapter = pydantic.TypeAdapter(typ)
    enum_value = adapter.validate_python(value)
    assert isinstance(enum_value, typ)
    with warnings.catch_warnings():
        warnings.simplefilter("error", UserWarning)
        assert adapter.dump_python(enum_value) is enum_value
        assert adapter.dump_python(enum_value, mode="json") == value
        assert json.loads(adapter.dump_json(enum_value)) == value
    with pytest.raises(pydantic.ValidationError):
        adapter.validate_python(value, strict=True)


def test_unrecognized_enum_generated_model():
    model = shared.FieldEnvelope.model_validate({"label": "root", "futureKind": "third", "futureCode": 3})
    with warnings.catch_warnings():
        warnings.simplefilter("error", UserWarning)
        dumped = model.model_dump(by_alias=True)
        assert type(dumped["futureKind"]) is UnrecognizedStr
        assert type(dumped["futureCode"]) is UnrecognizedInt
        helper_dumped = model.to_dict()
        assert type(helper_dumped["futureKind"]) is str
        assert type(helper_dumped["futureCode"]) is int
        assert helper_dumped == {"label": "root", "futureKind": "third", "futureCode": 3}
        assert json.loads(model.to_json()) == helper_dumped
        assert json.loads(utils.marshal_json(model, shared.FieldEnvelope)) == {"label": "root", "futureKind": "third", "futureCode": 3}


def test_flattened_parameter_named_like_helper():
    with SDK(server_url=HTTPBIN_URL) as sdk:
        response = sdk.to_json.echo_grouped_fields(to_json="j")
    assert response.raw_response.request.url.params["toJson"] == "j"


def test_serialization_methods_return_raw_payload_of_nested_unknown_variants():
    choice = {"kind": "future", "weight": 9000}
    payload = {"label": "root", "choice": choice, "choices": [{"kind": "text", "text": "t", "future": True}, choice], "future": True}
    model = shared.FieldEnvelope.model_validate(payload, context={ALLOW_UNKNOWN_UNION_VARIANTS: True})
    assert model.choice.is_unknown and model.choices[1].is_unknown
    assert model.to_dict() == model.to_dict(mode="json") == payload
    assert json.loads(model.to_json()) == payload
    assert model.model_dump(by_alias=True)["choice"] == choice
    assert shared.FieldEnvelope.model_validate(model.to_dict(), context={ALLOW_UNKNOWN_UNION_VARIANTS: True}).to_dict() == payload
    assert shared.FieldEnvelope.model_validate(model.to_dict()).to_dict() == payload


@pytest.mark.parametrize("model, options, expected", [
    (shared.FieldNode(display_name="child", nullable="value"), {"include": {"nullable"}}, {"nullable": "value"}),
    (shared.FieldNode(display_name="child", nullable="value"), {"exclude": {"display_name"}}, {"nullable": "value"}),
    (shared.ConstFields(label="label"), {"include": {"label"}}, {"label": "label"}),
    (shared.ConstFields(label="label"), {"exclude": {"kind"}}, {"label": "label", "priority": 1}),
    (shared.ConstFields(label="label"), {"exclude_defaults": True}, {"label": "label"}),
    (shared.DefaultsAndConsts(normal_field="x", default_str_nullable="value"), {"include": {"default_str_nullable"}}, {"defaultStrNullable": "value"}),
    (shared.DefaultsAndConsts(normal_field="x", default_str_nullable="value"), {"include": {"normal_field"}}, {"normalField": "x"}),
])
def test_exclude_none_with_filters_leaves_filtered_fields_out(model, options, expected):
    assert model.model_dump(by_alias=True, exclude_none=True, **options) == expected
    assert json.loads(model.model_dump_json(by_alias=True, exclude_none=True, **options)) == expected


@pytest.mark.parametrize("model, options, excluded", [
    (shared.DefaultsAndConsts(normal_field="x", default_str_nullable="value"), {"exclude": {"normal_field", "const_str", "default_str_nullable"}}, {"normalField", "constStr", "defaultStrNullable"}),
    (shared.DefaultsAndConsts(normal_field="x"), {"exclude_unset": True}, {"defaultStrNullable"}),
    (shared.DefaultsAndConsts(normal_field="x"), {"exclude_defaults": True}, {"constStr", "defaultStrNullable"}),
])
def test_exclude_none_with_filters_leaves_filtered_null_default_out(model, options, excluded):
    dumped = model.model_dump(by_alias=True, exclude_none=True, **options)
    assert not excluded & dumped.keys()
    for mode in ("python", "json"):
        expected = {k: v for k, v in model.model_dump(mode=mode, by_alias=True, **options).items() if v is not None}
        assert model.model_dump(mode=mode, by_alias=True, exclude_none=True, **options) == expected
    assert json.loads(model.model_dump_json(by_alias=True, exclude_none=True, **options)) == expected


@pytest.mark.parametrize("options, expected", [
    ({"include": {"child": {"nullable"}, "children": {"__all__": {"nullable"}}, "constant": {"label"}}}, {"child": {"nullable": "value"}, "children": [{"nullable": "value"}], "constant": {"label": "label"}}),
    ({"exclude": {"child": {"display_name"}, "children": {"__all__": {"display_name"}}, "constant": {"kind"}}}, {"label": "root", "child": {"nullable": "value"}, "children": [{"nullable": "value"}], "constant": {"label": "label", "priority": 1}}),
    ({"exclude_defaults": True}, {"label": "root", "child": {"displayName": "child", "nullable": "value"}, "children": [{"displayName": "child", "nullable": "value"}], "constant": {"label": "label"}}),
])
def test_exclude_none_with_filters_leaves_nested_filtered_fields_out(options, expected):
    node = shared.FieldNode(display_name="child", nullable="value")
    model = shared.FieldEnvelope(label="root", child=node, children=[node], constant=shared.ConstFields(label="label"))
    assert model.model_dump(by_alias=True, exclude_none=True, **options) == expected
    assert json.loads(model.model_dump_json(by_alias=True, exclude_none=True, **options)) == expected


def test_exclude_none_without_filters_keeps_explicit_nulls():
    model = shared.FieldEnvelope(label="root", child=shared.FieldNode(display_name="child", nullable=None), constant=shared.ConstFields())
    expected = {"label": "root", "child": {"displayName": "child", "nullable": None}, "constant": {"kind": "fixed", "priority": 1}}
    assert model.model_dump(by_alias=True, exclude_none=True) == expected
    assert json.loads(utils.marshal_json(model, shared.FieldEnvelope)) == expected
