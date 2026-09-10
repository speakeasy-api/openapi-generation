package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InterruptedIOException;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.SocketTimeoutException;
import java.net.URI;
import java.util.Collections;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.utils.AsyncRetries;
import org.openapis.secondary.openapi.utils.BackoffStrategy;
import org.openapis.secondary.openapi.utils.Headers;
import org.openapis.secondary.openapi.utils.Retries;
import org.openapis.secondary.openapi.utils.RetryConfig;
import org.openapis.secondary.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;
import org.openapis.secondary.openapi.utils.transport.HttpResponse;

/**
 * IO-timeout retry classification on the okhttp arm. OkHttp raises timeouts
 * as SocketTimeoutException (socket reads) or InterruptedIOException with
 * message "timeout" (okio AsyncTimeout: call timeout, HTTP/2 stream
 * timeouts). Retryability must key on the exception type, not on JDK socket
 * message strings like "Read timed out" which okhttp never emits on those
 * paths.
 */
public class IoTimeoutClassificationAdditionalTest {

    private ScheduledExecutorService scheduler;

    @AfterEach
    public void tearDown() {
        if (scheduler != null) {
            scheduler.shutdownNow();
        }
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncSocketTimeoutIsRetried() throws Exception {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(attempts, new SocketTimeoutException("timeout"), true);

        assertEquals(200, retries.run().statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncOkioAsyncTimeoutIsRetried() throws Exception {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(attempts, new InterruptedIOException("timeout"), true);

        assertEquals(200, retries.run().statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncConnectTimeoutIsRetriedWithConnectOnly() throws Exception {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(
                attempts, new SocketTimeoutException("Connect timed out"), true, false);

        assertEquals(200, retries.run().statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncTimeoutNotRetriedWhenFlagsDisabled() {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(attempts, new SocketTimeoutException("timeout"), false);

        assertThrows(Exception.class, retries::run);
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncReadTimeoutNotRetriedWithConnectOnly() {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(
                attempts, new SocketTimeoutException("timeout"), true, false);

        assertThrows(Exception.class, retries::run);
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncConnectTimeoutNotRetriedWithReadOnly() {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(
                attempts, new SocketTimeoutException("Connect timed out"), false, true);

        assertThrows(Exception.class, retries::run);
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testSyncGenericIOExceptionNotRetried() {
        AtomicInteger attempts = new AtomicInteger();
        Retries retries = retriesFailingOnceWith(attempts, new IOException("stream reset"), true);

        assertThrows(Exception.class, retries::run);
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncSocketTimeoutIsRetried() throws Exception {
        scheduler = Executors.newSingleThreadScheduledExecutor();
        AsyncRetries retries = AsyncRetries.builder()
                .retryConfig(timeoutRetryConfig(true))
                .statusCodes(Collections.singletonList("503"))
                .scheduler(scheduler)
                .build();

        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = retries.retry(attempt -> {
            if (attempts.incrementAndGet() == 1) {
                CompletableFuture<HttpResponse<InputStream>> failed = new CompletableFuture<>();
                failed.completeExceptionally(new SocketTimeoutException("timeout"));
                return failed;
            }
            return CompletableFuture.completedFuture(okResponse());
        });

        assertEquals(200, result.get(5, TimeUnit.SECONDS).statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncOkioAsyncTimeoutIsRetried() throws Exception {
        scheduler = Executors.newSingleThreadScheduledExecutor();
        AsyncRetries retries = AsyncRetries.builder()
                .retryConfig(timeoutRetryConfig(true))
                .statusCodes(Collections.singletonList("503"))
                .scheduler(scheduler)
                .build();

        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = retries.retry(attempt -> {
            if (attempts.incrementAndGet() == 1) {
                CompletableFuture<HttpResponse<InputStream>> failed = new CompletableFuture<>();
                failed.completeExceptionally(new InterruptedIOException("timeout"));
                return failed;
            }
            return CompletableFuture.completedFuture(okResponse());
        });

        assertEquals(200, result.get(5, TimeUnit.SECONDS).statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncConnectTimeoutIsRetriedWithConnectOnly() throws Exception {
        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = asyncRetriesFailingOnceWith(
                attempts, new SocketTimeoutException("Connect timed out"), true, false);

        assertEquals(200, result.get(5, TimeUnit.SECONDS).statusCode());
        assertEquals(2, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncReadTimeoutNotRetriedWithConnectOnly() {
        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = asyncRetriesFailingOnceWith(
                attempts, new SocketTimeoutException("timeout"), true, false);

        assertThrows(Exception.class, () -> result.get(5, TimeUnit.SECONDS));
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncConnectTimeoutNotRetriedWithReadOnly() {
        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = asyncRetriesFailingOnceWith(
                attempts, new SocketTimeoutException("Connect timed out"), false, true);

        assertThrows(Exception.class, () -> result.get(5, TimeUnit.SECONDS));
        assertEquals(1, attempts.get());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncConnectFailureFromOkHttpIsRetried() throws Exception {
        URI closedPort = closedLoopbackPort();
        HttpRequest request = HttpRequest.builder()
                .method("GET")
                .uri(closedPort)
                .build();
        SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();

        scheduler = Executors.newSingleThreadScheduledExecutor();
        AsyncRetries retries = AsyncRetries.builder()
                .retryConfig(connectRetryConfig())
                .statusCodes(Collections.singletonList("503"))
                .scheduler(scheduler)
                .build();

        AtomicInteger attempts = new AtomicInteger();
        CompletableFuture<HttpResponse<InputStream>> result = retries.retry(attempt -> {
            attempts.incrementAndGet();
            return client.sendAsync(request);
        });

        assertThrows(Exception.class, () -> result.get(5, TimeUnit.SECONDS));
        assertTrue(attempts.get() > 1,
                "AsyncRetries should retry OkHttp connection failures, attempts=" + attempts.get());
    }

    private Retries retriesFailingOnceWith(AtomicInteger attempts, IOException failure, boolean retryTimeouts) {
        return retriesFailingOnceWith(attempts, failure, false, retryTimeouts);
    }

    private Retries retriesFailingOnceWith(
            AtomicInteger attempts,
            IOException failure,
            boolean retryConnectErrors,
            boolean retryReadTimeouts) {
        return Retries.builder()
                .action(attempt -> {
                    if (attempts.incrementAndGet() == 1) {
                        throw failure;
                    }
                    return okResponse();
                })
                .retryConfig(timeoutRetryConfig(retryConnectErrors, retryReadTimeouts))
                .statusCodes(Collections.singletonList("503"))
                .build();
    }

    private static RetryConfig timeoutRetryConfig(boolean retryTimeouts) {
        return timeoutRetryConfig(false, retryTimeouts);
    }

    private static RetryConfig timeoutRetryConfig(
            boolean retryConnectErrors, boolean retryReadTimeouts) {
        return RetryConfig.builder().backoff(BackoffStrategy.builder()
                .initialInterval(1, TimeUnit.MILLISECONDS)
                .maxInterval(10, TimeUnit.MILLISECONDS)
                .baseFactor(1.5)
                .maxElapsedTime(5000, TimeUnit.MILLISECONDS)
                .retryConnectError(retryConnectErrors)
                .retryReadTimeoutError(retryReadTimeouts)
                .build()).build();
    }

    private CompletableFuture<HttpResponse<InputStream>> asyncRetriesFailingOnceWith(
            AtomicInteger attempts,
            IOException failure,
            boolean retryConnectErrors,
            boolean retryReadTimeouts) {
        scheduler = Executors.newSingleThreadScheduledExecutor();
        AsyncRetries retries = AsyncRetries.builder()
                .retryConfig(timeoutRetryConfig(retryConnectErrors, retryReadTimeouts))
                .statusCodes(Collections.singletonList("503"))
                .scheduler(scheduler)
                .build();
        return retries.retry(attempt -> {
            if (attempts.incrementAndGet() == 1) {
                CompletableFuture<HttpResponse<InputStream>> failed = new CompletableFuture<>();
                failed.completeExceptionally(failure);
                return failed;
            }
            return CompletableFuture.completedFuture(okResponse());
        });
    }

    private static RetryConfig connectRetryConfig() {
        return RetryConfig.builder().backoff(BackoffStrategy.builder()
                .initialInterval(1, TimeUnit.MILLISECONDS)
                .maxInterval(10, TimeUnit.MILLISECONDS)
                .baseFactor(1.5)
                .maxElapsedTime(100, TimeUnit.MILLISECONDS)
                .retryConnectError(true)
                .retryReadTimeoutError(false)
                .build()).build();
    }

    private static URI closedLoopbackPort() throws IOException {
        try (ServerSocket socket = new ServerSocket(0, 1, InetAddress.getLoopbackAddress())) {
            return URI.create("http://127.0.0.1:" + socket.getLocalPort() + "/");
        }
    }

    private static HttpResponse<InputStream> okResponse() {
        HttpRequest request = HttpRequest.builder()
                .method("GET")
                .uri(URI.create("http://localhost/test"))
                .build();
        return new HttpResponse<>(request, 200, new Headers(),
                new ByteArrayInputStream(new byte[0]));
    }
}
