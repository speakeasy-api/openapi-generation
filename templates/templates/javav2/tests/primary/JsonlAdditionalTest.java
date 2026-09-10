package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.CommonHelpers.recordTest;

import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import com.fasterxml.jackson.core.type.TypeReference;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.openapi.models.operations.JsonlStreamResponse;
import org.openapis.openapi.models.operations.JsonlStreamResponseBody;
import org.openapis.openapi.models.operations.JsonlStreamChunksResponse;
import org.openapis.openapi.models.operations.JsonlStreamChunksResponseBody;
import org.openapis.openapi.models.operations.JsonlDeserializationVerificationResponse;
import org.openapis.openapi.models.operations.JsonlDeserializationVerificationResponseBody;
import org.openapis.openapi.utils.JsonLStream;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.openapi.utils.Utils;

public class JsonlAdditionalTest {
    @Test
    public void testJsonlStreamData() throws Exception {
        recordTest("jsonl-stream-data-envelope-http-responses");
        SpeakeasyHTTPClient.setDebugLogging(true);
        try {
            SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
            JsonlStreamResponse res = s.jsonl().jsonlStream().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

            try (JsonLStream<JsonlStreamResponseBody> events = res.events()) {
                List<JsonlStreamResponseBody> list = events.stream().collect(Collectors.toList());
                assertEquals(2, list.size());

                // Assert first event
                JsonlStreamResponseBody firstEvent = list.get(0);
                assertEquals("Peter", firstEvent.name().get());
                assertEquals(List.of("Go", "Python"), firstEvent.skills().get());

                // Assert second event
                JsonlStreamResponseBody secondEvent = list.get(1);
                assertEquals("John", secondEvent.name().get());
                assertEquals(List.of("Go", "Rust"), secondEvent.skills().get());
            }
        } finally {
            SpeakeasyHTTPClient.setDebugLogging(false);
        }
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    public void testJsonlStreamDataNotBufferedWithDebugLogging() throws Exception {
        SpeakeasyHTTPClient.setDebugLogging(true);
        try {
            SpeakeasyHTTPClient client = new SpeakeasyHTTPClient();
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(Helpers.API_TEST_SERVICE_URL + "/jsonl/hold"))
                    .GET()
                    .build();
            long startTime = System.currentTimeMillis();
            HttpResponse<InputStream> response = client.send(request);
            assertEquals(HttpURLConnection.HTTP_OK, response.statusCode());

            try (JsonLStream<JsonlStreamResponseBody> events = new JsonLStream<>(
                    response.body(), new TypeReference<JsonlStreamResponseBody>() {}, Utils.mapper())) {
                Optional<JsonlStreamResponseBody> firstEventOpt = events.next();
                long firstEventTime = System.currentTimeMillis() - startTime;
                assertTrue(firstEventOpt.isPresent(), "First event missing or unparseable");
                // The endpoint writes both events immediately and then holds the
                // connection open for 10s. A client that buffers the body to EOF
                // (e.g. for debug logging) cannot deliver the first event before
                // the hold expires, so this bound fails under buffered delivery
                // without depending on tight timing.
                assertTrue(firstEventTime < 5000,
                        "First event took " + firstEventTime + "ms; response body was buffered to EOF");
                JsonlStreamResponseBody firstEvent = firstEventOpt.get();
                assertEquals("Peter", firstEvent.name().get());
                assertEquals(List.of("Go", "Python"), firstEvent.skills().get());

                Optional<JsonlStreamResponseBody> secondEventOpt = events.next();
                assertTrue(secondEventOpt.isPresent(), "Second event missing or unparseable");
                JsonlStreamResponseBody secondEvent = secondEventOpt.get();
                assertEquals("John", secondEvent.name().get());
                assertEquals(List.of("Go", "Rust"), secondEvent.skills().get());
            }
        } finally {
            SpeakeasyHTTPClient.setDebugLogging(false);
        }
    }

    @Test
    public void testJsonlStreamDataChunks() throws Exception {
        recordTest("jsonl-stream-data-chunks-envelope-http-responses");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        
        long startTime = System.currentTimeMillis();
        JsonlStreamChunksResponse res = s.jsonl().jsonlStreamChunks().call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        try (JsonLStream<JsonlStreamChunksResponseBody> events = res.events()) {
            // Get first event
            Optional<JsonlStreamChunksResponseBody> firstEventOpt = events.next();
            assertTrue(firstEventOpt.isPresent(), "First event should be present");
            long firstEventTime = System.currentTimeMillis() - startTime;
            
            // First event should take longer than 100ms because the server sends a chunk and then waits
            // Using more tolerant bounds to account for CI environment variability
            assertTrue(firstEventTime >= 80, "First event should take at least 80ms");
            // The upper bound is a liveness check only; it does not distinguish incremental delivery
            // from buffered delivery (the full response completes well within it).
            assertTrue(firstEventTime < 1000, "First event should take less than 1000ms");            

            // Assert first event
            JsonlStreamChunksResponseBody firstEvent = firstEventOpt.get();
            assertEquals("Peter", firstEvent.name().get());
            assertEquals(List.of("Go", "Python"), firstEvent.skills().get());

            // Get second event
            Optional<JsonlStreamChunksResponseBody> secondEventOpt = events.next();
            assertTrue(secondEventOpt.isPresent(), "Second event should be present");

            // Assert second event
            JsonlStreamChunksResponseBody secondEvent = secondEventOpt.get();
            assertEquals("John", secondEvent.name().get());
            assertEquals(List.of("Go", "Rust"), secondEvent.skills().get());
        }
    }

    @Test
    public void testJsonlDeserializationWithCamelCaseProperties() throws Exception {
        recordTest("jsonl-deserialization-camel-case-properties");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        JsonlDeserializationVerificationResponse res = s.jsonl().jsonlDeserializationVerification().call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        try (JsonLStream<JsonlDeserializationVerificationResponseBody> events = res.events()) {
            List<JsonlDeserializationVerificationResponseBody> list = events.stream().collect(Collectors.toList());
            assertEquals(1, list.size());

            JsonlDeserializationVerificationResponseBody event = list.get(0);
            assertEquals("yes", event.isFinished());
        }
    }
}
