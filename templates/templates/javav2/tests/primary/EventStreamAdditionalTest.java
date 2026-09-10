package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.fail;
import static org.openapis.openapi.CommonHelpers.recordTest;
import static org.assertj.core.api.Assertions.assertThat;
import static org.openapis.openapi.Helpers.*;
import static org.openapis.openapi.Helpers.HTTPBIN_URL;

import java.io.IOException;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URISyntaxException;
import java.net.http.HttpResponse;
import java.net.http.HttpRequest;
import java.time.Duration;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CancellationException;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Flow;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;
import org.openapis.openapi.models.errors.TeapotJSONError;
import org.openapis.openapi.models.operations.ChatHeartbeatResponse;
import org.openapis.openapi.models.operations.ChatRequestBody;
import org.openapis.openapi.models.operations.ChatResponse;
import org.openapis.openapi.models.operations.ChatSkipSentinelRequestBody;
import org.openapis.openapi.models.operations.ChatSkipSentinelResponse;
import org.openapis.openapi.models.operations.DifferentDataSchemasResponse;
import org.openapis.openapi.models.operations.OptionalDataResponse;
import org.openapis.openapi.models.operations.JsonResponse;
import org.openapis.openapi.models.operations.MultilineResponse;
import org.openapis.openapi.models.operations.PartialWithCommentsResponse;
import org.openapis.openapi.models.operations.RichResponse;
import org.openapis.openapi.models.operations.StayOpenResponse;
import org.openapis.openapi.models.operations.TextRequestBuilder;
import org.openapis.openapi.models.operations.TextResponse;
import org.openapis.openapi.models.operations.WptComplianceResponse;
import org.openapis.openapi.models.shared.ChatCompletionEvent;
import org.openapis.openapi.models.shared.ChatCompletionStream;
import org.openapis.openapi.models.shared.ChatCompletionEventData;
import org.openapis.openapi.models.shared.DifferentDataSchemas;
import org.openapis.openapi.models.shared.DifferentDataSchemasData;
import org.openapis.openapi.models.shared.Event;
import org.openapis.openapi.models.shared.HeartbeatEvent;
import org.openapis.openapi.models.operations.LargeEventSmallChunksResponse;
import org.openapis.openapi.models.operations.MalformedStreamResponse;
import org.openapis.openapi.models.operations.SplitBoundariesResponse;
import org.openapis.openapi.models.shared.JsonEvent;
import org.openapis.openapi.models.shared.LargeChunkedEvent;
import org.openapis.openapi.models.shared.SplitBoundaryEvent;
import org.openapis.openapi.models.shared.MessageEvent;
import org.openapis.openapi.models.shared.MixedDataEvent;
import org.openapis.openapi.models.shared.MixedDataEventEvent;
import org.openapis.openapi.models.shared.OptionalDataEvent;
import org.openapis.openapi.models.shared.OptionalDataEventPayload;
import org.openapis.openapi.models.shared.PartialWithCommentsEvent;
import org.openapis.openapi.models.shared.PartialWithCommentsEventData;
import org.openapis.openapi.models.shared.RichCompletionEvent;
import org.openapis.openapi.models.shared.RichCompletionEventData;
import org.openapis.openapi.models.shared.RichStream;
import org.openapis.openapi.models.shared.SentinelEvent;
import org.openapis.openapi.models.shared.SseEvent;
import org.openapis.openapi.models.shared.StopReason;
import org.openapis.openapi.models.shared.TextEvent;
import org.openapis.openapi.models.shared.UrlEvent;
import org.openapis.openapi.utils.AsyncResponse;
import org.openapis.openapi.utils.EventStream;
import org.openapis.openapi.utils.EventStreamMessage;
import org.openapis.openapi.utils.Helpers;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.openapi.utils.Utils;
import org.openapis.openapi.utils.Blob;
import reactor.core.publisher.Flux;
import reactor.test.StepVerifier;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;


public class EventStreamAdditionalTest {
    private ObjectMapper objectMapper;
    private TypeReference<JsonNode> jsonTypeReference;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        jsonTypeReference = new TypeReference<>() {
        };
    }

    // ===== SYNCHRONOUS EVENT STREAM TESTS =====

    @Test
    public void testJsonFromFullEventMessage() throws JsonProcessingException {
        String json = Utils.json(
                new EventStreamMessage(Optional.of("event1"), Optional.of("id1"), Optional.of(123), Optional.of("some data")),
                new ObjectMapper(), true);
        assertEquals("{\"event\":\"event1\",\"id\":\"id1\",\"retry\":123,\"data\":\"some data\"}", json);
    }

    @Test
    public void testJsonFromEmptyEventMessage() throws JsonProcessingException {
        String json = Utils.json(
                new EventStreamMessage(Optional.empty(), Optional.empty(), Optional.empty(), Optional.empty()),
                new ObjectMapper(), true);
        assertEquals("{}", json);
    }

    @Test
    public void testEventStreamJSONData() throws Exception {
        recordTest("event-stream-json-data");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        JsonResponse res = s.eventstreams().json().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        try (EventStream<JsonEvent> events = res.events()) {
            List<JsonEvent> list = events.stream().collect(Collectors.toList());
            assertEquals(4, list.size());
            String content = list //
                    .stream() //
                    .map(x -> x.data().content()) //
                    .collect(Collectors.joining());
            assertEquals("Hello world!", content);
            // before closure will be empty once stream traversed once
            assertFalse(events.next().isPresent());
        }
    }

    @Test
    public void testEventStreamJSONDataWithoutJavaUtilStream() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        JsonResponse res = s.eventstreams().json().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        try (EventStream<JsonEvent> events = res.events()) {
            List<JsonEvent> list = new ArrayList<>();
            Optional<JsonEvent> event;
            String content = "";
            while ((event = events.next()).isPresent()) {
                list.add(event.get());
                content += event.get().data().content();
            }
            assertEquals(4, list.size());
            assertEquals("Hello world!", content);
        }
    }

    @Test
    public void testEventStreamTextData() throws Exception {
        recordTest("event-stream-text-data");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        TextResponse res = s.eventstreams().text().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        List<TextEvent> list = res.events().toList();
        assertEquals(4, list.size());
        String content = list //
                .stream() //
                .map(x -> x.data()) //
                .collect(Collectors.joining());
        assertEquals("Hello world!", content);
    }

    @Test
    public void testEventStreamMultilineData() throws Exception {
        recordTest("event-stream-multiline-data");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        MultilineResponse res = s.eventstreams().multiline().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        List<TextEvent> list = res.events().toList();
        assertEquals(1, list.size());
        String content = list //
                .stream() //
                .map(x -> x.data()) //
                .collect(Collectors.joining());
        assertEquals("YHOO\n+2\n10", content);
    }

    @Test
    public void testEventStreamRichEvents() throws Exception {
        recordTest("event-stream-rich-events");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        RichResponse res = s.eventstreams().rich().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        List<RichStream> list = res.events().toList();
        assertEquals(3, list.size());
        assertEquals(RichCompletionEvent.builder() //
                .id("job-1") //
                .data(RichCompletionEventData.builder() //
                        .completion("Hello") //
                        .model("jeeves-1") //
                        .build()) //
                .build(), list.get(0));
        assertEquals(HeartbeatEvent.builder() //
                .data("ping") //
                .retry(3000L) //
                .build(), //
                list.get(1));
        assertEquals(RichCompletionEvent.builder() //
                .id("job-1") //
                .data(RichCompletionEventData.builder() //
                        .completion("world!") //
                        .model("jeeves-1") //
                        .stopReason(StopReason.STOP_SEQUENCE) //
                        .build()) //
                .build(), //
                list.get(2));
    }

    @Test
    public void testEventStreamWithSentinelEvents() throws Exception {
        recordTest("event-stream-chat-sentinel-event");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        ChatResponse res = s.eventstreams().chat() //
                .request(ChatRequestBody.builder() //
                        .prompt("Print test content") //
                        .build()) //
                .call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        List<ChatCompletionStream> list = res.events().toList();
        assertEquals(5, list.size());
        assertEquals(ChatCompletionEvent.builder() //
                .data(ChatCompletionEventData.builder() //
                        .content("Hello") //
                        .build()) //
                .build(), //
                list.get(0).chatCompletionEvent().get());
        assertEquals(ChatCompletionEvent.builder() //
                .data(ChatCompletionEventData.builder() //
                        .content(" ") //
                        .build()) //
                .build(), //
                list.get(1).chatCompletionEvent().get());
        assertEquals(ChatCompletionEvent.builder() //
                .data(ChatCompletionEventData.builder() //
                        .content("world") //
                        .build()) //
                .build(), //
                list.get(2).chatCompletionEvent().get());
        assertEquals(ChatCompletionEvent.builder() //
                .data(ChatCompletionEventData.builder() //
                        .content("!") //
                        .build()) //
                .build(), //
                list.get(3).chatCompletionEvent().get());
        assertEquals(SentinelEvent.builder().build(), list.get(4).sentinelEvent().get());
    }

    @Test
    public void testEventStreamSkipSentinel() throws Exception {
        recordTest("event-stream-chat-skip-sentinel");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        ChatSkipSentinelResponse res = s.eventstreams().chatSkipSentinel() //
                .request(ChatSkipSentinelRequestBody.builder().prompt("Print test content").build()) //
                .call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        try (EventStream<ChatCompletionEvent> events = res.events()) {
            List<String> list = events.stream().map(x -> x.data().content()).collect(Collectors.toList());
            assertEquals(List.of("Hello", " ", "world", "!"), list);
        }
    }

    @Test
    public void testEventStreamWithDifferentDataSchemas() throws Exception {
        recordTest("event-stream-different-data-schemas");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        DifferentDataSchemasResponse res = s.eventstreams().differentDataSchemas().call();
        List<DifferentDataSchemas> list = res.events().toList();
        assertEquals(6, list.size());
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-1") //
                .data(DifferentDataSchemasData.of( //
                        MessageEvent.builder() //
                                .id(123L) //
                                .content("Here is your url") //
                                .build())) //
                .event(Event.MESSAGE) //
                .build(), //
                list.get(0));
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-2") //
                .data(DifferentDataSchemasData.of( //
                        UrlEvent.builder() //
                                .url("https://example.com") //
                                .build())) //
                .event(Event.URL) //
                .build(), //
                list.get(1));
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-3") //
                .data(DifferentDataSchemasData.of( //
                        MessageEvent.builder() //
                                .content("Have a great day!") //
                                .build())) //
                .event(Event.MESSAGE) //
                .build(), //
                list.get(2));
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-4") //
                .data(DifferentDataSchemasData.of( //
                        List.of(1L, 2L, 3L, 4L))) //
                .event(Event.ARRAY) //
                .build(), //
                list.get(3));
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-5") //
                .data(DifferentDataSchemasData.of(true)) //
                .event(Event.PRIMITIVE) //
                .build(), //
                list.get(4));
        assertEquals(DifferentDataSchemas.builder() //
                .id("event-6") //
                .data(DifferentDataSchemasData.of(3.14159f)) //
                .event(Event.PRIMITIVE) //
                .build(), //
                list.get(5));
    }

    @Test
    public void testEventStreamWithMixedData() throws Exception {
        recordTest("event-stream-mixed-data");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        var res = s.eventstreams().mixedData().call();
        List<MixedDataEvent> list = res.events().toList();
        assertEquals(4, list.size());

        // First event: JSON object (completion)
        assertEquals(MixedDataEventEvent.COMPLETION, list.get(0).event());
        assertTrue(list.get(0).data().messageEvent().isPresent());
        assertEquals("Hello world", list.get(0).data().messageEvent().get().content());

        // Second event: plain text
        assertEquals(MixedDataEventEvent.TEXT, list.get(1).event());
        assertTrue(list.get(1).data().string().isPresent());
        assertEquals("Processing your request...", list.get(1).data().string().get());

        // Third event: plain text
        assertEquals(MixedDataEventEvent.LOADING, list.get(2).event());
        assertTrue(list.get(2).data().string().isPresent());
        assertEquals("Almost done", list.get(2).data().string().get());

        // Fourth event: JSON object (completion)
        assertEquals(MixedDataEventEvent.COMPLETION, list.get(3).event());
        assertTrue(list.get(3).data().messageEvent().isPresent());
        assertEquals("Done!", list.get(3).data().messageEvent().get().content());
    }

    @Test
    public void testEventStreamErrorResponse() throws Exception {
        recordTest("event-stream-error-response");
        SpeakeasyHTTPClient defaultClient = new SpeakeasyHTTPClient();
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                try {
                    HttpRequest r = Helpers.copy(request) //
                            .header("x-teapot", "json") //
                            .build();
                    return defaultClient.send(r);
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }
        }).build();
        TeapotJSONError e = assertThrows(TeapotJSONError.class, () -> s.eventstreams().text().call());
        assertEquals("I'm a teapot", e.data().get().message().get());
        assertEquals("API error occurred", e.getMessage());
    }

    // ===== ASYNCHRONOUS/REACTIVE EVENT STREAM TESTS =====

    // Helper method to create SSE flux
    private Flux<JsonNode> createSSEFlux(String sseData, String terminalMessage) {
        byte[] data = sseData.getBytes(StandardCharsets.UTF_8);
        Blob blob = Blob.from(data);
        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                CompletableFuture.completedFuture(new MockAsyncResponse(blob)),
                jsonTypeReference, objectMapper, terminalMessage
        );
        return Flux.from(eventStream);
    }

    private Flux<JsonNode> createSSEFlux(String sseData) {
        return createSSEFlux(sseData, null);
    }

    // Helper method to create JSONL flux
    private Flux<JsonNode> createJsonLFlux(String jsonlData) {
        byte[] data = jsonlData.getBytes(StandardCharsets.UTF_8);
        Blob blob = Blob.from(data);
        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forJsonL(
                CompletableFuture.completedFuture(new MockAsyncResponse(blob)),
                jsonTypeReference, objectMapper
        );
        return Flux.from(eventStream);
    }

    // Parameterized test data providers
    static Stream<Arguments> basicSSEMessageData() {
        return Stream.of(
                Arguments.of(
                        "Basic message with event and id",
                        "event: test-event\nid: msg-001\ndata: {\"message\":\"hello\",\"value\":42}\n\n",
                        "test-event", "msg-001", null, "hello", 42
                ),
                Arguments.of(
                        "Message with retry field",
                        "event: retry-event\nid: retry-456\nretry: 1000\ndata: {\"message\":\"retry test\",\"value\":99}\n\n",
                        "retry-event", "retry-456", 1000, "retry test", 99
                ),
                Arguments.of(
                        "Simple data only message",
                        "data: {\"message\":\"simple\",\"value\":123}\n\n",
                        null, null, null, "simple", 123
                )
        );
    }

    static Stream<Arguments> dataTypeTestData() {
        return Stream.of(
                Arguments.of("String data", "data: \"hello world\"\n\n", "hello world"),
                Arguments.of("Number data", "data: 42\n\n", 42),
                Arguments.of("Boolean data", "data: true\n\n", true),
                Arguments.of("Null data", "data: null\n\n", null)
        );
    }

    static Stream<Arguments> dataTypeMalformedJson() {
        return Stream.of(
                Arguments.of("data: malformed json\n\n", "malformed json"),
                Arguments.of("data: {incomplete\n\n", "{incomplete"),
                Arguments.of("data: \"unclosed string\n\n", "\"unclosed string")
        );
    }

    @ParameterizedTest(name = "{0}")
    @MethodSource("basicSSEMessageData")
    void testSSEMessagesWithFields(String description, String sseData, String expectedEvent,
                                   String expectedId, Integer expectedRetry, String expectedMessage, int expectedValue) {
        Flux<JsonNode> flux = createSSEFlux(sseData);

        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    if (expectedEvent != null) {
                        assertThat(json.get("event").asText()).isEqualTo(expectedEvent);
                    }
                    if (expectedId != null) {
                        assertThat(json.get("id").asText()).isEqualTo(expectedId);
                    }
                    if (expectedRetry != null) {
                        assertThat(json.get("retry").asInt()).isEqualTo(expectedRetry);
                    }
                    assertThat(json.get("data").get("message").asText()).isEqualTo(expectedMessage);
                    assertThat(json.get("data").get("value").asInt()).isEqualTo(expectedValue);
                    return true;
                })
                .verifyComplete();
    }

    @ParameterizedTest(name = "{0}")
    @MethodSource("dataTypeTestData")
    void testDifferentDataTypes(String description, String sseData, Object expectedValue) {
        Flux<JsonNode> flux = createSSEFlux(sseData);

        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    JsonNode dataNode = json.get("data");
                    if (expectedValue == null) {
                        return dataNode.isNull();
                    } else if (expectedValue instanceof String) {
                        return expectedValue.equals(dataNode.asText());
                    } else if (expectedValue instanceof Integer) {
                        return expectedValue.equals(dataNode.asInt());
                    } else if (expectedValue instanceof Boolean) {
                        return expectedValue.equals(dataNode.asBoolean());
                    }
                    return false;
                })
                .verifyComplete();
    }

    @ParameterizedTest
    @MethodSource("dataTypeMalformedJson")
    void testMalformedJSONCoercedToString(String sseData, String expectedData) {
        Flux<JsonNode> flux = createSSEFlux(sseData);

        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    assertThat(json.get("data").asText()).isEqualTo(expectedData);
                    return true;
                })
                .verifyComplete();
    }

    @Test
    void testEventStreamMalformedFrameStrict() throws Exception {
        recordTest("event-stream-malformed-frame-strict");
        SDK s = SDK.builder().build();

        MalformedStreamResponse res = s.eventstreams().malformedStream().call();

        // The server emits one valid frame followed by a frame whose data is
        // missing the required `content` field. The typed JsonEvent model
        // enforces the schema, so the first frame decodes but iterating onto the
        // malformed frame surfaces the deserialization failure rather than
        // yielding a partial frame.
        try (EventStream<JsonEvent> events = res.events()) {
            JsonEvent first = events.next().orElseThrow();
            assertEquals("Hello", first.data().content());
            RuntimeException ex = assertThrows(RuntimeException.class, events::next);
            assertInstanceOf(JsonProcessingException.class, ex.getCause());
        }
    }

    @ParameterizedTest
    @ValueSource(strings = {
            ": This is a comment\ndata: {\"message\":\"test\",\"value\":1}\n\n",
            ": Comment line 1\n: Comment line 2\ndata: {\"message\":\"test\",\"value\":1}\n\n"
    })
    void testCommentsIgnored(String sseData) {
        Flux<JsonNode> flux = createSSEFlux(sseData);

        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    assertThat(json.get("data").get("message").asText()).isEqualTo("test");
                    assertThat(json.get("data").get("value").asInt()).isEqualTo(1);
                    return true;
                })
                .verifyComplete();
    }

    @Test
    void testMultipleMessages() {
        Flux<JsonNode> flux = createSSEFlux(
                "data: {\"message\":\"first\",\"value\":1}\n\n" +
                        "data: {\"message\":\"second\",\"value\":2}\n\n" +
                        "data: {\"message\":\"third\",\"value\":3}\n\n"
        );

        StepVerifier.create(flux)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 1)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 2)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 3)
                .verifyComplete();
    }

    @Test
    void testTerminalMessage() {
        Flux<JsonNode> flux = createSSEFlux(
                "data: {\"message\":\"first\",\"value\":1}\n\n" +
                        "data: [DONE]\n\n",
                "[DONE]"
        );

        StepVerifier.create(flux)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 1)
                .verifyComplete();
    }

    @Test
    void testBackpressure() {
        Flux<JsonNode> flux = createSSEFlux(
                "data: {\"message\":\"msg1\",\"value\":1}\n\n" +
                        "data: {\"message\":\"msg2\",\"value\":2}\n\n" +
                        "data: {\"message\":\"msg3\",\"value\":3}\n\n"
        );

        StepVerifier.create(flux, 0)
                .expectSubscription()
                .thenRequest(1)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 1)
                .thenRequest(1)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 2)
                .thenRequest(1)
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 3)
                .verifyComplete();
    }

    @Test
    void testCancellation() {
        ControllableStream stream = new ControllableStream();
        Blob blob = stream.asBlob();

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                CompletableFuture.completedFuture(new MockAsyncResponse(blob)),
                jsonTypeReference, objectMapper, null
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        StepVerifier.create(flux)
                .then(() -> stream.sendData("data: {\"message\":\"msg1\",\"value\":1}\n\n"))
                .expectNextMatches(json -> json.get("data").get("value").asInt() == 1)
                .then(() -> stream.sendData("data: {\"message\":\"msg2\",\"value\":2}\n\n"))
                .thenCancel()
                .verify();
    }

    @Test
    void testFailedCompletableFuture() {
        RuntimeException exception = new RuntimeException("Failed to get AsyncResponse");
        CompletableFuture<MockAsyncResponse> failedFuture = CompletableFuture.failedFuture(exception);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                failedFuture, jsonTypeReference, objectMapper, null
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        StepVerifier.create(flux)
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void testSlowCompletableFuture() {
        CompletableFuture<MockAsyncResponse> slowFuture = new CompletableFuture<>();

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                slowFuture, jsonTypeReference, objectMapper, null
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        StepVerifier.create(flux)
                .then(() -> {
                    byte[] data = "data: {\"message\":\"delayed\",\"value\":42}\n\n".getBytes(StandardCharsets.UTF_8);
                    Blob blob = Blob.from(data);
                    slowFuture.complete(new MockAsyncResponse(blob));
                })
                .expectNextMatches(json -> {
                    assertThat(json.get("data").get("message").asText()).isEqualTo("delayed");
                    assertThat(json.get("data").get("value").asInt()).isEqualTo(42);
                    return true;
                })
                .verifyComplete();
    }

    @Test
    void testHttpAttributes() {
        byte[] data = "data: {\"message\":\"test\",\"value\":42}\n\n".getBytes(StandardCharsets.UTF_8);
        Blob blob = Blob.from(data);

        Map<String, List<String>> headers = Map.of(
                "Content-Type", List.of("text/event-stream; charset=utf-8"),
                "Cache-Control", List.of("no-cache")
        );
        MockAsyncResponse mockResponse = new MockAsyncResponse(blob, 200, headers);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                CompletableFuture.completedFuture(mockResponse), jsonTypeReference, objectMapper, null
        );

        // Test HTTP attribute methods using Helpers utility
        assertEventStreamHttpAttributes(eventStream);

        // Verify SSE functionality still works
        Flux<JsonNode> flux = Flux.from(eventStream);
        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    assertThat(json.get("data").get("message").asText()).isEqualTo("test");
                    assertThat(json.get("data").get("value").asInt()).isEqualTo(42);
                    return true;
                })
                .verifyComplete();
    }

    // ===== JSONL EVENT STREAM TESTS =====

    @Test
    void testJsonLBasicParsing() {
        Flux<JsonNode> flux = createJsonLFlux(
                "{\"message\":\"hello\",\"value\":1}\n" +
                        "{\"message\":\"world\",\"value\":2}\n"
        );

        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    assertThat(json.get("message").asText()).isEqualTo("hello");
                    assertThat(json.get("value").asInt()).isEqualTo(1);
                    return true;
                })
                .expectNextMatches(json -> {
                    assertThat(json.get("message").asText()).isEqualTo("world");
                    assertThat(json.get("value").asInt()).isEqualTo(2);
                    return true;
                })
                .verifyComplete();
    }

    @Test
    void testJsonLWithEmptyLines() {
        Flux<JsonNode> flux = createJsonLFlux(
                "{\"message\":\"first\"}\n\n{\"message\":\"second\"}\n"
        );

        StepVerifier.create(flux)
                .expectNextMatches(json -> json.get("message").asText().equals("first"))
                .expectNextMatches(json -> json.get("message").asText().equals("second"))
                .verifyComplete();
    }

    @Test
    void testJsonLBackpressure() {
        Flux<JsonNode> flux = createJsonLFlux(
                "{\"value\":1}\n{\"value\":2}\n{\"value\":3}\n"
        );

        StepVerifier.create(flux, 0)
                .expectSubscription()
                .thenRequest(1)
                .expectNextMatches(json -> json.get("value").asInt() == 1)
                .thenRequest(1)
                .expectNextMatches(json -> json.get("value").asInt() == 2)
                .thenRequest(1)
                .expectNextMatches(json -> json.get("value").asInt() == 3)
                .verifyComplete();
    }

    @Test
    void testJsonLCancellation() {
        ControllableStream stream = new ControllableStream();
        Blob blob = stream.asBlob();

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forJsonL(
                CompletableFuture.completedFuture(new MockAsyncResponse(blob)),
                jsonTypeReference, objectMapper
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        StepVerifier.create(flux)
                .then(() -> stream.sendData("{\"message\":\"msg1\"}\n"))
                .expectNextMatches(json -> json.get("message").asText().equals("msg1"))
                .then(() -> stream.sendData("{\"message\":\"msg2\"}\n"))
                .thenCancel()
                .verify();
    }

    @Test
    void testJsonLFailedCompletableFuture() {
        RuntimeException exception = new RuntimeException("Failed to get AsyncResponse");
        CompletableFuture<MockAsyncResponse> failedFuture = CompletableFuture.failedFuture(exception);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forJsonL(
                failedFuture, jsonTypeReference, objectMapper
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        StepVerifier.create(flux)
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void testImmediateErrorSignalingWithoutDemand() {
        RuntimeException exception = new RuntimeException("HTTP request failed");
        CompletableFuture<MockAsyncResponse> failedFuture = CompletableFuture.failedFuture(exception);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                failedFuture, jsonTypeReference, objectMapper, null
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        // Test that error is signaled immediately without requesting any items
        StepVerifier.create(flux, 0)
                .expectSubscription()
                // Don't request any items - error should still be signaled
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void testImmediateErrorSignalingVsDelayedRequest() {
        RuntimeException exception = new RuntimeException("HTTP request failed");
        CompletableFuture<MockAsyncResponse> failedFuture = CompletableFuture.failedFuture(exception);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forSSE(
                failedFuture, jsonTypeReference, objectMapper, null
        );

        Flux<JsonNode> flux = Flux.from(eventStream);

        // Test that error is signaled immediately even when request comes later
        StepVerifier.create(flux, 0)
                .expectSubscription()
                .then(() -> {
                    // Simulate a delay before requesting - error should already be signaled
                    try {
                        Thread.sleep(100);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                })
                .thenRequest(1)
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void testJsonLHttpAttributes() {
        byte[] data = "{\"message\":\"test\"}\n".getBytes(StandardCharsets.UTF_8);
        Blob blob = Blob.from(data);

        Map<String, List<String>> headers = Map.of(
                "Content-Type", List.of("application/x-ndjson"),
                "Cache-Control", List.of("no-cache")
        );
        MockAsyncResponse mockResponse = new MockAsyncResponse(blob, 200, headers);

        org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream = org.openapis.openapi.utils.reactive.EventStream.forJsonL(
                CompletableFuture.completedFuture(mockResponse), jsonTypeReference, objectMapper
        );

        // Test HTTP attribute methods
        assertEventStreamHttpAttributes(eventStream);

        // Verify JSONL functionality still works
        Flux<JsonNode> flux = Flux.from(eventStream);
        StepVerifier.create(flux)
                .expectNextMatches(json -> {
                    assertThat(json.get("message").asText()).isEqualTo("test");
                    return true;
                })
                .verifyComplete();
    }

    // Helper method for testing HTTP attributes
    private void assertEventStreamHttpAttributes(org.openapis.openapi.utils.reactive.EventStream<MockAsyncResponse, JsonNode> eventStream) {
        // Test contentType
        assertThat(eventStream.contentType())
                .succeedsWithin(java.time.Duration.ofSeconds(1))
                .satisfies(contentType -> assertThat(contentType).isNotEmpty());

        // Test statusCode
        assertThat(eventStream.statusCode())
                .succeedsWithin(java.time.Duration.ofSeconds(1))
                .satisfies(statusCode -> assertThat(statusCode).isEqualTo(200));

        // Test rawResponse
        assertThat(eventStream.rawResponse())
                .succeedsWithin(java.time.Duration.ofSeconds(1))
                .satisfies(response -> assertThat(response).isNotNull());

        // Test body
        assertThat(eventStream.body())
                .succeedsWithin(java.time.Duration.ofSeconds(1))
                .satisfies(body -> assertThat(body).isNotNull());
    }

    @Test
    public void testEventStreamWithAbortSignal() throws Exception {
        recordTest("event-stream-with-abort-signal");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        // Create a CompletableFuture that we can cancel to simulate abort
        CompletableFuture<Void> abortFuture = new CompletableFuture<>();

        // Create a custom HTTP client that respects cancellation
        SpeakeasyHTTPClient defaultClient = new SpeakeasyHTTPClient();
        SDK abortableSDK = SDK.builder().serverURL(HTTPBIN_URL).client(new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) throws IOException, InterruptedException, URISyntaxException {
                if (abortFuture.isCancelled()) {
                    throw new CancellationException("Request aborted");
                }
                return defaultClient.send(request);
            }
        }).build();

        TextResponse res = abortableSDK.eventstreams().text().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        try (EventStream<TextEvent> events = res.events()) {
            int chunks = 0;
            StringBuilder message = new StringBuilder();
            boolean caughtAbortError = false;

            try {
                for (TextEvent event : events) {
                    if (chunks == 1) {
                        // Cancel after first chunk
                        abortFuture.cancel(true);
                        events.close(); // Close the stream to trigger cancellation
                        break;
                    }
                    message.append(event.data());
                    chunks++;
                }
            } catch (Exception e) {
                caughtAbortError = e instanceof CancellationException ||
                                  e.getCause() instanceof CancellationException ||
                                  e.getMessage() != null && e.getMessage().contains("abort");
            }

            // In Java, we may not always get the exception when closing early
            // but we should have only received one chunk
            assertTrue(chunks <= 2, "Should have received at most 2 chunks");
            assertTrue(message.toString().startsWith("Hello"), "Should have received partial message");
        }
    }

    @Test
    public void testEventStreamStayOpenBreakingEarlyExitsStream() throws Exception {
        recordTest("event-stream-stay-open-break-early");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        StayOpenResponse res = s.eventstreams().stayOpen().call();

        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<String> events = new ArrayList<>();
        EventStream<TextEvent> streamRef;
        try (EventStream<TextEvent> stream = res.events()) {
            streamRef = stream;
            for (TextEvent event : stream) {
                events.add(event.data());
                break;
            }

            assertEquals(1, events.size());

            // In Java, `break` does not automatically close the stream.
            assertFalse(stream.isClosed(), "Stream should still be open after break");
        }
        // Stream is properly closed by try-with-resources
        assertTrue(streamRef.isClosed(), "Stream should be closed after try-with-resources");
    }

    @Test
    public void testEventStreamStayOpenSentinelDetectionClosesStream() throws Exception {
        recordTest("event-stream-stay-open-sentinel");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        StayOpenResponse res = s.eventstreams().stayOpen().call();

        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<String> events = new ArrayList<>();
        try (EventStream<TextEvent> stream = res.events()) {
            for (TextEvent event : stream) {
                events.add(event.data());
            }
        }

        // Should receive all events until sentinel
        assertEquals(List.of("event 1", "event 2", "event 3", "event 4"), events);
    }

    @Test
    public void testEventStreamPartialWithComments() throws Exception {
        recordTest("event-stream-partial-with-comments");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        PartialWithCommentsResponse res = s.eventstreams().partialWithComments().call();

        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<PartialWithCommentsEvent> actual = new ArrayList<>();
        try (EventStream<PartialWithCommentsEvent> events = res.events()) {
            for (PartialWithCommentsEvent event : events) {
                actual.add(event);
            }
        }

        // Build expected events
        List<PartialWithCommentsEvent> expected = new ArrayList<>();

        // First event - only data
        expected.add(PartialWithCommentsEvent.builder()
                .data(PartialWithCommentsEventData.builder()
                    .message("Hello from SSE")
                    .build())
                .build());

        // Second event - with id and event type
        expected.add(PartialWithCommentsEvent.builder()
                .id("msg-2")
                .event("update")
                .data(PartialWithCommentsEventData.builder()
                    .status("processing")
                    .progress(50L)
                    .build())
                .build());

        // Third event - completion with result
        expected.add(PartialWithCommentsEvent.builder()
                .id("msg-3")
                .data(PartialWithCommentsEventData.builder()
                    .status("complete")
                    .progress(100L)
                    .result("Success")
                    .build())
                .build());

        // Fourth event - mixed
        expected.add(PartialWithCommentsEvent.builder()
                .id("msg-4")
                .event("mixed")
                .data(PartialWithCommentsEventData.builder()
                    .test("mixed boundaries")
                    .build())
                .build());

        // Fifth event
        expected.add(PartialWithCommentsEvent.builder()
                .id("msg-5")
                .data(PartialWithCommentsEventData.builder()
                    .another("test")
                    .build())
                .build());

        assertEquals(expected, actual);
    }

    @Test
    public void testEventStreamWPTCompliance() throws Exception {
        recordTest("event-stream-wpt-compliance");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        WptComplianceResponse res = s.eventstreams().wptCompliance().call();

        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<SseEvent> actual = new ArrayList<>();
        String expectedData = null;
        try (EventStream<SseEvent> events = res.events()) {
            for (SseEvent event : events) {
                if (event.event().isPresent() && "expected".equals(event.event().get())) {
                    expectedData = event.data();
                } else {
                    actual.add(event);
                }
            }
        }

        ObjectMapper mapper = new ObjectMapper();
        List<SseEvent> expected = mapper.readValue(expectedData,
            mapper.getTypeFactory().constructCollectionType(List.class, SseEvent.class));

        assertTrue(expected.size() > 0, "No expectations received from server");
        assertEquals(expected, actual);
    }

    @Test
    public void testEventStreamOptionalDataField() throws Exception {
        recordTest("event-stream-optional-data-field");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        OptionalDataResponse res = s.eventstreams().optionalData().call();

        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<OptionalDataEvent> actual = new ArrayList<>();
        try (EventStream<OptionalDataEvent> events = res.events()) {
            for (OptionalDataEvent event : events) {
                actual.add(event);
            }
        }

        List<OptionalDataEvent> expected = List.of(
            OptionalDataEvent.builder()
                .event("message")
                .id("event-1")
                .data(OptionalDataEventPayload.builder()
                    .content("Hello, this event has data")
                    .build())
                .build(),
            OptionalDataEvent.builder()
                .event("heartbeat")
                .id("event-2")
                .build(),
            OptionalDataEvent.builder()
                .event("message")
                .id("event-3")
                .data(OptionalDataEventPayload.builder()
                    .content("Another message with data")
                    .build())
                .build(),
            OptionalDataEvent.builder()
                .event("ping")
                .id("event-4")
                .build(),
            OptionalDataEvent.builder()
                .event("complete")
                .id("event-5")
                .data(OptionalDataEventPayload.builder()
                    .content("Stream finished")
                    .build())
                .build()
        );

        assertEquals(expected, actual);
    }

    @Test
    public void testEventStreamChatHeartbeatSkipsDataless() throws Exception {
        recordTest("event-stream-chat-heartbeat-skips-dataless");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        ChatHeartbeatResponse res = s.eventstreams().chatHeartbeat().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        // Server sends: data("Hello1"), heartbeat (skipped), data("Hello 2"), data("!"), [DONE]
        try (EventStream<ChatCompletionEvent> events = res.events()) {
            List<String> list = events.stream().map(x -> x.data().content()).collect(Collectors.toList());
            assertEquals(List.of("Hello1", "Hello 2", "!"), list);
        }
    }

    @Test
    public void testEventStreamLargeEventSmallChunks() throws Exception {
        recordTest("event-stream-large-event-small-chunks");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        long size = 2L * 1024 * 1024;
        LargeEventSmallChunksResponse res = s.eventstreams().largeEventSmallChunks().size(size).call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        try (EventStream<LargeChunkedEvent> events = res.events()) {
            List<LargeChunkedEvent> list = events.stream().collect(Collectors.toList());
            assertEquals(3, list.size());
            assertEquals("start", list.get(0).data().content());
            assertEquals("end", list.get(2).data().content());
            String big = list.get(1).data().content();
            assertEquals((int) size, big.length());
            assertEquals('S', big.charAt(0));
            assertEquals('E', big.charAt(big.length() - 1));
            assertEquals("a".repeat((int) size - 2), big.substring(1, big.length() - 1));
        }
    }

    @Test
    public void testEventStreamSplitBoundaries() throws Exception {
        recordTest("event-stream-split-boundaries");
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        SplitBoundariesResponse res = s.eventstreams().splitBoundaries().call();
        assertEquals("text/event-stream", res.contentType());
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        List<String> actualTags = new ArrayList<>();
        List<String> expectedTags = new ArrayList<>();
        try (EventStream<SplitBoundaryEvent> events = res.events()) {
            List<SplitBoundaryEvent> list = events.stream().collect(Collectors.toList());
            for (SplitBoundaryEvent ev : list) {
                if ("expected".equals(ev.data().kind().value())) {
                    expectedTags = ev.data().tags();
                } else {
                    actualTags.addAll(ev.data().tags());
                }
            }
        }

        assertTrue(expectedTags.size() > 0);
        assertEquals(expectedTags, actualTags);
    }

}
