package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertThrowsExactly;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.io.InputStream;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.http.HttpHeaders;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;

import org.assertj.core.util.Maps;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.openapis.openapi.Helpers.StubHttpResponse;
import org.openapis.openapi.models.errors.APIException;

public class APIExceptionAdditionalTest {

    @Test
    void testAPIExceptionPropagatesMessage() {
        APIException e = new APIException("custom error message", 500, new byte[0], new StubHttpResponse(), null);

        APIException thrown = assertThrowsExactly(
                APIException.class,
                () -> {
                    throw e;
                },
                "should throw APIException"
        );

        assertEquals("custom error message", thrown.getMessage());

        Exception thrownBase = assertThrows(
                Exception.class,
                () -> {
                    throw e;
                },
                "should throw Exception"
        );

        assertEquals("custom error message", thrownBase.getMessage());
    }
    
    @Test
    void testAPIExceptionToString() throws URISyntaxException {
        @SuppressWarnings("unchecked")
        HttpResponse<InputStream> response = mock(HttpResponse.class);
        HttpRequest request = mock(HttpRequest.class);
        when(response.request()).thenReturn(request);
        when(request.uri()).thenReturn(new URI("https://example.com"));
        when(request.method()).thenReturn("GET");
        HttpHeaders headers = mock(HttpHeaders.class);
        Map<String, List<String>> map = Maps.newHashMap("agent", List.of("speakeasy"));
        when(headers.map()).thenReturn(map);
        when(response.headers()).thenReturn(headers);
        APIException e = new APIException("custom error message", 500, 
                 "boo".getBytes(StandardCharsets.UTF_8), response, null);
        String expected = "APIException[requestMethod=GET, requestUri=https://example.com, code=500, responseHeaders={agent=[speakeasy]}, message=custom error message, body=boo]";
        assertEquals(expected, e.toString());
    }
    
    @Test
    void testAPIExceptionToStringWhenBodyIsNotUtf8() throws URISyntaxException {
        @SuppressWarnings("unchecked")
        HttpResponse<InputStream> response = mock(HttpResponse.class);
        HttpRequest request = mock(HttpRequest.class);
        when(response.request()).thenReturn(request);
        when(request.uri()).thenReturn(new URI("https://example.com"));
        when(request.method()).thenReturn("GET");
        HttpHeaders headers = mock(HttpHeaders.class);
        when(headers.map()).thenReturn(java.util.Collections.emptyMap());
        when(response.headers()).thenReturn(headers);
        APIException e = new APIException("custom error message", 500,
                new byte[] {1}, response, null);
        String expected = "APIException[requestMethod=GET, requestUri=https://example.com, code=500, responseHeaders={}, message=custom error message, body=]";
        assertEquals(expected, e.toString());
    }

}

