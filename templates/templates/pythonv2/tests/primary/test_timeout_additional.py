import httpx
import pytest
from openapi import SDK

from .common_helpers import *


def test_timeout_ms_override_is_respected():
    """Verify that a per-request timeout_ms override actually causes the request
    to time out when the server takes longer than the specified timeout."""
    record_test("timeout-ms-override-is-respected")

    s = SDK()

    # The delay endpoint waits for the specified number of seconds before
    # responding. With a 1-second timeout against a 3-second delay, the
    # request must time out.
    with pytest.raises(httpx.ReadTimeout):
        s.cancellation.cancelled_request(seconds=3, timeout_ms=1000)


def test_timeout_ms_override_allows_completion():
    """Verify that a per-request timeout_ms override does not interfere with
    requests that complete within the timeout window."""
    record_test("timeout-ms-override-allows-completion")

    s = SDK()

    # A 5-second timeout against a 1-second delay should succeed.
    res = s.cancellation.cancelled_request(seconds=1, timeout_ms=5000)

    assert res is not None
    assert res.object is not None
    assert res.object.delay_seconds == 1
