from dataclasses import dataclass
from typing import Annotated, Optional, Union

import pytest
from pydantic import ValidationError

from openapi.types import BaseModel
from openapi.utils import marshal_json, serializers, unmarshal


@pytest.fixture
def create_model_calls(monkeypatch):
    calls = []
    original = serializers.create_model

    def counting_create_model(*args, **kwargs):
        calls.append(args[0])
        return original(*args, **kwargs)

    monkeypatch.setattr(serializers, "create_model", counting_create_model)
    return calls


def test_unmarshal_builds_wrapper_model_once_per_type(create_model_calls):
    class Widget(BaseModel):
        name: str

    first = unmarshal({"name": "a"}, Widget)
    second = unmarshal({"name": "b"}, Widget)

    assert first == Widget(name="a")
    assert second == Widget(name="b")
    assert create_model_calls == ["Unmarshaller"]


def test_marshal_json_builds_wrapper_model_once_per_type(create_model_calls):
    class Widget(BaseModel):
        name: str

    assert marshal_json(Widget(name="a"), Widget) == '{"name":"a"}'
    assert marshal_json(Widget(name="b"), Widget) == '{"name":"b"}'
    assert create_model_calls == ["Marshaller"]


def test_distinct_types_build_distinct_wrapper_models(create_model_calls):
    class Widget(BaseModel):
        name: str

    class Gadget(BaseModel):
        size: int

    assert unmarshal({"name": "a"}, Widget) == Widget(name="a")
    assert unmarshal({"size": 1}, Gadget) == Gadget(size=1)
    assert unmarshal(None, Optional[Widget]) is None
    assert create_model_calls == ["Unmarshaller"] * 3


def test_union_member_order_is_preserved_per_call():
    class Widget(BaseModel):
        name: str

    class Gadget(BaseModel):
        name: str

    # Both members accept the payload, so the declared order decides the result.
    # Unions with the same members compare equal regardless of order, so each
    # ordering must still resolve to its own first member.
    assert isinstance(unmarshal({"name": "a"}, Union[Widget, Gadget]), Widget)
    assert isinstance(unmarshal({"name": "a"}, Union[Gadget, Widget]), Gadget)


def test_unhashable_type_falls_back_to_uncached_model(create_model_calls):
    @dataclass
    class Marker:
        value: str

    unhashable = Annotated[int, Marker("a")]

    with pytest.raises(TypeError):
        hash(unhashable)

    assert unmarshal(1, unhashable) == 1
    assert unmarshal(2, unhashable) == 2
    assert marshal_json(3, unhashable) == "3"
    assert create_model_calls == ["Unmarshaller", "Unmarshaller", "Marshaller"]


def test_validation_errors_still_name_wrapper_model():
    class Widget(BaseModel):
        name: str

    with pytest.raises(ValidationError, match="for Unmarshaller"):
        unmarshal({}, Widget)
    with pytest.raises(ValidationError, match="for Unmarshaller"):
        unmarshal({}, Widget)
