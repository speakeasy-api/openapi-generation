package org.openapis.quaternary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.quaternary.openapi.CommonHelpers.recordTest;

import java.util.List;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.quaternary.openapi.models.operations.JsonlStreamChunksResponseBody;
import org.openapis.quaternary.openapi.models.operations.JsonlStreamResponseBody;
import org.openapis.quaternary.openapi.models.operations.async.JsonlStreamChunksResponse;
import org.openapis.quaternary.openapi.models.operations.async.JsonlStreamResponse;
import org.openapis.quaternary.openapi.utils.reactive.EventStream;
import reactor.core.publisher.Flux;
import reactor.test.StepVerifier;

public class JsonlAdditionalTest {
    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    public void testJsonlStreamDataAsync() {
        recordTest("jsonl-stream-data-async-flat-response");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        EventStream<JsonlStreamResponse, JsonlStreamResponseBody> events = s.toAsync().jsonl().jsonlStreamDirect();

        StepVerifier.create(Flux.from(events).collectList())
                .assertNext(list -> {
                    assertEquals(2, list.size());

                    JsonlStreamResponseBody firstEvent = list.get(0);
                    assertEquals("Peter", firstEvent.name());
                    assertEquals(List.of("Go", "Python"), firstEvent.skills());

                    JsonlStreamResponseBody secondEvent = list.get(1);
                    assertEquals("John", secondEvent.name());
                    assertEquals(List.of("Go", "Rust"), secondEvent.skills());
                })
                .verifyComplete();
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    public void testJsonlStreamDataChunksAsync() {
        recordTest("jsonl-stream-data-async-chunks-flat-response");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        long startTime = System.currentTimeMillis();
        EventStream<JsonlStreamChunksResponse, JsonlStreamChunksResponseBody> events =
                s.toAsync().jsonl().jsonlStreamChunksDirect();

        StepVerifier.create(Flux.from(events).take(2))
                .assertNext(firstEvent -> {
                    long firstEventTime = System.currentTimeMillis() - startTime;
                    assertTrue(firstEventTime >= 80, "First event should take at least 80ms");
                    assertEquals("Peter", firstEvent.name());
                    assertEquals(List.of("Go", "Python"), firstEvent.skills());
                })
                .assertNext(secondEvent -> {
                    assertEquals("John", secondEvent.name());
                    assertEquals(List.of("Go", "Rust"), secondEvent.skills());
                })
                .verifyComplete();
    }
}
