package org.openapis.openapi;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.params.provider.ValueSource;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;
import org.openapis.openapi.utils.AsyncRetries;
import org.openapis.openapi.utils.BackoffStrategy;
import org.openapis.openapi.utils.RetryConfig;
import org.openapis.openapi.utils.AsyncRetryableException;
import org.openapis.openapi.utils.NonRetryableException;

import java.io.IOException;
import java.io.InputStream;
import java.net.ConnectException;
import java.net.http.HttpHeaders;
import java.net.http.HttpResponse;
import org.openapis.openapi.utils.Blob;
import java.time.Duration;
import java.util.List;
import java.util.Map;
import java.util.concurrent.*;
import java.util.function.Supplier;

import static org.assertj.core.api.Assertions.*;
import static org.awaitility.Awaitility.await;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;
import static org.openapis.openapi.Helpers.*;

@ExtendWith(MockitoExtension.class)
class AsyncRetriesAdditionalTest {

    @Mock
    private RetryConfig retryConfig;

    @Mock
    private BackoffStrategy backoffStrategy;

    @Spy
    private ScheduledExecutorService scheduler = Executors.newSingleThreadScheduledExecutor();

    @Mock
    private HttpResponse<Blob> httpResponse;

    @Mock
    private Supplier<CompletableFuture<HttpResponse<Blob>>> task;

    private AsyncRetries asyncRetries;

    @BeforeEach
    void setUp() {
        lenient().when(retryConfig.strategy()).thenReturn(RetryConfig.Strategy.BACKOFF);
        lenient().when(retryConfig.backoff()).thenReturn(java.util.Optional.of(backoffStrategy));
        lenient().when(backoffStrategy.initialIntervalMs()).thenReturn(100L);
        lenient().when(backoffStrategy.maxIntervalMs()).thenReturn(500L);
        lenient().when(backoffStrategy.baseFactor()).thenReturn(2.0);
        lenient().when(backoffStrategy.jitterFactor()).thenReturn(0.1);
        lenient().when(backoffStrategy.maxElapsedTimeMs()).thenReturn(2746L);
        lenient().when(backoffStrategy.retryConnectError()).thenReturn(true);
        lenient().when(backoffStrategy.retryReadTimeoutError()).thenReturn(true);
        lenient().when(httpResponse.headers()).thenReturn(HttpHeaders.of(Map.of(), (a, b) -> true));
    }

    @Test
    void testBuilderWithValidConfiguration() {
        List<String> statusCodes = List.of("500", "502", "503");

        AsyncRetries result = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(statusCodes)
                .scheduler(scheduler)
                .build();

        assertThat(result).isNotNull();
    }

    @Test
    void testBuilderWithEmptyStatusCodes() {
        assertThatThrownBy(() ->
                AsyncRetries.builder()
                        .retryConfig(retryConfig)
                        .statusCodes(List.of())
                        .scheduler(scheduler)
                        .build()
        ).isInstanceOf(IllegalArgumentException.class)
                .hasMessage("statusCodes list cannot be empty");
    }

    @Test
    void testBuilderWithNullScheduler() {
        List<String> statusCodes = List.of("500");

        assertThatThrownBy(() ->
                AsyncRetries.builder()
                        .retryConfig(retryConfig)
                        .statusCodes(statusCodes)
                        .scheduler(null)
                        .build()
        ).isInstanceOf(IllegalArgumentException.class)
                .hasMessage("scheduler cannot be null");
    }

    @Test
    void testRetryWithNoneStrategy() {
        when(retryConfig.strategy()).thenReturn(RetryConfig.Strategy.NONE);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> expectedResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(expectedResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertThat(result).isSameAs(expectedResult);
        verify(task, times(1)).get();
    }

    @Test
    void testRetryWithBackoffStrategyButNoBackoffDefined() {
        when(retryConfig.strategy()).thenReturn(RetryConfig.Strategy.BACKOFF);
        when(retryConfig.backoff()).thenReturn(java.util.Optional.empty());

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        assertThatThrownBy(() -> asyncRetries.retry(task))
                .isInstanceOf(IllegalArgumentException.class)
                .hasMessage("Backoff strategy is not defined");
    }

    @ParameterizedTest
    @CsvSource({
            "500, 500, true",
            "502, 502, true",
            "404, 404, true",
            "200, 500, false",
            "503, 502, false"
    })
    void testStatusCodeMatchingExact(int responseCode, String pattern, boolean retryExpected) {
        when(httpResponse.statusCode()).thenReturn(responseCode);
        lenient().when(backoffStrategy.initialIntervalMs()).thenReturn(100L);
        lenient().when(backoffStrategy.maxElapsedTimeMs()).thenReturn(200L);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of(pattern))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        if (retryExpected) {
            verify(scheduler, timeout(500)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
        }
        assertThat(result).succeedsWithin(DEFAULT_TIMEOUT).isEqualTo(httpResponse);
    }

    @ParameterizedTest
    @CsvSource({
            "500, 5XX, true",
            "502, 5XX, true",
            "503, 5XX, true",
            "400, 4XX, true",
            "404, 4XX, true",
            "200, 2XX, true",
            "500, 4XX, false",
            "404, 5XX, false",
            "200, 5XX, false"
    })
    void testStatusCodeMatchingPattern(int responseCode, String pattern, boolean retryExpected) {
        when(httpResponse.statusCode()).thenReturn(responseCode);
        lenient().when(backoffStrategy.initialIntervalMs()).thenReturn(100L);
        lenient().when(backoffStrategy.maxElapsedTimeMs()).thenReturn(200L);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of(pattern))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        if (retryExpected) {
            verify(scheduler, timeout(500)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
        }
        assertThat(result).succeedsWithin(DEFAULT_TIMEOUT).isEqualTo(httpResponse);
    }

    @Test
    void testSuccessfulRequestWithoutRetry() {
        when(httpResponse.statusCode()).thenReturn(200);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500", "502"))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertThat(result).succeedsWithin(DEFAULT_TIMEOUT).isEqualTo(httpResponse);

        verify(scheduler, never()).schedule(any(Runnable.class), anyLong(), any(TimeUnit.class));
    }

    @Test
    void testRetryableExceptionHandling() {
        AsyncRetryableException retryableException = new AsyncRetryableException(httpResponse);
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(retryableException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testConnectExceptionRetry() {
        ConnectException connectException = new ConnectException("Connection refused");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(connectException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testConnectTimeoutRetry() {
        IOException timeoutException = new IOException("Connect timed out");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(timeoutException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testReadTimeoutRetry() {
        IOException timeoutException = new IOException("Read timed out");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(timeoutException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testNonRetryableIOException() {
        IOException nonRetryableException = new IOException("File not found");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(nonRetryableException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertThat(result)
                .failsWithin(Duration.ofMillis(200))
                .withThrowableOfType(ExecutionException.class)
                .withCauseInstanceOf(NonRetryableException.class);

        verify(scheduler, never()).schedule(any(Runnable.class), anyLong(), any(TimeUnit.class));
    }

    @Test
    void testMaxElapsedTimeExceededWithRetryableException() {
        when(backoffStrategy.maxElapsedTimeMs()).thenReturn(50L);

        AsyncRetryableException retryableException = new AsyncRetryableException(httpResponse);
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(retryableException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertThat(result).succeedsWithin(DEFAULT_TIMEOUT);
    }

    @Test
    void testMaxElapsedTimeExceededWithOtherException() {
        RuntimeException runtimeException = new RuntimeException("Test exception");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(runtimeException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertCompletableFutureFails(result, DEFAULT_TIMEOUT, NonRetryableException.class);
    }

    @Test
    void testBackoffCalculationWithJitter() {
        when(backoffStrategy.initialIntervalMs()).thenReturn(100L);
        when(backoffStrategy.baseFactor()).thenReturn(2.0);
        when(backoffStrategy.jitterFactor()).thenReturn(0.1);
        when(backoffStrategy.maxIntervalMs()).thenReturn(500L);

        when(httpResponse.statusCode()).thenReturn(500);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), longThat(delay -> {
            return delay >= 90L && delay <= 110L;
        }), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testShutdown() {
        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        asyncRetries.shutdown();

        verify(scheduler).shutdown();
    }

    @Test
    void testRetryableExceptionProperties() {
        AsyncRetryableException exception = new AsyncRetryableException(httpResponse);

        assertThat(exception.response()).isSameAs(httpResponse);
    }

    @Test
    void testNonRetryableExceptionProperties() {
        RuntimeException cause = new RuntimeException("Original exception");
        NonRetryableException exception = new NonRetryableException(cause);

        assertThat(exception.exception()).isSameAs(cause);
        assertThat(exception.getCause()).isSameAs(cause);
    }

    @Test
    void testCompletionExceptionUnwrapping() {
        RuntimeException originalException = new RuntimeException("Original");
        CompletionException wrappedException = new CompletionException(originalException);
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(wrappedException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertThat(result)
                .failsWithin(Duration.ofMillis(200))
                .withThrowableOfType(ExecutionException.class)
                .withCauseInstanceOf(NonRetryableException.class);
    }

    @ParameterizedTest
    @ValueSource(strings = {"500", "502", "503", "5XX", "4XX"})
    void testMultipleStatusCodes(String statusCode) {
        when(httpResponse.statusCode()).thenReturn(500);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500", "502", "503", "5XX"))
                .scheduler(scheduler)
                .build();

        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.completedFuture(httpResponse);
        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        verify(scheduler, timeout(1000)).schedule(any(Runnable.class), anyLong(), eq(TimeUnit.MILLISECONDS));
    }

    @Test
    void testIOExceptionWithNullMessage() {
        IOException exceptionWithNullMessage = new IOException((String) null);
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(exceptionWithNullMessage);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertCompletableFutureFails(result, DEFAULT_TIMEOUT, NonRetryableException.class);
    }

    @Test
    void testConnectErrorDisabled() {
        when(backoffStrategy.retryConnectError()).thenReturn(false);

        ConnectException connectException = new ConnectException("Connection refused");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(connectException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertCompletableFutureFails(result, DEFAULT_TIMEOUT, NonRetryableException.class);
    }

    @Test
    void testReadTimeoutErrorDisabled() {
        when(backoffStrategy.retryReadTimeoutError()).thenReturn(false);

        IOException timeoutException = new IOException("Read timed out");
        CompletableFuture<HttpResponse<Blob>> taskResult = CompletableFuture.failedFuture(timeoutException);

        asyncRetries = AsyncRetries.builder()
                .retryConfig(retryConfig)
                .statusCodes(List.of("500"))
                .scheduler(scheduler)
                .build();

        when(task.get()).thenReturn(taskResult);

        CompletableFuture<HttpResponse<Blob>> result = asyncRetries.retry(task);

        assertCompletableFutureFails(result, DEFAULT_TIMEOUT, NonRetryableException.class);
    }
}
