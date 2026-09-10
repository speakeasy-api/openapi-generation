# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "json"

module OpenApiSDK
  class TestUnionAdditional < Minitest::Test

    def test_strongly_typed_one_of_discriminated_post_res_deserialization_tag1
      # Test deserialization with tag1 discriminator
      json_string = {
        json: {
          imageURL: "https://example.com/image.png",
          tag: "tag1"
        }
      }.to_json

      json_obj = JSON.parse(json_string)
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::StronglyTypedOneOfDiscriminatedPostRes)

      refute_nil(result)
      refute_nil(result.json)
      assert_instance_of(OpenApiSDK::Models::Shared::TaggedObject1, result.json)
      assert_equal("https://example.com/image.png", result.json.image_url)
      assert_equal(OpenApiSDK::Models::Shared::TaggedObject1Tag::TAG1, result.json.tag)
    end

    def test_strongly_typed_one_of_discriminated_post_res_deserialization_tag2
      # Test deserialization with tag2 discriminator
      json_string = {
        json: {
          profileId: "user-123",
          tag: "tag2"
        }
      }.to_json

      json_obj = JSON.parse(json_string)
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::StronglyTypedOneOfDiscriminatedPostRes)

      refute_nil(result)
      refute_nil(result.json)
      assert_instance_of(OpenApiSDK::Models::Shared::TaggedObject2, result.json)
      assert_equal("user-123", result.json.profile_id)
      assert_equal(OpenApiSDK::Models::Shared::TaggedObject2Tag::TAG2, result.json.tag)
    end

    def test_strongly_typed_one_of_discriminated_post_res_deserialization_tag3
      # Test deserialization with tag3 discriminator
      json_string = {
        json: {
          phone: "+1-555-123-4567",
          tag: "tag3"
        }
      }.to_json

      json_obj = JSON.parse(json_string)
      result = Crystalline.unmarshal_json(json_obj, Models::Operations::StronglyTypedOneOfDiscriminatedPostRes)

      refute_nil(result)
      refute_nil(result.json)
      assert_instance_of(OpenApiSDK::Models::Shared::TaggedObject3, result.json)
      assert_equal("+1-555-123-4567", result.json.phone)
      assert_equal("tag3", result.json.tag)
    end
  end
end
