package org.openapis.secondary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;
import org.openapis.secondary.openapi.models.components.AuthServiceRequestBody;
import org.openapis.secondary.openapi.models.components.AuthServiceRequestBodyBasicAuth;
import org.openapis.secondary.openapi.models.components.SchemeBasicAuth;
import org.openapis.secondary.openapi.models.components.SchemeClientCredentials;
import org.openapis.secondary.openapi.models.components.Security;
import org.openapis.secondary.openapi.models.operations.*;
import static org.openapis.secondary.openapi.Helpers.API_TEST_SERVICE_URL;
import static org.openapis.secondary.openapi.Helpers.HTTPBIN_URL;
import org.openapis.secondary.openapi.utils.RecordingClient;
import org.openapis.secondary.openapi.utils.Utils;

public class AuthAdditionalTest {
    @Test
    void testNoAuth() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-no-auth-retained");

        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        assertNotNull(s);

        NoAuthResponse res = s.auth().noAuthDirect();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testBasicAuth() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-basic-auth");

        SDK s = SDK.builder().serverURL(HTTPBIN_URL)
                .security(Security.builder().username("testUser").password("testPass").build())
                .build();
        assertNotNull(s);

        BasicAuthResponse res = s.auth().basicAuth().user("testUser").passwd("testPass").call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.user().authenticated());
    }

    @Test
    void testApiKeyAuthGlobal() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-operation-auth-retained");

        SDK s = SDK.builder().serverURL(HTTPBIN_URL).build();
        assertNotNull(s);

        MultipleMixedOptionsAuthResponse res = s.authNew().multipleMixedOptionsAuth()
                .security(MultipleMixedOptionsAuthSecurity.builder()
                        .basicAuth(SchemeBasicAuth.builder().username("testUser")
                                .password("testPass").build())
                        .build())
                .request(AuthServiceRequestBody.builder().basicAuth(
                        AuthServiceRequestBodyBasicAuth.builder().username("testUser").password("testPass").build())
                        .build())
                .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testOperationLevelOauth2() throws Exception {
        CommonHelpers.recordTest("auth-operation-level-oauth2");

        RecordingClient client = new RecordingClient();
        String clientID = "speakeasy-sdks";
        String clientSecret = "supersecret-" + CommonHelpers.randomId();
        SDK s = SDK.builder().serverURL(HTTPBIN_URL).client(client).build();
        assertNotNull(s);

        // A token should be requested with 'read', 'write' and 'erase' scopes.
        AuthenticatedRequestResponse res = s.hooks().authenticatedRequest()
                .security(AuthenticatedRequestSecurity.builder()
                        .clientID(clientID)
                        .clientSecret(clientSecret)
                        .audience("")
                        .build())
                .call();

        assertNotNull(res);

        // Requires 'read' and 'write' scopes. The same token should be reused
        // since [read, write, erase] is a superset of [read, write].
        AuthenticatedRequestUnflattenedResponse res2 = s.hooks().authenticatedRequestUnflattened()
                .security(AuthenticatedRequestUnflattenedSecurity.builder()
                        .clientCredentials(SchemeClientCredentials.builder()
                                .clientID(clientID)
                                .clientSecret(clientSecret)
                                .audience("")
                                .build())
                        .build())
                .call();

        assertNotNull(res2);

        // Check that only a single token request was made (token is reused for superset scopes)
        String tokenURL = "/clientcredentials/token";
        long tokenRequests = client.requests().stream()
            .filter(request -> request.uri().toString().contains(tokenURL))
            .filter(request -> {
                String body = request.body()
                        .map(b -> {
                            try {
                                return Utils.toUtf8AndClose(b.stream());
                            } catch (java.io.IOException e) {
                                throw new java.io.UncheckedIOException(e);
                            }
                        })
                        .orElse("");
                return body.contains("scope=");
            })
            .count();
        assertEquals(1, tokenRequests);
    }

    @Test
    void testAuthnew_CustomSchemeAppID() throws Exception {
        Utils.recordTest("auth-custom-security-scheme-app-id");

        var sdk = SDK.builder().serverURL(HTTPBIN_URL) //
                .client(Utils.createTestHTTPClient("customSchemeAppId")) //
                .build();

        var security = CustomSchemeAppIdSecurity.builder() //
                .appId("testAppID") //
                .secret("testSecret") //
                .build();

        var server = API_TEST_SERVICE_URL;

        CustomSchemeAppIdResponse res = sdk.authNew().customSchemeAppId()
                .security(security) //
                .serverURL(server) //
                .call();
        assertEquals(200, res.statusCode());
    }
}
