package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.UUID;
import java.util.concurrent.ExecutionException;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.RetriesGetResponse;
import org.openapis.openapi.utils.ResponseTooLargeException;

import com.sun.net.httpserver.HttpServer;

public class ResponseBodyLimitAdditionalTest {

    private static final int BODY_BYTES = 64 * 1024;
    private static HttpServer server;
    private static String baseUrl;

    @BeforeAll
    public static void startServer() throws IOException {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/", exchange -> {
            byte[] padding = new byte[BODY_BYTES];
            Arrays.fill(padding, (byte) 'a');
            byte[] body = ("{\"retries\":3,\"padding\":\"" + new String(padding, StandardCharsets.US_ASCII) + "\"}")
                    .getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().add("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, body.length);
            try (OutputStream out = exchange.getResponseBody()) {
                out.write(body);
            }
        });
        server.start();
        baseUrl = "http://127.0.0.1:" + server.getAddress().getPort();
    }

    @AfterAll
    public static void stopServer() {
        server.stop(0);
    }

    private static Throwable findCause(Throwable t, Class<? extends Throwable> type) {
        while (t != null && !type.isInstance(t)) {
            t = t.getCause();
        }
        return t;
    }

    @Test
    public void testOversizedBodyFailsSyncCall() {
        CommonHelpers.recordTest("response-body-limit-sync");

        SDK s = SDK.builder()
            .maxResponseBodyBytes(1024)
            .build();

        RuntimeException thrown = assertThrows(RuntimeException.class, () -> {
            s.retries().retriesGet()
                .requestId(UUID.randomUUID().toString())
                .serverURL(baseUrl)
                .call();
        });
        ResponseTooLargeException tooLarge = (ResponseTooLargeException) findCause(thrown, ResponseTooLargeException.class);
        assertNotNull(tooLarge, "expected ResponseTooLargeException in the cause chain: " + thrown);
        assertEquals(1024, tooLarge.maxBytes());
    }

    @Test
    public void testOversizedBodyFailsAsyncCall() {
        CommonHelpers.recordTest("response-body-limit-async");

        AsyncSDK s = SDK.builder()
            .maxResponseBodyBytes(1024)
            .build()
            .async();

        ExecutionException thrown = assertThrows(ExecutionException.class, () -> {
            s.retries().retriesGet()
                .requestId(UUID.randomUUID().toString())
                .serverURL(baseUrl)
                .call()
                .get();
        });
        assertNotNull(findCause(thrown, ResponseTooLargeException.class),
                "expected ResponseTooLargeException in the cause chain: " + thrown.getCause());
    }

    @Test
    public void testBodyWithinLimitSucceeds() throws Exception {
        CommonHelpers.recordTest("response-body-limit-within-limit");

        SDK s = SDK.builder()
            .maxResponseBodyBytes(BODY_BYTES * 2L)
            .enableHTTPDebugLogging(true)
            .build();

        RetriesGetResponse res = s.retries().retriesGet()
            .requestId(UUID.randomUUID().toString())
            .serverURL(baseUrl)
            .call();
        assertEquals(200, res.statusCode());
        assertEquals(3, res.retries().get().retries());
    }
}
