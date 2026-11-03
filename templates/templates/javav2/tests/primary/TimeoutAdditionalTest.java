package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.io.InputStream;
import java.io.UncheckedIOException;
import java.net.URISyntaxException;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.http.HttpTimeoutException;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.CancelledRequestResponse;
import org.openapis.openapi.models.operations.CancelledRequestWithOperationTimeoutResponse;
import org.openapis.openapi.models.operations.JsonResponse;
import org.openapis.openapi.models.operations.RetriesGetTimeoutResponse;
import org.openapis.openapi.models.shared.JsonEvent;
import org.openapis.openapi.utils.BackoffStrategy;
import org.openapis.openapi.utils.Blob;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.EventStream;
import org.openapis.openapi.utils.RetryConfig;

public class TimeoutAdditionalTest {

    private static void assertTimedOut(Throwable thrown) {
        Throwable cause = thrown;
        while (cause != null && !(cause instanceof HttpTimeoutException)) {
            cause = cause.getCause();
        }
        assertNotNull(cause, "expected an HttpTimeoutException in the cause chain: " + thrown);
    }

    @Test
    public void testTimeoutOverrideIsRespected() {
        CommonHelpers.recordTest("timeout-ms-override-is-respected");

        SDK s = SDK.builder().build();

        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            s.cancellation().cancelledRequest()
                .seconds(3L)
                .timeout(Duration.ofMillis(1000))
                .call();
        });
        assertTrue(thrown.getCause() instanceof HttpTimeoutException, String.valueOf(thrown.getCause()));

        ExecutionException async = assertThrows(ExecutionException.class, () -> {
            s.async().cancellation().cancelledRequest()
                .seconds(3L)
                .timeout(Duration.ofMillis(1000))
                .call()
                .get();
        });
        assertTimedOut(async);
    }

    @Test
    public void testTimeoutOverrideAllowsCompletion() throws Exception {
        CommonHelpers.recordTest("timeout-ms-override-allows-completion");

        SDK s = SDK.builder().build();

        CancelledRequestResponse res = s.cancellation().cancelledRequest()
            .seconds(1L)
            .timeout(Duration.ofSeconds(5))
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(1L, res.object().get().delaySeconds());
    }

    @Test
    public void testSdkDefaultTimeoutIsRespected() throws Exception {
        SDK s = SDK.builder().timeout(Duration.ofMillis(1000)).build();

        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            s.cancellation().cancelledRequest().seconds(3L).call();
        });
        assertTrue(thrown.getCause() instanceof HttpTimeoutException, String.valueOf(thrown.getCause()));

        CancelledRequestResponse overridden = s.cancellation().cancelledRequest()
            .seconds(2L)
            .timeout(Duration.ofSeconds(5))
            .call();
        assertEquals(2L, overridden.object().get().delaySeconds());
    }

    @Test
    public void testOperationTimeoutExtensionIsRespected() throws Exception {
        SDK s = SDK.builder().build();

        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            s.cancellation().cancelledRequestWithOperationTimeout().seconds(3L).call();
        });
        assertTrue(thrown.getCause() instanceof HttpTimeoutException, String.valueOf(thrown.getCause()));

        CancelledRequestWithOperationTimeoutResponse perCall = s.cancellation().cancelledRequestWithOperationTimeout()
            .seconds(2L)
            .timeout(Duration.ofSeconds(5))
            .call();
        assertEquals(2L, perCall.object().get().delaySeconds());

        SDK withDefault = SDK.builder().timeout(Duration.ofSeconds(5)).build();
        CancelledRequestWithOperationTimeoutResponse sdkDefault = withDefault.cancellation()
            .cancelledRequestWithOperationTimeout()
            .seconds(2L)
            .call();
        assertEquals(2L, sdkDefault.object().get().delaySeconds());
    }

    @Test
    public void testTimeoutBoundsTheResponseBody() throws Exception {
        SDK s = SDK.builder().timeout(Duration.ofMillis(1500)).build();

        JsonResponse res = s.eventstreams().json()
                .header("event-interval-ms", "1000")
                .call();
        long started = System.nanoTime();
        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            try (EventStream<JsonEvent> events = res.events()) {
                events.stream().forEach(x -> { });
            }
        });
        assertTimedOut(thrown);
        assertTrue(System.nanoTime() - started < TimeUnit.SECONDS.toNanos(3), "body read was not cut at the timeout");

        HttpResponse<Blob> asyncRaw = s.async().eventstreams().json()
                .header("event-interval-ms", "1000")
                .call()
                .rawResponse()
                .get();
        ExecutionException asyncThrown = assertThrows(ExecutionException.class,
                () -> asyncRaw.body().toByteArray().get());
        assertTimedOut(asyncThrown);
    }

    @Test
    public void testTimeoutIsRetriedPerAttempt() throws Exception {
        String faultSettings = "{\"delay_count\": 1, \"delay_ms\": 1500}";
        RetryConfig retryConfig = RetryConfig.builder().backoff(BackoffStrategy.builder()
                .initialInterval(10, TimeUnit.MILLISECONDS)
                .maxInterval(100, TimeUnit.MILLISECONDS)
                .baseFactor(1.5)
                .maxElapsedTime(10000, TimeUnit.MILLISECONDS)
                .retryReadTimeoutError(true)
                .build()).build();
        AttemptCountingClient client = new AttemptCountingClient();
        SDK s = SDK.builder().client(client).timeout(Duration.ofMillis(500)).build();

        String id = "timeout-retry-" + UUID.randomUUID();
        RetriesGetTimeoutResponse res = s.retries().retriesGetTimeout()
            .requestIdQueryParameter(id)
            .requestId(id)
            .faultSettings(faultSettings)
            .numRetries(1L)
            .retryConfig(retryConfig)
            .call();
        assertEquals(200, res.statusCode());
        assertTrue(client.attempts.get() >= 2, "timed-out attempt was not retried, attempts=" + client.attempts.get());

        client.attempts.set(0);
        String asyncId = "timeout-retry-async-" + UUID.randomUUID();
        org.openapis.openapi.models.operations.async.RetriesGetTimeoutResponse asyncRes = s.async().retries().retriesGetTimeout()
            .requestIdQueryParameter(asyncId)
            .requestId(asyncId)
            .faultSettings(faultSettings)
            .numRetries(1L)
            .retryConfig(retryConfig)
            .call()
            .get();
        assertEquals(200, asyncRes.statusCode());
        assertTrue(client.attempts.get() >= 2, "timed-out async attempt was not retried, attempts=" + client.attempts.get());
    }

    private static final class AttemptCountingClient implements HTTPClient {
        final AtomicInteger attempts = new AtomicInteger();

        @Override
        public HttpResponse<InputStream> send(HttpRequest request)
                throws IOException, InterruptedException, URISyntaxException {
            attempts.incrementAndGet();
            return HTTPClient.super.send(request);
        }

        @Override
        public CompletableFuture<HttpResponse<Blob>> sendAsync(HttpRequest request) {
            attempts.incrementAndGet();
            return HTTPClient.super.sendAsync(request);
        }
    }
}
