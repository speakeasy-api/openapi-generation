package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.InterruptedIOException;
import java.time.Duration;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.CancelledRequestResponse;
import org.openapis.secondary.openapi.models.operations.CancelledRequestWithOperationTimeoutResponse;

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
    }
}
