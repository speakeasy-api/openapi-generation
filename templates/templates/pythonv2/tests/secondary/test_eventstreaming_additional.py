import json
from openapi import SDK
from openapi.models.shared import (
    SseEvent,
    OptionalDataEvent,
    OptionalDataEventPayload,
    JSONEvent,
    JSONEventData,
)
from openapi.utils.unmarshal_json_response import sse_event_data
from typing import List, Union
from .common_helpers import record_test, HTTPBIN_URL


def test_event_stream_malformed_frame_yields_without_raising():
    """Disabled response validation: a malformed mid-stream frame is yielded
    best-effort (the raw decoded JSON) instead of raising. A typed stream never
    surfaces a validation error when validation is disabled."""
    s = SDK(server_url=HTTPBIN_URL)
    result = s.eventstreams.malformed_stream()

    events: List[Union[JSONEvent, dict]] = []
    with result as event_stream:
        for event in event_stream:  # must not raise mid-iteration
            events.append(event)

    assert len(events) == 2
    # Valid lead frame decodes to a typed model.
    assert events[0] == JSONEvent(data=JSONEventData(content="Hello"))
    # Unparseable frame is passed through as the raw decoded dict.
    assert events[1] == {"data": {"unexpected": True}}


def test_sse_event_data_reads_model_or_raw_dict():
    """With response-schema validation disabled, an unparseable SSE frame
    decodes to a raw dict instead of a typed model; sse_event_data must extract
    the inner `data` from either so flat-response streams never raise."""
    # Valid frame -> typed model: read the `.data` attribute.
    payload = OptionalDataEventPayload(content="hi")
    assert sse_event_data(OptionalDataEvent(event="message", data=payload)) is payload
    # Disabled-mode unparseable frame -> raw dict: fall back to key access.
    assert sse_event_data({"data": {"content": "hi"}}) == {"content": "hi"}
    # Missing `data` key -> None, not KeyError.
    assert sse_event_data({"event": "x"}) is None


def test_event_stream_wpt_compliance():
    record_test("event-stream-wpt-compliance")

    s = SDK(server_url=HTTPBIN_URL)
    result = s.eventstreams.wpt_compliance()

    assert result is not None

    actual: List[SseEvent] = []
    expected: List[SseEvent] = []

    with result as event_stream:
        for event in event_stream:
            if event.event == "expected":
                raw = json.loads(event.data)
                expected = [SseEvent(**e) for e in raw]
            else:
                actual.append(event)

    assert len(expected) > 0, "no expectations received from server"
    assert actual == expected


def test_event_stream_optional_data_field():
    record_test("event-stream-optional-data-field")

    s = SDK(server_url=HTTPBIN_URL)
    result = s.eventstreams.optional_data()

    assert result is not None

    events: List[OptionalDataEvent] = []
    with result as event_stream:
        for event in event_stream:
            events.append(event)

    assert events == [
        OptionalDataEvent(event="message", id="event-1", data=OptionalDataEventPayload(content="Hello, this event has data")),
        OptionalDataEvent(event="heartbeat", id="event-2"),
        OptionalDataEvent(event="message", id="event-3", data=OptionalDataEventPayload(content="Another message with data")),
        OptionalDataEvent(event="ping", id="event-4"),
        OptionalDataEvent(event="complete", id="event-5", data=OptionalDataEventPayload(content="Stream finished")),
    ]
