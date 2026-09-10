# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"

module AlphabeticallyEarly
  class TestJsonlStreaming < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_jsonl_stream_data
      record_test("jsonl-stream-data-flat-responses")

      res = @sdk.jsonl.jsonl_stream

      refute_nil(res)
      assert_instance_of(AlphabeticallyEarly::Utils::JsonLStream, res)

      result = []
      res.each { |item| result << item }

      assert_equal(2, result.length)

      assert_equal("Peter", result[0].name)
      assert_equal(["Go", "Python"], result[0].skills)

      assert_equal("John", result[1].name)
      assert_equal(["Go", "Rust"], result[1].skills)
    end

    def test_jsonl_stream_data_chunks
      record_test("jsonl-stream-data-chunks-flat-response")

      res = @sdk.jsonl.jsonl_stream_chunks

      refute_nil(res)
      assert_instance_of(AlphabeticallyEarly::Utils::JsonLStream, res)

      result = []
      res.each { |item| result << item }

      assert_equal(2, result.length)

      assert_equal("Peter", result[0].name)
      assert_equal(["Go", "Python"], result[0].skills)

      assert_equal("John", result[1].name)
      assert_equal(["Go", "Rust"], result[1].skills)
    end
  end
end
