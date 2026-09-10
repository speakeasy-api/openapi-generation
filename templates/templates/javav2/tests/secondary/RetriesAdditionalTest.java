package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import java.nio.charset.StandardCharsets;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.RetriesPostBinaryRequestBodyResponse;
import org.openapis.secondary.openapi.utils.Blob;

/**
 * SDK-level retry replay of a repeatable binary request body against the
 * retriesPostBinaryRequestBody operation (503 until the num-retries'th
 * attempt, then echoes the body back). Adapted for the
 * java11 + okhttp arm with the "raw" getter style (response body and its
 * fields are {@code @Nullable} raw values, not {@code Optional}).
 */
public class RetriesAdditionalTest {

    private static final byte[] PAYLOAD = "binary-retry-payload-éè"
            .getBytes(StandardCharsets.UTF_8);

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testSyncRetryReplaysRepeatableBinaryBody() throws Exception {
        // in-memory bytes are repeatable: the retried attempt must resend the
        // full body, and the server echo proves it arrived intact
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        RetriesPostBinaryRequestBodyResponse res = s.retries()
                .retriesPostBinaryRequestBody()
                .requestBody(Blob.from(PAYLOAD))
                .requestId("sync-replay-" + UUID.randomUUID())
                .numRetries(2L)
                .call();

        assertEquals(200, res.statusCode());
        assertNotNull(res.retriesBody());
        assertEquals(2L, res.retriesBody().retries());
        assertEquals(new String(PAYLOAD, StandardCharsets.UTF_8),
                res.retriesBody().body());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testAsyncRetryReplaysRepeatableBinaryBody() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        org.openapis.secondary.openapi.models.operations.async.RetriesPostBinaryRequestBodyResponse res =
                s.toAsync().retries()
                        .retriesPostBinaryRequestBody()
                        .requestBody(Blob.from(PAYLOAD))
                        .requestId("async-replay-" + UUID.randomUUID())
                        .numRetries(2L)
                        .call()
                        .get(10, TimeUnit.SECONDS);

        assertEquals(200, res.statusCode());
        assertNotNull(res.retriesBody());
        assertEquals(2L, res.retriesBody().retries());
        assertEquals(new String(PAYLOAD, StandardCharsets.UTF_8),
                res.retriesBody().body());
    }

}
