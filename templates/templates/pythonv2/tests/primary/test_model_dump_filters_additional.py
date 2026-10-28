import json
from typing import List, Optional

import pytest
from pydantic import TypeAdapter, model_serializer

from openapi import utils
from openapi.models import errors, shared
from openapi.types import Nullable


def filtered_fields(display_name="name", **kwargs):
    return shared.FilteredFields(
        display_name=display_name, required_nullable=None, **kwargs
    )


def filtered_values(**kwargs):
    return shared.FilteredFields(
        display_name="name", required_nullable="value", **kwargs
    )


ALL_VALUES = {
    "displayName": "name",
    "requiredNullable": "value",
    "kind": "filtered",
    "nullDefault": "value",
    "withDefault": "fallback",
}


def all_values_except(key):
    return {k: v for k, v in ALL_VALUES.items() if k != key}


class TestModelDumpFilters:
    def test_exclude_leaves_required_field_out(self):
        model = filtered_fields()

        assert "displayName" not in model.model_dump(exclude={"display_name"})
        assert "displayName" not in model.model_dump(
            by_alias=True, exclude={"display_name"}
        )
        assert "displayName" not in json.loads(
            model.model_dump_json(exclude={"display_name"})
        )

    def test_include_leaves_other_fields_out(self):
        model = filtered_fields(optional_nullable="x")

        assert model.model_dump(include={"optional_nullable"}) == {
            "optionalNullable": "x"
        }
        assert model.model_dump(by_alias=True, include={"display_name"}) == {
            "displayName": "name"
        }

    def test_exclude_unset_leaves_unset_fields_out(self):
        model = filtered_fields()

        assert model.model_dump(exclude_unset=True) == {
            "displayName": "name",
            "requiredNullable": None,
        }

    def test_exclude_defaults_leaves_default_fields_out(self):
        model = filtered_fields(with_default="other")

        assert model.model_dump(exclude_defaults=True) == {
            "displayName": "name",
            "requiredNullable": None,
            "withDefault": "other",
        }

    def test_nested_exclude(self):
        model = filtered_fields(
            child=filtered_fields("child"), children=[filtered_fields("item")]
        )

        dumped = model.model_dump(
            exclude={
                "child": {"display_name"},
                "children": {"__all__": {"display_name"}},
            }
        )

        assert dumped["displayName"] == "name"
        assert "displayName" not in dumped["child"]
        assert "displayName" not in dumped["children"][0]


class TestExcludeNoneWithFilters:
    @pytest.mark.parametrize(
        "kwargs, options, expected",
        [
            (
                {"null_default": "value"},
                {"include": {"with_default"}},
                {"withDefault": "fallback"},
            ),
            (
                {"null_default": "value"},
                {"include": {"display_name", "kind", "null_default"}},
                {"displayName": "name", "kind": "filtered", "nullDefault": "value"},
            ),
            (
                {"null_default": "value"},
                {"exclude": {"display_name"}},
                all_values_except("displayName"),
            ),
            (
                {"null_default": "value"},
                {"exclude": {"kind"}},
                all_values_except("kind"),
            ),
            (
                {"null_default": "value"},
                {"exclude": {"null_default"}},
                all_values_except("nullDefault"),
            ),
            (
                {"null_default": "value"},
                {"exclude_unset": True},
                {
                    "displayName": "name",
                    "requiredNullable": "value",
                    "nullDefault": "value",
                },
            ),
            (
                {},
                {"exclude_unset": True},
                {"displayName": "name", "requiredNullable": "value"},
            ),
            (
                {"null_default": "value"},
                {"exclude_defaults": True},
                {
                    "displayName": "name",
                    "requiredNullable": "value",
                    "nullDefault": "value",
                },
            ),
            (
                {},
                {"exclude_defaults": True},
                {"displayName": "name", "requiredNullable": "value"},
            ),
        ],
    )
    def test_filtered_fields_stay_out(self, kwargs, options, expected):
        model = filtered_values(**kwargs)

        assert model.model_dump(exclude_none=True, **options) == expected
        assert (
            json.loads(model.model_dump_json(exclude_none=True, **options)) == expected
        )

    @pytest.mark.parametrize(
        "kwargs, options, expected",
        [
            (
                {"null_default": "value"},
                {
                    "include": {
                        "child": {"with_default"},
                        "children": {"__all__": {"with_default"}},
                    }
                },
                {
                    "child": {"withDefault": "fallback"},
                    "children": [{"withDefault": "fallback"}],
                },
            ),
            (
                {"null_default": "value"},
                {
                    "exclude": {
                        "child": {"display_name", "kind", "null_default"},
                        "children": {
                            "__all__": {"display_name", "kind", "null_default"}
                        },
                    }
                },
                {
                    **ALL_VALUES,
                    "child": {"requiredNullable": "value", "withDefault": "fallback"},
                    "children": [
                        {"requiredNullable": "value", "withDefault": "fallback"}
                    ],
                },
            ),
            (
                {},
                {"exclude_unset": True},
                {
                    "displayName": "name",
                    "requiredNullable": "value",
                    "child": {"displayName": "name", "requiredNullable": "value"},
                    "children": [{"displayName": "name", "requiredNullable": "value"}],
                },
            ),
            (
                {},
                {"exclude_defaults": True},
                {
                    "displayName": "name",
                    "requiredNullable": "value",
                    "child": {"displayName": "name", "requiredNullable": "value"},
                    "children": [{"displayName": "name", "requiredNullable": "value"}],
                },
            ),
        ],
    )
    def test_nested_filtered_fields_stay_out(self, kwargs, options, expected):
        model = filtered_values(
            child=filtered_values(**kwargs),
            children=[filtered_values(**kwargs)],
            **kwargs,
        )

        assert model.model_dump(exclude_none=True, **options) == expected
        assert (
            json.loads(model.model_dump_json(exclude_none=True, **options)) == expected
        )

    def test_filter_applies_to_its_own_level(self):
        model = filtered_fields(
            child=filtered_fields("child"), children=[filtered_fields("item")]
        )

        dumped = model.model_dump(by_alias=True, exclude_none=True, exclude={"kind"})

        assert "kind" not in dumped
        assert "requiredNullable" not in dumped
        assert dumped["child"] == model.child.model_dump(
            by_alias=True, exclude_none=True
        )
        assert dumped["child"]["requiredNullable"] is None
        assert dumped["children"] == [
            model.children[0].model_dump(by_alias=True, exclude_none=True)
        ]
        assert model.model_dump(by_alias=True, exclude_none=True)[
            "requiredNullable"
        ] is None

    def test_empty_exclude_is_not_a_filter(self):
        model = filtered_fields(optional_nullable=None)

        assert model.model_dump(
            by_alias=True, exclude_none=True, exclude=set()
        ) == model.model_dump(by_alias=True, exclude_none=True)

    def test_without_filters_keeps_explicit_nulls(self):
        model = filtered_fields(optional_nullable=None)

        dumped = model.model_dump(by_alias=True, exclude_none=True)

        assert dumped == json.loads(utils.marshal_json(model, shared.FilteredFields))
        assert shared.FilteredFields.model_validate(dumped) == model


class TestNestedDefaultDumps:
    def test_nested_models_keep_api_names(self):
        model = filtered_fields(
            child=filtered_fields("child"), children=[filtered_fields("item")]
        )

        for dumped in (model.model_dump(), json.loads(model.model_dump_json())):
            assert dumped["child"]["displayName"] == "child"
            assert dumped["child"]["requiredNullable"] is None
            assert dumped["children"][0]["displayName"] == "item"

    def test_model_lists_keep_api_names(self):
        adapter = TypeAdapter(List[shared.FilteredFields])

        dumped = adapter.dump_python([filtered_fields("first")])

        assert dumped[0]["displayName"] == "first"
        assert dumped[0]["requiredNullable"] is None


class TestRequestBodies:
    def test_explicit_nulls(self):
        model = filtered_fields(optional_nullable=None)

        assert (
            utils.marshal_json(model, shared.FilteredFields)
            == '{"displayName":"name","requiredNullable":null,"kind":"filtered","nullDefault":null,"optionalNullable":null,"withDefault":"fallback"}'
        )

    def test_required_nullable_null(self):
        model = shared.PaginationCursorNullableLimitNextCursor(
            next_cursor=None, results=["a"]
        )

        assert (
            utils.marshal_json(model, shared.PaginationCursorNullableLimitNextCursor)
            == '{"next_cursor":null,"results":["a"]}'
        )

    def test_nested_nulls_and_api_names(self):
        model = filtered_fields(
            child=filtered_fields("child", optional_nullable=None),
            children=[filtered_fields("item")],
        )

        assert json.loads(utils.marshal_json(model, shared.FilteredFields)) == {
            "displayName": "name",
            "requiredNullable": None,
            "child": {
                "displayName": "child",
                "requiredNullable": None,
                "kind": "filtered",
                "nullDefault": None,
                "optionalNullable": None,
                "withDefault": "fallback",
            },
            "children": [
                {
                    "displayName": "item",
                    "requiredNullable": None,
                    "kind": "filtered",
                    "nullDefault": None,
                    "withDefault": "fallback",
                }
            ],
            "kind": "filtered",
            "nullDefault": None,
            "withDefault": "fallback",
        }

    def test_null_additional_property_omitted(self):
        model = shared.ObjWithAnyAdditionalProperties.model_validate(
            {"normalField": "x", "future": None}
        )

        assert (
            utils.marshal_json(model, shared.ObjWithAnyAdditionalProperties)
            == '{"normalField":"x"}'
        )

    def test_additional_property_under_python_name(self):
        model = shared.ObjWithAnyAdditionalProperties.model_validate(
            {"normalField": None, "normal_field": "extra"}
        )
        other = shared.FilteredFieldsWithAdditionalProperties.model_validate(
            {"displayName": None, "display_name": "extra"}
        )

        assert (
            utils.marshal_json(model, shared.ObjWithAnyAdditionalProperties)
            == '{"normalField":"extra"}'
        )
        assert (
            utils.marshal_json(other, shared.FilteredFieldsWithAdditionalProperties)
            == '{"displayName":"extra"}'
        )

    def test_top_level_none(self):
        assert utils.marshal_json(None, Optional[shared.FilteredFields]) == ""
        assert utils.marshal_json(None, Nullable[shared.FilteredFields]) == "null"


class TestErrorData:
    def test_plain_filtering(self):
        data = errors.FilteredFieldsErrorData()

        assert data.model_dump() == {"detail": None}
        assert data.model_dump(exclude_none=True) == {}
        assert data.model_dump(exclude_unset=True) == {}

    def test_nested_model_honours_exclude(self):
        data = errors.FilteredFieldsErrorData(detail=filtered_fields())

        dumped = data.model_dump(exclude={"detail": {"display_name"}})

        assert "displayName" not in dumped["detail"]
        assert dumped["detail"]["requiredNullable"] is None


class TestEmptyAlias:
    def test_value_kept_under_empty_api_name(self):
        model = shared.BinaryString(file_name="a", content=b"x")

        assert model.model_dump(by_alias=True)[""] == b"x"
        assert model.model_dump()[""] == b"x"
        assert "content" not in model.model_dump(by_alias=True)
        assert "" not in model.model_dump(by_alias=True, exclude={"content"})


class TestSerializerOverrides:
    def test_decorated_override_calls_parent(self):
        class Custom(shared.FilteredFields):
            @model_serializer(mode="wrap")
            def serialize_model(self, handler):
                return {**super().serialize_model(handler), "custom": True}

        dumped = Custom(display_name="name", required_nullable=None).model_dump()

        assert dumped["displayName"] == "name"
        assert dumped["requiredNullable"] is None
        assert dumped["custom"] is True

    def test_undecorated_override_keeps_filters(self):
        class Custom(shared.FilteredFields):
            def serialize_model(self, handler):
                return {**super().serialize_model(handler), "custom": True}

        dumped = Custom(display_name="name", required_nullable=None).model_dump(
            exclude={"display_name"}
        )

        assert "displayName" not in dumped
        assert dumped["custom"] is True
