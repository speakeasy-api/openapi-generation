"""Regression test: model_serializer dropping aliased fields.

The serialize_model method used serialized.get(k) where k is the alias
(e.g. "isArchived"). But pydantic's handler returns field names
(e.g. "is_archived") by default, so aliased optional fields were silently
dropped from model_dump().
"""

import pydantic
from pydantic import model_serializer
from typing import Optional
from typing_extensions import Annotated

from openapi.types import BaseModel, UNSET_SENTINEL


class _FakePreference(BaseModel):
    """Mimics the generated serialization pattern for a model with aliased fields."""

    enabled: Optional[bool] = True

    is_archived: Annotated[Optional[bool], pydantic.Field(alias="isArchived")] = False

    display_name: Annotated[Optional[str], pydantic.Field(alias="displayName")] = None

    @model_serializer(mode="wrap")
    def serialize_model(self, handler):
        optional_fields = set(["enabled", "isArchived", "displayName"])
        serialized = handler(self)
        m = {}

        for n, f in type(self).model_fields.items():
            k = f.alias or n
            val = serialized.get(k, serialized.get(n))

            if val != UNSET_SENTINEL:
                if val is not None or k not in optional_fields:
                    m[k] = val

        return m


try:
    _FakePreference.model_rebuild()
except NameError:
    pass


def test_model_dump_includes_aliased_fields():
    """model_dump() without by_alias must still include aliased fields."""

    model = _FakePreference(enabled=True, is_archived=False)
    dumped = model.model_dump()

    assert "isArchived" in dumped, (
        f"Aliased field 'isArchived' missing from model_dump(). Keys: {list(dumped.keys())}"
    )
    assert dumped["isArchived"] is False

    assert "enabled" in dumped
    assert dumped["enabled"] is True


def test_model_dump_by_alias_includes_aliased_fields():
    """model_dump(by_alias=True) must also include aliased fields."""

    model = _FakePreference(enabled=True, is_archived=True)
    dumped = model.model_dump(by_alias=True)

    assert "isArchived" in dumped, (
        f"Aliased field 'isArchived' missing from model_dump(by_alias=True). Keys: {list(dumped.keys())}"
    )
    assert dumped["isArchived"] is True


def test_model_dump_optional_none_aliased_field_omitted():
    """Optional aliased fields with None value should be omitted."""

    model = _FakePreference(enabled=True, is_archived=False, display_name=None)
    dumped = model.model_dump()

    # display_name is None and optional, so it should be omitted
    assert "displayName" not in dumped
    # is_archived is False (not None), so it must be present
    assert "isArchived" in dumped
