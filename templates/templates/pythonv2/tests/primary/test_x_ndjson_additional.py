import pytest
import httpx
from openapi import SDK
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.models.errors import TeapotJSONError
from openapi.utils import *
from openapi.utils import jsonl

from .common_helpers import *
from .test_helpers import *

from typing import List, Union

def test_x_ndjson_stream_data():
    record_test("x-ndjson-stream-data-envelope-http-responses")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.x_ndjson_stream()

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
            assert json_events[0].name == 'Peter'
            assert json_events[0].skills == ['Go', 'Python']
            
            
            # Assert second event
            assert json_events[1].name == 'John'
            assert json_events[1].skills == ['Go', 'Rust']


@pytest.mark.asyncio()
async def test_x_ndjson_stream_data_async():
    record_test("x-ndjson-stream-data-async-envelope-http-responses")
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.jsonl.x_ndjson_stream_async()

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
            assert json_events[0].name == 'Peter'
            assert json_events[0].skills == ['Go', 'Python']
            
            # Assert second event
            assert json_events[1].name == 'John'
            assert json_events[1].skills == ['Go', 'Rust']

def test_x_ndjson_stream_data_chunks():
    record_test("x-ndjson-stream-data-chunks-envelope-http-responses")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.x_ndjson_stream_chunks()

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
            assert json_events[0].name == 'Peter'
            assert json_events[0].skills == ['Go', 'Python']
            
            # Assert second event
            assert json_events[1].name == 'John'
            assert json_events[1].skills == ['Go', 'Rust']


@pytest.mark.asyncio()
async def test_x_ndjson_stream_data_chunks_async():
    record_test("x-ndjson-stream-data-async-chunks-envelope-http-responses")
    async with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = await s.jsonl.x_ndjson_stream_chunks_async()

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
            assert json_events[0].name == 'Peter'
            assert json_events[0].skills == ['Go', 'Python']
            
            # Assert second event
            assert json_events[1].name == 'John'
            assert json_events[1].skills == ['Go', 'Rust']
