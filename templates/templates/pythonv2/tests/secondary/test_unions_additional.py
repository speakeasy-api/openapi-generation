from openapi import SDK
from openapi.models import shared

from .common_helpers import HTTPBIN_URL


def test_discriminated_union_enum_tag_in_dict():
    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.nullable_one_of_schema_post(
        request={"type": shared.TypedObject1Type.OBJ1, "value": "one"}
    )

    assert res.json_ == shared.TypedObject1(
        type=shared.TypedObject1Type.OBJ1, value="one"
    )


def test_nested_union_enum_tag_in_dict():
    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.one_of_enum_tag_nested_post(
        request={
            "single": {"type": shared.TypedObject1Type.OBJ1, "value": "one"},
            "many": [
                {"type": shared.TypedObject1Type.OBJ1, "value": "two"},
                {"type": "obj2", "value": "three"},
            ],
        }
    )

    assert res.json_ == {
        "single": {"type": "obj1", "value": "one"},
        "many": [
            {"type": "obj1", "value": "two"},
            {"type": "obj2", "value": "three"},
        ],
    }
