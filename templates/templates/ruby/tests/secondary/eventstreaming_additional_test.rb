# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "json"
require "minitest/autorun"
require "minitest/focus"

module AlphabeticallyEarly
  class TestEventStreaming < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_event_stream_wpt_compliance
      record_test("event-stream-wpt-compliance")

      result = @sdk.eventstreams.wpt_compliance

      refute_nil(result)

      actual = []
      expected = []

      result.each do |event|
        if event.event == "expected"
          raw = JSON.parse(event.data)
          expected = raw.map do |e|
            Models::Shared::SseEvent.new(
              data: e["data"],
              event: e["event"],
              id: e["id"],
              retry_: e["retry"]
            )
          end
        else
          actual << event
        end
      end

      refute_empty(expected, "no expectations received from server")
      assert_equal(expected, actual)
    end

    def test_event_stream_optional_data_field
      record_test("event-stream-optional-data-field")

      result = @sdk.eventstreams.optional_data

      refute_nil(result)

      events = []
      result.each { |event| events << event }

      expected = [
        Models::Shared::OptionalDataEvent.new(event: "message", id: "event-1", data: Models::Shared::OptionalDataEventPayload.new(content: "Hello, this event has data")),
        Models::Shared::OptionalDataEvent.new(event: "heartbeat", id: "event-2"),
        Models::Shared::OptionalDataEvent.new(event: "message", id: "event-3", data: Models::Shared::OptionalDataEventPayload.new(content: "Another message with data")),
        Models::Shared::OptionalDataEvent.new(event: "ping", id: "event-4"),
        Models::Shared::OptionalDataEvent.new(event: "complete", id: "event-5", data: Models::Shared::OptionalDataEventPayload.new(content: "Stream finished"))
      ]

      assert_equal(expected, events)
    end
  end
end
