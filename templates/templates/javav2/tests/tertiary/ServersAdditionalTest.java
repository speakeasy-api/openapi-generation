package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.openapis.tertiary.openapi.CommonHelpers.recordTest;

import java.net.HttpURLConnection;

import org.openapis.tertiary.openapi.Helpers;

import org.junit.jupiter.api.Test;
import org.openapis.tertiary.openapi.models.operations.SelectGlobalServerResponse;

public class ServersAdditionalTest {

    @Test
    public void testSelectGlobalServerByIdValid() throws Exception {
        recordTest("servers-select-global-server-by-id-valid");
        recordTest("servers-select-global-server-by-id-valid-using-builder");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        SelectGlobalServerResponse res = s.servers().selectGlobalServerDirect();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
    }

    @Test
    public void testSelectGlobalServerByIdDefault() throws Exception {
        recordTest("servers-select-global-server-by-id-default");
        recordTest("servers-select-global-server-by-id-default-using-builder");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        SelectGlobalServerResponse res = s.servers().selectGlobalServerDirect();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
    }

    @Test
    public void testSelectGlobalServerValid() throws Exception {
        recordTest("servers-select-global-server-valid");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        SelectGlobalServerResponse res = s.servers().selectGlobalServerDirect();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
    }

    @Test
    public void testSelectGlobalServerBroken() throws Exception {
        recordTest("servers-select-global-server-broken");

        String url = SDK.SERVERS[1];
        assertEquals("http://broken", url);
        SDK s = SDK.builder().serverURL(url).build();

        assertThrows(Exception.class, () -> {
            s.servers().selectGlobalServerDirect();
        });
    }

    @Test
    public void testSelectGlobalServerByIdInvalid() throws Exception {
        recordTest("servers-select-global-server-by-id-invalid");
        recordTest("servers-select-global-server-by-id-invalid-using-builder");

        assertThrows(Exception.class, () -> {
            SDK.builder().serverIndex(2).build();
        });
    }
    
    @Test
     public void testSelectGlobalServerByIdBroken() throws Exception {
        recordTest("servers-select-global-server-by-id-broken");
        recordTest("servers-select-global-server-by-id-broken-using-builder");

        SDK s = SDK.builder().serverIndex(1).build();
        assertThrows(Exception.class, () -> {
            s.servers().selectGlobalServerDirect();
        });
    }
    
}
