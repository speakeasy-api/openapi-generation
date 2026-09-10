# frozen_string_literal: true

require_relative "../lib/openapi"
require "date"
require "uri"

def create_simple_object
  OpenApiSDK::Models::Shared::SimpleObject.new(
    str_: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32_enum: OpenApiSDK::Models::Shared::Int32Enum::FIFTY_FIVE,
    int_enum: OpenApiSDK::Models::Shared::IntEnum::SECOND,
    num: 1.1,
    float32: 1.1,
    enum: OpenApiSDK::Models::Shared::Enum::ONE,
    any: "any",
    date: Date.new(2020, 1, 1),
    date_time: DateTime.new(2020, 1, 1, 0, 0, 1 / 1_000_000_000.0),
    bool_opt: true,
    str_opt: "testOptional",
    int_opt_null: nil,
    num_opt_null: nil
  )
end

def create_simple_object_with_type
  OpenApiSDK::Models::Shared::SimpleObjectWithType.new(
    str_: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32_enum: OpenApiSDK::Models::Shared::SimpleObjectWithTypeInt32Enum::FIFTY_FIVE,
    int_enum: OpenApiSDK::Models::Shared::SimpleObjectWithTypeIntEnum::SECOND,
    num: 1.1,
    float32: 1.1,
    enum: OpenApiSDK::Models::Shared::Enum::ONE,
    any: "any",
    date: Date.new(2020, 1, 1),
    date_time: DateTime.new(2020, 1, 1, 0, 0, 1 / 1_000_000_000.0),
    bool_opt: true,
    str_opt: "testOptional",
    int_opt_null: nil,
    num_opt_null: nil,
    type: "simpleObjectWithType"
  )
end

def create_simple_object_with_non_standard_type_name
  OpenApiSDK::Models::Shared::SimpleObjectWithNonStandardTypeName.new(
    str_: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32_enum: OpenApiSDK::Models::Shared::SimpleObjectWithNonStandardTypeNameInt32Enum::FIFTY_FIVE,
    int_enum: OpenApiSDK::Models::Shared::SimpleObjectWithNonStandardTypeNameIntEnum::SECOND,
    num: 1.1,
    float32: 1.1,
    enum: OpenApiSDK::Models::Shared::Enum::ONE,
    any: "any",
    date: Date.new(2020, 1, 1),
    date_time: DateTime.new(2020, 1, 1, 0, 0, 1 / 1_000_000_000.0),
    bool_opt: true,
    str_opt: "testOptional",
    int_opt_null: nil,
    num_opt_null: nil,
    obj_type: "simpleObjectWithNonStandardTypeName"
  )
end

def create_deep_object
  OpenApiSDK::Models::Shared::DeepObject.new(
    any: create_simple_object,
    arr: [create_simple_object, create_simple_object],
    bool: true,
    int: 1,
    map: {"key" => create_simple_object},
    num: 1.1,
    obj: create_simple_object,
    str_: "test"
  )
end

def create_deep_object_with_type
  OpenApiSDK::Models::Shared::DeepObjectWithType.new(
    any: create_simple_object,
    arr: [create_simple_object, create_simple_object],
    bool: true,
    int: 1,
    map: {"key" => create_simple_object},
    num: 1.1,
    obj: create_simple_object,
    str_: "test",
    type: "deepObjectWithType"
  )
end

def unsure(to_replace, match)
  if match.include?("=")
    pairs = match.split(",")

  else
    values = match.split(",")
    pairs = []
    0.step(values.length - 1, 2) do |i|
      pairs[i / 2] = "#{values[i]},#{values[i + 1]}"
    end

  end

  pairs.sort_by { |pair| pair.split("=")[0] }

  updated_pairs = pairs.join(",")

  to_replace.sub(match, updated_pairs)
end

def sort_serialized_map(input, regex)
  input = URI.decode_www_form_component(input)

  r = Regexp.new(regex)

  replace_all_string_submatch_fx(r, input)
end

def replace_all_string_submatch_fx(re, str)
  result = ""
  last_index = 0
  matches = re.match(str)

  (1..matches.length - 1).each do |i|
    result += unsure(str[last_index..matches.end(i) - 1], matches[i])
    last_index = matches.end(i)
  end

  result + str[last_index..str.length]
end

def compare_hashes(hash1, hash2)
  hash1 = hash1.transform_keys(&:to_sym)
  hash2 = hash2.transform_keys(&:to_sym)
  hash1.each_key do |key|
    assert_equal(hash1[key], hash2[key])
  end
end

def end_to_end_object(in_obj)
  kls = in_obj.class
  json_val = in_obj.to_json
  return Crystalline.unmarshal_json(JSON.parse(json_val), kls)
end
