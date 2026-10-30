package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.UncheckedIOException;
import java.net.http.HttpTimeoutException;
import java.time.Duration;
import java.util.concurrent.ExecutionException;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.CancelledRequestResponse;

public class TimeoutAdditionalTest {

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
    public void testSdkLevelTimeoutAppliesToEveryCall() {
        CommonHelpers.recordTest("timeout-sdk-default-is-respected");

        SDK s = SDK.builder().timeout(Duration.ofMillis(1000)).build();

        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            s.cancellation().cancelledRequest().seconds(3L).call();
        });
        assertTrue(thrown.getCause() instanceof HttpTimeoutException, String.valueOf(thrown.getCause()));
    }

    @Test
    public void testTimeoutOverrideIsRespectedAsync() {
        CommonHelpers.recordTest("timeout-ms-override-is-respected-async");

        AsyncSDK s = SDK.builder().build().async();

        ExecutionException thrown = assertThrows(ExecutionException.class, () -> {
            s.cancellation().cancelledRequest()
                .seconds(3L)
                .timeout(Duration.ofMillis(1000))
                .call()
                .get();
        });
        Throwable cause = thrown.getCause();
        while (cause != null && !(cause instanceof HttpTimeoutException)) {
            cause = cause.getCause();
        }
        assertNotNull(cause, "expected an HttpTimeoutException in the cause chain: " + thrown.getCause());
    }
}
