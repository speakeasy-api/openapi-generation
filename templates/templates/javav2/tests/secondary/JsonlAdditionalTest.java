package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.net.HttpURLConnection;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

import org.junit.jupiter.api.Test;
import org.openapis.secondary.openapi.models.operations.*;
import org.openapis.secondary.openapi.utils.Utils;
import org.openapis.secondary.openapi.utils.JsonLStream;
import org.openapis.secondary.openapi.utils.SpeakeasyHTTPClient;

public class JsonlAdditionalTest {
    @Test
    public void testJsonlStreamData() throws Exception {
        Utils.recordTest("jsonl-stream-data-flat-responses");
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
                assertEquals("Peter", firstEvent.name());
                assertEquals(List.of("Go", "Python"), firstEvent.skills());

                // Assert second event
                JsonlStreamResponseBody secondEvent = list.get(1);
                assertEquals("John", secondEvent.name());
                assertEquals(List.of("Go", "Rust"), secondEvent.skills());
            }
        } finally {
            SpeakeasyHTTPClient.setDebugLogging(false);
        }
    }

    @Test
    public void testJsonlStreamDataWithXNdjson() throws Exception {
        Utils.recordTest("x-ndjson-stream-data-envelope-http-responses");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        XNdjsonStreamResponse res = s.jsonl().xNdjsonStream().call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        try (JsonLStream<XNdjsonStreamResponseBody> events = res.events()) {
            List<XNdjsonStreamResponseBody> list = events.stream().collect(Collectors.toList());
            assertEquals(2, list.size());

            // Assert first event
            XNdjsonStreamResponseBody firstEvent = list.get(0);
            assertEquals("Peter", firstEvent.name());
            assertEquals(List.of("Go", "Python"), firstEvent.skills());

            // Assert second event
            XNdjsonStreamResponseBody secondEvent = list.get(1);
            assertEquals("John", secondEvent.name());
            assertEquals(List.of("Go", "Rust"), secondEvent.skills());
        }
    }

    @Test
    public void testJsonlStreamDataChunks() throws Exception {
        Utils.recordTest("jsonl-stream-data-chunks-flat-response");
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
            assertEquals("Peter", firstEvent.name());
            assertEquals(List.of("Go", "Python"), firstEvent.skills());

            // Get second event
            Optional<JsonlStreamChunksResponseBody> secondEventOpt = events.next();
            assertTrue(secondEventOpt.isPresent(), "Second event should be present");

            // Assert second event
            JsonlStreamChunksResponseBody secondEvent = secondEventOpt.get();
            assertEquals("John", secondEvent.name());
            assertEquals(List.of("Go", "Rust"), secondEvent.skills());
        }
    }

    @Test
    public void testJsonlStreamDataChunksWithXNdjson() throws Exception {
        Utils.recordTest("x-ndjson-stream-data-chunks-envelope-http-responses");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        
        long startTime = System.currentTimeMillis();
        XNdjsonStreamChunksResponse res = s.jsonl().xNdjsonStreamChunks().call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());

        try (JsonLStream<XNdjsonStreamChunksResponseBody> events = res.events()) {
            // Get first event
            Optional<XNdjsonStreamChunksResponseBody> firstEventOpt = events.next();
            assertTrue(firstEventOpt.isPresent(), "First event should be present");
            long firstEventTime = System.currentTimeMillis() - startTime;
            
            // First event should take longer than 100ms because the server sends a chunk and then waits
            // Using more tolerant bounds to account for CI environment variability
            assertTrue(firstEventTime >= 80, "First event should take at least 80ms");
            // The upper bound is a liveness check only; it does not distinguish incremental delivery
            // from buffered delivery (the full response completes well within it).
            assertTrue(firstEventTime < 1000, "First event should take less than 1000ms");            

            // Assert first event
            XNdjsonStreamChunksResponseBody firstEvent = firstEventOpt.get();
            assertEquals("Peter", firstEvent.name());
            assertEquals(List.of("Go", "Python"), firstEvent.skills());

            // Get second event
            Optional<XNdjsonStreamChunksResponseBody> secondEventOpt = events.next();
            assertTrue(secondEventOpt.isPresent(), "Second event should be present");

            // Assert second event
            XNdjsonStreamChunksResponseBody secondEvent = secondEventOpt.get();
            assertEquals("John", secondEvent.name());
            assertEquals(List.of("Go", "Rust"), secondEvent.skills());
        }
    }
}
