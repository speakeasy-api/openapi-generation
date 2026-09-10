import pytest
import httpx
from pydantic import ValidationError
from openapi import SDK
from openapi.models import errors
from unittest.mock import Mock, patch

from .common_helpers import *
from .test_helpers import *


def test_response_validation_error_handling():
    """Test that ValidationError is properly handled when server returns data that doesn't match expected schema."""

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # Create a real httpx.Response object with malformed data
    import httpx
    
    # Create a request first (required for Response)
    req = httpx.Request("GET", "http://test.com/json")
    
    # Create response with status 200 and malformed JSON data
    mock_response = httpx.Response(
        status_code=200,
        headers={"Content-Type": "application/json"},
        content=b'{"invalid_field": "value", "missing_required_field": null}',  # Wrong structure
        request=req
    )
    
    # Patch the HTTP client to return our malformed response 
    with patch.object(s.sdk_configuration.client, 'send', return_value=mock_response):
        # This should trigger a ResponseValidationError when trying to parse the response
        # into the expected Pydantic model
        with pytest.raises(errors.ResponseValidationError, match=r"Response validation failed") as exc_info:
            s.response_body_json_get()
        
        # Verify the ResponseValidationError contains details about the validation failure
        assert exc_info.value is not None
        assert exc_info.value.status_code == 200
        assert isinstance(exc_info.value.cause, ValidationError)


def test_malformed_json_response_handling():
    """Test handling of malformed JSON responses that can't be parsed."""
    
    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    # Mock a response with invalid JSON syntax
    mock_response = Mock(spec=httpx.Response)
    mock_response.status_code = 200
    mock_response.headers = httpx.Headers({"Content-Type": "application/json"})
    mock_response.text = '{"incomplete_json": "missing_closing_brace"'
    mock_response.url = "http://test.com/json"
    mock_response.iter_text.return_value = iter(['{"incomplete_json": "missing_closing_brace"'])
    
    with patch.object(s.sdk_configuration.client, 'send', return_value=mock_response):
        # This should trigger a ResponseValidationError due to malformed JSON
        with pytest.raises(errors.ResponseValidationError, match=r"Response validation failed") as exc_info:
            s.response_body_json_get()
        
        assert exc_info.value is not None
        assert exc_info.value.status_code == 200
        assert isinstance(exc_info.value.cause, ValueError)

