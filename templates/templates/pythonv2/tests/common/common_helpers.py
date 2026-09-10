import os
import random
import string
import httpx
from typing import List

# Test service URLs - read from environment with defaults for local development
HTTPBIN_PORT = os.environ.get("HTTPBIN_PORT", "35123")
API_TEST_SERVICE_PORT = os.environ.get("API_TEST_SERVICE_PORT", "35456")
HTTPBIN_URL = f"http://localhost:{HTTPBIN_PORT}"
API_TEST_SERVICE_URL = f"http://localhost:{API_TEST_SERVICE_PORT}"


def record_test(test_id: str):
    with open("test-python-record.txt", "a", -1, "utf-8") as f:
        f.write(test_id + "\n")


def rand_seq(n) -> str:
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=n))


class RequestLogEntry:
    def __init__(self, request_url: str, request_body: str, status_code: int):
        self.request_url = request_url
        self.request_body = request_body
        self.status_code = status_code


class RequestRecorderClient(httpx.Client):
    """ Test client to record requests"""
    def __init__(self):
        super().__init__()
        self.log: List[RequestLogEntry] = []

    def send(self, request: httpx.Request, **kwargs) -> httpx.Response:
        request_body = ""
        if hasattr(request, 'content') and request.content:
            request_body = request.content.decode('utf-8')

        response = super().send(request, **kwargs)

        self.log.append(RequestLogEntry(
            request_url=str(request.url),
            request_body=request_body,
            status_code=response.status_code
        ))

        return response
