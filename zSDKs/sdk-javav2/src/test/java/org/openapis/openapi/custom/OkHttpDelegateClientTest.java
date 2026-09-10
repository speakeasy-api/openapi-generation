package org.openapis.openapi.custom;

import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import okio.Buffer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.utils.Blob;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

class OkHttpDelegateClientTest {

    private MockWebServer server;

    @BeforeEach
    void setup() throws IOException {
        server = new MockWebServer();
        server.start();
    }

    @AfterEach
    void teardown() throws IOException {
        server.shutdown();
    }

    @Test
    void sendReturnsBodyAndStatus() throws Exception {
        server.enqueue(new MockResponse()
                .setResponseCode(201)
                .setHeader("Content-Type", "application/json")
                .setBody("{\"ok\":true}"));

        OkHttpDelegateClient client = new OkHttpDelegateClient(
                Duration.ofSeconds(2), Duration.ofSeconds(2),
                Duration.ofSeconds(2), Duration.ofSeconds(5));

        HttpRequest req = HttpRequest.newBuilder(URI.create(server.url("/foo").toString()))
                .header("X-Custom", "abc")
                .POST(HttpRequest.BodyPublishers.ofString("{\"hi\":1}"))
                .header("Content-Type", "application/json")
                .build();

        HttpResponse<InputStream> resp = client.send(req);

        assertEquals(201, resp.statusCode());
        assertEquals("application/json", resp.headers().firstValue("Content-Type").orElse(""));
        String body = new String(resp.body().readAllBytes(), StandardCharsets.UTF_8);
        assertEquals("{\"ok\":true}", body);

        var recorded = server.takeRequest();
        assertEquals("POST", recorded.getMethod());
        assertEquals("abc", recorded.getHeader("X-Custom"));
        assertEquals("{\"hi\":1}", recorded.getBody().readUtf8());
    }

    @Test
    void readTimeoutFiresWhenServerStallsBetweenBytes() throws Exception {
        // Server returns headers immediately but stalls the body for 2s
        Buffer body = new Buffer().writeUtf8("delayed-payload");
        server.enqueue(new MockResponse()
                .setResponseCode(200)
                .setBody(body)
                .throttleBody(1, 2, TimeUnit.SECONDS));

        OkHttpDelegateClient client = new OkHttpDelegateClient(
                Duration.ofMillis(500), Duration.ofMillis(300),
                Duration.ofSeconds(5), Duration.ofSeconds(10));

        HttpRequest req = HttpRequest.newBuilder(URI.create(server.url("/slow").toString())).GET().build();

        long start = System.nanoTime();
        IOException ex = assertThrows(IOException.class, () -> {
            HttpResponse<InputStream> r = client.send(req);
            r.body().readAllBytes();
        });
        long elapsedMs = (System.nanoTime() - start) / 1_000_000;
        assertTrue(elapsedMs < 2000, "expected read timeout to fire well before 2s, took " + elapsedMs + "ms");
        assertTrue(ex.getMessage() != null, "expected an IOException with a message, got " + ex);
    }

    @Test
    void callTimeoutFiresOnLongResponse() throws Exception {
        server.enqueue(new MockResponse()
                .setResponseCode(200)
                .setHeadersDelay(2, TimeUnit.SECONDS)
                .setBody("eventual"));

        OkHttpDelegateClient client = new OkHttpDelegateClient(
                Duration.ofSeconds(2), Duration.ofSeconds(5),
                Duration.ofSeconds(5), Duration.ofMillis(400));

        HttpRequest req = HttpRequest.newBuilder(URI.create(server.url("/calltimeout").toString())).GET().build();

        long start = System.nanoTime();
        assertThrows(IOException.class, () -> client.send(req));
        long elapsedMs = (System.nanoTime() - start) / 1_000_000;
        assertTrue(elapsedMs < 1500, "expected call timeout to fire well before 2s, took " + elapsedMs + "ms");
    }

    @Test
    void sendAsyncReturnsBlob() throws Exception {
        server.enqueue(new MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "text/plain")
                .setBody("async-hello"));

        OkHttpDelegateClient client = new OkHttpDelegateClient(
                Duration.ofSeconds(2), Duration.ofSeconds(2),
                Duration.ofSeconds(2), Duration.ofSeconds(5));

        HttpRequest req = HttpRequest.newBuilder(URI.create(server.url("/async").toString())).GET().build();

        CompletableFuture<HttpResponse<Blob>> future = client.sendAsync(req);
        HttpResponse<Blob> resp = future.get(5, TimeUnit.SECONDS);

        assertEquals(200, resp.statusCode());
        byte[] bytes = resp.body().toByteArray().get(5, TimeUnit.SECONDS);
        assertEquals("async-hello", new String(bytes, StandardCharsets.UTF_8));
    }

    @Test
    void debugFlagRoundTrips() {
        OkHttpDelegateClient client = new OkHttpDelegateClient(
                Duration.ofSeconds(1), Duration.ofSeconds(1),
                Duration.ofSeconds(1), Duration.ofSeconds(1));
        assertEquals(false, client.isDebugLoggingEnabled());
        client.enableDebugLogging(true);
        assertEquals(true, client.isDebugLoggingEnabled());
    }
}
