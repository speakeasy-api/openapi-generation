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
require "timeout"

module OpenApiSDK
  class TestEventStreaming < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_event_stream_json_data
      record_test("event-stream-json-data")

      res = @sdk.eventstreams.json

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.json_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.json_event)

      json_events = []
      res.json_event.each { |event| json_events << event }

      assert_equal(4, json_events.length)

      message = ""
      json_events.each { |event| message += event.data.content }

      assert_equal("Hello world!", message)
    end

    def test_event_stream_text_data
      record_test("event-stream-text-data")

      res = @sdk.eventstreams.text

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.text_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.text_event)

      text_events = []
      res.text_event.each { |event| text_events << event }

      assert_equal(4, text_events.length)

      message = ""
      text_events.each { |event| message += event.data }

      assert_equal("Hello world!", message)
    end

    def test_event_stream_multiline_data
      record_test("event-stream-multiline-data")

      res = @sdk.eventstreams.multiline

      refute_nil(res)
      refute_nil(res.text_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.text_event)

      text_events = []
      res.text_event.each { |event| text_events << event }

      assert_equal(1, text_events.length)
      assert_equal("YHOO\n+2\n10", text_events[0].data)
    end

    def test_event_stream_rich_events
      record_test("event-stream-rich-events")

      res = @sdk.eventstreams.rich

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.rich_stream)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.rich_stream)

      rich_events = []
      res.rich_stream.each { |event| rich_events << event }

      assert_equal(3, rich_events.length)

      assert_instance_of(Models::Shared::RichCompletionEvent, rich_events[0])
      assert_equal("job-1", rich_events[0].id)
      assert_equal("Hello", rich_events[0].data.completion)
      assert_equal("jeeves-1", rich_events[0].data.model)
      assert_nil(rich_events[0].data.stop_reason)

      assert_instance_of(Models::Shared::HeartbeatEvent, rich_events[1])
      assert_equal("ping", rich_events[1].data)
      assert_equal(3000, rich_events[1].retry_)

      assert_instance_of(Models::Shared::RichCompletionEvent, rich_events[2])
      assert_equal("job-1", rich_events[2].id)
      assert_equal("world!", rich_events[2].data.completion)
      assert_equal("jeeves-1", rich_events[2].data.model)
      refute_nil(rich_events[2].data.stop_reason)
    end

    def test_event_stream_with_sentinel_events
      record_test("event-stream-chat-sentinel-event")

      res = @sdk.eventstreams.chat(
        request: Models::Operations::ChatRequestBody.new(prompt: "Print test content")
      )

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.chat_completion_stream)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.chat_completion_stream)

      chat_events = []
      res.chat_completion_stream.each { |event| chat_events << event }

      assert_equal(5, chat_events.length)

      assert_instance_of(Models::Shared::ChatCompletionEvent, chat_events[0])
      assert_equal("Hello", chat_events[0].data.content)
      assert_equal(" ", chat_events[1].data.content)
      assert_equal("world", chat_events[2].data.content)
      assert_equal("!", chat_events[3].data.content)
      assert_instance_of(Models::Shared::SentinelEvent, chat_events[4])
    end

    def test_event_stream_skip_sentinel
      record_test("event-stream-chat-skip-sentinel")

      res = @sdk.eventstreams.chat_skip_sentinel(
        request: Models::Operations::ChatSkipSentinelRequestBody.new(prompt: "Print test content")
      )

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.chat_completion_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.chat_completion_event)

      chat_events = []
      res.chat_completion_event.each { |event| chat_events << event }

      assert_equal(4, chat_events.length)

      assert_equal("Hello", chat_events[0].data.content)
      assert_equal(" ", chat_events[1].data.content)
      assert_equal("world", chat_events[2].data.content)
      assert_equal("!", chat_events[3].data.content)
    end

    def test_event_stream_with_different_data_schemas
      record_test("event-stream-different-data-schemas")

      res = @sdk.eventstreams.different_data_schemas

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.different_data_schemas)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.different_data_schemas)

      events = []
      res.different_data_schemas.each { |event| events << event }

      assert_equal(6, events.length)

      assert_instance_of(Models::Shared::MessageEvent, events[0].data)
      assert_equal(123, events[0].data.id)
      assert_equal("Here is your url", events[0].data.content)

      assert_instance_of(Models::Shared::UrlEvent, events[1].data)
      assert_equal("https://example.com", events[1].data.url)

      assert_instance_of(Models::Shared::MessageEvent, events[2].data)
      assert_equal("Have a great day!", events[2].data.content)

      assert_equal([1, 2, 3, 4], events[3].data)
      assert(events[4].data)
      assert_in_delta(3.14159, events[5].data)
    end

    def test_event_stream_mixed_data
      record_test("event-stream-mixed-data")

      res = @sdk.eventstreams.mixed_data

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.mixed_data_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.mixed_data_event)

      events = []
      res.mixed_data_event.each { |event| events << event }

      assert_equal(4, events.length)

      # First event: JSON object (completion)
      assert_instance_of(Models::Shared::MessageEvent, events[0].data)
      assert_equal("Hello world", events[0].data.content)

      # Second event: plain text (text)
      assert_equal("Processing your request...", events[1].data)

      # Third event: plain text (loading)
      assert_equal("Almost done", events[2].data)

      # Fourth event: JSON object (completion)
      assert_instance_of(Models::Shared::MessageEvent, events[3].data)
      assert_equal("Done!", events[3].data.content)
    end

    def test_event_stream_error_response
      record_test("event-stream-error-response")

      client = Faraday.new(headers: { "x-teapot" => "json" }) do |f|
        f.request(:multipart, {})
        f.adapter Faraday.default_adapter
      end

      sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url, client: client)
      refute_nil(sdk)

      assert_raises(Models::Errors::TeapotJSONError) do
        sdk.eventstreams.text
      end
    end

    def test_event_stream_partial_with_comments
      record_test("event-stream-partial-with-comments")

      res = @sdk.eventstreams.partial_with_comments

      refute_nil(res)
      refute_nil(res.partial_with_comments_event)

      actual = []
      res.partial_with_comments_event.each { |event| actual << event }

      assert_equal(5, actual.length)

      assert_equal("Hello from SSE", actual[0].data.message)
      assert_nil(actual[0].data.status)

      assert_equal("processing", actual[1].data.status)
      assert_equal(50, actual[1].data.progress)

      assert_equal("complete", actual[2].data.status)
      assert_equal(100, actual[2].data.progress)
      assert_equal("Success", actual[2].data.result)

      assert_equal("mixed boundaries", actual[3].data.test)

      assert_equal("test", actual[4].data.another)
    end

    def test_event_stream_union_with_standalone_comments
      record_test("event-stream-union-with-standalone-comments")

      res = @sdk.eventstreams.union_with_comments

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.union_with_comments_stream)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.union_with_comments_stream)

      events = []
      res.union_with_comments_stream.each { |event| events << event }

      assert_equal(4, events.length)

      assert_instance_of(Models::Shared::StatusEvent, events[0])
      assert_equal("started", events[0].data.status)
      assert_equal("Initializing", events[0].data.message)

      assert_instance_of(Models::Shared::ProgressEvent, events[1])
      assert_equal(50, events[1].data.percent)
      assert_equal("Half done", events[1].data.detail)

      assert_instance_of(Models::Shared::StatusEvent, events[2])
      assert_equal("running", events[2].data.status)
      assert_equal("Processing items", events[2].data.message)

      assert_instance_of(Models::Shared::ProgressEvent, events[3])
      assert_equal(100, events[3].data.percent)
      assert_equal("Complete", events[3].data.detail)
    end

    def test_event_stream_chat_heartbeat_skips_dataless
      record_test("event-stream-chat-heartbeat-skips-dataless")

      res = @sdk.eventstreams.chat_heartbeat

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(200, res.http_meta.response.status)
      refute_nil(res.chat_completion_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.chat_completion_event)

      events = []
      res.chat_completion_event.each { |event| events << event }

      # Server sends: data("Hello1"), heartbeat (skipped), data("Hello 2"), data("!"), [DONE]
      assert_equal(3, events.length)

      assert_equal("Hello1", events[0].data.content)
      assert_equal("Hello 2", events[1].data.content)
      assert_equal("!", events[2].data.content)
    end

    def test_event_stream_stay_open_breaking_early
      record_test("event-stream-stay-open-break-early")

      res = nil
      Timeout.timeout(2) do
        res = @sdk.eventstreams.stay_open
      end

      refute_nil(res)
      refute_nil(res.text_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.text_event)

      # Using .first internally iterates via each, which has ensure { close }.
      # This marks the stream as closed after the first event is consumed.
      first_event = res.text_event.first
      refute_nil(first_event)

      # After .first, the stream is closed. Attempting to iterate again
      # returns immediately with no events (like Python's StopIteration).
      events = []
      res.text_event.each { |event| events << event }
      assert_empty(events)
    end

    def test_event_stream_raises_mid_stream_error
      response = ::Faraday::Response.new
      response.finish(
        status: 200,
        response_headers: { "Content-Type" => "text/event-stream" },
        body: StringIO.new("data: {\"content\":\"ok\"}\n\n")
      )
      response.env[:stream_error] = StandardError.new("simulated stream failure")

      stream = OpenApiSDK::Utils::EventStream.new(
        response,
        ->(raw) { JSON.parse(raw) },
        sdk_ref: self
      )

      events = []
      err = assert_raises(StandardError) do
        stream.each { |event| events << event }
      end

      assert_equal("simulated stream failure", err.message)
      assert_equal(1, events.length)
    end

    def test_event_stream_stay_open_sentinel_detection
      record_test("event-stream-stay-open-sentinel")

      res = @sdk.eventstreams.stay_open

      refute_nil(res)
      refute_nil(res.text_event)
      assert_instance_of(OpenApiSDK::Utils::EventStream, res.text_event)

      events = []
      res.text_event.each { |event| events << event }

      assert_equal(4, events.length)

      assert_equal("event 1", events[0].data)
      assert_equal("event 2", events[1].data)
      assert_equal("event 3", events[2].data)
      assert_equal("event 4", events[3].data)
    end

    def test_event_stream_large_event_small_chunks
      record_test("event-stream-large-event-small-chunks")

      size = 2 * 1024 * 1024
      contents = []
      Timeout.timeout(60) do
        res = @sdk.eventstreams.large_event_small_chunks(size: size)

        refute_nil(res)
        assert_equal(200, res.http_meta.response.status)
        refute_nil(res.large_chunked_event)

        res.large_chunked_event.each { |event| contents << event.data.content }
      end

      assert_equal(3, contents.length)
      assert_equal("start", contents[0])
      assert_equal("end", contents[2])

      big = contents[1]
      assert_equal(size, big.length)
      assert_equal("S", big[0])
      assert_equal("E", big[-1])
      assert_equal("a" * (size - 2), big[1..-2])
    end

    def test_event_stream_split_boundaries
      record_test("event-stream-split-boundaries")

      actual_tags = []
      expected_tags = []
      Timeout.timeout(60) do
        res = @sdk.eventstreams.split_boundaries

        refute_nil(res)
        assert_equal(200, res.http_meta.response.status)
        refute_nil(res.split_boundary_event)

        res.split_boundary_event.each do |event|
          if event.data.kind.serialize == "expected"
            expected_tags = event.data.tags
          else
            actual_tags.concat(event.data.tags)
          end
        end
      end

      refute_empty(expected_tags)
      assert_equal(expected_tags, actual_tags)
    end
  end
end
