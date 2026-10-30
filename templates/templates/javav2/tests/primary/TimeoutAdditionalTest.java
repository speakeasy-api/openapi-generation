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
import org.openapis.openapi.models.operations.CancelledRequestWithOperationTimeoutResponse;

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
        CommonHelpers.recordTest("timeout-sdk-default-is-respected");

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
        CommonHelpers.recordTest("timeout-operation-extension-is-respected");

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
}
