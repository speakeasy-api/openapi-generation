package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.secondary.openapi.Helpers.API_TEST_SERVICE_URL;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import com.fasterxml.jackson.databind.JsonNode;
import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.utils.JSON;
import org.openapis.secondary.openapi.utils.HTTPClient;
import org.openapis.secondary.openapi.utils.Headers;
import org.openapis.secondary.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.secondary.openapi.utils.transport.HttpBody;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;
import org.openapis.secondary.openapi.utils.transport.HttpResponse;

/**
 * Wire-level checks for the okhttp transport adapter. TransportTypesTest pins
 * immutable value-type behavior; these tests make sure SpeakeasyHTTPClient
 * preserves that behavior when converting to and from okhttp requests.
 */
public class TransportClientAdditionalTest {

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testDuplicateHeadersRoundTripThroughOkHttp() throws Exception {
        SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
        HttpRequest request = HttpRequest.builder()
                .uri(URI.create(API_TEST_SERVICE_URL + "/method/get"))
                .header("X-Inject-Header", "first")
                .header("x-inject-header", "second")
                .build();

        HttpResponse<InputStream> response = client.send(request);

        try (InputStream ignored = response.body()) {
            assertEquals(200, response.statusCode());
            assertEquals(List.of("first", "second"),
                    response.headers().get("X-INJECT-HEADER"));
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testBodylessPostGetsSyntheticEmptyBody() throws Exception {
        SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
        HttpRequest request = HttpRequest.builder()
                .uri(URI.create(HTTPBIN_URL + "/anything/transport/bodyless-post"))
                .method("POST")
                .build();

        HttpResponse<InputStream> response = client.send(request);
        JsonNode echoed = readJson(response);

        assertEquals(200, response.statusCode());
        assertEquals("POST", echoed.get("method").asText());
        assertEquals("", echoed.get("data").asText());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testDebugLoggingBuffersResponseBody() throws Exception {
        SpeakeasyHTTPClient.setDebugLogging(true);
        try {
            SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
            HttpRequest request = HttpRequest.builder()
                    .uri(URI.create(HTTPBIN_URL + "/anything/transport/debug-repeatable"))
                    .method("POST")
                    .header("Content-Type", "application/json")
                    .header("Authorization", "secret-token")
                    .body(HttpBody.of("{\"ok\":true}"))
                    .build();

            HttpResponse<InputStream> response = client.send(request);
            JsonNode echoed = readJson(response);

            assertEquals(200, response.statusCode());
            assertEquals("{\"ok\":true}", echoed.get("data").asText());
        } finally {
            SpeakeasyHTTPClient.setDebugLogging(false);
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testDebugLoggingDoesNotConsumeOneShotRequestBody() throws Exception {
        SpeakeasyHTTPClient.setDebugLogging(true);
        try {
            byte[] bytes = "{\"oneshot\":true}".getBytes(StandardCharsets.UTF_8);
            SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
            HttpRequest request = HttpRequest.builder()
                    .uri(URI.create(HTTPBIN_URL + "/anything/transport/debug-oneshot"))
                    .method("POST")
                    .header("Content-Type", "application/json")
                    .body(HttpBody.ofInputStream(new ByteArrayInputStream(bytes), bytes.length))
                    .build();

            HttpResponse<InputStream> response = client.send(request);
            JsonNode echoed = readJson(response);

            assertEquals(200, response.statusCode());
            assertEquals("{\"oneshot\":true}", echoed.get("data").asText());
        } finally {
            SpeakeasyHTTPClient.setDebugLogging(false);
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testNoContentSdkResponseReleasesConnection() throws Exception {
        AtomicBoolean closed = new AtomicBoolean(false);
        InputStream body = new ByteArrayInputStream(new byte[0]) {
            @Override
            public void close() throws java.io.IOException {
                closed.set(true);
                super.close();
            }
        };
        HTTPClient httpClient = new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                return new HttpResponse<>(request, 204, Headers.EMPTY, body);
            }

            @Override
            public CompletableFuture<HttpResponse<InputStream>> sendAsync(HttpRequest request) {
                return CompletableFuture.completedFuture(
                        new HttpResponse<>(request, 204, Headers.EMPTY, new ByteArrayInputStream(new byte[0])));
            }
        };

        SDK sdk = SDK.builder().client(httpClient).build();

        assertEquals(204, sdk.generation().dateParamWithDefault().call().statusCode());
        assertTrue(closed.get(), "no-content response body must be closed");
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testSdkCloseReleasesOwnedOkHttpResources() throws Exception {
        SpeakeasyHTTPClient httpClient = new SpeakeasyHTTPClient();
        java.lang.reflect.Field clientField = SpeakeasyHTTPClient.class.getDeclaredField("client");
        clientField.setAccessible(true);
        okhttp3.OkHttpClient okHttpClient = (okhttp3.OkHttpClient) clientField.get(httpClient);
        java.lang.reflect.Field streamingField = SpeakeasyHTTPClient.class.getDeclaredField("streamingClient");
        streamingField.setAccessible(true);
        okhttp3.OkHttpClient streamingClient = (okhttp3.OkHttpClient) streamingField.get(httpClient);

        // close() only tears down `client`; that is sufficient only while the
        // streaming client shares its dispatcher, connection pool, and cache.
        assertSame(okHttpClient.dispatcher(), streamingClient.dispatcher(),
                "streaming client must share the dispatcher, or close() leaks its executor");
        assertSame(okHttpClient.connectionPool(), streamingClient.connectionPool(),
                "streaming client must share the connection pool, or close() leaks its connections");
        assertSame(okHttpClient.cache(), streamingClient.cache(),
                "streaming client must share the cache, or close() leaks it");

        SDK sdk = SDK.builder().client(httpClient).build();
        HttpResponse<InputStream> response = httpClient.send(HttpRequest.builder()
                .uri(URI.create(API_TEST_SERVICE_URL + "/method/get"))
                .build());
        try (InputStream ignored = response.body()) {
            assertEquals(200, response.statusCode());
        }
        assertTrue(okHttpClient.connectionPool().idleConnectionCount() > 0,
                "request must leave an idle keep-alive connection in the pool");

        sdk.close();

        assertTrue(okHttpClient.dispatcher().executorService().isShutdown());
        assertEquals(0, okHttpClient.connectionPool().idleConnectionCount());
    }

    @Test
    public void testBorrowedClientInheritsNoOpClose() throws Exception {
        AtomicInteger sends = new AtomicInteger();
        HTTPClient borrowingClient = new StubHTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                sends.incrementAndGet();
                return super.send(request);
            }
        };

        SDK sdk = SDK.builder().client(borrowingClient).build();
        sdk.close();

        // The client does not override close(), so it inherits the interface's
        // no-op default and must remain fully usable after the SDK is closed.
        assertEquals(200, borrowingClient.send(HttpRequest.builder()
                .uri(URI.create("https://example.com"))
                .build()).statusCode());
        assertEquals(1, sends.get());
    }

    @Test
    public void testBuilderReuseClosesReplacementClient() throws Exception {
        AtomicInteger firstCloses = new AtomicInteger();
        AtomicInteger secondCloses = new AtomicInteger();
        HTTPClient first = new StubHTTPClient() {
            @Override
            public void close() {
                firstCloses.incrementAndGet();
            }
        };
        HTTPClient second = new StubHTTPClient() {
            @Override
            public void close() {
                secondCloses.incrementAndGet();
            }
        };

        SDK.Builder builder = SDK.builder();
        SDK sdk1 = builder.client(first).build();
        sdk1.close();
        assertEquals(1, firstCloses.get());

        SDK sdk2 = builder.client(second).build();
        sdk2.close();

        assertEquals(1, firstCloses.get(), "already-closed client must not be closed again");
        assertEquals(1, secondCloses.get(),
                "client installed after a close must still be closed");
    }

    @Test
    public void testSyncAndAsyncSdkCloseClientExactlyOnce() throws Exception {
        AtomicInteger closeCount = new AtomicInteger();
        HTTPClient countingClient = new StubHTTPClient() {
            @Override
            public void close() {
                closeCount.incrementAndGet();
            }
        };

        try (SDK sdk = SDK.builder().client(countingClient).build();
                AsyncSDK asyncSdk = sdk.toAsync()) {
            sdk.close();
            asyncSdk.close();
        }

        assertEquals(1, closeCount.get());
    }

    private abstract static class StubHTTPClient implements HTTPClient {
        @Override
        public HttpResponse<InputStream> send(HttpRequest request) {
            return new HttpResponse<>(request, 200, Headers.EMPTY, new ByteArrayInputStream(new byte[0]));
        }

        @Override
        public CompletableFuture<HttpResponse<InputStream>> sendAsync(HttpRequest request) {
            return CompletableFuture.completedFuture(send(request));
        }
    }

    private static JsonNode readJson(HttpResponse<InputStream> response) throws Exception {
        try (InputStream in = response.body()) {
            return JSON.getMapper().readTree(in);
        }
    }

}
