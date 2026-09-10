from openapi import SDK

from .common_helpers import *
from .test_helpers import *
import httpx


def test_globals_query_parameter_get_uses_global():
    record_test("globals-query-parameter-get-uses-global")

    s = SDK(server_url=HTTPBIN_URL, global_query_param="test")
    assert s is not None

    res = s.globals.globals_query_parameter_get()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.args.global_query_param == "test"


def test_globals_query_parameter_get_uses_local():
    record_test("globals-query-parameter-get-uses-local")

    s = SDK(server_url=HTTPBIN_URL, global_query_param="test")
    assert s is not None

    res = s.globals.globals_query_parameter_get(global_query_param="local")
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.args.global_query_param == "local"


def test_global_path_parameter_get_uses_global():
    record_test("globals-path-parameter-get-uses-global")

    s = SDK(server_url=HTTPBIN_URL, global_path_param=1)
    assert s is not None

    res = s.globals.global_path_parameter_get()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.url == f"{HTTPBIN_URL}/anything/globals/pathParameter/1"


def test_global_path_parameter_get_uses_local():
    record_test("globals-path-parameter-get-uses-local")

    s = SDK(server_url=HTTPBIN_URL, global_path_param=1)
    assert s is not None

    res = s.globals.global_path_parameter_get(global_path_param=2)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.url == f"{HTTPBIN_URL}/anything/globals/pathParameter/2"


def test_global_header_get_uses_global():
    record_test("globals-header-get-uses-global")

    s = SDK(server_url=HTTPBIN_URL, global_header_param=True)
    assert s is not None

    res = s.globals.globals_header_get()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.headers is not None
    assert res.res.headers["Globalheaderparam"] == "true"


def test_global_header_get_uses_local():
    record_test("globals-header-get-uses-local")

    s = SDK(server_url=HTTPBIN_URL, global_header_param=True)
    assert s is not None

    res = s.globals.globals_header_get(global_header_param=False)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.headers is not None
    assert res.res.headers["Globalheaderparam"] == "false"


def test_global_header_keeps_custom_client_headers():
    record_test("globals-header-keeps-custom-client-headers")

    http_client = httpx.Client(headers={"x-custom-header": "someValue"})

    s = SDK(server_url=HTTPBIN_URL, global_header_param=True, client=http_client)
    assert s is not None

    res = s.globals.globals_header_get()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.headers is not None
    assert res.res.headers["Globalheaderparam"] == "true"
    assert res.res.headers["X-Custom-Header"] == "someValue"


def test_globals_hidden_post():
    record_test("globals-hidden-post")

    s = SDK(
        server_url=HTTPBIN_URL,
        global_hidden_query_param="hello",
        global_hidden_header_param="world",
        global_hidden_path_param="test",
    )
    assert s is not None

    res = s.globals.globals_hidden_post(
        test="friend",
        other=37,
    )
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.args.global_hidden_query_param == "hello"
    assert res.res.json_.test == "friend"
    assert res.res.json_.other == 37
    assert res.res.headers["Globalhiddenheaderparam"] == "world"
    assert (
        res.res.url
        == f"{HTTPBIN_URL}/anything/globals/hidden/test?globalHiddenQueryParam=hello")


def test_globals_operation_params_only():
    record_test("globals-operation-params-only")

    # Initialize SDK with ALL global parameters
    s = SDK(
        server_url=HTTPBIN_URL,
        global_query_param="globalQueryValue",
        global_path_param=999,
        global_header_param=True,
        global_hidden_query_param="hiddenQueryValue",
        global_hidden_header_param="hiddenHeaderValue",
        global_hidden_path_param="hiddenPathValue",
    )
    assert s is not None

    # Call operation with operation-specific parameters
    res = s.globals.globals_operation_scoped_exclusive(
        operation_query_param="operationQuery",
        operation_path_param="operationPath",
        operation_header_param="operationHeader",
    )
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None

    # Verify that ONLY operation-specific parameters are present in the request
    # Query params should NOT contain global params
    assert (
        "globalQueryParam" not in res.res.args
    ), "Global query param should NOT be included when operation defines its own params"
    assert (
        "globalHiddenQueryParam" not in res.res.args
    ), "Global hidden query param should NOT be included when operation defines its own params"

    # URL should contain operation path param, NOT global path param
    assert (
        res.res.url
        == f"{HTTPBIN_URL}/anything/globals/operationScopedExclusive/operationPath?operationQueryParam=operationQuery"
    ), "URL should use operation-specific path parameter, not global"

    # Headers should NOT contain global headers
    assert (
        "Globalheaderparam" not in res.res.headers
    ), "Global header param should NOT be included when operation defines its own params"
    assert (
        "Globalhiddenheaderparam" not in res.res.headers
    ), "Global hidden header param should NOT be included when operation defines its own params"


def test_globals_kebab_case_param_get():
    record_test("globals-kebab-case-param-get")

    s = SDK(server_url=HTTPBIN_URL, kebab_case_param="kebab-case-value")
    assert s is not None

    res = s.globals.globals_kebab_case_param_get()
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.res is not None
    assert res.res.args["kebab-case-param"] == "kebab-case-value"
