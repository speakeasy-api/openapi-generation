# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "securerandom"

module OpenApiSDK
  class TestUnions < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_strongly_typed_one_of_post_basic
      record_test("unions-strongly-typed-one-of-post-basic")

      obj = create_simple_object_with_type
      res = @sdk.unions.strongly_typed_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
      assert_equal(obj.type, res.res.json.type)
    end

    def test_strongly_typed_one_of_post_deep
      record_test("unions-strongly-typed-one-of-post-deep")

      obj = create_deep_object_with_type
      res = @sdk.unions.strongly_typed_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
      assert_equal(obj.type, res.res.json.type)
    end

    def test_collection_one_of_post
      record_test("unions-collections-one-of-post")

      res = @sdk.unions.collection_one_of_post(request: ["one", "two"])
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(["one", "two"], res.res.json)

      res = @sdk.unions.collection_one_of_post(request: {"1" => "one", "2" => "two"})
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal({"1" => "one", "2" => "two"}, res.res.json)
    end

    def test_strongly_typed_one_of_post_with_non_standard_discriminator_name
      record_test("unions-strongly-typed-one-of-post-with-non-standard-discriminator-name")

      obj = create_simple_object_with_non_standard_type_name
      res = @sdk.unions.strongly_typed_one_of_post_with_non_standard_discriminator_name(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_weakly_typed_one_of_post_basic
      record_test("unions-weakly-typed-one-of-post-basic")

      obj = create_simple_object
      res = @sdk.unions.weakly_typed_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_weakly_typed_one_of_post_deep
      record_test("unions-weakly-typed-one-of-post-deep")

      obj = create_deep_object
      res = @sdk.unions.weakly_typed_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_typed_object_one_of_post_obj1
      record_test("unions-typed-object-one-of-post-obj1")

      obj = Models::Shared::TypedObject1.new(value: "obj1", type: Models::Shared::TypedObject1Type::OBJ1)
      res = @sdk.unions.typed_object_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_typed_object_one_of_post_obj2
      record_test("unions-typed-object-one-of-post-obj2")

      obj = Models::Shared::TypedObject2.new(value: "obj2", type: Models::Shared::TypedObject2Type::OBJ2)
      res = @sdk.unions.typed_object_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_typed_object_one_of_post_obj3
      record_test("unions-typed-object-one-of-post-obj3")

      obj = Models::Shared::TypedObject3.new(value: "obj3", type: Models::Shared::TypedObject3Type::OBJ3)
      res = @sdk.unions.typed_object_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_typed_object_one_of_post_null
      record_test("unions-typed-object-one-of-post-null")

      assert_raises(TypeError) do
        @sdk.unions.typed_object_one_of_post(request: nil)
      end
    end

    def test_typed_object_nullable_one_of_post_obj1
      record_test("unions-typed-object-nullable-one-of-post-obj1")

      obj = Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
      res = @sdk.unions.typed_object_nullable_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(Models::Shared::TypedObject1, res.res.json)
      assert_equal("one", res.res.json.value)
    end

    def test_typed_object_nullable_one_of_post_obj2
      record_test("unions-typed-object-nullable-one-of-post-obj2")

      obj = Models::Shared::TypedObject2.new(value: "two", type: Models::Shared::TypedObject2Type::OBJ2)
      res = @sdk.unions.typed_object_nullable_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(Models::Shared::TypedObject2, res.res.json)
      assert_equal("two", res.res.json.value)
    end

    def test_typed_object_nullable_one_of_post_null
      record_test("unions-typed-object-nullable-one-of-post-null")

      res = @sdk.unions.typed_object_nullable_one_of_post(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_nil(res.res.json)
    end

    def test_flattened_typed_object_post_obj1
      record_test("unions-flattened-typed-object-post-obj1")

      obj = Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
      res = @sdk.unions.flattened_typed_object_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_unions_nullable_typed_object_post_obj_1
      record_test("unions-nullable-typed-object-post-obj1")

      obj = Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
      res = @sdk.unions.nullable_typed_object_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_unions_nullable_typed_object_post_null
      record_test("unions-nullable-typed-object-post-null")

      res = @sdk.unions.nullable_typed_object_post(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_nil(res.res.json)
    end

    def test_nullable_one_of_schema_post_obj1
      record_test("unions-nullable-oneof-schema-post-obj1")

      obj = Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
      res = @sdk.unions.nullable_one_of_schema_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_nullable_one_of_schema_post_obj2
      record_test("unions-nullable-oneof-schema-post-obj2")

      obj = Models::Shared::TypedObject2.new(value: "two", type: Models::Shared::TypedObject2Type::OBJ2)
      res = @sdk.unions.nullable_one_of_schema_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_nullable_one_of_schema_post_null
      record_test("unions-nullable-oneof-schema-post-null")

      res = @sdk.unions.nullable_one_of_schema_post(request: nil)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_nil(res.res.json)
    end

    def test_nullable_one_of_type_in_object_post
      record_test("unions-nullable-oneof-type-in-object-post")

      tests = [
        {
          name: "Nullable fields set to null",
          obj: Models::Shared::NullableOneOfTypeInObject.new(
            nullable_one_of_one: nil,
            nullable_one_of_two: nil,
            one_of_one: true
          ),
          want: "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":true}"
        },
        {
          name: "All fields set to non-null values",
          obj: Models::Shared::NullableOneOfTypeInObject.new(
            nullable_one_of_one: true,
            nullable_one_of_two: 2,
            one_of_one: true
          ),
          want: "{\"NullableOneOfOne\":true,\"NullableOneOfTwo\":2,\"OneOfOne\":true}"
        }
      ]

      tests.each do |test|
        _, req, __ = Utils.serialize_request_body(test[:obj], true, true, :request, :json)
        refute_nil(req)
        assert_equal(test[:want], req)
        res = @sdk.unions.nullable_one_of_type_in_object_post(request: test[:obj])
        refute_nil(res)
        assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
        assert_equal(test[:obj], res.res.json)
      end
    end

    def test_nullable_one_of_ref_in_object
      record_test("unions-nullable-oneof-ref-in-object-post")

      tests = [
        {
          name: "Non-nullable field set only",
          obj: Models::Shared::NullableOneOfRefInObject.new(
            one_of_one: Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
          ),
          want: "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
        },
        {
          name: "Nullable fields set to null",
          obj: Models::Shared::NullableOneOfRefInObject.new(
            nullable_one_of_one: nil,
            nullable_one_of_two: nil,
            one_of_one: Models::Shared::TypedObject1.new(value: "one", type: Models::Shared::TypedObject1Type::OBJ1)
          ),
          want: "{\"NullableOneOfOne\":null,\"NullableOneOfTwo\":null,\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"}}"
        },
        {
          name: "All fields set to non-null values",
          obj: Models::Shared::NullableOneOfRefInObject.new(
            nullable_one_of_one: Models::Shared::TypedObject1.new(
              value: "one",
              type: Models::Shared::TypedObject1Type::OBJ1
            ),
            nullable_one_of_two: Models::Shared::TypedObject2.new(
              value: "two",
              type: Models::Shared::TypedObject2Type::OBJ2
            ),
            one_of_one: Models::Shared::TypedObject1.new(value: "", type: Models::Shared::TypedObject1Type::OBJ1)
          ),
          want: "{\"NullableOneOfOne\":{\"type\":\"obj1\",\"value\":\"one\"},\"NullableOneOfTwo\":{\"type\":\"obj2\",\"value\":\"two\"},\"OneOfOne\":{\"type\":\"obj1\",\"value\":\"\"}}"
        }
      ]

      tests.each do |test|
        _, req, __ = Utils.serialize_request_body(test[:obj], true, true, :request, :json)
        refute_nil(req)
        assert_equal(test[:want], req)
        res = @sdk.unions.nullable_one_of_ref_in_object_post(request: test[:obj])
        refute_nil(res)
        assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
        assert_equal(test[:obj], res.res.json)
      end
    end

    def test_primitive_type_one_of_post_string
      record_test("unions-primitive-type-one-of-post-string")

      res = @sdk.unions.primitive_type_one_of_post(request: "str")
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("str", res.res.json)
    end

    def test_primitive_type_one_of_post_integer
      record_test("unions-primitive-type-one-of-post-integer")

      res = @sdk.unions.primitive_type_one_of_post(request: 1)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(1, res.res.json)
    end

    def test_primitive_type_one_of_post_number
      record_test("unions-primitive-type-one-of-post-number")

      res = @sdk.unions.primitive_type_one_of_post(request: 1.1)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(1.1, res.res.json)
    end

    def test_primitive_type_one_of_post_boolean
      record_test("unions-primitive-type-one-of-post-boolean")

      res = @sdk.unions.primitive_type_one_of_post(request: true)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert(res.res.json)
    end

    def test_mixed_type_one_of_post_string
      record_test("unions-mixed-type-one-of-post-string")

      res = @sdk.unions.mixed_type_one_of_post(request: "str")
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("str", res.res.json)
    end

    def test_mixed_type_one_of_post_integer
      record_test("unions-mixed-type-one-of-post-integer")

      res = @sdk.unions.mixed_type_one_of_post(request: 1)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(1, res.res.json)
    end

    def test_mixed_type_one_of_post_object
      record_test("unions-mixed-type-one-of-post-object")

      obj = create_simple_object
      res = @sdk.unions.mixed_type_one_of_post(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end

    def test_date_null_union
      record_test("unions-date-null")

      res = @sdk.unions.union_date_null(request: Date.new(2020, 1, 1))
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(Date, res.res.json)
    end

    def test_date_time_null_union
      record_test("unions-datetime-null")

      res = @sdk.unions.union_date_time_null(request: DateTime.new(2020, 1, 1, 0, 0, 0, 1 / 1_000_000))
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(DateTime, res.res.json)
    end

    # TODO: needs bigint support
    # def test_date_time_bigint_union
    #   record_test('unions-datetime-bigint')

    #   res = @sdk.unions.union_date_time_bigint(DateTime.new(2020, 1, 1, 0, 0, 0, 1 / 1_000_000))
    #   refute_nil res
    #   assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    #   assert_instance_of(DateTime, res.res.json)

    #   big_int_val = # ??
    #   res = @sdk.unions.union_date_time_bigint(big_int_val)
    #   refute_nil res
    #   assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    #   assert_equal(big_int_val, res.res.json)
    # end

    def test_union_map
      record_test("unions-union-map")

      res = @sdk
        .unions
        .union_map(request: Models::Operations::UnionMapRequestBody.new(input: {"str" => "test", "bool" => true}))
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal({"str" => "test", "bool" => true}, res.res.json.input)
    end

    def test_optional_union_map
      record_test("unions-optional-union-map")

      res = @sdk
        .unions
        .union_map_optional(
          request: Models::Operations::UnionMapOptionalRequestBody.new(input: {"str" => "test", "bool" => true})
        )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal({"str" => "test", "bool" => true}, res.res.json.input)

      res = @sdk.unions.union_map_optional(request: Models::Operations::UnionMapOptionalRequestBody.new(input: nil))
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_nil(res.res.json.input)
    end

    def test_mixed_union_types
      record_test("unions-mixed-union-types")

      bike1 = Models::Shared::Bike.new(colour: "white", vehicle_type: "bike", wheels_type: "two")
      bike2 = Models::Shared::Bike.new(colour: "brown", vehicle_type: "bike", wheels_type: "two")
      res = @sdk.unions.mixed_union_types(request: [bike1, bike2])
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal([bike1, bike2], res.res.json)

      res = @sdk.unions.mixed_union_types(request: bike1)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(bike1, res.res.json)
    end

    def test_unions_extra_json_properties
      record_test("unions-extra-json-properties")

      req = Models::Operations::OneOfOverlappingObjectsRequestBody.new(field1: "test1", field3: 1.0)
      res = @sdk.unions.one_of_overlapping_objects(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(Models::Shared::Obj1, res.res.json)
      assert_equal("test1", res.res.json.field1)

      req2 = Models::Operations::OneOfOverlappingObjectsRequestBody.new(field1: "test2", field2: true, field3: 1.0)
      res = @sdk.unions.one_of_overlapping_objects(request: req2)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_instance_of(Models::Shared::Obj2, res.res.json)
      assert_equal("test2", res.res.json.field1)
      assert(res.res.json.field2)
    end

    def test_enum_nested_in_array_in_union
      record_test("union-enum-nested-in-array-in-union")
      val = Models::Operations::OneOfCollectionEnumRes.new(json: [Models::Shared::Nestedenum::ABC])
      dump = Crystalline.to_json(val)
      assert_equal(dump, "{\"json\":[\"abc\"]}")
    end

    def test_unions_mixed_type_one_of_boolean_and_string_enum_with_boolean_response
      record_test("unions-one-of-boolean-and-string-enum-with-response-boolean")
      res = @sdk
        .unions
        .one_of_boolean_and_string_enum(
          request: Models::Operations::OneOfBooleanAndStringEnumRequestBody.new(active: true)
        )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert(
        res.res.json.active.is_a?(TrueClass) || res.res.json.active.is_a?(FalseClass),
        "Expected boolean, got #{res.res.json.active.class}"
      )
    end

    def test_unions_mixed_type_one_of_boolean_and_string_enum_with_response_enum
      record_test("unions-one-of-boolean-and-string-enum-with-response-enum")
      obj = Models::Operations::OneOfBooleanAndStringEnumRequestBody.new(
        active: Models::Operations::OneOfBooleanAndStringEnum2::TRUE
      )
      res = @sdk.unions.one_of_boolean_and_string_enum(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(
        Models::Operations::OneOfBooleanAndStringEnumUnions2::TRUE,
        res.res.json.active,
        "Expected enum got #{res.res.json.active.class}"
      )
    end

    def test_array_of_discriminated_unions
      record_test("unions-array-of-discriminated-unions")

      simple_object = create_simple_object_with_type
      deep_object = create_deep_object_with_type

      res = @sdk.unions.array_of_discriminated_unions(request: [simple_object, deep_object])
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Array, res.res.json)
      assert_equal(2, res.res.json.length)
      assert_instance_of(Models::Shared::SimpleObjectWithType, res.res.json[0])
      assert_equal(simple_object, res.res.json[0])
      assert_instance_of(Models::Shared::DeepObjectWithType, res.res.json[1])
      assert_equal(deep_object, res.res.json[1])
    end

    def test_array_of_discriminated_unions_map
      record_test("unions-array-of-discriminated-unions-map")

      simple_object = create_simple_object_with_type
      deep_object = create_deep_object_with_type

      res = @sdk.unions.array_of_discriminated_unions_map(
        request: Models::Shared::ArrayOfDiscriminatedUnionsMap.new(
          array_map: { "item" => [simple_object, deep_object] }
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Hash, res.res.json.array_map)
      assert(res.res.json.array_map.key?("item"))
      item = res.res.json.array_map["item"]
      assert_instance_of(Array, item)
      assert_equal(2, item.length)
      assert_instance_of(Models::Shared::SimpleObjectWithType, item[0])
      assert_equal(simple_object, item[0])
      assert_instance_of(Models::Shared::DeepObjectWithType, item[1])
      assert_equal(deep_object, item[1])
    end

    def test_nested_array_of_discriminated_unions
      record_test("unions-nested-array-of-discriminated-unions")

      simple_object = create_simple_object_with_type
      deep_object = create_deep_object_with_type

      res = @sdk.unions.nested_array_of_discriminated_unions(
        request: Models::Shared::NestedArrayOfDiscriminatedUnions.new(
          nested_array: [[simple_object], [deep_object, simple_object]]
        )
      )
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      nested = res.res.json.nested_array
      assert_instance_of(Array, nested)
      assert_equal(2, nested.length)
      assert_instance_of(Array, nested[0])
      assert_equal(1, nested[0].length)
      assert_instance_of(Models::Shared::SimpleObjectWithType, nested[0][0])
      assert_equal(simple_object, nested[0][0])
      assert_instance_of(Array, nested[1])
      assert_equal(2, nested[1].length)
      assert_instance_of(Models::Shared::DeepObjectWithType, nested[1][0])
      assert_equal(deep_object, nested[1][0])
      assert_instance_of(Models::Shared::SimpleObjectWithType, nested[1][1])
      assert_equal(simple_object, nested[1][1])
    end

    def test_const_discriminator
      record_test("unions-const-discriminator")
      obj = Models::Shared::ConstObject1.new(image_url: "http://boo")
      res = @sdk.unions.const_discriminated_one_of(request: obj)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal(obj, res.res.json)
    end


    def test_circular_reference_recursive_one_of
      record_test("unions-circular-reference-recursive-one-of")

      payload = ["hello", { "nested" => ["world"] }]

      req = Models::Operations::CircularReferenceRecursiveOneOfRequestBody.new(
        value: payload
      )
      res = @sdk.unions.circular_reference_recursive_one_of(request: req)

      refute_nil(res)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.object)
      assert_equal(payload, res.object.json.value)
    end
  end
end
