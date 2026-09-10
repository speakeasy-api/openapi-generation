package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.tertiary.openapi.Helpers.HTTPBIN_URL;

import java.nio.charset.StandardCharsets;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.tertiary.openapi.models.operations.RetriesPostBinaryRequestBodyResponse;
import org.openapis.tertiary.openapi.utils.Blob;

/**
 * SDK-level retry replay of a repeatable binary request body against the
 * retriesPostBinaryRequestBody operation (503 until the num-retries'th
 * attempt, then echoes the body back).
 */
public class RetriesAdditionalTest {

    private static final byte[] PAYLOAD = "binary-retry-payload-éè"
            .getBytes(StandardCharsets.UTF_8);

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testSyncRetryReplaysRepeatableBinaryBody() throws Exception {
        CommonHelpers.recordTest("retries-binary-request-body");

        // in-memory bytes are repeatable: the retried attempt must resend the
        // full body, and the server echo proves it arrived intact
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        RetriesPostBinaryRequestBodyResponse res = s.retries()
                .retriesPostBinaryRequestBody()
                .requestBody(Blob.from(PAYLOAD))
                .requestId("sync-replay-" + UUID.randomUUID())
                .numRetries(2)
                .call();

        assertEquals(200, res.statusCode());
        assertTrue(res.retriesBody().isPresent());
        assertEquals(2, res.retriesBody().get().retries().get());
        assertEquals(new String(PAYLOAD, StandardCharsets.UTF_8),
                res.retriesBody().get().body().get());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncRetryReplaysRepeatableBinaryBody() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        org.openapis.tertiary.openapi.models.operations.async.RetriesPostBinaryRequestBodyResponse res =
                s.toAsync().retries()
                        .retriesPostBinaryRequestBody()
                        .requestBody(Blob.from(PAYLOAD))
                        .requestId("async-replay-" + UUID.randomUUID())
                        .numRetries(2)
                        .call()
                        .get(10, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertTrue(res.retriesBody().isPresent());
        assertEquals(2, res.retriesBody().get().retries().get());
        assertEquals(new String(PAYLOAD, StandardCharsets.UTF_8),
                res.retriesBody().get().body().get());
    }

}
