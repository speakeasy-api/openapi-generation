# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "json"

module OpenApiSDK
  class TestSmartUnions < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    # --- smart-union-open-enums ---
    # Types: { kind: OpenEnum["cat"] } | { kind: OpenEnum["dog"] }
    # Payload: { kind: "dog" }
    # Expect: SmartUnionOpenEnumsDog (exact enum match)
    def test_smart_union_open_enums
      record_test("smart-union-open-enums")

      req = Models::Shared::SmartUnionOpenEnumsDog.new(
        kind: Models::Shared::SmartUnionOpenEnumsDogKind::DOG
      )
      res = @sdk.unions.smart_union_open_enums(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionOpenEnumsDog, res.res.json)
    end

    # --- smart-union-open-enums-and-size ---
    # Types: { kind: OpenEnum["dog"] } | { kind: OpenEnum["cat"], name: string }
    # Payload: { kind: "bat", name: "asdf" }
    # Expect: CatWithName (has more matched fields despite unrecognized enum)
    def test_smart_union_open_enums_and_size
      record_test("smart-union-open-enums-and-size")

      json_obj = JSON.parse("{\"json\": {\"kind\": \"bat\", \"name\": \"asdf\"}}")
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::SmartUnionOpenEnumsAndSizeRes)
      refute_nil(result)
      assert_instance_of(Models::Shared::SmartUnionOpenEnumsAndSizeCatWithName, result.json)
      assert_equal("asdf", result.json.name)
    end

    # --- smart-union-deeply-nested-array ---
    # Types: Array<{ x: { a: string } }> | Array<{ x: { a: string, b: boolean } }>
    # Payload: [{ x: { a: "", b: false } }]
    # Expect: array deserialized successfully
    def test_smart_union_deeply_nested_array
      record_test("smart-union-deeply-nested-array")

      json_obj = JSON.parse("{\"json\": [{\"x\": {\"a\": \"\", \"b\": false}}]}")
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::SmartUnionDeeplyNestedArrayRes)
      refute_nil(result)
      refute_nil(result.json)
      assert_equal(1, result.json.length)
    end

    # --- smart-union-empty-string ---
    # Types: { a: string } | { b: string }
    # Payload: { b: "" }
    # Expect: SmartUnionEmptyStringObjectB (field b matches)
    def test_smart_union_empty_string
      record_test("smart-union-empty-string")

      req = Models::Shared::SmartUnionEmptyStringObjectB.new(b: "")
      res = @sdk.unions.smart_union_empty_string(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionEmptyStringObjectB, res.res.json)
      assert_equal("", res.res.json.b)
    end

    # --- smart-union-nested-union ---
    # Outer: { data: InnerUnion } | { data: { kind: OpenEnum, name: OpenEnum } }
    # InnerUnion: { kind: OpenEnum["cat"] } | { kind: OpenEnum["dog"] } | { kind: OpenEnum["bird"] }
    # Payload: { data: { kind: "unknown", name: "also_unknown" } }
    # Expect: OuterBWrapper (has 2 matched fields vs inner union's 1)
    def test_smart_union_nested_union
      record_test("smart-union-nested-union")

      json_obj = JSON.parse("{\"json\": {\"data\": {\"kind\": \"unknown\", \"name\": \"also_unknown\"}}}")
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::SmartUnionNestedUnionRes)
      refute_nil(result)
      assert_instance_of(Models::Operations::SmartUnionNestedOuterBWrapper, result.json)
    end

    # --- unions-discriminated-open-enum ---
    # Discriminator on "status" field with open enum values
    # active/pending -> ObjectWithOpenEnumStatus1, inactive/archived -> ObjectWithOpenEnumStatus2
    def test_unions_discriminated_open_enum
      record_test("unions-discriminated-open-enum")

      # Step 1: active status -> Status1
      req_active = Models::Shared::ObjectWithOpenEnumStatus1.new(
        status: Models::Shared::ObjectWithOpenEnumStatus1Status::ACTIVE,
        user_id: "user-123",
        active_at: DateTime.parse("2024-01-15T10:30:00Z")
      )
      res = @sdk.unions.discriminated_open_enum(request: req_active)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::ObjectWithOpenEnumStatus1, res.res.json)
      assert_equal("user-123", res.res.json.user_id)

      # Step 2: inactive status -> Status2
      req_inactive = Models::Shared::ObjectWithOpenEnumStatus2.new(
        status: Models::Shared::ObjectWithOpenEnumStatus2Status::INACTIVE,
        reason: "User requested deactivation",
        inactive_since: DateTime.parse("2024-01-10T15:45:00Z")
      )
      res = @sdk.unions.discriminated_open_enum(request: req_inactive)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::ObjectWithOpenEnumStatus2, res.res.json)
      assert_equal("User requested deactivation", res.res.json.reason)

      # Step 3: pending status -> Status1
      req_pending = Models::Shared::ObjectWithOpenEnumStatus1.new(
        status: Models::Shared::ObjectWithOpenEnumStatus1Status::PENDING,
        user_id: "user-456",
        active_at: DateTime.parse("2024-01-16T08:00:00Z")
      )
      res = @sdk.unions.discriminated_open_enum(request: req_pending)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::ObjectWithOpenEnumStatus1, res.res.json)
      assert_equal("user-456", res.res.json.user_id)

      # Step 4: archived status -> Status2
      req_archived = Models::Shared::ObjectWithOpenEnumStatus2.new(
        status: Models::Shared::ObjectWithOpenEnumStatus2Status::ARCHIVED,
        reason: "Account archived",
        inactive_since: DateTime.parse("2024-01-05T12:00:00Z")
      )
      res = @sdk.unions.discriminated_open_enum(request: req_archived)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::ObjectWithOpenEnumStatus2, res.res.json)
      assert_equal("Account archived", res.res.json.reason)
    end

    # --- smart-union-all-consts ---
    # Types: { a: const "A", b: const "B" } | { b: const "B", c: const "C" }
    # Payload: { b: "B", c: "C" }
    # Expect: SmartUnionAllConstsB (matches both b and c)
    def test_smart_union_all_consts
      record_test("smart-union-all-consts")

      req = Models::Shared::SmartUnionAllConstsB.new(b: "B", c: "C")
      res = @sdk.unions.smart_union_all_consts(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionAllConstsB, res.res.json)
    end

    # --- smart-union-any-field-type ---
    # Types: { a: optional string } | { b: any }
    # Payload: { b: "asdf" }
    # Expect: SmartUnionAnyFieldB (field b matches)
    def test_smart_union_any_field_type
      record_test("smart-union-any-field-type")

      req = Models::Shared::SmartUnionAnyFieldB.new(b: "asdf")
      res = @sdk.unions.smart_union_any_field_type(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionAnyFieldB, res.res.json)
    end

    # --- smart-union-array-fields ---
    # Types: Array<{ name: string }> | Array<{ name: string, value: string }>
    # Payload: [{ name: "a", value: "1" }, { name: "b", value: "2" }]
    # Expect: array deserialized with items
    def test_smart_union_array_fields
      record_test("smart-union-array-fields")

      json_obj = JSON.parse("{\"json\": [{\"name\": \"a\", \"value\": \"1\"}, {\"name\": \"b\", \"value\": \"2\"}]}")
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::SmartUnionArrayFieldsRes)
      refute_nil(result)
      refute_nil(result.json)
      assert_equal(2, result.json.length)
      assert_equal("a", result.json[0].name)
      assert_equal("b", result.json[1].name)
    end

    # --- smart-union-const-field-discrimination ---
    # Types: { a: const "x", b: const "1" } | { a: const "x", c: const "1" } | { b: const "1", c: const "1" }
    # Payload: { b: "1", c: "1" }
    # Expect: SmartUnionConstFieldC (matches both b and c)
    def test_smart_union_const_field_discrimination
      record_test("smart-union-const-field-discrimination")

      req = Models::Shared::SmartUnionConstFieldC.new(b: "1", c: "1")
      res = @sdk.unions.smart_union_const_field_discrimination(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionConstFieldC, res.res.json)
    end

    # --- smart-union-nested-structs ---
    # Types: { nested: { value: string } } | { nested: { value: string, extra: string } }
    # Payload: { nested: { value: "test", extra: "data" } }
    # Expect: SmartUnionNestedStructsB (nested has more fields)
    def test_smart_union_nested_structs
      record_test("smart-union-nested-structs")

      inner = Models::Shared::SmartUnionNestedStructsInnerB.new(value: "test", extra: "data")
      req = Models::Shared::SmartUnionNestedStructsB.new(nested: inner)
      res = @sdk.unions.smart_union_nested_structs(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionNestedStructsB, res.res.json)
      assert_equal("test", res.res.json.nested.value)
      assert_equal("data", res.res.json.nested.extra)
    end

    # --- smart-union-optional-pointer-fields ---
    # Types: { foo: optional string } | { foo: optional string, bar: optional string }
    # Payload: { foo: "test", bar: "value" }
    # Expect: SmartUnionOptionalPointerB (has more matched fields)
    def test_smart_union_optional_pointer_fields
      record_test("smart-union-optional-pointer-fields")

      req = Models::Shared::SmartUnionOptionalPointerB.new(foo: "test", bar: "value")
      res = @sdk.unions.smart_union_optional_pointer_fields(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionOptionalPointerB, res.res.json)
    end

    # --- smart-union-optional-pointer-structs ---
    # Types: { nested: optional { name: string } } | { nested: optional { name: string, value: string } }
    # Payload: { nested: { name: "test", value: "data" } }
    # Expect: SmartUnionOptionalPointerStructsB (nested has more fields)
    def test_smart_union_optional_pointer_structs
      record_test("smart-union-optional-pointer-structs")

      inner = Models::Shared::SmartUnionOptionalPointerStructsInnerB.new(name: "test", value: "data")
      req = Models::Shared::SmartUnionOptionalPointerStructsB.new(nested: inner)
      res = @sdk.unions.smart_union_optional_pointer_structs(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionOptionalPointerStructsB, res.res.json)
      assert_equal("test", res.res.json.nested.name)
      assert_equal("data", res.res.json.nested.value)
    end

    # --- smart-union-prefers-fewer-unmatched-fields ---
    # Types: { foo: string, bar: string } | { foo: string }
    # Payload: { foo: "test" }
    # Expect: SmartUnionFewerUnmatchedB (fewer unmatched fields)
    def test_smart_union_prefers_fewer_unmatched_fields
      record_test("smart-union-prefers-fewer-unmatched-fields")

      req = Models::Shared::SmartUnionFewerUnmatchedB.new(foo: "test")
      res = @sdk.unions.smart_union_prefers_fewer_unmatched_fields(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionFewerUnmatchedB, res.res.json)
      assert_equal("test", res.res.json.foo)
    end

    # --- smart-union-preserves-order-on-tie ---
    # Types: { foo: string } | { foo: string }
    # Payload: { foo: "test" }
    # Expect: SmartUnionTieA (first wins on tie)
    def test_smart_union_preserves_order_on_tie
      record_test("smart-union-preserves-order-on-tie")

      req = Models::Shared::SmartUnionTieA.new(foo: "test")
      res = @sdk.unions.smart_union_preserves_order_on_tie(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionTieA, res.res.json)
      assert_equal("test", res.res.json.foo)
    end

    # --- smart-union-selects-more-matched-fields ---
    # Types: { foo: string } | { foo: string, bar: string }
    # Payload: { foo: "test", bar: "value" }
    # Expect: SmartUnionMoreFieldsB (matches both foo and bar)
    def test_smart_union_selects_more_matched_fields
      record_test("smart-union-selects-more-matched-fields")

      req = Models::Shared::SmartUnionMoreFieldsB.new(foo: "test", bar: "value")
      res = @sdk.unions.smart_union_selects_more_matched_fields(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionMoreFieldsB, res.res.json)
      assert_equal("test", res.res.json.foo)
      assert_equal("value", res.res.json.bar)
    end

    # --- smart-union-three-way-field-discrimination ---
    # Types: { a: string, b: string } | { a: string, c: string } | { b: string, c: string }
    # Payload: { b: "test", c: "value" }
    # Expect: SmartUnionThreeWayC (matches both b and c)
    def test_smart_union_three_way_field_discrimination
      record_test("smart-union-three-way-field-discrimination")

      req = Models::Shared::SmartUnionThreeWayC.new(b: "test", c: "value")
      res = @sdk.unions.smart_union_three_way_field_discrimination(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionThreeWayC, res.res.json)
      assert_equal("test", res.res.json.b)
      assert_equal("value", res.res.json.c)
    end

    # --- smart-union-nested-union-vs-flat-struct ---
    # Outer: { data: Union[x-only | y-only | z-only] } | { data: { x: string, y: string } }
    # Payload: { data: { x: "", y: "" } }
    # Expect: OuterB (flat struct matches 2 fields vs nested union's 1)
    def test_smart_union_nested_union_vs_flat_struct
      record_test("smart-union-nested-union-vs-flat-struct")

      data = Models::Shared::SmartUnionNestedVsFlatB.new(x: "", y: "")
      req = Models::Operations::SmartUnionNestedVsFlatOuterB.new(data: data)
      res = @sdk.unions.smart_union_nested_union_vs_flat_struct(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(
        Models::Operations::SmartUnionNestedUnionVsFlatStructSmartUnionNestedVsFlatOuterB,
        res.res.json
      )
    end

    # --- smart-union-union-vs-union ---
    # Outer: Union[{ a } | { b }] | Union[{ a, b } | { a, b, c }]
    # Payload: { a: "", b: "", c: "" }
    # Expect: SmartUnionVsUnionB2 (matches all 3 fields)
    def test_smart_union_union_vs_union
      record_test("smart-union-union-vs-union")

      req = Models::Shared::SmartUnionVsUnionB2.new(a: "", b: "", c: "")
      res = @sdk.unions.smart_union_union_vs_union(request: req)
      refute_nil(res)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.res)
      assert_instance_of(Models::Shared::SmartUnionVsUnionB2, res.res.json)
    end
  end
end
