import asyncio
import uuid

import httpx
import pytest
from openapi import SDK
from openapi.models import errors
from openapi.models.operations import *
from openapi.utils import BackoffStrategy, Retries, RetryConfig

from .common_helpers import *
from .test_helpers import *


def test_retries_succeeds():
    record_test("retries-succeeds")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_get(request_id=str(uuid.uuid4()))
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


@pytest.mark.asyncio()
async def test_retries_succeeds_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = await s.retries.retries_get_async(request_id=str(uuid.uuid4()))
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_request_timeout():
    record_test("retries-request-timeout")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        Exception,
    ):
        s.retries.retries_get_timeout(
            request_id_query_parameter=str(uuid.uuid4()),
            num_retries=10,
            timeout_ms=1,
        )


def test_retries_timeout():
    record_test("retries-timeout")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        s.retries.retries_get(
            request_id=str(uuid.uuid4()),
            num_retries=1000000000,
            retries=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False),
        )

    assert exc_info.value.status_code == 503


@pytest.mark.asyncio()
async def test_global_retry_config_disable_async():
    s = SDK(server_url=HTTPBIN_URL, retry_config=None)

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        await s.retries.retries_get_async(request_id=str(uuid.uuid4()), num_retries=2)

    assert exc_info.value.status_code == 503


def test_global_retry_config_disable():
    record_test("retries-global-config-disable")

    s = SDK(server_url=HTTPBIN_URL, retry_config=None)

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        s.retries.retries_get(request_id=str(uuid.uuid4()), num_retries=2)

    assert exc_info.value.status_code == 503


@pytest.mark.asyncio()
async def test_retries_timeout_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        await s.retries.retries_get_async(
            request_id=str(uuid.uuid4()),
            num_retries=1000000000,
            retries=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False),
        )

    assert exc_info.value.status_code == 503


def test_global_retry_config_success():
    record_test("retries-global-config-success")

    s = SDK(
        server_url=HTTPBIN_URL,
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 1000), False)
    )

    res = s.retries.retries_get(request_id=str(uuid.uuid4()), num_retries=10)
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 10


@pytest.mark.asyncio()
async def test_global_retry_config_success_async():
    s = SDK(
        server_url=HTTPBIN_URL,
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 1000), False)
    )

    res = await s.retries.retries_get_async(
        request_id=str(uuid.uuid4()), num_retries=10
    )
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 10



def test_global_retry_config_timeout():
    record_test("retries-global-config-timeout")

    s = SDK(
        server_url=HTTPBIN_URL,
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False)
    )

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        s.retries.retries_get(request_id=str(uuid.uuid4()), num_retries=30)

    assert exc_info.value.status_code == 503


@pytest.mark.asyncio()
async def test_global_retry_config_timeout_async():
    s = SDK(
        server_url=HTTPBIN_URL,
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 100), False)
    )

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        await s.retries.retries_get_async(request_id=str(uuid.uuid4()), num_retries=30)

    assert exc_info.value.status_code == 503


def test_retries_connect_error():
    record_test("retries-connect-error")

    s = SDK(
        server_url=HTTPBIN_URL,
        retry_config=RetryConfig("backoff", BackoffStrategy(1, 50, 1.1, 1000), False)
    )
    assert s is not None

    with pytest.raises(httpx.ConnectError, match="Connection refused"):
        s.retries.retries_connect_error_get()


def test_retries_header():
    record_test("retries-header")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_after(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        retry_after_val=1,
        retries=RetryConfig("backoff", BackoffStrategy(5000, 10000, 1.1, 5000), False),
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_attempt_count_backoff():
    record_test("retries-attempt-count-backoff")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_attempt_count(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        retry_after_ms_val=1,
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_attempt_count_backoff_zero():
    record_test("retries-attempt-count-zero")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        s.retries.retries_attempt_count_zero(
            request_id=str(uuid.uuid4()),
            num_retries=2,
            retry_after_ms_val=1,
        )

    assert exc_info.value.status_code == 503


@pytest.mark.asyncio()
async def test_retries_attempt_count_backoff_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = await s.retries.retries_attempt_count_async(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        retry_after_ms_val=1,
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


@pytest.mark.asyncio()
async def test_retries_attempt_count_backoff_zero_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    with pytest.raises(
        errors.APIError, match="API error occurred: Status 503"
    ) as exc_info:
        await s.retries.retries_attempt_count_zero_async(
            request_id=str(uuid.uuid4()),
            num_retries=2,
            retry_after_ms_val=1,
        )

    assert exc_info.value.status_code == 503


@pytest.mark.asyncio()
async def test_retries_header_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = await s.retries.retries_after_async(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        retry_after_val=1,
        retries=RetryConfig("backoff", BackoffStrategy(5000, 10000, 1.1, 5000), False),
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_succeeds_with_body():
    record_test("retries-succeeds-with-body")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_post(request_id=str(uuid.uuid4()), field_one="one")
    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def _override_config():
    return RetryConfig(
        "attempt-count-backoff",
        BackoffStrategy(1, 50, 1.1, 100),
        False,
        max_retries=5,
        status_codes_override=["521"],
    )


def test_status_codes_override_resolution():
    cfg = _override_config()
    assert Retries(cfg, ["503"]).status_codes == ["521"]


def test_status_codes_override_default():
    cfg = RetryConfig(
        "attempt-count-backoff",
        BackoffStrategy(1, 50, 1.1, 100),
        False,
        status_codes_override=[],
    )
    assert Retries(cfg, ["503"]).status_codes == ["503"]


def test_status_codes_override_none_falls_back():
    cfg = RetryConfig(
        "attempt-count-backoff",
        BackoffStrategy(1, 50, 1.1, 100),
        False,
    )
    assert Retries(cfg, ["503"]).status_codes == ["503"]


def test_retries_status_codes_override():
    record_test("retries-status-codes-override")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_attempt_count(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        status_code=521,
        retry_after_ms_val=1,
        retries=_override_config(),
    )

    assert res is not None
    assert res.http_meta is not None
    assert res.http_meta.response is not None
    assert res.http_meta.response.status_code == 200
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_status_codes_override_default():
    record_test("retries-status-codes-override-default")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.retries.retries_attempt_count(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        status_code=503,
        retry_after_ms_val=1,
        retries=RetryConfig(
            "attempt-count-backoff",
            BackoffStrategy(1, 50, 1.1, 100),
            False,
            max_retries=5,
            status_codes_override=[],
        ),
    )

    assert res is not None
    assert res.retries is not None
    assert res.retries.retries == 3


def test_retries_status_codes_override_global():
    record_test("retries-status-codes-override-global")

    s = SDK(server_url=HTTPBIN_URL, retry_config=_override_config())
    assert s is not None

    res = s.retries.retries_attempt_count(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        status_code=521,
        retry_after_ms_val=1,
    )

    assert res is not None
    assert res.retries is not None
    assert res.retries.retries == 3


@pytest.mark.asyncio()
async def test_retries_status_codes_override_applied_async():
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = await s.retries.retries_attempt_count_async(
        request_id=str(uuid.uuid4()),
        num_retries=3,
        status_code=521,
        retry_after_ms_val=1,
        retries=_override_config(),
    )

    assert res is not None
    assert res.retries is not None
    assert res.retries.retries == 3


def test_get_sleep_interval_jitter_additive(monkeypatch):
    """jitter set: sleep += uniform(0, jitter/1000); asserts bound passed + additive."""
    from openapi.utils.retries import _get_sleep_interval

    calls = []
    monkeypatch.setattr(
        "openapi.utils.retries.random.uniform",
        lambda a, b: calls.append((a, b)) or b,
    )
    sleep = _get_sleep_interval(
        Exception(), 1000, 60000, 2.0, 0, attempt_count_backoff=True, jitter_ms=500
    )
    assert calls == [(0, 0.5)]
    assert sleep == 1.5


def test_get_sleep_interval_jitter_unset_attempt_count(monkeypatch):
    """jitter unset, attempt-count: unchanged multiplicative jitter *(1 - rand*0.25)."""
    from openapi.utils.retries import _get_sleep_interval

    monkeypatch.setattr("openapi.utils.retries.random.random", lambda: 1.0)
    sleep = _get_sleep_interval(
        Exception(), 1000, 60000, 2.0, 0, attempt_count_backoff=True
    )
    assert sleep == 0.75


def test_get_sleep_interval_jitter_unset_backoff(monkeypatch):
    """jitter unset, backoff: unchanged additive jitter +uniform(0, 1)."""
    from openapi.utils.retries import _get_sleep_interval

    monkeypatch.setattr("openapi.utils.retries.random.uniform", lambda a, b: b)
    sleep = _get_sleep_interval(Exception(), 1000, 60000, 2.0, 0)
    assert sleep == 2.0


def test_get_sleep_interval_jitter_capped(monkeypatch):
    """sleep capped at max_interval/1000 even when jitter overshoots."""
    from openapi.utils.retries import _get_sleep_interval

    monkeypatch.setattr("openapi.utils.retries.random.uniform", lambda a, b: b)
    sleep = _get_sleep_interval(Exception(), 1000, 5000, 2.0, 0, jitter_ms=100000)
    assert sleep == 5.0
