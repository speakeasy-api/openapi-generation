package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.InterruptedIOException;
import java.time.Duration;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.CancelledRequestResponse;
import org.openapis.secondary.openapi.models.operations.CancelledRequestWithOperationTimeoutResponse;
import org.openapis.secondary.openapi.models.operations.JsonResponse;
import org.openapis.secondary.openapi.models.components.JsonEvent;
import org.openapis.secondary.openapi.utils.EventStream;

/**
 * Timeouts on the okhttp arm. The resolved timeout is applied to the OkHttp
 * call, which bounds the whole exchange and fails with an
 * InterruptedIOException rather than the JDK client's HttpTimeoutException.
 */
public class TimeoutAdditionalTest {

    private static void assertTimedOut(Throwable thrown) {
        Throwable cause = thrown;
        while (cause != null && !(cause instanceof InterruptedIOException)) {
            cause = cause.getCause();
        }
        assertNotNull(cause, "expected an InterruptedIOException in the cause chain: " + thrown);
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testTimeoutOverrideIsRespected() throws Exception {
        SDK s = SDK.builder().build();

        assertTimedOut(assertThrows(RuntimeException.class, () -> {
            s.cancellation().cancelledRequest()
                .seconds(3L)
                .timeout(Duration.ofMillis(1000))
                .call();
        }));

        CancelledRequestResponse res = s.cancellation().cancelledRequest()
            .seconds(1L)
            .timeout(Duration.ofSeconds(5))
            .call();
        assertEquals(200, res.statusCode());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testSdkDefaultAndOperationExtensionTimeouts() throws Exception {
        SDK s = SDK.builder().timeout(Duration.ofMillis(1000)).build();
        assertTimedOut(assertThrows(RuntimeException.class, () -> {
            s.cancellation().cancelledRequest().seconds(3L).call();
        }));

        SDK plain = SDK.builder().build();
        assertTimedOut(assertThrows(RuntimeException.class, () -> {
            plain.cancellation().cancelledRequestWithOperationTimeout().seconds(3L).call();
        }));

        CancelledRequestWithOperationTimeoutResponse res = plain.cancellation().cancelledRequestWithOperationTimeout()
            .seconds(2L)
            .timeout(Duration.ofSeconds(5))
            .call();
        assertEquals(200, res.statusCode());

        SDK withDefault = SDK.builder().timeout(Duration.ofSeconds(5)).build();
        CancelledRequestWithOperationTimeoutResponse sdkDefault = withDefault.cancellation()
            .cancelledRequestWithOperationTimeout()
            .seconds(2L)
            .call();
        assertEquals(200, sdkDefault.statusCode());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testStreamingResponseOutlivesTimeout() throws Exception {
        SDK s = SDK.builder().timeout(Duration.ofMillis(2000)).build();

        JsonResponse res = s.eventstreams().json()
                .header("event-interval-ms", "1000")
                .call();
        try (EventStream<JsonEvent> events = res.events()) {
            assertEquals("Hello world!", events.stream()
                    .map(x -> x.data().content())
                    .collect(Collectors.joining()));
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testStreamingHeaderWaitIsBounded() throws Exception {
        SDK s = SDK.builder().timeout(Duration.ofMillis(500)).build();
        String delayHeaders = "{\"delay_count\": 1, \"delay_ms\": 3000}";

        assertTimedOut(assertThrows(RuntimeException.class, () -> s.eventstreams().json()
                .header("request-id", "stream-header-wait-" + UUID.randomUUID())
                .header("fault-settings", delayHeaders)
                .call()));

        assertTimedOut(assertThrows(ExecutionException.class, () -> s.toAsync().eventstreams().json()
                .header("request-id", "stream-header-wait-async-" + UUID.randomUUID())
                .header("fault-settings", delayHeaders)
                .call()
                .get()));
    }
}
