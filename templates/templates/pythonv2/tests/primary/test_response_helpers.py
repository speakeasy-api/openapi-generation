import httpx
import pytest
from pydantic import BaseModel
from typing import Any, Optional, TypeVar
from typing_extensions import Annotated, TypeAliasType

from openapi.utils.response_helpers import APIResponse, AsyncAPIResponse
from openapi.utils.eventstreaming import EventStream, EventStreamAsync

Stream = EventStream
AsyncStream = EventStreamAsync


class _Payload(BaseModel):
    name: str
    count: int


_JSON_BODY = b'{"name":"alice","count":3}'
_NDJSON_BODY = b'{"name":"alice","count":3}\n'
_SSE_BODY = b"data: {\"name\":\"alice\",\"count\":3}\n\n"


def _make_response(
    *, content_type: Optional[str], body: bytes, status: int = 200
) -> httpx.Response:
    request = httpx.Request("GET", "http://test.invalid/resource")
    headers = {"content-type": content_type} if content_type is not None else {}
    return httpx.Response(
        status_code=status,
        headers=headers,
        content=body,
        request=request,
    )


def _default_parser(_: httpx.Response) -> dict[str, Any]:
    return {"parsed_via_default": True}


async def _async_default_parser(_: httpx.Response) -> dict[str, Any]:
    return {"parsed_via_default": True}


def test_parse_to_buffered_json():
    """to=Model, mode="buffered", application/json -> unmarshal into Model."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_default_parser,
        mode="buffered",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_buffered_json():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_async_default_parser,
        mode="buffered",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_streamed_json():
    """to=Model, mode="raw_stream", application/json -> body read, unmarshalled."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_default_parser,
        mode="raw_stream",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_streamed_json():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_async_default_parser,
        mode="raw_stream",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_streamed_ndjson():
    """to=Model, mode="raw_stream", application/x-ndjson -> body read, unmarshalled."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/x-ndjson", body=_NDJSON_BODY),
        parser=_default_parser,
        mode="raw_stream",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_streamed_ndjson():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/x-ndjson", body=_NDJSON_BODY),
        parser=_async_default_parser,
        mode="raw_stream",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_sse_non_stream_class_raises_typeerror():
    """mode="event_stream" + to=<BaseModel> -> TypeError guard fires (intent mismatch).
    Caller asked for a JSON model on an op the spec marked as SSE — wrong tool."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError) as exc_info:
        raw.parse(to=_Payload)
    msg = str(exc_info.value)
    assert "SSE" in msg
    assert "EventStream[T]" in msg
    assert "iter_lines" in msg and "iter_bytes" in msg


@pytest.mark.asyncio()
async def test_async_parse_to_sse_non_stream_class_raises_typeerror():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError) as exc_info:
        await raw.parse(to=_Payload)
    msg = str(exc_info.value)
    assert "SSE" in msg
    assert "EventStreamAsync[T]" in msg
    assert "aiter_lines" in msg and "aiter_bytes" in msg


def test_parse_to_event_stream_synthesizes_decoder():
    """to=EventStream[T], no decoder -> synthesized model_validate_json decoder."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=EventStream[_Payload])
    assert isinstance(stream, EventStream)
    events = list(stream)
    assert events == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_synthesizes_decoder():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=EventStreamAsync[_Payload])
    assert isinstance(stream, EventStreamAsync)
    events = [event async for event in stream]
    assert events == [_Payload(name="alice", count=3)]


def test_parse_to_event_stream_with_explicit_decoder():
    """Caller-supplied decoder beats synthesis. ``stream_events`` always wraps
    the SSE ``data:`` value in a ``{"data": <wire>, ...}`` envelope before
    handing the JSON-serialized frame to the decoder, so ``_Envelope`` models
    that envelope shape and the decoder peels ``.data`` to recover ``_Payload``.
    Mirrors what the generator emits at op call sites for SSE-capable methods."""

    class _Envelope(BaseModel):
        data: _Payload

    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(
        to=EventStream[_Payload],
        decoder=lambda frame: _Envelope.model_validate_json(frame).data,
    )
    events = list(stream)
    assert events == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_with_explicit_decoder():
    """Caller-supplied decoder beats synthesis on async path."""

    class _Envelope(BaseModel):
        data: _Payload

    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(
        to=EventStreamAsync[_Payload],
        decoder=lambda frame: _Envelope.model_validate_json(frame).data,
    )
    events = [event async for event in stream]
    assert events == [_Payload(name="alice", count=3)]


def test_parse_to_stream_alias_dispatches_same_as_event_stream():
    """``Stream`` alias collapses to ``EventStream`` via ``get_origin``."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=Stream[_Payload])
    assert isinstance(stream, EventStream)
    events = list(stream)
    assert events == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_async_stream_alias_dispatches_same():
    """``AsyncStream`` alias collapses to ``EventStreamAsync`` via ``get_origin``."""
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=AsyncStream[_Payload])
    assert isinstance(stream, EventStreamAsync)
    events = [event async for event in stream]
    assert events == [_Payload(name="alice", count=3)]


def test_parse_to_event_stream_sentinel_terminates():
    """sentinel= short-circuits iteration before EOF."""
    body = (
        b'data: {"name":"alice","count":3}\n\n'
        b"data: [DONE]\n\n"
        b'data: {"name":"bob","count":9}\n\n'
    )
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=body),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=EventStream[_Payload], sentinel="[DONE]")
    events = list(stream)
    assert events == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_sentinel_terminates():
    """sentinel= short-circuits async iteration before EOF."""
    body = (
        b'data: {"name":"alice","count":3}\n\n'
        b"data: [DONE]\n\n"
        b'data: {"name":"bob","count":9}\n\n'
    )
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=body),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=EventStreamAsync[_Payload], sentinel="[DONE]")
    events = [event async for event in stream]
    assert events == [_Payload(name="alice", count=3)]


def test_parse_to_event_stream_not_cached():
    """Streams are single-pass; repeated parse(to=EventStream[X]) must return
    a fresh instance each call so callers can re-iterate (assuming the body
    has not been exhausted upstream)."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    first = raw.parse(to=EventStream[_Payload])
    second = raw.parse(to=EventStream[_Payload])
    assert first is not second


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_not_cached():
    """Async streams are single-pass; repeated parse(to=EventStreamAsync[X])
    must return a fresh instance each call."""
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    first = await raw.parse(to=EventStreamAsync[_Payload])
    second = await raw.parse(to=EventStreamAsync[_Payload])
    assert first is not second


def test_parse_to_event_stream_requires_basemodel_when_no_decoder():
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError, match="BaseModel"):
        raw.parse(to=EventStream[dict])


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_requires_basemodel_when_no_decoder():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError, match="BaseModel"):
        await raw.parse(to=EventStreamAsync[dict])


def test_parse_to_async_event_stream_rejected_on_sync_response():
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError, match="EventStream\\["):
        raw.parse(to=EventStreamAsync[_Payload])


@pytest.mark.asyncio()
async def test_async_parse_to_sync_event_stream_rejected_on_async_response():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
    )
    with pytest.raises(TypeError, match="EventStreamAsync\\["):
        await raw.parse(to=EventStream[_Payload])


def test_parse_to_caches_result():
    """Repeated parse(to=X) returns same instance from ``_parsed_by_type``."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_default_parser,
        mode="buffered",
    )
    first = raw.parse(to=_Payload)
    second = raw.parse(to=_Payload)
    assert first is second


@pytest.mark.asyncio()
async def test_async_parse_to_caches_result():
    """Repeated parse(to=X) returns same instance from ``_parsed_by_type``."""
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_async_default_parser,
        mode="buffered",
    )
    first = await raw.parse(to=_Payload)
    second = await raw.parse(to=_Payload)
    assert first is second


def test_parse_default_invokes_parser_and_caches():
    """to=None -> generator-emitted parser runs, result cached under default key."""
    calls = {"n": 0}

    def parser(_: httpx.Response) -> dict[str, Any]:
        calls["n"] += 1
        return {"hit": calls["n"]}

    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=parser,
        mode="buffered",
    )
    first = raw.parse()
    second = raw.parse()
    assert first == {"hit": 1}
    assert first is second
    assert calls["n"] == 1


@pytest.mark.asyncio()
async def test_async_parse_default_invokes_parser_and_caches():
    calls = {"n": 0}

    async def parser(_: httpx.Response) -> dict[str, Any]:
        calls["n"] += 1
        return {"hit": calls["n"]}

    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=parser,
        mode="buffered",
    )
    first = await raw.parse()
    second = await raw.parse()
    assert first == {"hit": 1}
    assert first is second
    assert calls["n"] == 1


def test_parse_to_event_stream_decoder_error_propagates():
    """Decoder raising mid-stream surfaces through __next__."""

    def boom(_: str) -> _Payload:
        raise ValueError("boom")

    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=EventStream[_Payload], decoder=boom)
    with pytest.raises(ValueError, match="boom"):
        next(iter(stream))


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_decoder_error_propagates():
    """Decoder raising mid-stream surfaces through __anext__."""

    def boom(_: str) -> _Payload:
        raise ValueError("boom")

    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=EventStreamAsync[_Payload], decoder=boom)
    with pytest.raises(ValueError, match="boom"):
        await stream.__anext__()


def test_parse_to_event_stream_dispatch_ignores_content_type():
    """``to=EventStream[T]`` is the caller-intent override: SSE dispatch must
    succeed regardless of wire Content-Type (proxy stripping, server misconfig,
    charset suffix, etc.) when the caller has explicitly opted into a stream
    return type."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_SSE_BODY
        ),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=EventStream[_Payload])
    assert isinstance(stream, EventStream)
    events = list(stream)
    assert events == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_dispatch_ignores_content_type():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_SSE_BODY
        ),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=EventStreamAsync[_Payload])
    assert isinstance(stream, EventStreamAsync)
    events = [event async for event in stream]
    assert events == [_Payload(name="alice", count=3)]


def test_parse_to_non_sse_skips_event_stream_dispatch():
    """mode="buffered"/raw_stream + to=Model -> never enters _build_event_stream even if the
    wire Content-Type happens to be text/event-stream."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_JSON_BODY),
        parser=_default_parser,
        mode="buffered",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_non_sse_skips_event_stream_dispatch():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_JSON_BODY),
        parser=_async_default_parser,
        mode="buffered",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_model_demotes_event_stream_when_wire_is_json():
    """mode="event_stream" can be over-eagerly emitted by the
    generator when a spec aggregates an SSE error branch alongside a JSON
    success branch (or pairs a 200 SSE with a 201 JSON success). When the
    caller did NOT opt into a stream return type and the wire is not actually
    SSE, the wrapper must demote to buffered, read the body, and unmarshal
    into the caller's model rather than raising the SSE-intent TypeError."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_default_parser,
        mode="event_stream",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_model_demotes_event_stream_when_wire_is_json():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="application/json", body=_JSON_BODY),
        parser=_async_default_parser,
        mode="event_stream",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_model_demotes_event_stream_for_non_sse_non_json_wire():
    """Demote path covers any non-SSE wire when caller passed a non-stream
    ``to=`` — binary download endpoints, plain-text responses, and proxy-
    stripped success bodies that ride alongside a spec-aggregated SSE error
    variant. Wire here is ``application/octet-stream`` but body is JSON so
    we can assert the demote produced a successful unmarshal rather than the
    SSE-intent TypeError."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_JSON_BODY
        ),
        parser=_default_parser,
        mode="event_stream",
    )
    result = raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


@pytest.mark.asyncio()
async def test_async_parse_to_model_demotes_event_stream_for_non_sse_non_json_wire():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_JSON_BODY
        ),
        parser=_async_default_parser,
        mode="event_stream",
    )
    result = await raw.parse(to=_Payload)
    assert isinstance(result, _Payload)
    assert result == _Payload(name="alice", count=3)


def test_parse_to_none_does_not_demote_event_stream_on_non_sse_wire():
    """``to=None`` defers to the spec-driven default parser instead of
    demoting to the buffered branch. Demote would route through the
    ``mode == "buffered"`` eager-read path and break streaming semantics when
    the wire Content-Type is unreliable (proxy strip, server misconfig).
    Asserting the parser ran (and returned its marker) proves the dispatch
    stayed on the event_stream branch."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_SSE_BODY
        ),
        parser=_default_parser,
        mode="event_stream",
    )
    assert raw.parse() == {"parsed_via_default": True}


@pytest.mark.asyncio()
async def test_async_parse_to_none_does_not_demote_event_stream_on_non_sse_wire():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(
            content_type="application/octet-stream", body=_SSE_BODY
        ),
        parser=_async_default_parser,
        mode="event_stream",
    )
    assert await raw.parse() == {"parsed_via_default": True}




_PayloadStreamAlias = TypeAliasType(
    "_PayloadStreamAlias", EventStream[_Payload]
)
_PayloadStreamAsyncAlias = TypeAliasType(
    "_PayloadStreamAsyncAlias", EventStreamAsync[_Payload]
)


_TChunk = TypeVar("_TChunk")


class _MyEventStream(EventStream[_TChunk]):
    """Generic-parameterized subclass used to validate ``issubclass``
    acceptance in ``_event_stream_origin`` (vs the old identity check).
    Keeps the chunk TypeVar open so the call site can parametrize as
    ``_MyEventStream[_Payload]``."""


class _MyEventStreamAsync(EventStreamAsync[_TChunk]):
    pass


class _MyChatStream(EventStream[_Payload]):
    """Pre-bound subclass: chunk type ``_Payload`` is fixed on the base, not
    on the caller's parametrization site. Decoder synthesis must walk
    ``__orig_bases__`` to recover it."""


class _MyChatStreamAsync(EventStreamAsync[_Payload]):
    pass


def test_parse_to_type_alias_event_stream():
    """``parse(to=TypeAliasType(..., EventStream[T]))`` must unwrap the alias
    and dispatch into the SSE branch with chunk type preserved."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=_PayloadStreamAlias)
    assert isinstance(stream, EventStream)
    assert list(stream) == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_type_alias_event_stream():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=_PayloadStreamAsyncAlias)
    assert isinstance(stream, EventStreamAsync)
    assert [event async for event in stream] == [_Payload(name="alice", count=3)]


def test_parse_to_annotated_event_stream():
    """``Annotated[EventStream[T], <meta>]`` must unwrap to ``EventStream[T]``
    so caller-side metadata (e.g. pydantic field annotations) doesn't break
    dispatch."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=Annotated[EventStream[_Payload], "metadata"])
    assert isinstance(stream, EventStream)
    assert list(stream) == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_annotated_event_stream():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(
        to=Annotated[EventStreamAsync[_Payload], "metadata"]
    )
    assert isinstance(stream, EventStreamAsync)
    assert [event async for event in stream] == [_Payload(name="alice", count=3)]


def test_parse_to_event_stream_subclass_returns_subclass_instance():
    """``issubclass`` acceptance: caller-defined subclass instantiated, not
    the base class. Allows users to attach extra methods/state to streams
    while still flowing through the generated dispatch."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=_MyEventStream[_Payload])
    assert isinstance(stream, _MyEventStream)
    assert list(stream) == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_event_stream_subclass_returns_subclass_instance():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=_MyEventStreamAsync[_Payload])
    assert isinstance(stream, _MyEventStreamAsync)
    assert [event async for event in stream] == [_Payload(name="alice", count=3)]


def test_parse_to_prebound_event_stream_subclass():
    """Chunk type baked into the subclass's base (``class X(EventStream[T])``)
    must be recovered via ``__orig_bases__`` so the synthesized decoder still
    knows what model to validate frames into."""
    raw = APIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = raw.parse(to=_MyChatStream)
    assert isinstance(stream, _MyChatStream)
    assert list(stream) == [_Payload(name="alice", count=3)]


@pytest.mark.asyncio()
async def test_async_parse_to_prebound_event_stream_subclass():
    raw = AsyncAPIResponse[dict[str, Any]](
        raw=_make_response(content_type="text/event-stream", body=_SSE_BODY),
        parser=_async_default_parser,
        mode="event_stream",
        client_ref=object(),
    )
    stream = await raw.parse(to=_MyChatStreamAsync)
    assert isinstance(stream, _MyChatStreamAsync)
    assert [event async for event in stream] == [_Payload(name="alice", count=3)]
