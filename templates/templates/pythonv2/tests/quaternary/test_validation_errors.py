"""Tests for lenient response construction helpers."""

from functools import partial
from typing import Dict, List, Literal, Optional, Union

import httpx
import pytest
from pydantic import BeforeValidator
from typing_extensions import Annotated

from openapi.models import errors
from openapi.models.shared import (
    ConstDiscriminatedOneOf,
    ConstObject1,
    UnknownConstDiscriminatedOneOf,
)
from openapi.types import BaseModel
from openapi.utils.serializers import construct_unvalidated
from openapi.utils.unions import parse_open_union
from openapi.utils.unmarshal_json_response import unmarshal_json_response


class _Inner(BaseModel):
    label: Optional[str] = None


class _Parent(BaseModel):
    id: str
    inner: _Inner
    items: Optional[List[_Inner]] = None
    tags: Optional[Dict[str, _Inner]] = None


# --- construct_unvalidated ---------------------------------------------------


def test_valid_payload_takes_typed_fast_path():
    result = construct_unvalidated({"id": "x", "inner": {"label": "ok"}}, _Parent)
    assert isinstance(result, _Parent)
    assert result.id == "x"
    assert isinstance(result.inner, _Inner)
    assert result.inner.label == "ok"


def test_missing_required_scalar_defaults_to_none():
    result = construct_unvalidated({"inner": {"label": "y"}}, _Parent)
    assert isinstance(result, _Parent)
    assert result.id is None
    assert isinstance(result.inner, _Inner)


def test_missing_required_model_defaults_to_empty_typed_instance():
    result = construct_unvalidated({"id": "x"}, _Parent)
    assert isinstance(result, _Parent)
    assert isinstance(result.inner, _Inner)
    assert result.inner.label is None


def test_explicit_null_for_required_model_stays_none():
    # A present-but-null required model field cannot be coerced into a typed
    # instance; lenient construction preserves None instead of raising.
    result = construct_unvalidated({"inner": None}, _Parent)
    assert isinstance(result, _Parent)
    assert result.inner is None


def test_lenient_returns_scalar_as_is_for_model_field():
    # Structural mismatch (scalar where a model is expected) must not raise out of
    # lenient construction; the unconvertible value is returned as-is.
    result = construct_unvalidated({"id": "x", "inner": "str-not-model"}, _Parent)
    assert result.inner == "str-not-model"


def test_lenient_returns_scalar_as_is_for_list_field():
    result = construct_unvalidated({"id": "x", "inner": {}, "items": "str-not-list"}, _Parent)
    assert result.items == "str-not-list"


def test_lenient_returns_scalar_as_is_for_mapping_field():
    result = construct_unvalidated({"id": "x", "inner": {}, "tags": "str-not-dict"}, _Parent)
    assert result.tags == "str-not-dict"


def test_nested_list_of_models_is_recursively_typed():
    result = construct_unvalidated(
        {"id": "x", "inner": {}, "items": [{"label": "a"}, {"label": "b"}]},
        _Parent,
    )
    assert [type(i) for i in result.items] == [_Inner, _Inner]
    assert [i.label for i in result.items] == ["a", "b"]


def test_nested_mapping_of_models_is_recursively_typed():
    result = construct_unvalidated(
        {"id": "x", "inner": {}, "tags": {"k": {"label": "v"}}},
        _Parent,
    )
    assert isinstance(result.tags["k"], _Inner)
    assert result.tags["k"].label == "v"


def test_wrong_type_leaf_passes_through():
    result = construct_unvalidated({"id": 123, "inner": {}}, _Parent)
    assert result.id == 123


def test_unknown_extra_fields_do_not_break_construction():
    result = construct_unvalidated(
        {"id": "x", "inner": {"label": "z"}, "future_field": {"kept": True}},
        _Parent,
    )
    assert isinstance(result, _Parent)
    assert result.id == "x"
    assert result.inner.label == "z"


# --- unmarshal_json_response -------------------------------------------------


def _json_response(body: str) -> httpx.Response:
    return httpx.Response(
        200, content=body, headers={"content-type": "application/json"}
    )


def test_unmarshal_json_response_validate_false_uses_lenient_construction():
    result = unmarshal_json_response(
        _Parent, _json_response('{"id": 123}'), validate=False
    )
    assert isinstance(result, _Parent)
    assert result.id == 123
    assert isinstance(result.inner, _Inner)
    assert result.inner.label is None


def test_unmarshal_json_response_validate_true_still_raises_validation_error():
    with pytest.raises(errors.ResponseValidationError) as exc_info:
        unmarshal_json_response(_Parent, _json_response('{"id": 123}'))
    assert exc_info.value.raw_response.status_code == 200
    assert exc_info.value.cause is not None


def test_unmarshal_json_response_non_json_body_raises_default_error():
    # A body that is not parseable JSON has no best-effort typed representation,
    # so even lenient construction (validate=False) surfaces the default
    # ResponseValidationError rather than leaking a raw JSON decode error.
    with pytest.raises(errors.ResponseValidationError) as exc_info:
        unmarshal_json_response(
            _Parent, _json_response("not json at all"), validate=False
        )
    assert exc_info.value.raw_response.status_code == 200
    assert exc_info.value.cause is not None


# --- parse_open_union --------------------------------------------------------


class _Args(BaseModel):
    queries: Optional[List[str]] = None


class _SearchStep(BaseModel):
    type: str
    id: str
    arguments: _Args


class _UnknownStep(BaseModel):
    raw: dict
    is_unknown: bool = True

    def __init__(self, **kwargs):
        super().__init__(raw=kwargs.get("raw"), is_unknown=True)


_VARIANTS = {"search": _SearchStep}


def _dispatch(value, *, lenient):
    return parse_open_union(
        value,
        disc_key="type",
        variants=_VARIANTS,
        unknown_cls=_UnknownStep,
        union_name="_Step",
        lenient=lenient,
    )


def test_open_union_valid_payload_is_typed_variant():
    result = _dispatch(
        {"type": "search", "id": "c1", "arguments": {"queries": ["q"]}},
        lenient=True,
    )
    assert isinstance(result, _SearchStep)
    assert result.arguments.queries == ["q"]


def test_open_union_known_variant_missing_required_stays_typed_when_lenient():
    result = _dispatch({"type": "search", "id": "c1"}, lenient=True)
    assert isinstance(result, _SearchStep)
    assert result.id == "c1"
    assert isinstance(result.arguments, _Args)
    assert result.arguments.queries is None


def test_open_union_known_variant_missing_required_is_unknown_when_strict():
    result = _dispatch({"type": "search", "id": "c1"}, lenient=False)
    assert isinstance(result, _UnknownStep)
    assert result.is_unknown is True


def test_open_union_incompatible_shape_stays_typed_when_lenient():
    # Known discriminator with a field the SDK can't coerce: lenient construction
    # keeps the typed variant and leaves the unconvertible field as-is (recursive
    # leniency), rather than discarding the whole value to Unknown.
    result = _dispatch(
        {"type": "search", "arguments": "not-an-object"}, lenient=True
    )
    assert isinstance(result, _SearchStep)
    assert result.arguments == "not-an-object"
    assert result.id is None


def test_open_union_incompatible_shape_falls_back_to_unknown_when_strict():
    result = _dispatch(
        {"type": "search", "arguments": "not-an-object"}, lenient=False
    )
    assert isinstance(result, _UnknownStep)
    assert result.raw == {"type": "search", "arguments": "not-an-object"}


def test_open_union_unknown_discriminator_is_unknown():
    result = _dispatch({"type": "future-step", "id": "c1"}, lenient=True)
    assert isinstance(result, _UnknownStep)
    assert result.raw["type"] == "future-step"


def test_open_union_already_constructed_model_passes_through():
    step = _SearchStep.model_construct(type="search", id="c1", arguments=_Args())
    assert _dispatch(step, lenient=True) is step


def test_open_union_missing_discriminator_raises():
    with pytest.raises(ValueError):
        _dispatch({"id": "c1"}, lenient=True)


class _ConstEnvelope(BaseModel):
    id: str  # required; absent so the parent falls to lenient reconstruction
    choice: ConstDiscriminatedOneOf  # actual generator-emitted discriminated union


def test_generated_open_union_unknown_discriminator_preserved():
    result = construct_unvalidated({"choice": {"tag": "tag99"}}, _ConstEnvelope)
    assert result.id is None
    assert isinstance(result.choice, UnknownConstDiscriminatedOneOf)
    assert result.choice.raw == {"tag": "tag99"}


def test_generated_open_union_known_discriminator_typed():
    result = construct_unvalidated(
        {"choice": {"tag": "tag1", "imageURL": "u"}}, _ConstEnvelope
    )
    assert isinstance(result.choice, ConstObject1)
    assert result.choice.image_url == "u"


# --- lenient union resolution respects declaration order ---------------------
# When a response payload for an open discriminated union does not contain the
# discriminator field `parse_open_union` raises by design so pydantic can try sibling
# branches. The value then falls through to `_construct_lenient` on the bare union.
# Resolution must follow declaration order (first variant via lenient construct), not
# prefer a later variant that happens to validate against the payload with its defaults.


class _LeadEvent(BaseModel):
    """First-declared variant with a required field, so it only validates a
    conforming payload. Lenient construct must still resolve here by order."""
    type: Literal["lead"] = "lead"
    payload: str  # required


class _FallbackEvent(BaseModel):
    """Later variant whose fields are all optional, so it validates against any
    payload via defaults. Must NOT be preferred over the first-declared variant."""
    type: Literal["fallback"] = "fallback"
    detail: Optional[str] = None


_CATCHALL_VARIANTS = {"lead": _LeadEvent, "fallback": _FallbackEvent}

_OpenEvent = Annotated[
    Union[_LeadEvent, _FallbackEvent, _UnknownStep],
    BeforeValidator(
        partial(
            parse_open_union,
            disc_key="type",
            variants=_CATCHALL_VARIANTS,
            unknown_cls=_UnknownStep,
            union_name="_OpenEvent",
            lenient=True,
        )
    ),
]


class _EventEnvelope(BaseModel):
    event: _OpenEvent


def test_lenient_union_no_discriminator_yields_first_declared_not_catchall():
    result = construct_unvalidated(
        {"event": {"totally": "unconforming"}}, _EventEnvelope
    )
    assert isinstance(result.event, _LeadEvent)


def test_lenient_union_no_discriminator_ignores_later_exact_match():
    # Order wins over best-match: payload fully validates as the later
    # _FallbackEvent, yet still resolves to the first-declared _LeadEvent.
    result = construct_unvalidated(
        {"event": {"detail": "looks exactly like fallback"}}, _EventEnvelope
    )
    assert isinstance(result.event, _LeadEvent)


# --- plain (non-open) union field regression guards --------------------------
# These exercise the bare `is_union` branch of `_construct_lenient`, reached only
# after the strict unmarshal in `construct_unvalidated` has failed.

class _ChoiceA(BaseModel):
    a: str


class _ChoiceB(BaseModel):
    b: str


class _ParentUnion(BaseModel):
    id: str  # required; absent here so the parent falls to lenient construction
    choice: Union[_ChoiceA, _ChoiceB]


def test_lenient_union_field_strict_matches_later_variant():
    # Parent union deserialization fails (missing `id`), so `choice` is reconstructed
    # leniently. Value strictly matches the later _ChoiceB so strict-first recursion in
    # construct_unvalidated must still select it, not force-construct first _ChoiceA.
    result = construct_unvalidated({"choice": {"b": "x"}}, _ParentUnion)
    assert isinstance(result.choice, _ChoiceB)
    assert result.choice.b == "x"
    assert result.id is None


class _ListUnionParent(BaseModel):
    id: str  # required; absent so the list field is reconstructed leniently
    items: List[Union[_ChoiceA, _ChoiceB]]


def test_lenient_union_nested_in_list_resolves_per_item_by_order():
    # None of the items match any variant strictly, so the list branch recurses and
    # the union resolves to the first-declared _ChoiceA, with defaults filled in.
    result = construct_unvalidated(
        {"items": [{"unknown": "z"}, {"unknown": "y"}]}, _ListUnionParent
    )
    assert [type(i) for i in result.items] == [_ChoiceA, _ChoiceA]
    assert result.items[0].a is None


class _DictUnionParent(BaseModel):
    id: str
    tags: Dict[str, Union[_ChoiceA, _ChoiceB]]


def test_lenient_union_nested_in_mapping_resolves_per_value_by_order():
    result = construct_unvalidated(
        {"tags": {"k": {"unknown": "z"}}}, _DictUnionParent
    )
    assert isinstance(result.tags["k"], _ChoiceA)
    assert result.tags["k"].a is None
