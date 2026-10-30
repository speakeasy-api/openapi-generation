package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;

import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Timeout;
import org.openapis.secondary.openapi.models.operations.ResponseBodyBytesGetResponse;
import org.openapis.secondary.openapi.models.operations.ResponseBodyStringGetResponse;


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
    public void testResponseBodyLimit() throws Exception {
        SDK capped = SDK.builder().serverURL(HTTPBIN_URL).maxResponseBodyBytes(50).build();

        ResponseBodyBytesGetResponse res = capped.responseBodies().responseBodyBytesGet().call();
        assertEquals(200, res.statusCode());
        assertThrows(org.openapis.secondary.openapi.utils.ResponseTooLargeException.class,
                () -> IOUtils.toByteArray(res.bytes()));

        SDK roomy = SDK.builder().serverURL(HTTPBIN_URL).maxResponseBodyBytes(200).build();
        ResponseBodyBytesGetResponse ok = roomy.responseBodies().responseBodyBytesGet().call();
        assertEquals(100, IOUtils.toByteArray(ok.bytes()).length);
    }
}
