package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;

import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.openapis.tertiary.openapi.utils.transport.HttpBody;

/**
 * Unit-level contract for the SDK-owned request-body carrier. End-to-end retry
 * replay is covered by tertiary AsyncAdditionalTest/RetriesAdditionalTest; this
 * pins the repeatable-vs-one-shot semantics those e2e tests assume correct.
 */
public class HttpBodyTest {

    private static byte[] drain(HttpBody body) throws Exception {
        try (InputStream in = body.stream()) {
            return IOUtils.toByteArray(in);
        }
    }

    @Test
    public void testOfBytesIsRepeatable() throws Exception {
        byte[] bytes = "payload".getBytes(StandardCharsets.UTF_8);
        HttpBody body = HttpBody.of(bytes);
        assertTrue(body.isRepeatable());
        assertEquals(bytes.length, body.contentLength());
        assertArrayEquals(bytes, drain(body));
        assertArrayEquals(bytes, drain(body));
    }

    @Test
    public void testOfStringIsRepeatable() throws Exception {
        HttpBody body = HttpBody.of("héllo");
        byte[] expected = "héllo".getBytes(StandardCharsets.UTF_8);
        assertTrue(body.isRepeatable());
        assertArrayEquals(expected, drain(body));
        assertArrayEquals(expected, drain(body));
    }

    @Test
    public void testEmptyIsRepeatableAndZeroLength() throws Exception {
        HttpBody body = HttpBody.empty();
        assertTrue(body.isRepeatable());
        assertEquals(0, body.contentLength());
        assertEquals(0, drain(body).length);
    }

    @Test
    public void testOfFileIsRepeatable() throws Exception {
        Path path = Files.createTempFile("httpbody", ".bin");
        try {
            byte[] bytes = "file-content".getBytes(StandardCharsets.UTF_8);
            Files.write(path, bytes);
            HttpBody body = HttpBody.ofFile(path);
            assertTrue(body.isRepeatable());
            assertEquals(bytes.length, body.contentLength());
            assertArrayEquals(bytes, drain(body));
            assertArrayEquals(bytes, drain(body));
        } finally {
            Files.deleteIfExists(path);
        }
    }

    @Test
    public void testOfInputStreamReplayReturnsConsumedStream() throws Exception {
        byte[] bytes = "stream".getBytes(StandardCharsets.UTF_8);
        HttpBody body = HttpBody.ofInputStream(new ByteArrayInputStream(bytes), bytes.length);
        assertFalse(body.isRepeatable());
        assertEquals(bytes.length, body.contentLength());
        assertArrayEquals(bytes, drain(body));
        assertEquals(0, drain(body).length);
    }

    @Test
    public void testOfInputStreamUnknownLength() {
        HttpBody body = HttpBody.ofInputStream(new ByteArrayInputStream(new byte[0]));
        assertEquals(-1, body.contentLength());
    }

    @Test
    public void testOfInputStreamRejectsInvalidContentLength() {
        assertThrows(
                IllegalArgumentException.class,
                () -> HttpBody.ofInputStream(new ByteArrayInputStream(new byte[0]), -2));
    }
}
