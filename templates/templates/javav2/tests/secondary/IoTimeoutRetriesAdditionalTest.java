package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import java.io.IOException;
import java.io.InputStream;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.RetriesGetTimeoutResponse;
import org.openapis.secondary.openapi.utils.BackoffStrategy;
import org.openapis.secondary.openapi.utils.HTTPClient;
import org.openapis.secondary.openapi.utils.Headers;
import org.openapis.secondary.openapi.utils.RetryConfig;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;
import org.openapis.secondary.openapi.utils.transport.HttpResponse;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

/**
 * IO-timeout-driven retry against the retriesGetTimeout operation: the fault
 * middleware delays the first request past a short client call timeout, so the
 * first attempt fails with okio's AsyncTimeout exception
 * (InterruptedIOException with message "timeout" — the same shape okhttp
 * produces for HTTP/2 stream timeouts), and the SDK must retry until an
 * undelayed attempt succeeds. Guards against timeout classification relying on
 * JDK socket message strings ("Read timed out"), which okhttp does not emit on
 * these paths.
 */
public class IoTimeoutRetriesAdditionalTest {

    private static final String FAULT_SETTINGS = "{\"delay_count\": 1, \"delay_ms\": 1500}";

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testSyncRetriesAfterIoTimeout() throws Exception {
        CallTimeoutHttpClient client = new CallTimeoutHttpClient();
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(client).build();

        String id = "sync-io-timeout-" + UUID.randomUUID();
        RetriesGetTimeoutResponse res = s.retries().retriesGetTimeout()
                .requestIdQueryParameter(id)
                .requestId(id)
                .faultSettings(FAULT_SETTINGS)
                .numRetries(1L)
                .retryConfig(timeoutRetryConfig())
                .call();

        assertEquals(200, res.statusCode());
        assertTrue(client.attempts.get() >= 2,
                "expected the timed-out first attempt to be retried, attempts=" + client.attempts.get());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncRetriesAfterIoTimeout() throws Exception {
        CallTimeoutHttpClient client = new CallTimeoutHttpClient();
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(client).build();

        String id = "async-io-timeout-" + UUID.randomUUID();
        org.openapis.secondary.openapi.models.operations.async.RetriesGetTimeoutResponse res =
                s.toAsync().retries().retriesGetTimeout()
                        .requestIdQueryParameter(id)
                        .requestId(id)
                        .faultSettings(FAULT_SETTINGS)
                        .numRetries(1L)
                        .retryConfig(timeoutRetryConfig())
                        .call()
                        .get(12, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertTrue(client.attempts.get() >= 2,
                "expected the timed-out first attempt to be retried, attempts=" + client.attempts.get());
    }

    private static RetryConfig timeoutRetryConfig() {
        return RetryConfig.builder().backoff(BackoffStrategy.builder()
                .initialInterval(10, TimeUnit.MILLISECONDS)
                .maxInterval(100, TimeUnit.MILLISECONDS)
                .baseFactor(1.5)
                .maxElapsedTime(10000, TimeUnit.MILLISECONDS)
                .retryReadTimeoutError(true)
                .build()).build();
    }

    /**
     * Attempt-counting client whose call timeout is far shorter than the fault
     * middleware delay. The call timeout is enforced by okio's AsyncTimeout,
     * not the socket read timeout, so the failure carries the "timeout"
     * message rather than the JDK's "Read timed out".
     */
    private static final class CallTimeoutHttpClient implements HTTPClient {

        final AtomicInteger attempts = new AtomicInteger();

        private final OkHttpClient client = new OkHttpClient.Builder()
                .callTimeout(500, TimeUnit.MILLISECONDS)
                .readTimeout(10, TimeUnit.SECONDS)
                .build();

        @Override
        public HttpResponse<InputStream> send(HttpRequest request) throws IOException {
            attempts.incrementAndGet();
            Request.Builder builder = new Request.Builder()
                    .url(request.uri().toString())
                    .method(request.method(), null);
            request.headers().forEach((name, values) ->
                    values.forEach(value -> builder.addHeader(name, value)));
            Response response = client.newCall(builder.build()).execute();
            return new HttpResponse<>(request, response.code(),
                    new Headers(response.headers().toMultimap()),
                    response.body().byteStream());
        }

        @Override
        public CompletableFuture<HttpResponse<InputStream>> sendAsync(HttpRequest request) {
            CompletableFuture<HttpResponse<InputStream>> future = new CompletableFuture<>();
            CompletableFuture.runAsync(() -> {
                try {
                    future.complete(send(request));
                } catch (Exception e) {
                    future.completeExceptionally(e);
                }
            });
            return future;
        }
    }
}
