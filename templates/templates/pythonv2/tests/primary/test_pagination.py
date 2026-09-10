import json
from typing import Any, Dict, List, Tuple
import uuid
import httpx
import logging
import pytest
from openapi import SDK
from openapi.models.operations import *
from openapi.models.shared import *
from openapi.utils import *

from .common_helpers import *
from .test_helpers import *


def test_pagination_limit_offset_page_params():
    record_test("pagination-limit-offset-page-params")
    logging.basicConfig(level=logging.DEBUG)
    debug_logger = logging.getLogger("sdk")
    s = SDK(server_url=HTTPBIN_URL, debug_logger=debug_logger)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_page_params(page=1)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 0

    null_res = next_res.next()
    assert null_res is None


@pytest.mark.asyncio()
async def test_pagination_limit_offset_page_params_async():
    logging.basicConfig(level=logging.DEBUG)
    debug_logger = logging.getLogger("sdk")
    async with SDK(server_url=HTTPBIN_URL, debug_logger=debug_logger) as s:
        assert s is not None
        server_limit = 20

        res = await s.pagination.pagination_limit_offset_page_params_async(page=1)

        assert res is not None
        assert res.http_meta is not None
        assert res.http_meta.response is not None
        assert res.http_meta.response.status_code == 200
        assert res.res is not None
        assert len(res.res.result_array) == server_limit

        next_res = await res.next()
        assert next_res is not None
        assert next_res.http_meta is not None
        assert next_res.http_meta.response is not None
        assert next_res.http_meta.response.status_code == 200
        assert next_res.res is not None
        assert len(next_res.res.result_array) == 0

        null_res = await next_res.next()
        assert null_res is None


def test_pagination_limit_offset_union_output_page_params():
    record_test("pagination-limit-offset-union-output-page-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_union_output_page_params(page=1)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert isinstance(res.res, ResultObject)
    assert len(res.res.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert isinstance(next_res.res, ResultObject)
    assert len(next_res.res.result_array) == 0

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_nil_page_params():
    record_test("pagination-limit-offset-nil-page-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_optional_page_params()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 0

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_zero_page_params():
    record_test("pagination-limit-offset-zero-page-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_optional_page_params(page=0)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == server_limit

    assert res.http_meta.request is not None
    assert res.http_meta.request.url.params.get("page") == "0"

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 20

    assert next_res.http_meta.request is not None
    assert next_res.http_meta.request.url.params.get("page") == "1"


def test_pagination_limit_offset_page_body():
    record_test("pagination-limit-offset-page-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_limit_offset_page_body(
        limit=limit, page=1
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_page_body_nullable():
    record_test("pagination-limit-offset-page-body-nullable")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # first request sends a null body; later pages materialize one with the
    # advanced page (wire sequence enforced by the test service)
    res = s.pagination.pagination_limit_offset_page_body_nullable(request=None)

    assert res is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.http_meta.request is not None
    assert json.loads(res.http_meta.request.content.decode("utf-8")) is None
    assert res.res is not None
    assert res.res.result_array == [0, 1, 2, 3, 4, 5, 6]

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200

    body: Dict[str, Any] = json.loads(
        next_res.http_meta.request.content.decode("utf-8")
    )
    assert body is not None
    assert body.get("page") == 2
    assert next_res.res is not None
    assert next_res.res.result_array == [7, 8, 9, 10, 11, 12, 13]

    last_res = next_res.next()
    assert last_res is not None
    assert last_res.http_meta.response is not None
    assert last_res.http_meta.response.status_code == 200

    body = json.loads(last_res.http_meta.request.content.decode("utf-8"))
    assert body is not None
    assert body.get("page") == 3
    assert last_res.res is not None
    assert last_res.res.result_array == [14, 15, 16, 17, 18, 19]

    null_res = last_res.next()
    assert null_res is None


def test_pagination_wrapped_optional_body():
    record_test("pagination-wrapped-optional-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # the body is omitted entirely; later pages materialize one carrying the
    # advanced offset
    res = s.pagination.pagination_wrapped_optional_body()

    assert res is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.http_meta.request is not None
    assert not res.http_meta.request.content
    assert res.res is not None
    assert len(res.res.result_array) == 20

    empty_res = res.next()
    assert empty_res is not None
    assert empty_res.http_meta.response is not None
    assert empty_res.http_meta.response.status_code == 200

    body: Dict[str, Any] = json.loads(
        empty_res.http_meta.request.content.decode("utf-8")
    )
    assert body is not None
    assert body.get("offset") == 20
    assert empty_res.res is not None
    assert empty_res.res.result_array == []

    null_res = empty_res.next()
    assert null_res is None


def test_pagination_limit_offset_deep_outputs_page_body():
    record_test("pagination-limit-offset-deep-outputs-page-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_limit_offset_deep_outputs_page_body(
        limit=limit, page=1
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_offset_params():
    record_test("pagination-limit-offset-offset-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_limit_offset_offset_params(limit=limit, offset=0)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_nil_offset_params():
    record_test("pagination-limit-offset-nil-offset-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    default_limit = 20

    res = s.pagination.pagination_limit_offset_offset_params()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == default_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < default_limit

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_offset_body():
    record_test("pagination-limit-offset-offset-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_limit_offset_offset_body(
        limit=limit, offset=0
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_nullable():
    record_test("pagination-limit-offset-nullable")

    # Passing None for the required+nullable offset/limit forces the SDK to
    # serialise null values; the generated paginator must narrow them to int
    # before computing the next offset and comparing against limit, otherwise
    # mypy/pyright reject the arithmetic on a potentially-None operand.
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 10

    res = s.pagination.pagination_offset_nullable(limit=None, offset=None)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.res is not None
    assert len(next_res.res.result_array) == server_limit

    empty_res = next_res.next()
    assert empty_res is not None
    assert empty_res.res is not None
    assert len(empty_res.res.result_array) == 0

    null_res = empty_res.next()
    assert null_res is None


def test_pagination_limit_offset_default_offset_body():
    record_test("pagination-limit-offset-default-offset-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_limit_offset_default_offset_body()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    assert res.http_meta.request is not None

    body: Dict[str, Any] = json.loads(res.http_meta.request.content.decode("utf-8"))
    assert body is not None
    assert body.get("limit") == 15
    assert body.get("offset") == 10


def test_pagination_limit_offset_default_offset_params():
    record_test("pagination-limit-offset-default-offset-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_limit_offset_default_offset_params()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200

    assert res.http_meta.request is not None
    assert res.http_meta.request.url.params.get("limit") == "15"
    assert res.http_meta.request.url.params.get("offset") == "10"


def test_pagination_cursor_params():
    record_test("pagination-cursor-params")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_cursor_params(cursor=-1)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    penultimate_res = next_res.next()
    assert penultimate_res is not None
    assert penultimate_res.http_meta is not None
    assert penultimate_res.http_meta.response is not None
    assert penultimate_res.http_meta.response.status_code == 200
    assert penultimate_res.res is not None
    assert len(penultimate_res.res.result_array) == 0

    null_res = penultimate_res.next()
    assert null_res is None


def test_pagination_cursor_body():
    record_test("pagination-cursor-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    limit = 15

    res = s.pagination.pagination_cursor_body(
        cursor=-1,
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) < limit

    penultimate_res = next_res.next()
    assert penultimate_res is not None
    assert penultimate_res.http_meta is not None
    assert penultimate_res.http_meta.response is not None
    assert penultimate_res.http_meta.response.status_code == 200
    assert penultimate_res.res is not None
    assert len(penultimate_res.res.result_array) == 0

    null_res = penultimate_res.next()
    assert null_res is None


def test_pagination_cursor_non_numeric():
    record_test("pagination-cursor-non-numeric")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_cursor_non_numeric()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == 15

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 5

    penultimate_res = next_res.next()
    assert penultimate_res is not None
    assert penultimate_res.http_meta is not None
    assert penultimate_res.http_meta.response is not None
    assert penultimate_res.http_meta.response.status_code == 200
    assert penultimate_res.res is not None
    assert len(penultimate_res.res.result_array) == 0

    null_res = penultimate_res.next()
    assert null_res is None


def test_pagination_cursor_non_numeric_nullable():
    record_test("pagination-cursor-non-numeric-nullable")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_cursor_non_numeric_nullable(cursor="2")

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == 15
    assert res.res.cursor == "17"

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 2
    assert next_res.res.cursor is None

    null_res = next_res.next()
    assert null_res is None


def test_pagination_cursor_non_numeric_empty_string():
    record_test("pagination-cursor-non-numeric-empty-string")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_cursor_non_numeric_empty_string(cursor="2", end_cursor="")

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == 15
    assert res.res.cursor == "17"

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 2
    assert next_res.res.cursor == ""

    null_res = next_res.next()
    assert null_res is None


def test_pagination_with_retries():
    record_test("pagination-with-retries")

    recorder = PaginationRecorder()
    http_client = httpx.Client()
    http_client.event_hooks["response"].append(recorder.log_response)

    s = SDK(server_url=HTTPBIN_URL, client=http_client)

    assert s is not None

    count = 0

    res = s.pagination.pagination_with_retries(
        request_id=str(uuid.uuid4()),
        fault_settings='{"error_code": 503, "error_count": 3}',
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    count += len(res.res.result_array)

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    count += len(next_res.res.result_array)

    penultimate_res = next_res.next()
    assert penultimate_res is not None
    assert penultimate_res.http_meta is not None
    assert penultimate_res.http_meta.response is not None
    assert penultimate_res.http_meta.response.status_code == 200
    assert penultimate_res.res is not None
    assert len(penultimate_res.res.result_array) == 0

    null_res = penultimate_res.next()
    assert null_res is None

    assert count == 20
    assert recorder.log == [
        (503, "GET", "/pagination/cursor_non_numeric"),
        (503, "GET", "/pagination/cursor_non_numeric"),
        (503, "GET", "/pagination/cursor_non_numeric"),
        (200, "GET", "/pagination/cursor_non_numeric"),
        (200, "GET", "/pagination/cursor_non_numeric"),
        (200, "GET", "/pagination/cursor_non_numeric"),
    ]


class PaginationRecorder:
    log: List[Tuple[int, str, str]] = []

    def log_response(self, res: httpx.Response):
        self.log.append((res.status_code, str(res.request.method), res.url.path))


def test_pagination_cursor_nullable():
    record_test("pagination-cursor-nullable-limit")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    default_limit = 10

    # Call without passing limit — limit is OptionalNullable[int] and defaults
    # to UNSET. The pagination template must handle UNSET (not just None).
    res = s.pagination.pagination_cursor_nullable_limit()

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.pagination_cursor_nullable_limit_next_cursor is not None
    assert len(res.pagination_cursor_nullable_limit_next_cursor.results) == default_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.pagination_cursor_nullable_limit_next_cursor is not None
    assert len(next_res.pagination_cursor_nullable_limit_next_cursor.results) == default_limit

    # Keep paginating until exhausted
    last_res = next_res
    while True:
        n = last_res.next()
        if n is None:
            break
        last_res = n

    # Final page should have been the termination
    assert last_res.pagination_cursor_nullable_limit_next_cursor is not None


def test_pagination_url():
    record_test("pagination-url")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.pagination.pagination_url_params(attempts=3)

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert len(res.res.result_array) == 9

    next_res = res.next()
    assert next_res is not None
    assert next_res.http_meta is not None
    assert next_res.http_meta.response is not None
    assert next_res.http_meta.response.status_code == 200
    assert next_res.res is not None
    assert len(next_res.res.result_array) == 6

    penultimate_res = next_res.next()
    assert penultimate_res is not None
    assert penultimate_res.http_meta is not None
    assert penultimate_res.http_meta.response is not None
    assert penultimate_res.http_meta.response.status_code == 200
    assert penultimate_res.res is not None
    assert len(penultimate_res.res.result_array) == 3

    null_res = penultimate_res.next()
    assert null_res is None

    res2 = s.pagination.pagination_url_params(attempts=3, is_reference_path="true")

    assert res2 is not None
    assert res2.http_meta is not None
    assert res2.http_meta.response is not None
    assert res2.http_meta.response.status_code == 200
    assert res2.res is not None
    assert len(res2.res.result_array) == 9

    next_res2 = res2.next()
    assert next_res2 is not None
    assert next_res2.http_meta is not None
    assert next_res2.http_meta.response is not None
    assert next_res2.http_meta.response.status_code == 200
    assert next_res2.res is not None
    assert len(next_res2.res.result_array) == 6

    penultimate_res2 = next_res2.next()
    assert penultimate_res2 is not None
    assert penultimate_res2.http_meta is not None
    assert penultimate_res2.http_meta.response is not None
    assert penultimate_res2.http_meta.response.status_code == 200
    assert penultimate_res2.res is not None
    assert len(penultimate_res2.res.result_array) == 3

    null_res2 = penultimate_res2.next()
    assert null_res2 is None
