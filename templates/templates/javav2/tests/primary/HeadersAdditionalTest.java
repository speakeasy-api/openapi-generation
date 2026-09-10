package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.stream.Collectors;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.errors.APIException;
import org.openapis.openapi.models.operations.HeaderParamsNilResponse;

public class HeadersAdditionalTest {

    @Test
    public void testHeadersOverrideRequestHeaders() throws Exception {
        CommonHelpers.recordTest("headers-override-request-headers");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        var res = sdk.methods().methodGet()
                .header("x-inject-header-1", "foo")
                .header("x-inject-header-2", "bar")
                .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.object());
        assertEquals("OK", res.object().get().status().get());
        assertEquals("foo", res.rawResponse().request().headers().firstValue("x-inject-header-1").orElse(null));
        assertEquals("bar", res.rawResponse().request().headers().firstValue("x-inject-header-2").orElse(null));
    }

    @Test
    public void testHeadersEmptyResponseBodyWithHeaders() throws Exception {
        CommonHelpers.recordTest("headers-empty-response-body-with-headers");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        var res = sdk.responseHeaders().responseBodyEmptyWithHeaders(1.1, "hello");

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.rawResponse().headers());
        assertEquals("hello", res.rawResponse().headers().firstValue("x-string-header").orElse(null));
        assertEquals("1.1", res.rawResponse().headers().firstValue("x-number-header").orElse(null));

        // In Java, response headers are collected in a `Map<String, List<String>>`, regardless of `responseFormat`.
        assertNotNull(res.headers());
        assertEquals(List.of("hello"), res.headers().get("x-string-header"));
        assertEquals(List.of("1.1"), res.headers().get("x-number-header"));
    }

    @Test
    public void testHeadersResponseWithoutHeaders() throws Exception {
        CommonHelpers.recordTest("headers-response-headers-none");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        // The 200 response in `errorResponseHeaders` does not include any headers.
        // but other responses in the same operation do.
        var res200 = sdk.responseHeaders().errorResponseHeaders(true, 200, null);

        assertNotNull(res200);
        assertEquals(200, res200.statusCode());
        assertNotNull(res200.authToken());
        assertEquals("test-token", res200.authToken().get().token());
    }

    @Test
    public void testHeadersErrorResponsesWithoutHeaders() throws Exception {
        CommonHelpers.recordTest("headers-error-response-headers-none");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        // 1. Basic error response (no body nor custom headers) whose
        // operation includes another response that has headers.
        APIException thrown1 = assertThrows(APIException.class, () -> {
            sdk.responseHeaders().responseHeaders(true, 400, null);
        });
        assertEquals(400, thrown1.code());
        assertEquals("", thrown1.bodyAsString().orElse(""));

        // 2. Error response with a body but no custom headers whose
        // operation includes another response that has headers.
        APIException thrown2 = assertThrows(APIException.class, () -> {
            sdk.responseHeaders().responseHeaders(true, 500, null);
        });
        assertEquals(500, thrown2.code());
        assertEquals("\"Internal server error.\"\n", thrown2.bodyAsString().orElse(""));
    }

    @Test
    public void testHeadersResponseHeadersOptional() throws Exception {
        CommonHelpers.recordTest("headers-response-headers-optional");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        // 1. Success response with required header included
        var res1 = sdk.responseHeaders().responseHeaders(true, 200, null);
        assertNotNull(res1);
        assertEquals(200, res1.statusCode());
        assertNotNull(res1.authToken());
        assertEquals("test-token", res1.authToken().get().token());
        assertNotNull(res1.rawResponse().headers());
        assertEquals("required", res1.rawResponse().headers().firstValue("x-required-header").orElse(null));
        assertTrue(res1.rawResponse().headers().firstValue("x-optional-header").isEmpty());
        assertNotNull(res1.headers());
        assertEquals(List.of("required"), res1.headers().get("x-required-header"));
        assertNull(res1.headers().get("x-optional-header"));

        // 2. Success response with required header omitted - SDK should not throw
        var res2 = sdk.responseHeaders().responseHeaders(false, 200, null);
        assertNotNull(res2);
        assertEquals(200, res2.statusCode());
        assertNotNull(res2.authToken());
        assertEquals("test-token", res2.authToken().get().token());
        assertTrue(res2.rawResponse().headers().firstValue("x-required-header").isEmpty());
        assertTrue(res2.rawResponse().headers().firstValue("x-optional-header").isEmpty());
        assertNull(res2.headers().get("x-required-header"));
        assertNull(res2.headers().get("x-optional-header"));
    }

    @Test
    public void testHeadersErrorResponseHeadersOptional() throws Exception {
        CommonHelpers.recordTest("headers-error-response-headers-optional");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(sdk);

        // 1. Error response with Retry-After header included
        APIException thrown1 = assertThrows(APIException.class, () -> {
            sdk.responseHeaders().errorResponseHeaders(true, 429, null);
        });
        assertEquals(429, thrown1.code());
        assertEquals("60", thrown1.rawResponse().headers().firstValue("retry-after").orElse(null));
        assertEquals("\"Too many attempts. Please try again later.\"\n", thrown1.bodyAsString().orElse(""));

        // 2. Error response with Retry-After header omitted - SDK should not throw
        APIException thrown2 = assertThrows(APIException.class, () -> {
            sdk.responseHeaders().errorResponseHeaders(false, 429, null);
        });
        assertEquals(429, thrown2.code());
        assertTrue(thrown2.rawResponse().headers().firstValue("retry-after").isEmpty());
        assertEquals("\"Too many attempts. Please try again later.\"\n", thrown2.bodyAsString().orElse(""));
    }
}
