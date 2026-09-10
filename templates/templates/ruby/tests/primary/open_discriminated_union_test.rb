# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"

# Simple test models for open discriminated union tests.
# These mirror the inline schemas used by the TypeScript test suite.
module OpenApiSDK
  module TestOpenUnionModels
    class Cat
      include Crystalline::MetadataFields

      field :type, String, { "format_json": { "letter_case": proc { "type" } } }
      field :name, String, { "format_json": { "letter_case": proc { "name" } } }

      def initialize(type: nil, name: nil)
        @type = type
        @name = name
      end

      def ==(other)
        other.is_a?(Cat) && @type == other.type && @name == other.name
      end
    end

    class Dog
      include Crystalline::MetadataFields

      field :type, String, { "format_json": { "letter_case": proc { "type" } } }
      field :age, Integer, { "format_json": { "letter_case": proc { "age" } } }

      def initialize(type: nil, age: nil)
        @type = type
        @age = age
      end

      def ==(other)
        other.is_a?(Dog) && @type == other.type && @age == other.age
      end
    end

    class CatStrict
      include Crystalline::MetadataFields

      field :type, String, { "format_json": { "letter_case": proc { "type" } } }
      field :name, String, { "format_json": { "letter_case": proc { "name" } } }
      field :required_field, Integer, { "format_json": { "letter_case": proc { "requiredField" } } }

      def initialize(type: nil, name: nil, required_field: nil)
        @type = type
        @name = name
        @required_field = required_field
      end
    end

    class ValueObject
      include Crystalline::MetadataFields

      field :value, Integer, { "format_json": { "letter_case": proc { "value" } } }

      def initialize(value: nil)
        @value = value
      end

      def ==(other)
        other.is_a?(ValueObject) && @value == other.value
      end
    end

    class FooObject
      include Crystalline::MetadataFields

      field :foo, String, { "format_json": { "letter_case": proc { "foo" } } }

      def initialize(foo: nil)
        @foo = foo
      end

      def ==(other)
        other.is_a?(FooObject) && @foo == other.foo
      end
    end

    class Container
      include Crystalline::MetadataFields

      field :label, String, { "format_json": { "letter_case": proc { "label" } } }
      field :animal, Crystalline::DiscriminatedUnion.new("type", {
        "cat" => Cat,
        "dog" => Dog
      }), { "format_json": { "letter_case": proc { "animal" } } }

      def initialize(label: nil, animal: nil)
        @label = label
        @animal = animal
      end
    end
  end

  class TestOpenDiscriminatedUnion < Minitest::Test
    # Test: Known discriminator values parse correctly
    def test_known_variant
      record_test("open-union-known-variant")

      union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat,
        "dog" => TestOpenUnionModels::Dog
      })

      cat_payload = { "type" => "cat", "name" => "whiskers" }
      cat_result = union.parse(cat_payload)

      refute Crystalline.unknown?(cat_result)
      assert_instance_of TestOpenUnionModels::Cat, cat_result
      assert_equal "cat", cat_result.type
      assert_equal "whiskers", cat_result.name

      dog_payload = { "type" => "dog", "age" => 5 }
      dog_result = union.parse(dog_payload)

      refute Crystalline.unknown?(dog_result)
      assert_instance_of TestOpenUnionModels::Dog, dog_result
      assert_equal "dog", dog_result.type
      assert_equal 5, dog_result.age
    end

    # Test: Unknown discriminator values produce Unknown fallback with raw payload
    def test_unknown_discriminator
      record_test("open-union-unknown-discriminator")

      union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat,
        "dog" => TestOpenUnionModels::Dog
      })

      payload = { "type" => "bird", "wingspan" => 5 }
      result = union.parse(payload)

      assert Crystalline.unknown?(result)
      assert result.unknown?
      assert_equal payload, result.raw
    end

    # Test: Capture payload with missing discriminator property
    def test_missing_discriminator
      record_test("open-union-missing-discriminator")

      union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat
      })

      # Object without the 'type' discriminator
      payload = { "name" => "whiskers" }
      result = union.parse(payload)

      assert Crystalline.unknown?(result)
      assert_equal payload, result.raw
    end

    # Test: Invalid payloads (non-object, non-string discriminator, null, array) -> Unknown
    def test_invalid_payload
      record_test("open-union-invalid-payload")

      union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat
      })

      # String payload instead of object
      str_result = union.parse("not an object")
      assert Crystalline.unknown?(str_result)
      assert_equal "not an object", str_result.raw

      # Non-string discriminator
      num_disc_result = union.parse({ "type" => 123, "name" => "whiskers" })
      assert Crystalline.unknown?(num_disc_result)

      # Null payload
      null_result = union.parse(nil)
      assert Crystalline.unknown?(null_result)
      assert_nil null_result.raw

      # Array payload
      array_result = union.parse([1, 2, 3])
      assert Crystalline.unknown?(array_result)
      assert_equal [1, 2, 3], array_result.raw
    end

    # Test: Known discriminator but schema validation fails -> falls back to Unknown
    def test_known_disc_invalid_schema
      record_test("open-union-known-disc-invalid-schema")

      union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::CatStrict
      })

      # Has correct discriminator but missing requiredField
      payload = { "type" => "cat", "name" => "whiskers" }
      result = union.parse(payload)

      assert Crystalline.unknown?(result)
      assert_equal payload, result.raw
    end

    # Test: smartUnion with discriminatedUnion interop
    def test_smart_union_interop
      record_test("open-union-smart-union-interop")

      disc_union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat,
        "dog" => TestOpenUnionModels::Dog
      })

      # smartUnion(discriminatedUnion | String) with string payload -> matches String
      string_payload = "hello world"
      union_types = [disc_union, String]
      result = Crystalline.unmarshal_union_populated_fields(string_payload, union_types)
      assert_equal string_payload, result
      assert_instance_of String, result

      # smartUnion(discriminatedUnion | String) with known object payload -> matches discriminatedUnion
      object_payload = { "type" => "cat", "name" => "whiskers" }
      result2 = Crystalline.unmarshal_union_populated_fields(object_payload, union_types)
      refute Crystalline.unknown?(result2)
      assert_instance_of TestOpenUnionModels::Cat, result2

      # smartUnion(discriminatedUnion | FooObject) with simple object -> matches FooObject
      simple_payload = { "foo" => "bar" }
      union_types2 = [disc_union, TestOpenUnionModels::FooObject]
      result3 = Crystalline.unmarshal_union_populated_fields(simple_payload, union_types2)
      refute Crystalline.unknown?(result3)
      assert_instance_of TestOpenUnionModels::FooObject, result3

      # smartUnion(discriminatedUnion | FooObject) with known disc -> matches discriminatedUnion
      disc_payload = { "type" => "dog", "age" => 3 }
      result4 = Crystalline.unmarshal_union_populated_fields(disc_payload, union_types2)
      refute Crystalline.unknown?(result4)
      assert_instance_of TestOpenUnionModels::Dog, result4

      # smartUnion prefers exact match over Unknown from discriminatedUnion
      value_payload = { "value" => 42 }
      union_types3 = [disc_union, TestOpenUnionModels::ValueObject]
      result5 = Crystalline.unmarshal_union_populated_fields(value_payload, union_types3)
      assert_instance_of TestOpenUnionModels::ValueObject, result5
      refute Crystalline.unknown?(result5)
    end

    # Test: Open union works when composed in parent models and lists
    def test_embedded
      record_test("open-union-embedded")

      animal_union = Crystalline::DiscriminatedUnion.new("type", {
        "cat" => TestOpenUnionModels::Cat,
        "dog" => TestOpenUnionModels::Dog
      })

      # Known variant embedded in parent
      known_data = { "label" => "my pet", "animal" => { "type" => "cat", "name" => "whiskers" } }
      known_result = TestOpenUnionModels::Container.from_dict(known_data)
      refute Crystalline.unknown?(known_result.animal)
      assert_instance_of TestOpenUnionModels::Cat, known_result.animal
      assert_equal "whiskers", known_result.animal.name

      # Unknown variant embedded in parent
      unknown_data = { "label" => "mystery", "animal" => { "type" => "bird", "wingspan" => 5 } }
      unknown_result = TestOpenUnionModels::Container.from_dict(unknown_data)
      assert Crystalline.unknown?(unknown_result.animal)

      # List of mixed known and unknown
      list_data = [
        { "type" => "cat", "name" => "whiskers" },
        { "type" => "bird", "wingspan" => 5 },
        { "type" => "dog", "age" => 3 }
      ]
      list_result = list_data.map { |item| animal_union.parse(item) }
      assert_equal 3, list_result.length
      refute Crystalline.unknown?(list_result[0])
      assert Crystalline.unknown?(list_result[1])
      refute Crystalline.unknown?(list_result[2])
    end
  end
end
