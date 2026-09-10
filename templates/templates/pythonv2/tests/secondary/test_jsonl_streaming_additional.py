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


def test_jsonl_stream_data():
    record_test("jsonl-stream-data-flat-responses")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.jsonl_stream()

        assert res is not None
        
        assert isinstance(res, jsonl.JsonLStream)

        result: List[JsonlStreamResponseBody] = []
        with res as jsonl_stream:
            for item in jsonl_stream:
                result.append(item)
        assert len(result) == 2


        # Assert first event
        assert result[0].name == 'Peter'
        assert result[0].skills == ['Go', 'Python']
        
        # Assert second event
        assert result[1].name == 'John'
        assert result[1].skills == ['Go', 'Rust']


def test_jsonl_stream_data_chunks():
    record_test("jsonl-stream-data-chunks-flat-response")

    with SDK(server_url=HTTPBIN_URL) as s:
        assert s is not None

        res = s.jsonl.jsonl_stream_chunks()

        assert isinstance(res, jsonl.JsonLStream)

        result: List[JsonlStreamChunksResponseBody] = []
        with res as jsonl_stream:
            for item in jsonl_stream:
                result.append(item)
        assert len(result) == 2


        # Assert first event
        assert result[0].name == 'Peter'
        assert result[0].skills == ['Go', 'Python']
        
        # Assert second event
        assert result[1].name == 'John'
        assert result[1].skills == ['Go', 'Rust']
