from datetime import datetime

from decimal import Decimal
import re
import pytest
import pydantic
from openapi import SDK

from openapi.models import shared
from openapi import utils

from .common_helpers import *
from .test_helpers import *


def test_strongly_typed_one_of_post_basic():
    record_test("unions-strongly-typed-one-of-post-basic")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_simple_object_with_type()

    res = s.unions.strongly_typed_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.SimpleObjectWithType)
    compare_simple_object_with_type(res.res.json_, obj)


def test_collection_one_of_post():
    record_test("unions-collections-one-of-post")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    req = ["one", "two"]

    res = s.unions.collection_one_of_post(request=req)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == req

    req2 = {"1": "one", "2": "two"}

    res2 = s.unions.collection_one_of_post(request=req2)
    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.res is not None
    assert res2.res.json_ == req2


def test_strongly_typed_one_of_post_with_non_standard_discriminator_name():
    record_test(
        "unions-strongly-typed-one-of-post-with-non-standard-discriminator-name"
    )

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_simple_object_with_non_standard_type_name()

    res = s.unions.strongly_typed_one_of_post_with_non_standard_discriminator_name(
        request=obj
    )
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.SimpleObjectWithNonStandardTypeName)
    compare_simple_object_with_non_standard_type_name(res.res.json_, obj)


def test_strongly_typed_one_of_post_deep():
    record_test("unions-strongly-typed-one-of-post-deep")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_deep_object_with_type()

    res = s.unions.strongly_typed_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.DeepObjectWithType)
    compare_deep_object_with_type(res.res.json_, obj)


def test_weakly_typed_one_of_post_basic():
    record_test("unions-weakly-typed-one-of-post-basic")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_simple_object()

    res = s.unions.weakly_typed_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.SimpleObject)


def test_weakly_typed_one_of_post_deep():
    record_test("unions-weakly-typed-one-of-post-deep")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_deep_object()

    res = s.unions.weakly_typed_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.DeepObject)


def test_typed_object_one_of_post_obj1():
    record_test("unions-typed-object-one-of-post-obj1")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = shared.TypedObject1(value="obj1", type="obj1")

    res = s.unions.typed_object_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.TypedObject1)
    assert res.res.json_.value == "obj1"


def test_typed_object_one_of_post_obj2():
    record_test("unions-typed-object-one-of-post-obj2")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = shared.TypedObject2(value="obj2", TYPE="obj2")

    res = s.unions.typed_object_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.TypedObject2)
    assert res.res.json_.value == "obj2"


def test_typed_object_one_of_post_obj3():
    record_test("unions-typed-object-one-of-post-obj3")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = shared.TypedObject3(value="obj3", TYPE="obj3")

    res = s.unions.typed_object_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.TypedObject3)
    assert res.res.json_.value == "obj3"


def test_typed_object_one_of_post_null():
    record_test("unions-typed-object-one-of-post-null")

    s = SDK(server_url=HTTPBIN_URL)
    with pytest.raises(
        pydantic.ValidationError,
        match=r"expected object with 'type' field",
    ):
        s.unions.typed_object_one_of_post(
            request=None  # pyright: ignore[reportArgumentType]
        )


def test_typed_object_nullable_one_of_post_obj1():
    record_test("unions-typed-object-nullable-one-of-post-obj1")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject1(value="one", type="obj1")

    res = s.unions.typed_object_nullable_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.TypedObject1)
    assert res.res.json_.value == "one"


def test_typed_object_nullable_one_of_post_obj2():
    record_test("unions-typed-object-nullable-one-of-post-obj2")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject2(value="two", TYPE="obj2")

    res = s.unions.typed_object_nullable_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.TypedObject2)
    assert res.res.json_.value == "two"


def test_typed_object_nullable_one_of_post_null():
    record_test("unions-typed-object-nullable-one-of-post-null")

    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.typed_object_nullable_one_of_post(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ is None


def test_flattened_typed_object_post_obj1():
    record_test("unions-flattened-typed-object-post-obj1")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject1(value="one", type="obj1")

    res = s.unions.flattened_typed_object_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_.value == "one"


def test_unions_nullable_typed_object_post_obj1():
    record_test("unions-nullable-typed-object-post-obj1")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject1(value="one", type="obj1")

    res = s.unions.nullable_typed_object_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == obj


def test_unions_nullable_typed_object_post_null():
    record_test("unions-nullable-typed-object-post-null")

    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.nullable_typed_object_post(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ is None


def test_nullable_one_of_schema_post_obj1():
    record_test("unions-nullable-oneof-schema-post-obj1")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject1(value="one", type="obj1")

    res = s.unions.nullable_one_of_schema_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == obj


def test_nullable_one_of_schema_post_obj2():
    record_test("unions-nullable-oneof-schema-post-obj2")

    s = SDK(server_url=HTTPBIN_URL)

    obj = shared.TypedObject2(value="two", TYPE="obj2")

    res = s.unions.nullable_one_of_schema_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == obj


def test_nullable_one_of_schema_post_null():
    record_test("unions-nullable-oneof-schema-post-null")

    s = SDK(server_url=HTTPBIN_URL)

    res = s.unions.nullable_one_of_schema_post(request=None)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ is None


class MicroMock:
    def __init__(self, name, obj, typ, want_json):
        self.name = name
        self.obj = obj
        self.typ = typ
        self.want_json = want_json


def test_nullable_one_of_type_in_object_post():
    record_test("unions-nullable-oneof-type-in-object-post")
    tests = [
        MicroMock(
            name="Nullable fields set to null",
            obj=shared.NullableOneOfTypeInObject(
                nullable_one_of_one=None, nullable_one_of_two=None, one_of_one=True
            ),
            typ=shared.NullableOneOfTypeInObject,
            want_json='{"NullableOneOfOne":null,"NullableOneOfTwo":null,"OneOfOne":true}',
        ),
        MicroMock(
            name="All fields set to non-null values",
            obj=shared.NullableOneOfTypeInObject(
                nullable_one_of_one=True, nullable_one_of_two=2, one_of_one=True
            ),
            typ=shared.NullableOneOfTypeInObject,
            want_json='{"NullableOneOfOne":true,"NullableOneOfTwo":2,"OneOfOne":true}',
        ),
    ]

    s = SDK(server_url=HTTPBIN_URL)
    for tt in tests:
        body = utils.serialize_request_body(tt.obj, False, False, "json", tt.typ)
        assert body is not None
        assert body.content == tt.want_json
        res = s.unions.nullable_one_of_type_in_object_post(
            one_of_one=tt.obj.one_of_one,
            nullable_one_of_one=tt.obj.nullable_one_of_one,
            nullable_one_of_two=tt.obj.nullable_one_of_two,
        )
        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.res is not None
        assert res.res.json_ == tt.obj


def test_nullable_one_of_ref_in_object_post():
    record_test("unions-nullable-oneof-ref-in-object-post")

    tests = [
        MicroMock(
            name="Nullable fields set to null",
            obj=shared.NullableOneOfRefInObject(
                nullable_one_of_one=None,
                nullable_one_of_two=None,
                one_of_one=shared.TypedObject1(value="one", type="obj1"),
            ),
            typ=shared.NullableOneOfRefInObject,
            want_json='{"NullableOneOfOne":null,"NullableOneOfTwo":null,"OneOfOne":{"type":"obj1","value":"one"}}',
        ),
        MicroMock(
            name="All fields set to non-null values",
            obj=shared.NullableOneOfRefInObject(
                nullable_one_of_one=shared.TypedObject1(value="one", type="obj1"),
                nullable_one_of_two=shared.TypedObject2(value="two", TYPE="obj2"),
                one_of_one=shared.TypedObject1(value="", type="obj1"),
            ),
            typ=shared.NullableOneOfRefInObject,
            want_json='{"NullableOneOfOne":{"type":"obj1","value":"one"},"NullableOneOfTwo":{"value":"two","type":"obj2"},"OneOfOne":{"type":"obj1","value":""}}',
        ),
    ]

    s = SDK(server_url=HTTPBIN_URL)
    for tt in tests:
        body = utils.serialize_request_body(tt.obj, False, False, "json", tt.typ)
        assert body is not None
        assert body.content == tt.want_json
        res = s.unions.nullable_one_of_ref_in_object_post(
            one_of_one=tt.obj.one_of_one,
            nullable_one_of_one=tt.obj.nullable_one_of_one,
            nullable_one_of_two=tt.obj.nullable_one_of_two,
        )
        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.res is not None
        assert res.res.json_ == tt.obj


def test_primitive_type_one_of_post_string():
    record_test("unions-primitive-type-one-of-post-string")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.primitive_type_one_of_post(request="test")
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == "test"


def test_primitive_type_one_of_post_integer():
    record_test("unions-primitive-type-one-of-post-integer")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.primitive_type_one_of_post(request=1)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == 1


def test_primitive_type_one_of_post_number():
    record_test("unions-primitive-type-one-of-post-number")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.primitive_type_one_of_post(request=1.1)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == 1.1


def test_primitive_type_one_of_post_boolean():
    record_test("unions-primitive-type-one-of-post-boolean")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.primitive_type_one_of_post(request=True)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ is True


def test_mixed_type_one_of_post_string():
    record_test("unions-mixed-type-one-of-post-string")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.mixed_type_one_of_post(request="str")
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == "str"


def test_mixed_type_one_of_post_integer():
    record_test("unions-mixed-type-one-of-post-integer")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.mixed_type_one_of_post(request=1)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_ == 1


def test_mixed_type_one_of_post_object():
    record_test("unions-mixed-type-one-of-post-object")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    obj = create_simple_object()

    res = s.unions.mixed_type_one_of_post(request=obj)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.SimpleObject)
    compare_simple_object(res.res.json_, obj)


def test_date_null_union():
    record_test("unions-date-null")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.union_date_null(request=datetime.now().date())
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, date)


def test_date_time_null_union():
    record_test("unions-datetime-null")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.union_date_time_null(request=datetime.now())
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, datetime)


def test_date_time_bigint_union():
    record_test("unions-datetime-bigint")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.union_date_time_big_int(request=datetime.now())
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, datetime)

    res = s.unions.union_date_time_big_int(request=9007199254740991)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, int)


def test_union_map():
    record_test("unions-union-map")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.union_map(input={"str": "test", "bool": True})
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.json_.input["str"] == "test"
    assert res.res.json_.input["bool"] is True


def test_unions_extra_json_properties():
    record_test("unions-extra-json-properties")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.one_of_overlapping_objects(
        field1="test1",
        field3=1,
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.Obj1)
    assert res.res.json_.field1 == "test1"

    res = s.unions.one_of_overlapping_objects(
        field1="test2",
        field2=True,
        field3=1,
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, shared.Obj2)
    assert res.res.json_.field1 == "test2"
    assert res.res.json_.field2 is True


def test_unions_nested_enums_form():
    record_test("unions-nested-enums-form")

    s = SDK(server_url=HTTPBIN_URL)

    res1 = s.unions.union_nested_enums_form(
        request={
            "enums": ["one", "two"],
            "tags": "one,two",
        }
    )

    assert res1 is not None
    assert res1.http_meta is not None
    assert res1.http_meta.response is not None
    assert res1.http_meta.response.status_code == 200
    assert res1.res is not None
    assert res1.res.form == {"enums": ["one", "two"], "tags": "one,two"}

    res2 = s.unions.union_nested_enums_form(
        request={
            "enums": {
                "key2": "two",
                "key3": "three",
            },
            "tags": "two,three",
        }
    )

    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.res is not None
    assert res2.res.form == {"enums": '{"key2":"two","key3":"three"}', "tags": "two,three"}


def test_unions_nested_enums_multipart():
    record_test("unions-nested-enums-multipart")

    s = SDK(server_url=HTTPBIN_URL)

    res1 = s.unions.union_nested_enums_multipart(enums=["one", "two"])

    assert res1 is not None
    assert res1.http_meta is not None
    assert res1.http_meta.response is not None
    assert res1.http_meta.response.status_code == 200
    assert res1.res is not None
    assert res1.res.form == {"enums": '["one","two"]'}

    res2 = s.unions.union_nested_enums_multipart(
        enums={
            "key2": "two",
            "key3": "three",
        }
    )

    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.res is not None
    assert res2.res.form == {"enums": '{"key2":"two","key3":"three"}'}


def test_const_discriminator():
    record_test("unions-const-discriminator")
    s = SDK()
    req = shared.ConstObject1(image_url="http://boo")
    res = s.unions.const_discriminated_one_of(request=req)
    assert res is not None
    assert res.http_meta.response.status_code == 200
    assert isinstance(res.res.json_, shared.ConstObject1)
    assert res.res.json_.image_url == "http://boo"


def test_bigint_str_decimal_union():
    record_test("unions-bigint-str-decimal")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.unions.union_big_int_str_decimal(request=Decimal("3.141592653589793"))
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, Decimal)

    res = s.unions.union_big_int_str_decimal(request=9223372036854775807)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, int)


def test_array_of_discriminated_unions():
    record_test("unions-array-of-discriminated-unions")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    simple_object = create_simple_object_with_type()
    deep_object = create_deep_object_with_type()

    res = s.unions.array_of_discriminated_unions(request=[simple_object, deep_object])

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_, list)
    assert len(res.res.json_) == 2
    assert isinstance(res.res.json_[0], shared.SimpleObjectWithType)
    compare_simple_object_with_type(res.res.json_[0], simple_object)
    assert isinstance(res.res.json_[1], shared.DeepObjectWithType)
    compare_deep_object_with_type(res.res.json_[1], deep_object)


def test_array_of_discriminated_unions_map():
    record_test("unions-array-of-discriminated-unions-map")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    simple_object = create_simple_object_with_type()
    deep_object = create_deep_object_with_type()

    res = s.unions.array_of_discriminated_unions_map(
        array_map={
            "item": [simple_object, deep_object],
        }
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    array_map = res.res.json_.array_map
    assert isinstance(array_map, dict) and "item" in array_map
    item = res.res.json_.array_map.get("item")
    assert isinstance(item, list)
    assert len(item) == 2
    assert isinstance(item[0], shared.SimpleObjectWithType)
    compare_simple_object_with_type(item[0], simple_object)
    assert isinstance(item[1], shared.DeepObjectWithType)
    compare_deep_object_with_type(item[1], deep_object)


def test_nested_array_of_discriminated_unions():
    record_test("unions-nested-array-of-discriminated-unions")
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    simple_object = create_simple_object_with_type()
    deep_object = create_deep_object_with_type()

    res = s.unions.nested_array_of_discriminated_unions(
        nested_array=[[simple_object], [deep_object, simple_object]]
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res.json_.nested_array, list)
    assert len(res.res.json_.nested_array) == 2
    assert isinstance(res.res.json_.nested_array[0], list)
    assert len(res.res.json_.nested_array[0]) == 1
    assert isinstance(res.res.json_.nested_array[0][0], shared.SimpleObjectWithType)
    compare_simple_object_with_type(res.res.json_.nested_array[0][0], simple_object)
    assert isinstance(res.res.json_.nested_array[1], list)
    assert len(res.res.json_.nested_array[1]) == 2
    assert isinstance(res.res.json_.nested_array[1][0], shared.DeepObjectWithType)
    compare_deep_object_with_type(res.res.json_.nested_array[1][0], deep_object)
    assert isinstance(res.res.json_.nested_array[1][1], shared.SimpleObjectWithType)
    compare_simple_object_with_type(res.res.json_.nested_array[1][1], simple_object)


def test_circular_reference_recursive_one_of():
    record_test("unions-circular-reference-recursive-one-of")

    s = SDK(server_url=HTTPBIN_URL)

    payload = ["hello", {"nested": ["world"]}]

    res = s.unions.circular_reference_recursive_one_of(value=payload)

    assert res is not None
    assert res.http_meta.response.status_code == 200
    assert res.object is not None
    assert res.object.json_.value == payload


def test_whole_number_float_literals_survive_nullable_union_serialization():
    """Whole-number float default and const values (80.0) must serialize as
    floats so the populated list inside an OptionalNullable union is not
    dropped in favor of the Unset sentinel branch."""
    holder = shared.WholeNumberDefaultHolder(items=[shared.WholeNumberDefaultItem()])

    assert isinstance(holder.items[0].score, float)
    assert isinstance(holder.items[0].constant_score, float)
    assert holder.model_dump(mode="json", by_alias=True) == {
        "items": [{"constant_score": 80.0, "score": 80.0}]
    }

    assert shared.WholeNumberDefaultHolder().model_dump(mode="json", by_alias=True) == {}
    assert shared.WholeNumberDefaultHolder(items=None).model_dump(
        mode="json", by_alias=True
    ) == {"items": None}
