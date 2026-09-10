import pytest
import httpx
from openapi import AsyncSDK
from openapi.models import operations
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.models.errors import TeapotJSONError
from openapi.utils import *
from openapi.utils import jsonl

from .common_helpers import *

from typing import List, Union


@pytest.mark.asyncio()
async def test_jsonl_stream_data_async():
    record_test("jsonl-stream-data-async-flat-response")
    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token") as s:
        assert s is not None

        res = await s.jsonl.jsonl_stream()

        assert res is not None
        
        assert isinstance(res, operations.JsonlStreamResponse)
        assert isinstance(res.object, jsonl.JsonLStreamAsync)

        result: List[JsonlStreamResponseBody] = []
        async with res.object as jsonl_stream:
            async for item in jsonl_stream:
                result.append(item)
        assert len(result) == 2


        # Assert first event
        assert result[0].name == 'Peter'
        assert result[0].skills == ['Go', 'Python']
        
        # Assert second event
        assert result[1].name == 'John'
        assert result[1].skills == ['Go', 'Rust']


@pytest.mark.asyncio()
async def test_jsonl_stream_data_chunks_async():
    record_test("jsonl-stream-data-async-chunks-flat-response")
    async with AsyncSDK(server_url=HTTPBIN_URL, api_key_auth="token") as s:
        assert s is not None

        res = await s.jsonl.jsonl_stream_chunks()

        assert isinstance(res, operations.JsonlStreamChunksResponse)
        assert isinstance(res.object, jsonl.JsonLStreamAsync)

        result: List[JsonlStreamChunksResponseBody] = []
        async with res.object as jsonl_stream:
            async for item in jsonl_stream:
                result.append(item)
        assert len(result) == 2


        # Assert first event
        assert result[0].name == 'Peter'
        assert result[0].skills == ['Go', 'Python']

        # Assert second event
        assert result[1].name == 'John'
        assert result[1].skills == ['Go', 'Rust']
