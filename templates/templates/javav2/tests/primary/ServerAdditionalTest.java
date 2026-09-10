package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.util.HashMap;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.SDK.Builder.ServerSomething;
import org.openapis.openapi.models.operations.SelectGlobalServerResponse;
import org.openapis.openapi.models.operations.SelectServerWithIDResponse;
import org.openapis.openapi.models.operations.ServersByIDWithTemplatesResponse;
import org.openapis.openapi.operations.SelectServerWithID;
import org.openapis.openapi.operations.ServerWithTemplates;
import org.openapis.openapi.models.operations.ServerWithTemplatesGlobalResponse;
import org.openapis.openapi.models.operations.ServerWithTemplatesResponse;
import org.openapis.openapi.utils.Utils;


public class ServerAdditionalTest {
    @Test
    void testSelectGlobalServerValid() throws Exception {
        CommonHelpers.recordTest("servers-select-global-server-valid");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        SelectGlobalServerResponse res = s.servers().selectGlobalServer().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testSelectGlobalServerBroken() throws Exception {
        CommonHelpers.recordTest("servers-select-global-server-broken");

        SDK s = SDK.builder().serverIndex(1).build();
        assertNotNull(s);

        assertThrows(Exception.class, () -> s.servers().selectGlobalServer().call());
    }

    @Test
    void testSelectServerWithIDDefault() throws Exception {
        CommonHelpers.recordTest("servers-select-server-with-id-default");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        SelectServerWithIDResponse res = s.servers().selectServerWithID().serverURL(Helpers.HTTPBIN_URL).call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testSelectServerWithIDValid() throws Exception {
        CommonHelpers.recordTest("servers-select-server-with-id-valid");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        SelectServerWithIDResponse res = s.servers()
            .selectServerWithID()
            .serverURL(Helpers.HTTPBIN_URL)
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testSelectServerWithIBroken() throws Exception {
        CommonHelpers.recordTest("servers-select-server-with-id-broken");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        assertThrows(Exception.class, () -> s.servers()
            .selectServerWithID()
            .serverURL(SelectServerWithID.SELECT_SERVER_WITH_ID_SERVERS.get(SelectServerWithID.SelectServerWithIDServers.BROKEN))
            .call());
    }

    @Test
    void testServerWithTemplatesGlobal() throws Exception {
        CommonHelpers.recordTest("servers-server-with-templates-global");

        SDK s = SDK.builder().serverIndex(2).hostname("localhost").port(Helpers.HTTPBIN_PORT).build();
        assertNotNull(s);

        ServerWithTemplatesGlobalResponse res = s.servers().serverWithTemplatesGlobal().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testServerWithTemplatesGlobalDefaults() throws Exception {
        CommonHelpers.recordTest("servers-server-with-templates-global-defaults");

        SDK s = SDK.builder().serverIndex(2).hostname("localhost").port(Helpers.HTTPBIN_PORT).build();
        assertNotNull(s);

        ServerWithTemplatesGlobalResponse res = s.servers().serverWithTemplatesGlobal().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testServerWithTemplatesGlobalEnum() throws Exception {
        CommonHelpers.recordTest("servers-server-with-templates-global-enum");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL + "/anything/" + ServerSomething.SOMETHING_ELSE_AGAIN.value()).build();
        assertNotNull(s);

        ServerWithTemplatesGlobalResponse res = s.servers().serverWithTemplatesGlobal().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testServerWithTemplates() throws Exception {
        CommonHelpers.recordTest("servers-server-with-templates");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        ServerWithTemplatesResponse res = s.servers()
            .serverWithTemplates()
            .serverURL(Utils.templateUrl(ServerWithTemplates.SERVER_WITH_TEMPLATES_SERVERS[0],
                                         new HashMap<String, String>() {
                                             {
                                                 put("hostname", "localhost");
                                                 put("port", Helpers.HTTPBIN_PORT);
                                             }
                                         }))
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testServerWithTemplatesDefaults() throws Exception {
        CommonHelpers.recordTest("servers-server-with-templates-defaults");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        ServerWithTemplatesResponse res = s.servers()
            .serverWithTemplates()
            .serverURL(Utils.templateUrl(ServerWithTemplates.SERVER_WITH_TEMPLATES_SERVERS[0],
                                         new HashMap<String, String>() {
                                             {
                                                 put("hostname", "localhost");
                                                 put("port", Helpers.HTTPBIN_PORT);
                                             }
                                         }))
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testServerByIDWithTemplates() throws Exception {
        CommonHelpers.recordTest("servers-server-by-id-with-templates");

        SDK s = SDK.builder().build();
        assertNotNull(s);

        ServersByIDWithTemplatesResponse res = s.servers()
            .serversByIDWithTemplates()
            .serverURL(Utils.templateUrl("http://{hostname}:{port}",
                                         new HashMap<String, String>() {
                                             {
                                                 put("hostname", "localhost");
                                                 put("port", Helpers.HTTPBIN_PORT);
                                             }
                                         }))
            .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }
}
