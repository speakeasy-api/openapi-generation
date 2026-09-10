import uuid

import httpx
import pytest
from openapi import SDK
from openapi.utils import BackoffStrategy, Retries, RetryConfig, retry, retry_async

from .common_helpers import API_TEST_SERVICE_URL


# A client supplied by the caller may come from an httpx-compatible
# distribution that mirrors httpx's exception names without subclassing them.
# These stand in for that hierarchy so the test needs no extra dependency.
class HTTPError(Exception):
    pass


class RequestError(HTTPError):
    pass


class TransportError(RequestError):
    pass


class NetworkError(TransportError):
    pass


class ConnectError(NetworkError):
    pass


class TimeoutException(TransportError):
    pass


class ReadTimeout(TimeoutException):
    pass


# A library that reuses one of the names without reproducing httpx's hierarchy
# (httpcore is the real-world case). These must keep failing fast.
class UnrelatedNetworkError(Exception):
    pass


UnrelatedNetworkError.__name__ = "NetworkError"


def _retries(retry_connection_errors: bool = True) -> Retries:
    return Retries(
        RetryConfig(
            "backoff",
            BackoffStrategy(1, 5, 1.1, 200),
            retry_connection_errors,
        ),
        ["5XX"],
    )


def test_fork_transport_errors_are_not_subclasses_of_httpx():
    # Guards the premise of this module: if a fork's errors ever did subclass
    # httpx's, these tests would pass for the wrong reason.
    assert not issubclass(NetworkError, httpx.NetworkError)
    assert not issubclass(TimeoutException, httpx.TimeoutException)


@pytest.mark.parametrize(
    "exception",
    [ConnectError("connection failed"), ReadTimeout("timed out")],
)
def test_retries_transport_errors_from_compatible_client(exception):
    attempts = 0

    def do_request(*_args):
        nonlocal attempts
        attempts += 1
        raise exception

    with pytest.raises(type(exception)):
        retry(do_request, _retries())

    assert attempts > 1, "transport error from a compatible client was not retried"


@pytest.mark.asyncio()
async def test_retries_transport_errors_from_compatible_client_async():
    attempts = 0

    async def do_request(*_args):
        nonlocal attempts
        attempts += 1
        raise ConnectError("connection failed")

    with pytest.raises(ConnectError):
        await retry_async(do_request, _retries())

    assert attempts > 1, "transport error from a compatible client was not retried"


def test_does_not_retry_transport_errors_when_disabled():
    attempts = 0

    def do_request(*_args):
        nonlocal attempts
        attempts += 1
        raise ConnectError("connection failed")

    with pytest.raises(ConnectError):
        retry(do_request, _retries(retry_connection_errors=False))

    assert attempts == 1


def test_does_not_retry_name_collision_without_httpx_hierarchy():
    attempts = 0

    def do_request(*_args):
        nonlocal attempts
        attempts += 1
        raise UnrelatedNetworkError("not an httpx-shaped transport error")

    with pytest.raises(UnrelatedNetworkError):
        retry(do_request, _retries())

    assert attempts == 1


def test_does_not_retry_unrelated_errors():
    attempts = 0

    def do_request(*_args):
        nonlocal attempts
        attempts += 1
        raise ValueError("not a transport failure")

    with pytest.raises(ValueError):
        retry(do_request, _retries())

    assert attempts == 1


def _attempt_count_retries(max_retries: int = 2) -> Retries:
    return Retries(
        RetryConfig(
            "attempt-count-backoff",
            BackoffStrategy(1, 5, 1.1, 200),
            False,
            max_retries,
        ),
        ["5XX"],
    )


def _streamed_response(status: int) -> httpx.Response:
    # A response whose body has not been read holds its connection until it is
    # closed, which is what a streaming operation returns.
    return httpx.Response(status, content=iter([b"streamed body"]))


def _async_streamed_response(status: int) -> httpx.Response:
    async def body():
        yield b"streamed body"

    return httpx.Response(status, content=body())


def _assert_only_returned_response_is_open(responses, result, expected_attempts):
    assert len(responses) == expected_attempts
    assert all(res.is_closed for res in responses[:-1]), (
        "a retried response was left open, pinning its connection in the pool"
    )
    assert not responses[-1].is_closed, (
        "the returned response must stay open for the caller to read"
    )
    assert result is responses[-1]


def _single_connection_limits() -> httpx.Limits:
    # One connection, so a retried response that is never closed strands the
    # only connection in the pool and the next attempt fails on acquire.
    return httpx.Limits(max_connections=1, max_keepalive_connections=1)


_POOL_TIMEOUT = httpx.Timeout(10.0, pool=5.0)

_STREAMING_RETRIES = RetryConfig("backoff", BackoffStrategy(10, 100, 1.1, 5000), False)


def test_streaming_retries_do_not_exhaust_the_connection_pool():
    # The SDK only closes clients it created itself, so the caller-supplied
    # client is scoped here to keep this test from leaking what it asserts about.
    with httpx.Client(
        limits=_single_connection_limits(), timeout=_POOL_TIMEOUT
    ) as client:
        s = SDK(server_url=API_TEST_SERVICE_URL, client=client)

        with s.retries.with_streaming_response.retries_get(
            request_id=str(uuid.uuid4()),
            num_retries=3,
            retries=_STREAMING_RETRIES,
        ) as res:
            assert res.http_response.status_code == 200
            assert res.parse().retries.retries == 3


@pytest.mark.asyncio()
async def test_streaming_retries_do_not_exhaust_the_connection_pool_async():
    async with httpx.AsyncClient(
        limits=_single_connection_limits(), timeout=_POOL_TIMEOUT
    ) as client:
        s = SDK(server_url=API_TEST_SERVICE_URL, async_client=client)

        async with s.retries.with_streaming_response.retries_get_async(
            request_id=str(uuid.uuid4()),
            num_retries=3,
            retries=_STREAMING_RETRIES,
        ) as res:
            assert res.http_response.status_code == 200
            assert (await res.parse()).retries.retries == 3


def test_leaves_exhausted_response_open_for_caller():
    responses = []

    def do_request(*_args):
        responses.append(_streamed_response(500))
        return responses[-1]

    result = retry(do_request, _attempt_count_retries())

    _assert_only_returned_response_is_open(responses, result, 3)


@pytest.mark.asyncio()
async def test_leaves_exhausted_response_open_for_caller_async():
    responses = []

    async def do_request(*_args):
        responses.append(_async_streamed_response(500))
        return responses[-1]

    result = await retry_async(do_request, _attempt_count_retries())

    _assert_only_returned_response_is_open(responses, result, 3)
