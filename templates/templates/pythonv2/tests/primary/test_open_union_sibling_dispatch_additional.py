import json

import httpx
import pytest
from pydantic import ValidationError

from openapi import SDK, utils
from openapi.models import shared
from openapi.utils.unmarshal_json_response import unmarshal_json_response

from .common_helpers import HTTPBIN_URL

CIRCLE = {"kind": "circle", "radius": 1}
SQUARE = {"kind": "square", "side": 2}
TEXT = {"kind": "text", "text": "t"}
NUMBER = {"kind": "number", "number": 3}
INVALID_CIRCLE = {"kind": "circle", "radius": "oops"}
TRIANGLE = {"kind": "triangle", "corners": 3}


def parse_response(payload):
    return unmarshal_json_response(
        shared.OpenUnionSiblings, httpx.Response(200, text=json.dumps(payload))
    )


def parse_request(payload):
    return shared.OpenUnionSiblings.model_validate(payload)


def wire(model):
    return json.loads(utils.marshal_json(model, shared.OpenUnionSiblings))


class TestResponseObjectMember:
    @pytest.mark.parametrize(
        "payload,expected",
        [
            (CIRCLE, shared.SiblingCircle),
            (SQUARE, shared.SiblingSquare),
            (TEXT, shared.SiblingTextLabel),
            (NUMBER, shared.SiblingNumberLabel),
        ],
    )
    def test_known_value_resolves_to_owning_union(self, payload, expected):
        model = parse_response({"labelOrShape": payload})

        assert isinstance(model.label_or_shape, expected)
        assert wire(model) == {"labelOrShape": payload}

    def test_invalid_payload_falls_back_to_owning_union(self):
        model = parse_response({"labelOrShape": INVALID_CIRCLE})

        assert isinstance(model.label_or_shape, shared.UnknownSiblingShape)
        assert model.label_or_shape.raw == INVALID_CIRCLE

    def test_value_no_union_knows_falls_back_to_first_member(self):
        model = parse_response({"labelOrShape": TRIANGLE})

        assert isinstance(model.label_or_shape, shared.UnknownSiblingLabel)
        assert model.label_or_shape.raw == TRIANGLE


class TestResponseArrayMember:
    def test_known_values_resolve_to_owning_union(self):
        payload = {"labelsOrShapes": [CIRCLE, SQUARE]}
        model = parse_response(payload)

        assert [type(item) for item in model.labels_or_shapes] == [
            shared.SiblingCircle,
            shared.SiblingSquare,
        ]
        assert wire(model) == payload

    def test_second_member_with_unknown_item(self):
        model = parse_response({"labelsOrShapes": [CIRCLE, TRIANGLE]})

        circle, unknown = model.labels_or_shapes
        assert isinstance(circle, shared.SiblingCircle)
        assert isinstance(unknown, shared.UnknownSiblingShape)
        assert unknown.raw == TRIANGLE

    def test_invalid_payload_falls_back_to_owning_union(self):
        model = parse_response({"labelsOrShapes": [INVALID_CIRCLE]})

        (item,) = model.labels_or_shapes
        assert isinstance(item, shared.UnknownSiblingShape)
        assert item.raw == INVALID_CIRCLE

    def test_first_member_values(self):
        payload = {"labelsOrShapes": [TEXT, NUMBER]}
        model = parse_response(payload)

        assert [type(item) for item in model.labels_or_shapes] == [
            shared.SiblingTextLabel,
            shared.SiblingNumberLabel,
        ]
        assert wire(model) == payload

    def test_unknown_items_fall_back_to_first_member(self):
        model = parse_response({"labelsOrShapes": [TRIANGLE, TRIANGLE]})

        assert [type(item) for item in model.labels_or_shapes] == [
            shared.UnknownSiblingLabel,
            shared.UnknownSiblingLabel,
        ]

    def test_empty(self):
        assert parse_response({"labelsOrShapes": []}).labels_or_shapes == []

    def test_items_owned_by_different_unions_use_union_validation(self):
        model = parse_response({"labelsOrShapes": [CIRCLE, TEXT]})

        unknown, text = model.labels_or_shapes
        assert isinstance(unknown, shared.UnknownSiblingLabel)
        assert unknown.raw == CIRCLE
        assert isinstance(text, shared.SiblingTextLabel)


class TestResponseMapMember:
    def test_known_values_resolve_to_owning_union(self):
        payload = {"labelOrShapeMap": {"a": CIRCLE, "b": SQUARE}}
        model = parse_response(payload)

        assert isinstance(model.label_or_shape_map["a"], shared.SiblingCircle)
        assert isinstance(model.label_or_shape_map["b"], shared.SiblingSquare)
        assert wire(model) == payload

    def test_invalid_payload_falls_back_to_owning_union(self):
        model = parse_response({"labelOrShapeMap": {"a": INVALID_CIRCLE}})

        item = model.label_or_shape_map["a"]
        assert isinstance(item, shared.UnknownSiblingShape)
        assert item.raw == INVALID_CIRCLE

    def test_first_member_values(self):
        model = parse_response({"labelOrShapeMap": {"a": TEXT}})

        assert isinstance(model.label_or_shape_map["a"], shared.SiblingTextLabel)

    def test_second_member_with_unknown_value(self):
        model = parse_response({"labelOrShapeMap": {"a": CIRCLE, "b": TRIANGLE}})

        assert isinstance(model.label_or_shape_map["a"], shared.SiblingCircle)
        assert isinstance(model.label_or_shape_map["b"], shared.UnknownSiblingShape)
        assert model.label_or_shape_map["b"].raw == TRIANGLE

    def test_values_owned_by_different_unions_use_union_validation(self):
        model = parse_response({"labelOrShapeMap": {"a": CIRCLE, "b": TEXT}})

        assert isinstance(model.label_or_shape_map["a"], shared.UnknownSiblingLabel)
        assert isinstance(model.label_or_shape_map["b"], shared.SiblingTextLabel)

    def test_empty(self):
        assert parse_response({"labelOrShapeMap": {}}).label_or_shape_map == {}


class TestRequest:
    @pytest.mark.parametrize(
        "payload,expected",
        [
            (CIRCLE, shared.SiblingCircle),
            (SQUARE, shared.SiblingSquare),
            (TEXT, shared.SiblingTextLabel),
            (NUMBER, shared.SiblingNumberLabel),
        ],
    )
    def test_known_value_resolves_to_owning_union(self, payload, expected):
        request = {
            "labelOrShape": payload,
            "labelsOrShapes": [payload],
            "labelOrShapeMap": {"a": payload},
        }
        model = parse_request(request)

        assert isinstance(model.label_or_shape, expected)
        assert isinstance(model.labels_or_shapes[0], expected)
        assert isinstance(model.label_or_shape_map["a"], expected)
        assert wire(model) == request

    @pytest.mark.parametrize(
        "request_body",
        [
            {"labelOrShape": INVALID_CIRCLE},
            {"labelsOrShapes": [INVALID_CIRCLE]},
            {"labelOrShapeMap": {"a": INVALID_CIRCLE}},
            {"labelOrShape": TRIANGLE},
            {"labelsOrShapes": [TRIANGLE]},
            {"labelOrShapeMap": {"a": TRIANGLE}},
            {"labelsOrShapes": [CIRCLE, TEXT]},
        ],
    )
    def test_invalid_value_raises(self, request_body):
        with pytest.raises(ValidationError):
            parse_request(request_body)

    @pytest.mark.parametrize(
        "request_body,location",
        [
            ({"labelOrShape": INVALID_CIRCLE}, ("labelOrShape", "radius")),
            ({"labelsOrShapes": [INVALID_CIRCLE]}, ("labelsOrShapes", 0, "radius")),
        ],
    )
    def test_invalid_value_reports_owning_union_error(self, request_body, location):
        with pytest.raises(ValidationError) as exc_info:
            parse_request(request_body)

        assert [error["loc"] for error in exc_info.value.errors()] == [location]

    def test_model_instances(self):
        model = shared.OpenUnionSiblings(
            label_or_shape=shared.SiblingCircle(radius=1),
            labels_or_shapes=[
                shared.SiblingSquare(side=2),
                shared.SiblingCircle(radius=1),
            ],
            label_or_shape_map={
                "a": shared.SiblingTextLabel(text="t"),
                "b": shared.SiblingNumberLabel(number=3),
            },
        )

        assert wire(model) == {
            "labelOrShape": CIRCLE,
            "labelsOrShapes": [SQUARE, CIRCLE],
            "labelOrShapeMap": {"a": TEXT, "b": NUMBER},
        }


class TestObjectArraysAndString:
    @pytest.mark.parametrize("parse", [parse_response, parse_request])
    @pytest.mark.parametrize(
        "payload,expected",
        [
            (TEXT, shared.SiblingTextLabel),
            ([TEXT, NUMBER], [shared.SiblingTextLabel, shared.SiblingNumberLabel]),
            ([CIRCLE, SQUARE], [shared.SiblingCircle, shared.SiblingSquare]),
            ("plain", str),
        ],
    )
    def test_value_resolves_to_owning_member(self, parse, payload, expected):
        model = parse({"labelOrListsOrText": payload})

        value = model.label_or_lists_or_text
        if isinstance(expected, list):
            assert [type(item) for item in value] == expected
        else:
            assert isinstance(value, expected)
        assert wire(model) == {"labelOrListsOrText": payload}


class TestOtherMember:
    @pytest.mark.parametrize("parse", [parse_response, parse_request])
    def test_known_value_resolves_to_owning_union(self, parse):
        model = parse({"labelShapeOrFreeform": CIRCLE})

        assert isinstance(model.label_shape_or_freeform, shared.SiblingCircle)

    def test_response_invalid_payload_falls_back_to_owning_union(self):
        model = parse_response({"labelShapeOrFreeform": INVALID_CIRCLE})

        assert isinstance(model.label_shape_or_freeform, shared.UnknownSiblingShape)
        assert model.label_shape_or_freeform.raw == INVALID_CIRCLE

    def test_request_invalid_payload_resolves_to_other_member(self):
        model = parse_request({"labelShapeOrFreeform": INVALID_CIRCLE})

        assert isinstance(model.label_shape_or_freeform, shared.SiblingFreeform)

    def test_request_invalid_item_resolves_to_other_member(self):
        model = parse_request({"labelsShapesOrFreeforms": [INVALID_CIRCLE]})

        assert [type(item) for item in model.labels_shapes_or_freeforms] == [
            shared.SiblingFreeform
        ]

    @pytest.mark.parametrize("parse", [parse_response, parse_request])
    def test_item_without_discriminator_resolves_to_other_member(self, parse):
        model = parse({"labelsShapesOrFreeforms": [CIRCLE, {"note": "x"}]})

        assert [type(item) for item in model.labels_shapes_or_freeforms] == [
            shared.SiblingFreeform,
            shared.SiblingFreeform,
        ]

    @pytest.mark.parametrize("parse", [parse_response, parse_request])
    def test_object_with_discriminated_property_is_not_a_map(self, parse):
        payload = {"labelOrShapeMapOrSettings": {"title": "t", "primary": CIRCLE}}
        model = parse(payload)

        settings = model.label_or_shape_map_or_settings
        assert isinstance(settings, shared.SiblingSettings)
        assert isinstance(settings.primary, shared.SiblingCircle)
        assert wire(model) == payload

    def test_map_resolves_to_owning_union(self):
        model = parse_response({"labelOrShapeMapOrSettings": {"a": CIRCLE}})

        assert isinstance(model.label_or_shape_map_or_settings["a"], shared.SiblingCircle)

    def test_unknown_object_with_discriminated_property_is_not_a_map(self):
        payload = {"kind": "triangle", "inner": CIRCLE}
        model = parse_response({"labelOrShapeMapEntry": payload})

        assert isinstance(model.label_or_shape_map_entry, shared.UnknownSiblingLabel)
        assert model.label_or_shape_map_entry.raw == payload


def test_round_trip_through_sdk():
    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.open_union_siblings_post(
        label_or_shape=CIRCLE,
        labels_or_shapes=[SQUARE, CIRCLE],
        label_or_shape_map={"a": CIRCLE},
    )

    assert res.object is not None
    echoed = res.object.json_
    assert echoed is not None
    assert isinstance(echoed.label_or_shape, shared.SiblingCircle)
    assert [type(item) for item in echoed.labels_or_shapes] == [
        shared.SiblingSquare,
        shared.SiblingCircle,
    ]
    assert isinstance(echoed.label_or_shape_map["a"], shared.SiblingCircle)
