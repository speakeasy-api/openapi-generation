import gc
import json
import pytest
import httpx
from openapi import SDK
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.models.errors import TeapotJSONError, ResponseValidationError
from openapi.utils import *
from openapi.utils import eventstreaming

from .common_helpers import *
from .test_helpers import *

from typing import List, Union, cast
import signal


def test_parse_event_data_types():
    cases = [
        ('data: {"key": "value"}', {"key": "value"}),
        ("data: [1, 2, 3]", [1, 2, 3]),
        ("data: true", True),
        ("data: false", False),
        ("data: null", None),
        ("data: 123", 123),
        ("data: 3.14", 3.14),
        ('data: "hello"', "hello"),
        ("data: hello world", "hello world"),
        ("data: {invalid json", "{invalid json"),
        ("data:", ""),
        ("data: a\ndata: b", "a\nb"),
    ]
    for sse_block, expected in cases:
        raw = bytearray(sse_block.encode())
        result, _, _ = eventstreaming._parse_event(raw=raw, decoder=json.loads)
        assert result["data"] == expected, f"Failed for {sse_block!r}: got {result['data']!r}"


class _ChunkedStreamResponse:
    """Minimal httpx.Response stand-in for stream_events() unit tests.

    Yields a pre-baked sequence of byte chunks via iter_bytes / aiter_bytes
    so we can exercise the buffer-scan logic without a live HTTP connection.
    """

    def __init__(self, chunks):
        self._chunks = chunks
        self.closed = False

    def iter_bytes(self):
        for c in self._chunks:
            yield c

    async def aiter_bytes(self):
        for c in self._chunks:
            yield c

    def close(self):
        self.closed = True

    async def aclose(self):
        self.closed = True


@pytest.mark.parametrize(
    "chunks,description",
    [
        # 2-byte boundary (b"\n\n") split 1/1 across chunks
        ([b"data: x\n", b"\n"], "LF|LF split"),
        ([b"data: x\r", b"\r"], "CR|CR split"),
        # 3-byte boundary (b"\r\n\r"): split shapes 1/2 and 2/1
        ([b"data: x\r", b"\n\r"], "CR|LFCR (3-byte \\r\\n\\r split 1/2)"),
        ([b"data: x\r\n", b"\r"], "CRLF|CR (3-byte \\r\\n\\r split 2/1)"),
        # 4-byte boundary (b"\r\n\r\n"): every split shape
        ([b"data: x\r", b"\n\r\n"], "CR|LFCRLF (4-byte split 1/3)"),
        ([b"data: x\r\n", b"\r\n"], "CRLF|CRLF (4-byte split 2/2)"),
        ([b"data: x\r\n\r", b"\n"], "CRLFCR|LF (4-byte split 3/1)"),
        # Boundary split across many 1-byte chunks (worst-case fragmentation)
        ([b"data: x", b"\r", b"\n", b"\r", b"\n"], "4-byte boundary split 1/1/1/1"),
    ],
)
def test_event_stream_boundary_split_across_chunks(chunks, description):
    """Boundary detection must work when the delimiter straddles chunk seams.

    The buffer-scan loop skips bytes it already examined on prior chunks but it must still
    catch boundaries whose first bytes landed in the previous chunk by re-examining the
    trailing MAX_BOUNDARY_LEN-1 bytes when the next chunk arrives.
    """
    fake = _ChunkedStreamResponse(chunks)
    events = list(eventstreaming.stream_events(cast(httpx.Response, fake), decoder=json.loads))
    assert len(events) == 1, f"expected exactly one event for {description}: got {events}"
    assert events[0]["data"] == "x", f"data mismatch for {description}"


@pytest.mark.asyncio()
@pytest.mark.parametrize(
    "chunks",
    [
        [b"data: x\n", b"\n"],
        [b"data: x\r\n", b"\r\n"],
        [b"data: x\r\n\r", b"\n"],
        [b"data: x", b"\r", b"\n", b"\r", b"\n"],
    ],
)
async def test_event_stream_async_boundary_split_across_chunks(chunks):
    fake = _ChunkedStreamResponse(chunks)
    events = [e async for e in eventstreaming.stream_events_async(cast(httpx.Response, fake), decoder=json.loads)]
    assert len(events) == 1
    assert events[0]["data"] == "x"


def test_event_stream_large_payload_small_chunks_scan_is_linear(monkeypatch):
    """Buffer scan must stay linear when a single large event is split across
    many small chunks with no boundary visible until the final chunk.

    Each new chunk should only cause re-examination of MAX_BOUNDARY_LEN-1
    trailing bytes from the previously scanned buffer plus the new bytes
    themselves. A naive implementation that re-scans the whole accumulated
    buffer on every chunk grows the work quadratically with payload size.

    Payload deliberately seeds isolated CR bytes (no matching LF after) so
    that every CR forces an 8-boundary _peek_sequence sweep on each re-scan.
    """
    payload_size = 240_000
    chunk_size = 1500
    block = b"a" * 99 + b"\r"  # Insert a lone CR every 100 bytes.
    big_data = block * (payload_size // len(block))
    event_bytes = b"data: " + big_data + b"\n\n"
    chunks = [event_bytes[i : i + chunk_size] for i in range(0, len(event_bytes), chunk_size)]

    call_count = 0
    original_peek = eventstreaming._peek_sequence

    def counting_peek(position, buffer, sequence):
        nonlocal call_count
        call_count += 1
        return original_peek(position, buffer, sequence)

    monkeypatch.setattr(eventstreaming, "_peek_sequence", counting_peek)

    fake = _ChunkedStreamResponse(chunks)
    events = list(eventstreaming.stream_events(cast(httpx.Response, fake), decoder=json.loads))

    assert len(events) == 1
    # Linear: each CR sweeps all 8 boundaries once across the full payload.
    # Expected ~ (payload_size / 100) * 8 ~= 19k peek calls.
    # The MAX_BOUNDARY_LEN-1 chunk-overlap re-scan adds at most 3 bytes per chunk
    # so the 2x cushion accounts for that plus any minor overhead from the main loop.
    # In contrast, a quadratic implementation would ~3M peeks for this payload.
    cr_count = big_data.count(b"\r")
    linear_upper_bound = cr_count * len(eventstreaming.MESSAGE_BOUNDARIES) * 2
    assert call_count <= linear_upper_bound, (
        f"_peek_sequence called {call_count} times for {payload_size}-byte payload "
        f"in {len(chunks)} chunks; linear upper bound is {linear_upper_bound}"
    )


def test_parse_event_dataless_heartbeat_skipped():
    """Data-less SSE events (heartbeats) are skipped when data is required.

    ChatCompletionEvent has `data: ChatCompletionEventData` as a required Pydantic field.
    An SSE heartbeat event ("event: heartbeat") has no data: line. The parser skips
    data-less events when data_required=True (default) to prevent deserialization crashes.
    """

    def chat_decoder(json_str):
        return ChatCompletionEvent.model_validate(json.loads(json_str))

    raw = bytearray(b"event: heartbeat")
    event, discard, _ = eventstreaming._parse_event(raw=raw, decoder=chat_decoder)
    assert event is None, "data-less event should be skipped"
    assert discard is False


def test_event_stream_malformed_frame_raises_response_validation_error():
    """End-to-end: a live SSE stream whose mid-stream frame violates the event
    schema surfaces ResponseValidationError during iteration, not a raw
    pydantic.ValidationError. The server yields one valid frame then a frame
    missing the required `content` field."""
    record_test("event-stream-malformed-frame-strict")

    with SDK(server_url=HTTPBIN_URL) as s:
        res = s.eventstreams.malformed_stream()
        assert res.json_event is not None
        assert isinstance(res.json_event, eventstreaming.EventStream)

        with pytest.raises(ResponseValidationError) as exc_info:
            # pylint: disable=not-context-manager
            with res.json_event as event_stream:
                for _ in event_stream:
                    pass

        assert exc_info.value.raw_response is not None
        assert exc_info.value.body is not None


@pytest.mark.asyncio()
async def test_event_stream_malformed_frame_raises_response_validation_error_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        res = await s.eventstreams.malformed_stream_async()
        assert res.json_event is not None
        assert isinstance(res.json_event, eventstreaming.EventStreamAsync)

        with pytest.raises(ResponseValidationError) as exc_info:
            # pylint: disable=not-context-manager
            async with res.json_event as event_stream:
                async for _ in event_stream:
                    pass

        assert exc_info.value.raw_response is not None
        assert exc_info.value.body is not None


def test_event_stream_json_data():
    record_test("event-stream-json-data")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.eventstreams.json()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.json_event is not None
        assert isinstance(res.json_event, eventstreaming.EventStream)

        # pylint: disable=not-context-manager
        with res.json_event as event_stream:
            json_events: List[JSONEventData] = []
            for event in event_stream:
                json_events.append(event)

            assert len(json_events) == 4

            message = ""
            for event in json_events:
                message += event.content

            assert message == "Hello world!"


@pytest.mark.asyncio()
async def test_event_stream_json_data_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.json_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.json_event is not None
        assert isinstance(res.json_event, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.json_event as event_stream:
            json_events: List[JSONEventData] = []
            async for event in event_stream:
                json_events.append(event)

            assert len(json_events) == 4

            message = ""
            for event in json_events:
                message += event.content

            assert message == "Hello world!"


def test_event_stream_text_data():
    record_test("event-stream-text-data")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.text()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.text_event is not None
    assert isinstance(res.text_event, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.text_event as event_stream:
        text_events: List[str] = []
        for event in event_stream:
            text_events.append(event)

        assert len(text_events) == 4

        message = ""
        for event in text_events:
            message += event

        assert message == "Hello world!"


@pytest.mark.asyncio()
async def test_event_stream_text_data_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.text_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.text_event is not None
        assert isinstance(res.text_event, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.text_event as event_stream:
            text_events: List[str] = []
            async for event in event_stream:
                text_events.append(event)

            assert len(text_events) == 4

            message = ""
            for event in text_events:
                message += event

            assert message == "Hello world!"


def test_event_stream_multiline_data():
    record_test("event-stream-multiline-data")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.multiline()
    assert res is not None
    assert res.text_event is not None
    assert isinstance(res.text_event, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.text_event as event_stream:
        text_events: List[str] = []
        for event in event_stream:
            text_events.append(event)

        assert len(text_events) == 1

        assert text_events[0] == "YHOO\n+2\n10"


@pytest.mark.asyncio()
async def test_event_stream_multiline_data_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.multiline_async()
        assert res is not None
        assert res.text_event is not None
        assert isinstance(res.text_event, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.text_event as event_stream:
            text_events: List[str] = []
            async for event in event_stream:
                text_events.append(event)

            assert len(text_events) == 1

            assert text_events[0] == "YHOO\n+2\n10"


def test_event_stream_rich_events():
    record_test("event-stream-rich-events")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.rich()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.rich_stream is not None
    assert isinstance(res.rich_stream, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.rich_stream as event_stream:
        rich_events: List[Union[RichCompletionEvent, HeartbeatEvent]] = []
        for event in event_stream:
            rich_events.append(event)

        assert len(rich_events) == 3

        assert rich_events == [
            RichCompletionEvent(
                id="job-1",
                data=RichCompletionEventData(
                    completion="Hello", model="jeeves-1", stop_reason=None
                ),
            ),
            HeartbeatEvent(data="ping", retry=3000),
            RichCompletionEvent(
                id="job-1",
                data=RichCompletionEventData(
                    completion="world!",
                    model="jeeves-1",
                    stop_reason="stop_sequence",
                ),
            ),
        ]


@pytest.mark.asyncio()
async def test_event_stream_rich_events_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.rich_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.rich_stream is not None
        assert isinstance(res.rich_stream, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.rich_stream as event_stream:
            rich_events: List[Union[RichCompletionEvent, HeartbeatEvent]] = []
            async for event in event_stream:
                rich_events.append(event)

            assert len(rich_events) == 3

            assert rich_events == [
                RichCompletionEvent(
                    id="job-1",
                    data=RichCompletionEventData(
                        completion="Hello", model="jeeves-1", stop_reason=None
                    ),
                ),
                HeartbeatEvent(data="ping", retry=3000),
                RichCompletionEvent(
                    id="job-1",
                    data=RichCompletionEventData(
                        completion="world!",
                        model="jeeves-1",
                        stop_reason="stop_sequence",
                    ),
                ),
            ]


def test_event_stream_with_sentinel_events():
    record_test("event-stream-chat-sentinel-event")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.chat(
        prompt="Print test content",
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.chat_completion_stream is not None
    assert isinstance(res.chat_completion_stream, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.chat_completion_stream as event_stream:
        chat_events: List[Union[ChatCompletionEvent, SentinelEvent]] = []
        for event in event_stream:
            chat_events.append(event)

        assert len(chat_events) == 5

        assert chat_events == [
            ChatCompletionEvent(
                data=ChatCompletionEventData(content="Hello"),
            ),
            ChatCompletionEvent(
                data=ChatCompletionEventData(content=" "),
            ),
            ChatCompletionEvent(
                data=ChatCompletionEventData(content="world"),
            ),
            ChatCompletionEvent(
                data=ChatCompletionEventData(content="!"),
            ),
            SentinelEvent(),
        ]


@pytest.mark.asyncio()
async def test_event_stream_with_sentinel_events_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.chat_async(
            prompt="Print test content",
        )

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.chat_completion_stream is not None
        assert isinstance(res.chat_completion_stream, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.chat_completion_stream as event_stream:
            chat_events: List[Union[ChatCompletionEvent, SentinelEvent]] = []
            async for event in event_stream:
                chat_events.append(event)

            assert len(chat_events) == 5

            assert chat_events == [
                ChatCompletionEvent(
                    data=ChatCompletionEventData(content="Hello"),
                ),
                ChatCompletionEvent(
                    data=ChatCompletionEventData(content=" "),
                ),
                ChatCompletionEvent(
                    data=ChatCompletionEventData(content="world"),
                ),
                ChatCompletionEvent(
                    data=ChatCompletionEventData(content="!"),
                ),
                SentinelEvent(),
            ]


def test_event_stream_skip_sentinel():
    record_test("event-stream-chat-skip-sentinel")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.chat_skip_sentinel(
        prompt="Print test content",
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.chat_completion_event is not None
    assert isinstance(res.chat_completion_event, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.chat_completion_event as event_stream:
        chat_events: List[ChatCompletionEventData] = []
        for event in event_stream:
            chat_events.append(event)

        assert len(chat_events) == 4

        assert chat_events == [
            ChatCompletionEventData(content="Hello"),
            ChatCompletionEventData(content=" "),
            ChatCompletionEventData(content="world"),
            ChatCompletionEventData(content="!"),
        ]


@pytest.mark.asyncio()
async def test_event_stream_skip_sentinel_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.chat_skip_sentinel_async(
            prompt="Print test content",
        )

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.chat_completion_event is not None
        assert isinstance(res.chat_completion_event, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.chat_completion_event as event_stream:
            chat_events: List[ChatCompletionEventData] = []
            async for event in event_stream:
                chat_events.append(event)

            assert len(chat_events) == 4

            assert chat_events == [
                ChatCompletionEventData(content="Hello"),
                ChatCompletionEventData(content=" "),
                ChatCompletionEventData(content="world"),
                ChatCompletionEventData(content="!"),
            ]


def test_event_stream_with_different_data_schemas():
    record_test("event-stream-different-data-schemas")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.different_data_schemas()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.different_data_schemas is not None
    assert isinstance(res.different_data_schemas, eventstreaming.EventStream)

    events: List[Union[MessageEvent, URLEvent, List[int], bool, float]] = []
    # pylint: disable=not-an-iterable
    for event in res.different_data_schemas:
        events.append(event)

    assert len(events) == 6

    assert events == [
        MessageEvent(id=123, content="Here is your url"),
        URLEvent(url="https://example.com"),
        MessageEvent(content="Have a great day!"),
        [1, 2, 3, 4],
        True,
        3.14159,
    ]


@pytest.mark.asyncio()
async def test_event_stream_with_different_data_schemas_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.different_data_schemas_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.different_data_schemas is not None
        assert isinstance(res.different_data_schemas, eventstreaming.EventStreamAsync)

        events: List[Union[MessageEvent, URLEvent, List[int], bool, float]] = []
        # pylint: disable=not-an-iterable
        async for event in res.different_data_schemas:
            events.append(event)

        assert len(events) == 6

        assert events == [
            MessageEvent(id=123, content="Here is your url"),
            URLEvent(url="https://example.com"),
            MessageEvent(content="Have a great day!"),
            [1, 2, 3, 4],
            True,
            3.14159,
        ]


def test_event_stream_mixed_data():
    record_test("event-stream-mixed-data")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.mixed_data()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.mixed_data_event is not None
    assert isinstance(res.mixed_data_event, eventstreaming.EventStream)

    events = []
    # pylint: disable=not-an-iterable
    for event in res.mixed_data_event:
        events.append(event)

    assert len(events) == 4

    # First event: JSON object (completion)
    assert events[0].content == "Hello world"

    # Second event: plain text
    assert events[1] == "Processing your request..."

    # Third event: plain text
    assert events[2] == "Almost done"

    # Fourth event: JSON object (completion)
    assert events[3].content == "Done!"


@pytest.mark.asyncio()
async def test_event_stream_mixed_data_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.mixed_data_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.mixed_data_event is not None
        assert isinstance(res.mixed_data_event, eventstreaming.EventStreamAsync)

        events = []
        # pylint: disable=not-an-iterable
        async for event in res.mixed_data_event:
            events.append(event)

        assert len(events) == 4

        # First event: JSON object (completion)
        assert events[0].content == "Hello world"

        # Second event: plain text
        assert events[1] == "Processing your request..."

        # Third event: plain text
        assert events[2] == "Almost done"

        # Fourth event: JSON object (completion)
        assert events[3].content == "Done!"


def test_event_stream_chat_heartbeat_skips_dataless():
    record_test("event-stream-chat-heartbeat-skips-dataless")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.chat_heartbeat()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.chat_completion_event is not None
    assert isinstance(res.chat_completion_event, eventstreaming.EventStream)

    # pylint: disable=not-context-manager
    with res.chat_completion_event as event_stream:
        events: List[ChatCompletionEventData] = []
        for event in event_stream:
            events.append(event)

        # Server sends: data("Hello1"), heartbeat (no data — skipped), data("Hello 2"), data("!"), [DONE]
        assert len(events) == 3

        assert events == [
            ChatCompletionEventData(content="Hello1"),
            ChatCompletionEventData(content="Hello 2"),
            ChatCompletionEventData(content="!"),
        ]


@pytest.mark.asyncio()
async def test_event_stream_chat_heartbeat_skips_dataless_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.chat_heartbeat_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.chat_completion_event is not None
        assert isinstance(res.chat_completion_event, eventstreaming.EventStreamAsync)

        # pylint: disable=not-context-manager
        async with res.chat_completion_event as event_stream:
            events: List[ChatCompletionEventData] = []
            async for event in event_stream:
                events.append(event)

            # Server sends: data("Hello1"), heartbeat (no data — skipped), data("Hello 2"), data("!"), [DONE]
            assert len(events) == 3

            assert events == [
                ChatCompletionEventData(content="Hello1"),
                ChatCompletionEventData(content="Hello 2"),
                ChatCompletionEventData(content="!"),
            ]


def test_event_stream_error_response():
    record_test("event-stream-error-response")

    http_client = httpx.Client(headers={"x-teapot": "json"})

    s = SDK(server_url=HTTPBIN_URL, client=http_client)
    assert s is not None

    with pytest.raises(TeapotJSONError, match="""I'm a teapot"""):
        s.eventstreams.text()


@pytest.mark.asyncio()
async def test_event_stream_error_response_async():
    http_client = httpx.AsyncClient(headers={"x-teapot": "json"})

    s = SDK(server_url=HTTPBIN_URL, async_client=http_client)
    assert s is not None

    with pytest.raises(TeapotJSONError, match="""I'm a teapot"""):
        await s.eventstreams.text_async()


def test_event_stream_stay_open_breaking_early():
    record_test("event-stream-stay-open-break-early")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.stay_open()
    assert res is not None
    assert res.text_event is not None

    events: List[str] = []
    with res.text_event as event_stream:
        for event in event_stream:
            events.append(event)
            break  # Exit early

    assert len(events) == 1

    # In Python, exiting the `with` block calls `__exit__`, which marks the
    # stream as closed. The iterator is therefore expected to be exhausted.
    with pytest.raises(StopIteration):
        next(event_stream)


@pytest.mark.asyncio()
async def test_event_stream_stay_open_breaking_early_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.stay_open_async()

        assert res is not None
        assert res.text_event is not None

        events: List[str] = []
        async with res.text_event as event_stream:
            async for event in event_stream:
                events.append(event)
                break  # Exit early

        assert len(events) == 1

        # In Python, exiting the `async with` block calls `__aexit__`, which marks
        # the stream as closed. The iterator is therefore expected to be exhausted.
        with pytest.raises(StopAsyncIteration):
            await event_stream.__anext__()


def test_event_stream_stay_open_sentinel_detection():
    record_test("event-stream-stay-open-sentinel")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.stay_open()

    assert res is not None
    assert res.text_event is not None

    events: List[str] = []
    with res.text_event as event_stream:
        for event in event_stream:
            events.append(event)

    assert events == ["event 1", "event 2", "event 3", "event 4"]


@pytest.mark.asyncio()
async def test_event_stream_stay_open_sentinel_detection_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.stay_open_async()

        assert res is not None
        assert res.text_event is not None

        events: List[str] = []
        async with res.text_event as event_stream:
            async for event in event_stream:
                events.append(event)

        assert events == ["event 1", "event 2", "event 3", "event 4"]


def test_event_stream_partial_with_comments():
    record_test("event-stream-partial-with-comments")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.partial_with_comments()

    assert res is not None
    assert res.partial_with_comments_event is not None

    actual: List[PartialWithCommentsEventData] = []
    with res.partial_with_comments_event as event_stream:
        for event in event_stream:
            actual.append(event)

    expected = [
        PartialWithCommentsEventData(message="Hello from SSE"),
        PartialWithCommentsEventData(status="processing", progress=50),
        PartialWithCommentsEventData(status="complete", progress=100, result="Success"),
        PartialWithCommentsEventData(test="mixed boundaries"),
        PartialWithCommentsEventData(another="test"),
    ]

    assert actual == expected


@pytest.mark.asyncio()
async def test_event_stream_partial_with_comments_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.partial_with_comments_async()

        assert res is not None
        assert res.partial_with_comments_event is not None

        actual: List[PartialWithCommentsEventData] = []
        async with res.partial_with_comments_event as event_stream:
            async for event in event_stream:
                actual.append(event)

        expected = [
            PartialWithCommentsEventData(message="Hello from SSE"),
            PartialWithCommentsEventData(status="processing", progress=50),
            PartialWithCommentsEventData(status="complete", progress=100, result="Success"),
            PartialWithCommentsEventData(test="mixed boundaries"),
            PartialWithCommentsEventData(another="test"),
        ]

        assert actual == expected

def test_event_stream_holds_http_client():
    def create_generator():
        s = SDK(server_url=HTTPBIN_URL)
        res = s.eventstreams.json()
        return res.json_event

    generator = create_generator()

    gc.collect()

    nb_chunks = 0
    for _ in generator:
        nb_chunks += 1

    assert nb_chunks == 4, "The generator should yield exactly 4 chunks"


@pytest.mark.asyncio()
async def test_event_stream_holds_http_client_async():
    async def create_generator():
        s = SDK(server_url=HTTPBIN_URL)
        res = await s.eventstreams.json_async()
        return res.json_event

    generator = await create_generator()

    gc.collect()

    nb_chunks = 0
    async for _ in generator:
        nb_chunks += 1
    assert nb_chunks == 4, "The generator should yield exactly 2 chunks"


def test_event_stream_union_with_standalone_comments():
    record_test("event-stream-union-with-standalone-comments")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.eventstreams.union_with_comments()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.union_with_comments_stream is not None
    assert isinstance(res.union_with_comments_stream, eventstreaming.EventStream)

    with res.union_with_comments_stream as event_stream:
        events = []
        for event in event_stream:
            events.append(event)

        assert len(events) == 4

        assert isinstance(events[0], StatusEvent)
        assert events[0].data.status == "started"
        assert events[0].data.message == "Initializing"

        assert isinstance(events[1], ProgressEvent)
        assert events[1].data.percent == 50
        assert events[1].data.detail == "Half done"

        assert isinstance(events[2], StatusEvent)
        assert events[2].data.status == "running"
        assert events[2].data.message == "Processing items"

        assert isinstance(events[3], ProgressEvent)
        assert events[3].data.percent == 100
        assert events[3].data.detail == "Complete"


@pytest.mark.asyncio()
async def test_event_stream_union_with_standalone_comments_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.eventstreams.union_with_comments_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.union_with_comments_stream is not None
        assert isinstance(res.union_with_comments_stream, eventstreaming.EventStreamAsync)

        async with res.union_with_comments_stream as event_stream:
            events = []
            async for event in event_stream:
                events.append(event)

            assert len(events) == 4

            assert isinstance(events[0], StatusEvent)
            assert events[0].data.status == "started"
            assert events[0].data.message == "Initializing"

            assert isinstance(events[1], ProgressEvent)
            assert events[1].data.percent == 50
            assert events[1].data.detail == "Half done"

            assert isinstance(events[2], StatusEvent)
            assert events[2].data.status == "running"
            assert events[2].data.message == "Processing items"

            assert isinstance(events[3], ProgressEvent)
            assert events[3].data.percent == 100
            assert events[3].data.detail == "Complete"


def test_event_stream_sse_overload_streaming_response():
    record_test("event-stream-sse-overload-streaming-response")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    expected_events: List[Union[ChatCompletionEvent, SentinelEvent]] = [
        ChatCompletionEvent(data=ChatCompletionEventData(content="Hello")),
        ChatCompletionEvent(data=ChatCompletionEventData(content=" ")),
        ChatCompletionEvent(data=ChatCompletionEventData(content="world")),
        ChatCompletionEvent(data=ChatCompletionEventData(content="!")),
        SentinelEvent(),
    ]

    # 1. Parsed-mode (high-level) path: response matchers build the envelope.
    res = s.eventstreams.sse_overload_chat(
        prompt="Print test content",
        stream=True,
    )
    assert res is not None
    assert res.chat_completion_stream is not None
    assert isinstance(res.chat_completion_stream, eventstreaming.EventStream)
    # pylint: disable=not-context-manager
    with res.chat_completion_stream as event_stream:
        assert list(event_stream) == expected_events

    # 2. Raw-response path: stream=True threads into mode="event_stream".
    raw = s.eventstreams.with_raw_response.sse_overload_chat(
        prompt="Print test content",
        stream=True,
    )
    assert raw.status_code == 200
    assert raw._mode == "event_stream"
    with pytest.raises(TypeError, match=r"EventStream\["):
        raw.parse(to=ChatCompletionEvent)

    # 3. Streaming-response path: same mode= reaches StreamedAPIResponse.
    with s.eventstreams.with_streaming_response.sse_overload_chat(
        prompt="Print test content",
        stream=True,
    ) as streamed:
        assert streamed.status_code == 200
        assert streamed._mode == "event_stream"
        streamed_res = streamed.parse()
        assert streamed_res.chat_completion_stream is not None
        assert isinstance(streamed_res.chat_completion_stream, eventstreaming.EventStream)


class _FakeSyncResponse:
    """Minimal httpx.Response stand-in whose byte iterator does not self-close.

    A real httpx.Response closes itself when iter_bytes is fully consumed, which
    would mask whether the helper closes on its own. This isolates the helper's
    close responsibility on the non-sentinel termination paths.
    """

    def __init__(self, chunks: List[bytes]):
        self._chunks = chunks
        self.closed = False

    def iter_bytes(self):
        for chunk in self._chunks:
            yield chunk

    def close(self):
        self.closed = True


class _FakeAsyncResponse:
    def __init__(self, chunks: List[bytes]):
        self._chunks = chunks
        self.closed = False

    async def aiter_bytes(self):
        for chunk in self._chunks:
            yield chunk

    async def aclose(self):
        self.closed = True


def test_event_stream_closes_without_sentinel_or_context_manager():
    """Regression: the response is closed when the stream ends without a sentinel
    and is consumed via plain iteration (no `with`). Previously the helper only
    closed on the sentinel branch, leaking the connection on natural end."""
    response = _FakeSyncResponse([b"data: one\n\n", b"data: two\n\n"])
    stream = eventstreaming.EventStream(
        cast(httpx.Response, response), decoder=json.loads, sentinel="[DONE]"
    )

    events = list(stream)

    assert len(events) == 2
    assert response.closed, "response must be closed after natural end without sentinel"


def test_event_stream_closes_on_exception():
    """Regression: an exception raised mid-stream still closes the response."""

    def boom(_):
        raise ValueError("decode failed")

    response = _FakeSyncResponse([b"data: one\n\n"])
    stream = eventstreaming.EventStream(
        cast(httpx.Response, response), decoder=boom, sentinel="[DONE]"
    )

    with pytest.raises(ValueError):
        list(stream)

    assert response.closed, "response must be closed when iteration raises"


@pytest.mark.asyncio()
async def test_event_stream_closes_without_sentinel_or_context_manager_async():
    response = _FakeAsyncResponse([b"data: one\n\n", b"data: two\n\n"])
    stream = eventstreaming.EventStreamAsync(
        cast(httpx.Response, response), decoder=json.loads, sentinel="[DONE]"
    )

    events = [event async for event in stream]

    assert len(events) == 2
    assert response.closed, "response must be closed after natural end without sentinel"


@pytest.mark.asyncio()
async def test_event_stream_closes_on_exception_async():
    def boom(_):
        raise ValueError("decode failed")

    response = _FakeAsyncResponse([b"data: one\n\n"])
    stream = eventstreaming.EventStreamAsync(
        cast(httpx.Response, response), decoder=boom, sentinel="[DONE]"
    )

    with pytest.raises(ValueError):
        [event async for event in stream]

    assert response.closed, "response must be closed when iteration raises"


def test_event_stream_sse_overload_json_response():
    record_test("event-stream-sse-overload-json-response")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    expected_result = [
        ChatCompletionEvent(data=ChatCompletionEventData(content="Hello")),
        ChatCompletionEvent(data=ChatCompletionEventData(content=" ")),
        ChatCompletionEvent(data=ChatCompletionEventData(content="world")),
        ChatCompletionEvent(data=ChatCompletionEventData(content="!")),
    ]

    # 1. Parsed-mode (high-level) path.
    res = s.eventstreams.sse_overload_chat(
        prompt="Print test content",
        stream=False,
    )
    assert res is not None
    assert res.chat_completion_result == expected_result

    # 2. Raw-response path: stream=False threads into mode="buffered".
    raw = s.eventstreams.with_raw_response.sse_overload_chat(
        prompt="Print test content",
        stream=False,
    )
    assert raw.status_code == 200
    assert raw._mode == "buffered"
    raw_res = raw.parse()
    assert raw.parse() is raw_res  # buffered cache hit
    assert raw_res.chat_completion_result == expected_result

    # 3. Streaming-response path: StreamedAPIResponse still gets mode="buffered".
    with s.eventstreams.with_streaming_response.sse_overload_chat(
        prompt="Print test content",
        stream=False,
    ) as streamed:
        assert streamed.status_code == 200
        assert streamed._mode == "buffered"
        assert streamed.parse().chat_completion_result == expected_result


@pytest.mark.asyncio()
async def test_event_stream_sse_overload_streaming_response_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        expected_events: List[Union[ChatCompletionEvent, SentinelEvent]] = [
            ChatCompletionEvent(data=ChatCompletionEventData(content="Hello")),
            ChatCompletionEvent(data=ChatCompletionEventData(content=" ")),
            ChatCompletionEvent(data=ChatCompletionEventData(content="world")),
            ChatCompletionEvent(data=ChatCompletionEventData(content="!")),
            SentinelEvent(),
        ]

        # 1. Parsed-mode (high-level) path.
        res = await s.eventstreams.sse_overload_chat_async(
            prompt="Print test content",
            stream=True,
        )
        assert res is not None
        assert res.chat_completion_stream is not None
        assert isinstance(res.chat_completion_stream, eventstreaming.EventStreamAsync)
        # pylint: disable=not-context-manager
        async with res.chat_completion_stream as event_stream:
            events = [event async for event in event_stream]
            assert events == expected_events

        # 2. Raw-response path: stream=True threads into mode="event_stream".
        raw = await s.eventstreams.with_raw_response.sse_overload_chat_async(
            prompt="Print test content",
            stream=True,
        )
        assert raw.status_code == 200
        assert raw._mode == "event_stream"
        with pytest.raises(TypeError, match=r"EventStreamAsync\["):
            await raw.parse(to=ChatCompletionEvent)

        # 3. Streaming-response path: same mode= reaches AsyncStreamedAPIResponse.
        async with s.eventstreams.with_streaming_response.sse_overload_chat_async(
            prompt="Print test content",
            stream=True,
        ) as streamed:
            assert streamed.status_code == 200
            assert streamed._mode == "event_stream"
            streamed_res = await streamed.parse()
            assert streamed_res.chat_completion_stream is not None
            assert isinstance(streamed_res.chat_completion_stream, eventstreaming.EventStreamAsync)


@pytest.mark.asyncio()
async def test_event_stream_sse_overload_json_response_async():
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        expected_result = [
            ChatCompletionEvent(data=ChatCompletionEventData(content="Hello")),
            ChatCompletionEvent(data=ChatCompletionEventData(content=" ")),
            ChatCompletionEvent(data=ChatCompletionEventData(content="world")),
            ChatCompletionEvent(data=ChatCompletionEventData(content="!")),
        ]

        # 1. Parsed-mode (high-level) path.
        res = await s.eventstreams.sse_overload_chat_async(
            prompt="Print test content",
            stream=False,
        )
        assert res is not None
        assert res.chat_completion_result == expected_result

        # 2. Raw-response path: stream=False threads into mode="buffered".
        raw = await s.eventstreams.with_raw_response.sse_overload_chat_async(
            prompt="Print test content",
            stream=False,
        )
        assert raw.status_code == 200
        assert raw._mode == "buffered"
        raw_res = await raw.parse()
        assert (await raw.parse()) is raw_res  # buffered cache hit
        assert raw_res.chat_completion_result == expected_result

        # 3. Streaming-response path: AsyncStreamedAPIResponse still gets mode="buffered".
        async with s.eventstreams.with_streaming_response.sse_overload_chat_async(
            prompt="Print test content",
            stream=False,
        ) as streamed:
            assert streamed.status_code == 200
            assert streamed._mode == "buffered"
            assert (await streamed.parse()).chat_completion_result == expected_result


def test_event_stream_large_event_small_chunks():
    record_test("event-stream-large-event-small-chunks")

    size = 2 * 1024 * 1024
    with SDK(server_url=HTTPBIN_URL) as s:
        res = s.eventstreams.large_event_small_chunks(size=size)

        assert res is not None
        assert res.http_meta.response.status_code == 200
        assert res.large_chunked_event is not None

        # pylint: disable=not-context-manager
        with res.large_chunked_event as event_stream:
            contents = [event.content for event in event_stream]

    assert len(contents) == 3
    assert contents[0] == "start"
    assert contents[2] == "end"

    big = contents[1]
    assert len(big) == size
    assert big[0] == "S"
    assert big[-1] == "E"
    assert big[1:-1] == "a" * (size - 2)


def test_event_stream_split_boundaries():
    record_test("event-stream-split-boundaries")

    with SDK(server_url=HTTPBIN_URL) as s:
        res = s.eventstreams.split_boundaries()

        assert res is not None
        assert res.http_meta.response.status_code == 200
        assert res.split_boundary_event is not None

        actual_tags: List[str] = []
        expected_tags: List[str] = []

        # pylint: disable=not-context-manager
        with res.split_boundary_event as event_stream:
            for event in event_stream:
                if event.kind == "expected":
                    expected_tags = event.tags
                else:
                    actual_tags.extend(event.tags)

    assert len(expected_tags) > 0
    assert actual_tags == expected_tags
