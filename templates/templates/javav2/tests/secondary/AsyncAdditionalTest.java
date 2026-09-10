package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.time.Duration;
import java.time.Instant;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Base64;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;
import java.util.stream.Collectors;
import java.util.stream.LongStream;
import java.util.stream.Stream;

import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.errors.SDKException;
import org.openapis.secondary.openapi.models.operations.async.CustomClientPostResponse;
import org.openapis.secondary.openapi.models.operations.async.PaginationLimitOffsetOffsetParamsResponse;
import org.openapis.secondary.openapi.models.operations.async.PaginationLimitOffsetPageBodyResponse;
import org.openapis.secondary.openapi.models.operations.async.RequestBodyPostApplicationJsonSimpleResponse;
import org.openapis.secondary.openapi.models.operations.async.RequestBodyPutBytesResponse;
import org.openapis.secondary.openapi.models.operations.async.ResponseBodyBytesGetResponse;
import org.openapis.secondary.openapi.models.operations.async.RetriesAfterResponse;
import org.openapis.secondary.openapi.models.operations.async.RetriesAttemptCountResponse;
import org.openapis.secondary.openapi.models.operations.async.RetriesGetResponse;
import org.openapis.secondary.openapi.models.components.Enum;
import org.openapis.secondary.openapi.models.components.Int32Enum;
import org.openapis.secondary.openapi.models.components.IntEnum;
import org.openapis.secondary.openapi.models.components.JsonEvent;
import org.openapis.secondary.openapi.models.components.LimitOffsetConfig;
import org.openapis.secondary.openapi.models.components.SimpleObject;
import org.openapis.secondary.openapi.models.operations.JsonlStreamResponseBody;
import org.openapis.secondary.openapi.utils.Blob;
import org.openapis.secondary.openapi.utils.EventStream;
import org.openapis.secondary.openapi.utils.JsonLStream;
import org.openapis.secondary.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;
import org.openapis.secondary.openapi.utils.transport.HttpResponse;

/**
 * Runtime tests for the CompletableFuture async surface on the Java 11 / OkHttp
 * transport (secondary variant). This is the same OkHttp
 * {@code sendAsync} enqueue-callback transport that
 * tertiary (java8) exercises, but generated for the java11 arm with the "raw"
 * getter style: response bodies and nested optional fields are {@code @Nullable}
 * raw values, not {@code Optional}. Unary operations return
 * {@code CompletableFuture<XResponse>}, SSE streaming returns a
 * {@code CompletableFuture} of a blocking {@link EventStream}, and pagination
 * returns {@code CompletableFuture<Stream<XResponse>>} via the blocking
 * auto-pager. No Reactive Streams types are generated on this arm.
 */
public class AsyncAdditionalTest {

    private static SimpleObject createSimpleObject() {
        return SimpleObject.builder()
                .any("any")
                .bool(true)
                .date(LocalDate.parse("2020-01-01"))
                .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00.000000001Z"))
                .enum_(Enum.ONE)
                .float32(1.1f)
                .int_(1L)
                .int32(1)
                .int32Enum(Int32Enum.FIFTY_FIVE)
                .intEnum(IntEnum.Second)
                .num(1.1)
                .str("test")
                .strOpt("testOptional")
                .boolOpt(true)
                .build();
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncBasicCall() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        SimpleObject obj = createSimpleObject();

        CompletableFuture<RequestBodyPostApplicationJsonSimpleResponse> future =
                s.toAsync().requestBodies().requestBodyPostApplicationJsonSimple(obj);

        RequestBodyPostApplicationJsonSimpleResponse res = future.get(8, TimeUnit.SECONDS);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res());
        assertEquals(obj, res.res().json());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncBytesDownload() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        ResponseBodyBytesGetResponse res = s.toAsync().responseBodies()
                .responseBodyBytesGetDirect()
                .get(10, TimeUnit.SECONDS);
        assertEquals(200, res.statusCode());
        assertNotNull(res.bytes());

        // async response bodies are plain InputStreams, same as sync
        try (InputStream in = res.bytes()) {
            byte[] bytes = IOUtils.toByteArray(in);
            assertEquals(100, bytes.length);
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncEventStreamBlockingIteration() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // async SSE: future completes with a blocking-iterator EventStream
        List<JsonEvent> list = new ArrayList<>();
        try (EventStream<JsonEvent> events = s.toAsync().eventstreams().json()
                .call()
                .get(10, TimeUnit.SECONDS)) {
            for (JsonEvent event : events) {
                list.add(event);
            }
        }

        assertEquals(4, list.size());
        String content = list.stream()
                .map(x -> x.data().content())
                .collect(Collectors.joining());
        assertEquals("Hello world!", content);
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncEventStreamToList() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        try (EventStream<JsonEvent> events = s.toAsync().eventstreams().json()
                .call()
                .get(10, TimeUnit.SECONDS)) {
            List<JsonEvent> list = events.toList();
            assertEquals(4, list.size());
            String content = list.stream()
                    .map(x -> x.data().content())
                    .collect(Collectors.joining());
            assertEquals("Hello world!", content);
        }
    }

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    public void testAsyncPaginationCallAsStream() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        Stream<PaginationLimitOffsetPageBodyResponse> stream = s.toAsync().pagination()
                .paginationLimitOffsetPageBody()
                .request(LimitOffsetConfig.builder()
                        .page(1L)
                        .limit(5L)
                        .build())
                .callAsStream()
                .get(15, TimeUnit.SECONDS);

        List<PaginationLimitOffsetPageBodyResponse> pages = stream.collect(Collectors.toList());
        assertEquals(4, pages.size(), "20 items with limit 5 should yield 4 pages");
        pages.forEach(page -> assertEquals(200, page.statusCode()));
        assertEquals(
                LongStream.range(0, 20).boxed().collect(Collectors.toList()),
                pages.stream()
                        .flatMap(page -> page.paginationResponse().resultArray().stream())
                        .collect(Collectors.toList()),
                "flattened pages should contain every item in order");
    }

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    public void testAsyncPaginationCallAsStreamUnwrapped() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        Stream<Long> stream = s.toAsync().pagination()
                .paginationLimitOffsetPageBody()
                .request(LimitOffsetConfig.builder()
                        .page(1L)
                        .limit(5L)
                        .build())
                .callAsStreamUnwrapped()
                .get(15, TimeUnit.SECONDS);

        List<Long> items = stream.collect(Collectors.toList());
        assertEquals(
                LongStream.range(0, 20).boxed().collect(Collectors.toList()),
                items,
                "should stream every item across all pages in order");
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncErrorPropagation() throws Exception {
        // an unroutable server URL must fail the future with the underlying
        // IOException, not hang or wrap it beyond ExecutionException
        SDK s = SDK.builder().serverURL("http://localhost:1").build();
        CompletableFuture<RequestBodyPostApplicationJsonSimpleResponse> future =
                s.toAsync().requestBodies().requestBodyPostApplicationJsonSimple(createSimpleObject());
        ExecutionException ex = assertThrows(ExecutionException.class, () -> future.get(8, TimeUnit.SECONDS));
        assertInstanceOf(IOException.class, ex.getCause(),
                "connection failures must surface the underlying IOException as-is, got: " + ex.getCause());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncApiErrorPropagation() throws Exception {
        // API errors are the same error classes the sync surface throws,
        // surfaced through the future's exceptional completion
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        ExecutionException ex = assertThrows(ExecutionException.class, () -> s.toAsync().errors()
                .statusGetError()
                .statusCode(500L)
                .call()
                .get(8, TimeUnit.SECONDS));
        assertInstanceOf(SDKException.class, ex.getCause());
        assertEquals(500, ((SDKException) ex.getCause()).code());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncRetries() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        RetriesAfterResponse res = s.toAsync().retries().retriesAfter()
                .requestId("request-id")
                .numRetries(10L)
                .call()
                .get(8, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertNotNull(res.retries());
        assertEquals(10L, res.retries().retries());
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncAttemptCountBackoff() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        RetriesAttemptCountResponse res = s.toAsync().retries().retriesAttemptCount()
                .requestId("request-id-attempt-count")
                .numRetries(3L)
                .retryAfterMsVal(1L)
                .call()
                .get(8, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertNotNull(res.retries());
        assertEquals(3L, res.retries().retries());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testMultipleAsyncCallsInParallel() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        List<CompletableFuture<RetriesGetResponse>> futures = Arrays.asList(
                s.toAsync().retries().retriesGet("request-1"),
                s.toAsync().retries().retriesGet("request-2"),
                s.toAsync().retries().retriesGet("request-3"));

        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).get(10, TimeUnit.SECONDS);

        for (CompletableFuture<RetriesGetResponse> future : futures) {
            RetriesGetResponse res = future.get();
            assertEquals(200, res.statusCode());
            assertNotNull(res.retries());
        }
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testBinaryUpload() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        byte[] randomData = new byte[1024];
        new Random().nextBytes(randomData);
        // repeatable Blob upload: replayable across retry attempts
        Blob blob = Blob.from(randomData);

        RequestBodyPutBytesResponse res = s.toAsync().requestBodies()
                .requestBodyPutBytes(blob)
                .get(8, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertNotNull(res.res());
        String encoded = Base64.getEncoder().encodeToString(randomData);
        assertTrue(res.res().data().contains(encoded));
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncStreamToFile() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        ResponseBodyBytesGetResponse res = s.toAsync().responseBodies()
                .responseBodyBytesGet()
                .call()
                .get(10, TimeUnit.SECONDS);
        assertEquals(200, res.statusCode());
        assertNotNull(res.bytes());

        Path tempFile = Files.createTempFile("async-stream-to-file", ".bin");
        try (InputStream in = res.bytes()) {
            Files.copy(in, tempFile, StandardCopyOption.REPLACE_EXISTING);
            assertEquals(100, Files.readAllBytes(tempFile).length);
        } finally {
            Files.deleteIfExists(tempFile);
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncStreamToByteArray() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        ResponseBodyBytesGetResponse res = s.toAsync().responseBodies()
                .responseBodyBytesGet()
                .call()
                .get(10, TimeUnit.SECONDS);
        assertNotNull(res.bytes());

        try (InputStream in = res.bytes()) {
            assertEquals(100, IOUtils.toByteArray(in).length);
        }
    }

    @Test
    @Timeout(value = 20, unit = TimeUnit.SECONDS)
    public void testAsyncPaginationSlowConsumer() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // pages are fetched lazily by the blocking auto-pager: slow per-page
        // consumption on the caller's thread must not break paging
        Stream<PaginationLimitOffsetOffsetParamsResponse> stream = s.toAsync().pagination()
                .paginationLimitOffsetOffsetParams()
                .offset(0L)
                .limit(3L)
                .callAsStream()
                .get(15, TimeUnit.SECONDS);

        List<PaginationLimitOffsetOffsetParamsResponse> pages = new ArrayList<>();
        Iterator<PaginationLimitOffsetOffsetParamsResponse> it = stream.iterator();
        while (it.hasNext() && pages.size() < 2) {
            pages.add(it.next());
            Thread.sleep(500);
        }

        assertEquals(2, pages.size(), "should receive two pages under slow consumption");
        assertEquals(200, pages.get(0).statusCode());
        assertEquals(
                LongStream.range(0, 3).boxed().collect(Collectors.toList()),
                pages.get(0).paginationResponse().resultArray());
        assertEquals(200, pages.get(1).statusCode());
        assertEquals(
                LongStream.range(3, 6).boxed().collect(Collectors.toList()),
                pages.get(1).paginationResponse().resultArray());
    }

    @Test
    @Timeout(value = 25, unit = TimeUnit.SECONDS)
    public void testConcurrentPaginationStreams() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // two independent unwrapped streams consumed on parallel threads:
        // each pages independently; results must not interleave/corrupt
        CompletableFuture<Stream<Long>> futA = s.toAsync().pagination()
                .paginationLimitOffsetOffsetParams()
                .offset(0L)
                .callAsStreamUnwrapped();
        CompletableFuture<Stream<Long>> futB = s.toAsync().pagination()
                .paginationLimitOffsetOffsetParams()
                .offset(10L)
                .callAsStreamUnwrapped();

        AtomicReference<List<Long>> resultA = new AtomicReference<>();
        AtomicReference<List<Long>> resultB = new AtomicReference<>();
        AtomicReference<Throwable> failure = new AtomicReference<>();

        Thread tA = new Thread(() -> {
            try {
                resultA.set(futA.get(15, TimeUnit.SECONDS)
                        .limit(10)
                        .collect(Collectors.toList()));
            } catch (Throwable t) {
                failure.compareAndSet(null, t);
            }
        });
        Thread tB = new Thread(() -> {
            try {
                resultB.set(futB.get(15, TimeUnit.SECONDS)
                        .limit(10)
                        .collect(Collectors.toList()));
            } catch (Throwable t) {
                failure.compareAndSet(null, t);
            }
        });
        tA.start();
        tB.start();
        tA.join(TimeUnit.SECONDS.toMillis(20));
        tB.join(TimeUnit.SECONDS.toMillis(20));

        if (failure.get() != null) {
            throw new AssertionError("concurrent pagination stream failed", failure.get());
        }
        assertNotNull(resultA.get(), "stream A did not complete");
        assertNotNull(resultB.get(), "stream B did not complete");

        assertEquals(
                LongStream.range(0, 10).boxed().collect(Collectors.toList()),
                resultA.get(),
                "offset-0 stream should contain its own ordered range");
        assertEquals(
                LongStream.range(10, 20).boxed().collect(Collectors.toList()),
                resultB.get(),
                "offset-10 stream should contain its own ordered range");
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncOperationWithCustomHttpClient() throws Exception {
        String modifiedSinceValue = Instant.now().minus(Duration.ofMinutes(30)).toString();

        // custom okhttp-arm client overrides sendAsync(transport.HttpRequest)
        // and injects an `If-Modified-Since` header via toBuilder(); verifies
        // the request is threaded through the async transport seam untouched.
        SpeakeasyHTTPClient customClient = new SpeakeasyHTTPClient() {
            @Override
            public CompletableFuture<HttpResponse<InputStream>> sendAsync(HttpRequest request) {
                HttpRequest modified = request.toBuilder()
                        .header("If-Modified-Since", modifiedSinceValue)
                        .build();
                return super.sendAsync(modified);
            }
        };

        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(customClient).build();

        CustomClientPostResponse res = s.toAsync().customClient().customClientPost()
                .headerParam("some-header")
                .pathParam("some-path")
                .queryStringParam("some-query")
                .call()
                .get(5, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertNotNull(res.res());
        assertEquals("some-query", res.res().args().queryStringParam());
        assertTrue(res.res().url().contains("some-path"));
        Map<String, String> headers = res.res().headers();
        assertEquals("some-header", headers.get("Headerparam"));
        assertEquals(modifiedSinceValue, headers.get("If-Modified-Since"),
                "should contain `If-Modified-Since` injected by custom client");
    }

    @Test
    @Timeout(value = 10, unit = TimeUnit.SECONDS)
    public void testAsyncDebugLoggingReadsBodyOffDispatcher() throws Exception {
        AtomicReference<String> completionThread = new AtomicReference<>();
        java.util.concurrent.CountDownLatch releaseBody = new java.util.concurrent.CountDownLatch(1);
        byte[] responseBytes = "{\"ok\":true}".getBytes(java.nio.charset.StandardCharsets.UTF_8);
        com.sun.net.httpserver.HttpServer server = com.sun.net.httpserver.HttpServer.create(
                new java.net.InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/debug", exchange -> {
            exchange.getResponseHeaders().add("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, responseBytes.length);
            try (java.io.OutputStream body = exchange.getResponseBody()) {
                if (releaseBody.await(5, TimeUnit.SECONDS)) {
                    body.write(responseBytes);
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        });
        server.start();
        try {
            SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
            client.enableDebugLogging(true);
            HttpRequest request = HttpRequest.builder()
                    .method("GET")
                    .uri(java.net.URI.create("http://127.0.0.1:"
                            + server.getAddress().getPort() + "/debug"))
                    .build();

            CompletableFuture<HttpResponse<InputStream>> response = client.sendAsync(request)
                    .thenApply(value -> {
                        completionThread.set(Thread.currentThread().getName());
                        return value;
                    });
            releaseBody.countDown();

            byte[] body = IOUtils.toByteArray(response.get(5, TimeUnit.SECONDS).body());
            assertEquals("{\"ok\":true}", new String(body, java.nio.charset.StandardCharsets.UTF_8));
            assertFalse(completionThread.get().contains("OkHttp"),
                    "debug response buffering must not complete on an OkHttp dispatcher thread");
        } finally {
            releaseBody.countDown();
            server.stop(0);
        }
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncJsonlStreamBlockingIteration() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // async JSONL: future completes with a blocking-iterator JsonLStream
        try (JsonLStream<JsonlStreamResponseBody> events = s.toAsync().jsonl()
                .jsonlStreamDirect()
                .get(10, TimeUnit.SECONDS)) {
            List<JsonlStreamResponseBody> list = events.stream().collect(Collectors.toList());
            assertEquals(2, list.size());
            assertEquals("Peter", list.get(0).name());
            assertEquals(Arrays.asList("Go", "Python"), list.get(0).skills());
            assertEquals("John", list.get(1).name());
            assertEquals(Arrays.asList("Go", "Rust"), list.get(1).skills());
        }
    }

    @Test
    public void testAsyncStreamingClientSelectionByAcceptMediaType() throws Exception {
        // Async calls whose Accept advertises a streaming media type must use
        // the streaming client (read timeout disabled) so a stream that idles
        // between events longer than the default read timeout is not killed.
        SpeakeasyHTTPClient httpClient = new SpeakeasyHTTPClient();
        java.lang.reflect.Method selector = SpeakeasyHTTPClient.class
                .getDeclaredMethod("asyncCallClient", HttpRequest.class);
        selector.setAccessible(true);

        assertEquals(0, selectedReadTimeoutMillis(httpClient, selector, "text/event-stream"));
        assertEquals(0, selectedReadTimeoutMillis(httpClient, selector, "application/x-ndjson"));
        assertEquals(0, selectedReadTimeoutMillis(httpClient, selector, "application/jsonl"));
        // media-type parameters, case, and multi-valued Accept must not defeat detection
        assertEquals(0, selectedReadTimeoutMillis(httpClient, selector, "Application/JSONL; charset=utf-8"));
        assertEquals(0, selectedReadTimeoutMillis(httpClient, selector, "application/json, application/jsonl"));
        // unary calls stay on the default client with its read timeout intact
        assertTrue(selectedReadTimeoutMillis(httpClient, selector, "application/json") > 0,
                "unary Accept should select the default client (non-zero read timeout)");
    }

    private static int selectedReadTimeoutMillis(SpeakeasyHTTPClient httpClient,
            java.lang.reflect.Method selector, String accept) throws Exception {
        HttpRequest request = HttpRequest.builder()
                .method("GET")
                .uri(java.net.URI.create(HTTPBIN_URL))
                .header("Accept", accept)
                .build();
        okhttp3.OkHttpClient selected = (okhttp3.OkHttpClient) selector.invoke(httpClient, request);
        return selected.readTimeoutMillis();
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    public void testConcurrentEventStreams() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // several concurrent SSE calls to one host: futures complete on
        // headers and consumption happens on consumer threads, so streams
        // must all deliver independently
        int streams = 4;
        List<CompletableFuture<EventStream<JsonEvent>>> futures = new ArrayList<>();
        for (int i = 0; i < streams; i++) {
            futures.add(s.toAsync().eventstreams().json().call());
        }

        AtomicReference<Throwable> failure = new AtomicReference<>();
        List<Thread> consumers = new ArrayList<>();
        for (CompletableFuture<EventStream<JsonEvent>> future : futures) {
            Thread t = new Thread(() -> {
                try (EventStream<JsonEvent> events = future.get(15, TimeUnit.SECONDS)) {
                    String content = events.toList().stream()
                            .map(x -> x.data().content())
                            .collect(Collectors.joining());
                    assertEquals("Hello world!", content);
                } catch (Throwable t2) {
                    failure.compareAndSet(null, t2);
                }
            });
            consumers.add(t);
            t.start();
        }
        for (Thread t : consumers) {
            t.join(TimeUnit.SECONDS.toMillis(25));
            assertFalse(t.isAlive(), "concurrent SSE consumer did not terminate: " + t.getName());
        }
        if (failure.get() != null) {
            throw new AssertionError("concurrent SSE stream failed", failure.get());
        }
    }

    @Test
    @Timeout(value = 60, unit = TimeUnit.SECONDS)
    public void testAsyncPaginationBlockingContinuationsDoNotDeadlockDispatcher() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // regression: callAsStream futures used to complete on the okhttp
        // dispatcher thread with the per-host slot still held. Consuming the
        // stream inside thenAccept blocks on subsequent page fetches, so more
        // concurrent paginations than maxRequestsPerHost (default 5) would
        // deadlock all dispatcher slots. The executor handoff makes this safe.
        int concurrent = 8;
        List<CompletableFuture<Void>> futures = new ArrayList<>();
        for (int i = 0; i < concurrent; i++) {
            futures.add(s.toAsync().pagination()
                    .paginationLimitOffsetOffsetParams()
                    .offset(0L)
                    .limit(3L)
                    .callAsStream()
                    .thenAccept(pages -> pages.limit(3).forEach(page -> {
                        assertEquals(200, page.statusCode());
                        assertNotNull(page.paginationResponse());
                    })));
        }
        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).get(45, TimeUnit.SECONDS);
    }

    @Test
    @Timeout(value = 120, unit = TimeUnit.SECONDS)
    public void testAsyncPaginationBlockedConsumersUseStreamCompletionExecutor() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // Regression (deadlock): blocking consumers attached with thenAccept
        // used to occupy response-completion workers while waiting for later
        // pages. Complete pagination futures on the separate stream executor
        // so every consumer can start without pinning response or dispatcher
        // threads.
        int n = 24;
        java.util.concurrent.CountDownLatch started = new java.util.concurrent.CountDownLatch(n);
        java.util.concurrent.CountDownLatch release = new java.util.concurrent.CountDownLatch(1);
        java.util.Set<String> badThreads = java.util.concurrent.ConcurrentHashMap.newKeySet();
        List<CompletableFuture<Void>> futures = new ArrayList<>();
        try {
            for (int i = 0; i < n; i++) {
                futures.add(s.toAsync().pagination()
                        .paginationLimitOffsetOffsetParams()
                        .offset(0L)
                        .limit(3L)
                        .callAsStream()
                        .thenAccept(pages -> {
                            String name = Thread.currentThread().getName();
                            if (name.contains("ForkJoinPool") || name.contains("OkHttp")) {
                                badThreads.add(name);
                            }
                            started.countDown();
                            try {
                                release.await(60, TimeUnit.SECONDS);
                            } catch (InterruptedException e) {
                                Thread.currentThread().interrupt();
                            }
                            pages.limit(2).forEach(page -> assertEquals(200, page.statusCode()));
                        }));
            }
            assertTrue(started.await(60, TimeUnit.SECONDS),
                    "every pagination consumer should start on the stream-completion executor");
        } finally {
            release.countDown();
        }
        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).get(45, TimeUnit.SECONDS);
        assertTrue(badThreads.isEmpty(),
                "pagination continuations ran on restricted threads: " + badThreads);
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncCompletionNeverOnDispatcherThread() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // Dependent stages run on the completing thread. Async body parsing
        // must complete away from OkHttp's dispatcher.
        AtomicReference<String> unaryThread = new AtomicReference<>();
        s.toAsync().requestBodies().requestBodyPostApplicationJsonSimple(createSimpleObject())
                .thenApply(res -> {
                    unaryThread.set(Thread.currentThread().getName());
                    return res;
                })
                .get(8, TimeUnit.SECONDS);
        assertNotNull(unaryThread.get());
        assertFalse(unaryThread.get().contains("OkHttp"),
                "unary completion ran on dispatcher thread: " + unaryThread.get());

        AtomicReference<String> paginationThread = new AtomicReference<>();
        s.toAsync().pagination()
                .paginationLimitOffsetOffsetParams()
                .offset(0L)
                .limit(3L)
                .callAsStream()
                .thenApply(pages -> {
                    paginationThread.set(Thread.currentThread().getName());
                    return pages;
                })
                .get(10, TimeUnit.SECONDS);
        assertNotNull(paginationThread.get());
        assertFalse(paginationThread.get().contains("OkHttp"),
                "pagination completion ran on dispatcher thread: " + paginationThread.get());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncFutureCancellationCancelsCallAndReleasesConnection() throws Exception {
        // cancelling the returned CompletableFuture must propagate to the
        // underlying okhttp Call (whenComplete -> call.cancel()) and the
        // connection acquired for the in-flight request must be released, not
        // leaked. Covers the cancel branch and the connection-release side of
        // the late-complete guard (!future.complete(response) -> okResponse.close()).
        SpeakeasyHTTPClient httpClient = new SpeakeasyHTTPClient();

        java.util.concurrent.CountDownLatch acquired = new java.util.concurrent.CountDownLatch(1);
        java.util.concurrent.CountDownLatch canceled = new java.util.concurrent.CountDownLatch(1);
        AtomicInteger acquiredCount = new AtomicInteger();
        AtomicInteger releasedCount = new AtomicInteger();

        okhttp3.EventListener listener = new okhttp3.EventListener() {
            @Override
            public void connectionAcquired(okhttp3.Call call, okhttp3.Connection connection) {
                acquiredCount.incrementAndGet();
                acquired.countDown();
            }

            @Override
            public void connectionReleased(okhttp3.Call call, okhttp3.Connection connection) {
                releasedCount.incrementAndGet();
            }

            @Override
            public void canceled(okhttp3.Call call) {
                canceled.countDown();
            }
        };
        injectEventListener(httpClient, listener);

        // /delay holds the response so the future is cancelled while the request
        // is still in flight (cancel-before-headers ordering)
        HttpRequest request = HttpRequest.builder()
                .method("GET")
                .uri(java.net.URI.create(HTTPBIN_URL + "/delay/5"))
                .build();

        CompletableFuture<HttpResponse<InputStream>> future = httpClient.sendAsync(request);

        assertTrue(acquired.await(5, TimeUnit.SECONDS), "request should reach the wire before cancellation");
        assertTrue(future.cancel(true));
        assertTrue(future.isCancelled());

        assertTrue(canceled.await(5, TimeUnit.SECONDS),
                "cancelling the returned future must cancel the underlying okhttp Call");

        long deadline = System.currentTimeMillis() + 5000;
        while (releasedCount.get() < acquiredCount.get() && System.currentTimeMillis() < deadline) {
            Thread.sleep(20);
        }
        assertTrue(acquiredCount.get() >= 1, "a connection should have been acquired for the in-flight call");
        assertEquals(acquiredCount.get(), releasedCount.get(),
                "cancelled call must release its connection, not leak it");
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncOperationLevelCancellationCancelsCallAndReleasesConnection() throws Exception {
        // cancelling the CompletableFuture returned by an SDK operation — several
        // completion stages downstream of the transport-level future — must relay
        // to the underlying okhttp Call (CompletableFuture cancellation does not
        // propagate upstream on its own) and release the connection, not leak it.
        SpeakeasyHTTPClient httpClient = new SpeakeasyHTTPClient();

        java.util.concurrent.CountDownLatch acquired = new java.util.concurrent.CountDownLatch(1);
        java.util.concurrent.CountDownLatch canceled = new java.util.concurrent.CountDownLatch(1);
        AtomicInteger acquiredCount = new AtomicInteger();
        AtomicInteger releasedCount = new AtomicInteger();

        okhttp3.EventListener listener = new okhttp3.EventListener() {
            @Override
            public void connectionAcquired(okhttp3.Call call, okhttp3.Connection connection) {
                acquiredCount.incrementAndGet();
                acquired.countDown();
            }

            @Override
            public void connectionReleased(okhttp3.Call call, okhttp3.Connection connection) {
                releasedCount.incrementAndGet();
            }

            @Override
            public void canceled(okhttp3.Call call) {
                canceled.countDown();
            }
        };
        injectEventListener(httpClient, listener);

        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(httpClient).build();

        // /delay/{seconds} holds the response so the future is cancelled while
        // the request is still in flight
        CompletableFuture<org.openapis.secondary.openapi.models.operations.async.CancelledRequestResponse> future =
                s.toAsync().cancellation().cancelledRequest(5L);

        assertTrue(acquired.await(5, TimeUnit.SECONDS), "request should reach the wire before cancellation");
        assertTrue(future.cancel(true));
        assertTrue(future.isCancelled());

        assertTrue(canceled.await(5, TimeUnit.SECONDS),
                "cancelling the operation-level future must cancel the underlying okhttp Call");

        long deadline = System.currentTimeMillis() + 5000;
        while (releasedCount.get() < acquiredCount.get() && System.currentTimeMillis() < deadline) {
            Thread.sleep(20);
        }
        assertTrue(acquiredCount.get() >= 1, "a connection should have been acquired for the in-flight call");
        assertEquals(acquiredCount.get(), releasedCount.get(),
                "cancelled call must release its connection, not leak it");
    }

    // swaps the client's internal OkHttpClient for one carrying an EventListener
    // so the test can observe Call cancellation and connection release
    private static void injectEventListener(SpeakeasyHTTPClient httpClient, okhttp3.EventListener listener)
            throws Exception {
        java.lang.reflect.Field field = SpeakeasyHTTPClient.class.getDeclaredField("client");
        field.setAccessible(true);
        okhttp3.OkHttpClient base = (okhttp3.OkHttpClient) field.get(httpClient);
        field.set(httpClient, base.newBuilder().eventListener(listener).build());
    }
}
