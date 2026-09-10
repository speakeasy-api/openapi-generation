import time
import pytest
from openapi import SDK
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.utils import *
from openapi.utils import jsonl

from .common_helpers import *
from .test_helpers import *

from typing import List


def test_jsonl_stream_data():
    record_test("jsonl-stream-data-envelope-http-responses")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.jsonl_stream()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.object is not None
        assert isinstance(res.object, jsonl.JsonLStream)

        # pylint: disable=not-context-manager
        with res.object as jsonl_stream:
            json_events: List[JSONEvent] = []
            for event in jsonl_stream:
                json_events.append(event)

            assert len(json_events) == 2

            # Assert first event
            assert json_events[0].name == "Peter"
            assert json_events[0].skills == ["Go", "Python"]

            # Assert second event
            assert json_events[1].name == "John"
            assert json_events[1].skills == ["Go", "Rust"]


@pytest.mark.asyncio()
async def test_jsonl_stream_data_async():
    record_test("jsonl-stream-data-async-envelope-http-responses")
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.jsonl.jsonl_stream_async()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.object is not None
        assert isinstance(res.object, jsonl.JsonLStreamAsync)

        # pylint: disable=not-context-manager
        async with res.object as jsonl_stream:
            json_events: List[JSONEvent] = []
            async for event in jsonl_stream:
                json_events.append(event)

            assert len(json_events) == 2

            # Assert first event
            assert json_events[0].name == "Peter"
            assert json_events[0].skills == ["Go", "Python"]

            # Assert second event
            assert json_events[1].name == "John"
            assert json_events[1].skills == ["Go", "Rust"]


# This test checks that the stream of data is processed as soon as it comes. It makes sure the client is not waiting for the entire payload before starting the processing.
def test_jsonl_stream_data_chunks():
    record_test("jsonl-stream-data-chunks-envelope-http-responses")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None
        start_time = time.time()

        res = s.jsonl.jsonl_stream_chunks()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.object is not None
        assert isinstance(res.object, jsonl.JsonLStream)

        # pylint: disable=not-context-manager
        with res.object as jsonl_stream:
            json_events: List[JSONEvent] = []
            stream_iter = iter(jsonl_stream)
            first_event = next(stream_iter)
            first_event_time = time.time() - start_time
            assert 100 <= (first_event_time * 1000) < 190, (
                f"First event should take longer than 100ms because the server sends a chunk and then waits for 100ms. After receiving the 2nd chunk the first event should be received (because the 2nd chunk has the first jsonl separator).  Time taken -> {first_event_time * 1000}ms. This test can break if the implementation has a bug or if the chunks are sent slowly from api-test-service"
            )

            # Assert first event
            assert first_event.name == "Peter"
            assert first_event.skills == ["Go", "Python"]
            second_event = next(stream_iter)
            # Assert second event
            assert second_event.name == "John"
            assert second_event.skills == ["Go", "Rust"]


def test_jsonl_deserialization_with_camel_case_properties():
    record_test("jsonl-deserialization-camel-case-properties")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.jsonl_deserialization_verification()

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.object is not None
        assert isinstance(res.object, jsonl.JsonLStream)

        # pylint: disable=not-context-manager
        with res.object as jsonl_stream:
            stream_iter = iter(jsonl_stream)
            first_event = next(stream_iter)
            print(first_event)
            assert first_event.is_finished == "yes"


def test_jsonl_stream_holds_http_client():
    def create_generator():
        s = SDK(server_url=HTTPBIN_URL)
        res = s.jsonl.jsonl_stream()
        return res.object

    generator = create_generator()

    nb_chunks = 0
    for _ in generator:
        nb_chunks += 1

    assert nb_chunks == 2, "The generator should yield exactly 2 chunks"


@pytest.mark.asyncio()
async def test_jsonl_stream_holds_http_client_async():
    async def create_generator():
        s = SDK(server_url=HTTPBIN_URL)
        res = await s.jsonl.jsonl_stream_async()
        return res.object

    generator = await create_generator()

    nb_chunks = 0
    async for _ in generator:
        nb_chunks += 1

    assert nb_chunks == 2, "The generator should yield exactly 2 chunks"
