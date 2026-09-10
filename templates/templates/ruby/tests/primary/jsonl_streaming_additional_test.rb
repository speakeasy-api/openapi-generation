# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "json"
require "rack"
require "stringio"

module OpenApiSDK
  class TestJsonlStreaming < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_jsonl_stream_data
      record_test("jsonl-stream-data-envelope-http-responses")
      record_test("jsonl-stream-data-async-envelope-http-responses")

      res = @sdk.jsonl.jsonl_stream

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.object)
      assert_instance_of(OpenApiSDK::Utils::JsonLStream, res.object)

      json_events = []
      res.object.each { |event| json_events << event }

      assert_equal(2, json_events.length)

      assert_equal("Peter", json_events[0].name)
      assert_equal(["Go", "Python"], json_events[0].skills)

      assert_equal("John", json_events[1].name)
      assert_equal(["Go", "Rust"], json_events[1].skills)
    end

    def test_jsonl_stream_data_chunks
      record_test("jsonl-stream-data-chunks-envelope-http-responses")

      res = @sdk.jsonl.jsonl_stream_chunks

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.object)
      assert_instance_of(OpenApiSDK::Utils::JsonLStream, res.object)

      json_events = []
      res.object.each { |event| json_events << event }

      assert_equal(2, json_events.length)

      assert_equal("Peter", json_events[0].name)
      assert_equal(["Go", "Python"], json_events[0].skills)

      assert_equal("John", json_events[1].name)
      assert_equal(["Go", "Rust"], json_events[1].skills)
    end

    def test_jsonl_deserialization_with_camel_case_properties
      record_test("jsonl-deserialization-camel-case-properties")

      res = @sdk.jsonl.jsonl_deserialization_verification

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.object)
      assert_instance_of(OpenApiSDK::Utils::JsonLStream, res.object)

      first_event = res.object.first
      refute_nil(first_event)
      assert_equal("yes", first_event.is_finished)
    end

    def test_jsonl_raises_mid_stream_error
      response = ::Faraday::Response.new
      response.finish(
        status: 200,
        response_headers: { "Content-Type" => "application/jsonl" },
        body: StringIO.new("{\"name\":\"Peter\"}\n")
      )
      response.env[:stream_error] = StandardError.new("simulated jsonl stream failure")

      stream = OpenApiSDK::Utils::JsonLStream.new(
        response,
        ->(raw) { JSON.parse(raw) },
        sdk_ref: self
      )

      events = []
      err = assert_raises(StandardError) do
        stream.each { |event| events << event }
      end

      assert_equal("simulated jsonl stream failure", err.message)
      assert_equal(1, events.length)
    end
  end
end
