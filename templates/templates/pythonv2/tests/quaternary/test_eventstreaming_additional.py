from openapi import SDK
from openapi.models.shared import JSONEvent, JSONEventData
from typing import List
from .common_helpers import record_test, HTTPBIN_URL


def test_event_stream_malformed_frame_yields_lenient_typed_without_raising():
    """Lenient response validation: a malformed mid-stream SSE frame is
    best-effort constructed into the typed model (missing required field left
    None) and yielded, so the stream completes instead of raising. Contrast the
    disabled variant, which yields the raw decoded dict."""
    record_test("event-stream-malformed-frame-lenient")
    s = SDK(server_url=HTTPBIN_URL)
    res = s.eventstreams.malformed_stream()

    events: List[JSONEvent] = []
    with res.json_event as event_stream:
        for event in event_stream:  # must not raise mid-iteration
            events.append(event)

    assert len(events) == 2
    # Valid lead frame decodes to a typed model.
    assert events[0] == JSONEvent(data=JSONEventData(content="Hello"))
    # Malformed frame is best-effort constructed: still a typed model, with the
    # missing required `content` defaulted to None.
    assert isinstance(events[1], JSONEvent)
    assert events[1].data.content is None
