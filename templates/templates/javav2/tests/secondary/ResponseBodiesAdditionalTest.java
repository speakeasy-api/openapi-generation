package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import com.fasterxml.jackson.core.type.TypeReference;
import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.ResponseBodyBytesGetResponse;
import org.openapis.secondary.openapi.models.operations.ResponseBodyStringGetResponse;
import org.openapis.secondary.openapi.utils.Headers;
import org.openapis.secondary.openapi.utils.Utils;
import org.openapis.secondary.openapi.utils.transport.HttpRequest;
import org.openapis.secondary.openapi.utils.transport.HttpResponse;

import java.io.IOException;
import java.io.InputStream;
import java.io.InterruptedIOException;
import java.io.UncheckedIOException;
import java.net.URI;
import java.util.Map;
import java.util.concurrent.TimeUnit;

/**
 * Sync streaming response-body reads over the java11 + okhttp transport. The
 * body is exposed as a plain {@code InputStream} (raw getter style, so it is a
 * {@code @Nullable} value, not {@code Optional}); this exercises the okhttp
 * {@code send()} body-pump path that the async tests do not cover on the
 * blocking arm.
 */
public class ResponseBodiesAdditionalTest {

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testResponseBodyBytesGet() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        ResponseBodyBytesGetResponse res = s.responseBodies().responseBodyBytesGet().call();

        assertEquals(200, res.statusCode());
        assertNotNull(res.bytes());

        byte[] bytes = new byte[100];
        IOUtils.readFully(res.bytes(), bytes, 0, 100); // throws if 100 bytes not streamed
        assertEquals(-1, res.bytes().read());
    }

    @Test
    @Timeout(value = 15, unit = TimeUnit.SECONDS)
    public void testResponseBodyStringGet() throws Exception {
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();

        ResponseBodyStringGetResponse res = s.responseBodies().responseBodyStringGet().call();

        assertEquals(200, res.statusCode());
        assertNotNull(res.html());
        assertTrue(res.html().contains("Herman Melville - Moby-Dick"));
    }

    @Test
    public void testUnmarshalSurfacesBodyTimeoutsAsUncheckedIO() {
        HttpRequest request = HttpRequest.builder().method("GET").uri(URI.create("http://localhost/test")).build();
        InputStream timingOut = new InputStream() {
            @Override
            public int read() throws IOException {
                throw new InterruptedIOException("timeout");
            }
        };
        HttpResponse<InputStream> response = new HttpResponse<>(request, 200, new Headers(), timingOut);
        UncheckedIOException thrown = assertThrows(UncheckedIOException.class,
                () -> Utils.unmarshal(response, new TypeReference<Map<String, Object>>() {}));
        assertTrue(thrown.getCause() instanceof InterruptedIOException, String.valueOf(thrown.getCause()));
    }
}
