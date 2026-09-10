package org.openapis.openapi;

import java.io.UncheckedIOException;
import java.net.ConnectException;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import org.junit.jupiter.api.Test;

import org.openapis.openapi.models.errors.APIException;
import org.openapis.openapi.models.operations.RetriesAfterResponse;
import org.openapis.openapi.models.operations.RetriesAttemptCountResponse;
import org.openapis.openapi.models.operations.RetriesConnectErrorGetResponse;
import org.openapis.openapi.models.operations.RetriesGetResponse;
import org.openapis.openapi.models.operations.RetriesPostRequestBody;
import org.openapis.openapi.models.operations.RetriesPostResponse;
import org.openapis.openapi.utils.BackoffStrategy;
import org.openapis.openapi.utils.RetryConfig;

public class RetriesAdditionalTest {
    
    @Test
    public void testRetriesSucceeds() throws Exception {
        CommonHelpers.recordTest("retries-succeeds");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        RetriesGetResponse res = s.retries().retriesGet()
            .requestId(UUID.randomUUID().toString())
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(3, res.retries().get().retries());
    }

    @Test
    public void testRetriesSucceedsWithBody() throws Exception {
        CommonHelpers.recordTest("retries-succeeds-with-body");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        RetriesPostRequestBody req = RetriesPostRequestBody.builder()
            .fieldOne("one")
            .build();

        RetriesPostResponse res = s.retries().retriesPost()
            .requestBody(req)
            .requestId(UUID.randomUUID().toString())
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(3, res.retries().get().retries());
    }

    @Test
    public void testRetriesTimeout() throws Exception {
        CommonHelpers.recordTest("retries-timeout");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        APIException error = assertThrows(APIException.class, () -> {

            RetriesGetResponse res = s.retries().retriesGet()
                .requestId(UUID.randomUUID().toString())
                .numRetries(1000000000L)
                .retryConfig(RetryConfig.builder()
                                .backoff(BackoffStrategy.builder()
                                            .initialInterval(1L, TimeUnit.MILLISECONDS)
                                            .maxInterval(50L, TimeUnit.MILLISECONDS)
                                            .maxElapsedTime(100L, TimeUnit.MILLISECONDS)
                                            .exponent(1.1f)
                                    .build())
                            .build())
                    .call();

        });

        assertEquals("API error occurred", error.message());
        assertEquals(503, error.code());
    }

    @Test
    public void testRetriesConnectError() throws Exception {
        CommonHelpers.recordTest("retries-connect-error");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
                .retryConfig(RetryConfig.builder()
                        .backoff(BackoffStrategy.builder()
                                .retryConnectError(false)
                                .build())
                        .build())
                .build();
        assertNotNull(s);

        UncheckedIOException e = assertThrows(UncheckedIOException.class, () -> {
            RetriesConnectErrorGetResponse res = s.retries().retriesConnectErrorGet().call();
        });
        assertTrue(e.getCause() instanceof ConnectException);
    }

    @Test
    public void testGlobalRetryConfigDisable() throws Exception {
        CommonHelpers.recordTest("retries-global-config-disable");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
            .retryConfig(RetryConfig.noRetries())
            .build();
        assertNotNull(s);

        APIException error = assertThrows(APIException.class, () -> {

            RetriesGetResponse res = s.retries().retriesGet()
                .requestId(UUID.randomUUID().toString())
                .numRetries(2L)
                .call();

        });

        assertEquals("API error occurred", error.message());
        assertEquals(503, error.code());
    }

    @Test
    public void testGlobalRetryConfigSuccess() throws Exception {
        CommonHelpers.recordTest("retries-global-config-success");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
            .retryConfig(RetryConfig.builder()
                            .backoff(BackoffStrategy.builder()
                                        .initialInterval(1L, TimeUnit.MILLISECONDS)
                                        .maxInterval(50L, TimeUnit.MILLISECONDS)
                                        .maxElapsedTime(1000L, TimeUnit.MILLISECONDS)
                                        .exponent(1.1f)
                                        .build())
                            .build())
            .build();
        assertNotNull(s);

        RetriesGetResponse res = s.retries().retriesGet()
            .requestId(UUID.randomUUID().toString())
            .numRetries(20L)
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(20, res.retries().get().retries());
    }

    @Test
    public void testRespectRetryAfter() throws Exception {
        CommonHelpers.recordTest("retries-header");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
            .retryConfig(RetryConfig.builder()
                            .backoff(BackoffStrategy.builder()
                                        .initialInterval(5000L, TimeUnit.MILLISECONDS)
                                        .maxInterval(10000L, TimeUnit.MILLISECONDS)
                                        .maxElapsedTime(10000L, TimeUnit.MILLISECONDS)
                                        .exponent(1.1f)
                                        .build())
                            .build())
            .build();
        assertNotNull(s);

        RetriesAfterResponse res = s.retries().retriesAfter()
            .requestId(UUID.randomUUID().toString())
            .numRetries(3L)
            .retryAfterVal(1L)
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(3, res.retries().get().retries());
    }

    @Test
    public void testAttemptCountBackoff() throws Exception {
        CommonHelpers.recordTest("retries-attempt-count-backoff");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        RetriesAttemptCountResponse res = s.retries().retriesAttemptCount()
            .requestId(UUID.randomUUID().toString())
            .numRetries(3L)
            .retryAfterMsVal(1L)
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(3, res.retries().get().retries());
    }

    @Test
    public void testAttemptCountBackoffZero() throws Exception {
        CommonHelpers.recordTest("retries-attempt-count-zero");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        APIException error = assertThrows(APIException.class, () -> {
            s.retries().retriesAttemptCountZero()
                .requestId(UUID.randomUUID().toString())
                .numRetries(2L)
                .retryAfterMsVal(1L)
                .call();
        });

        assertEquals("API error occurred", error.message());
        assertEquals(503, error.code());
    }

    @Test
    public void testGlobalRetryConfigTimeout() throws Exception {
        CommonHelpers.recordTest("retries-global-config-timeout");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
            .retryConfig(RetryConfig.builder()
                            .backoff(BackoffStrategy.builder()
                                        .initialInterval(1L, TimeUnit.MILLISECONDS)
                                        .maxInterval(50L, TimeUnit.MILLISECONDS)
                                        .maxElapsedTime(100L, TimeUnit.MILLISECONDS)
                                        .exponent(1.1f)
                                        .build())
                            .build())
            .build();
        assertNotNull(s);

        APIException error = assertThrows(APIException.class, () -> {
                RetriesGetResponse res = s.retries().retriesGet()
                .requestId(UUID.randomUUID().toString())
                .numRetries(30L)
                .call();
        });

        assertEquals("API error occurred", error.message());
        assertEquals(503, error.code());
    }
}
