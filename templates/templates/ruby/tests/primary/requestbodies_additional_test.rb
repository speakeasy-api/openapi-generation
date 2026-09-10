# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "base64"
require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestRequestBodiesAdditional < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_request_bodies_base64_file_input_idempotent
      record_test("request-bodies-base64-file-input-idempotent")

      pre_encoded = Base64.strict_encode64(
        [0xff, 0xfe, 0x00, 0x01, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0xc3, 0x28].pack("C*")
      )

      request = Models::Shared::Base64InputFileModeRequest.new(
        data_byte: pre_encoded,
        data_content_encoding: pre_encoded,
        data_plain: "plain-idempotent"
      )

      first = @sdk.request_bodies.post_base64_input_mode(request: request)
      refute_nil(first)
      assert_equal(Rack::Utils.status_code(:ok), first.http_meta.response.status)
      refute_nil(first.res)
      assert_equal(pre_encoded, first.res.json.data_byte)
      assert_equal(pre_encoded, first.res.json.data_content_encoding)
      assert_equal("plain-idempotent", first.res.json.data_plain)

      second = @sdk.request_bodies.post_base64_input_mode(request: request)
      refute_nil(second)
      assert_equal(Rack::Utils.status_code(:ok), second.http_meta.response.status)
      refute_nil(second.res)
      assert_equal(first.res.json, second.res.json)
    end
  end
end
