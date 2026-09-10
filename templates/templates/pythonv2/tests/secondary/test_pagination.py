from openapi import SDK
from openapi.models import operations
from .common_helpers import record_test, HTTPBIN_URL


def test_pagination_limit_offset_page_params_flat():
    record_test("pagination-limit-offset-page-params-flat")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_page_params(page=1)

    assert res is not None
    assert res.result is not None
    assert len(res.result.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.result is not None
    assert len(next_res.result.result_array) == 0

    null_res = next_res.next()
    assert null_res is None


def test_pagination_limit_offset_union_output_page_params_flat():
    record_test("pagination-limit-offset-union-output-page-params-flat")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None
    server_limit = 20

    res = s.pagination.pagination_limit_offset_union_output_page_params(page=1)

    assert res is not None
    assert res.result is not None
    assert isinstance(res.result, operations.ResultObject)
    assert len(res.result.result_array) == server_limit

    next_res = res.next()
    assert next_res is not None
    assert next_res.result is not None
    assert isinstance(next_res.result, operations.ResultObject)
    assert len(next_res.result.result_array) == 0

    null_res = next_res.next()
    assert null_res is None
