package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.CommonHelpers.recordTest;

import java.util.List;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.CustomSchemeAppIdResponse;
import org.openapis.openapi.models.operations.GlobalBearerAuthResponse;
import org.openapis.openapi.models.operations.MultipleSimpleSchemeAuthResponse;
import org.openapis.openapi.models.operations.MultipleSimpleSchemeAuthSecurity;
import org.openapis.openapi.models.errors.APIException;
import org.openapis.openapi.models.shared.AuthServiceRequestBody;
import org.openapis.openapi.models.shared.HeaderAuth;
import org.openapis.openapi.models.shared.SchemeCustomSchemeAppID;
import org.openapis.openapi.models.shared.SchemeBasicHTTP;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.utils.Utils;
import static org.openapis.openapi.Helpers.API_TEST_SERVICE_URL;

public class AuthAdditionalTest {

    @Test
    void testMultipleSimpleSchemeAuth() throws Exception {
        recordTest("auth-multiple-simple-scheme-auth");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        MultipleSimpleSchemeAuthResponse res = s.authNew().multipleSimpleSchemeAuth()
                .request(AuthServiceRequestBody.builder()
                        .headerAuth(List.of(HeaderAuth.builder().headerName("x-api-key")
                                .expectedValue("test_api_key").build()))
                        .build())
                .security(MultipleSimpleSchemeAuthSecurity.builder().oauth2("Bearer testToken")
                        .apiKeyAuthNew("test_api_key").build())
                .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
    }

    @Test
    void testFunctionCallbacksForOAuthSupportGlobalSecurity() throws Exception {
        recordTest("auth-function-callbacks-oauth-global-security");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
                .securitySource(
                        SecuritySource.of(Security.builder().oauth2("Bearer global").build()))
                .build();
        assertNotNull(s);

        GlobalBearerAuthResponse res = s.auth().globalBearerAuth().call();

        assertNotNull(res);
        assertEquals("global", res.token().get().token());
    }

    @Test
    void testAuthnew_CustomOptionAppID() throws Exception {
        Utils.recordTest("auth-custom-security-option-app-id");

        var sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL) //
                .client(Utils.createTestHTTPClient("customSchemeAppId")) //
                .security(Security.builder()
                        .customSchemeAppId(SchemeCustomSchemeAppID.builder()
                                .appId("testAppID") //
                                .secret("testSecret") //
                                .build()) //
                        .build()) //
                .build();

        var server = API_TEST_SERVICE_URL;

        CustomSchemeAppIdResponse res = sdk.authNew().customSchemeAppId()
                .serverURL(server) //
                .call();
        assertEquals(200, res.statusCode());
    }

    @Test
    void testGlobalSecurityFieldsOrdering() {
        recordTest("auth-global-security-fields-ordering");

        // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
        // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
        // which is not valid authentication for this endpoint.
        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .apiKeyAuth("testApiKey")
                        .basicHttp(SchemeBasicHTTP.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .build())
                .build();

        APIException ex = assertThrows(APIException.class, () -> {
            s.auth().globalSecurityBasicHttp().call();
        });
        assertEquals(401, ex.code());
    }

    @Test
    void testHoistedSecurityAccessTokenOnly() throws Exception {
        recordTest("auth-hoisted-security-access-token-only");

        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .apiKeyAuth("testApiKey")
                        .basicHttp(SchemeBasicHTTP.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .accessToken("Bearer ghp_xxxx")
                        .build())
                .build();

        var res = s.auth().hoistedSecurityAccessTokenOnly().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.token().isPresent());
        assertEquals("Bearer ghp_xxxx", res.token().get().token());
    }

    @Test
    void testHoistedSecurityAccessTokenFirst() throws Exception {
        recordTest("auth-hoisted-security-access-token-first");

        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .apiKeyAuth("testApiKey")
                        .accessToken("Bearer ghp_xxxx")
                        .build())
                .build();

        var res = s.auth().hoistedSecurityAccessTokenFirst().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.token().isPresent());
        assertEquals("Bearer ghp_xxxx", res.token().get().token());
    }

    @Test
    void testHoistedSecurityApiKeyFirst() throws Exception {
        recordTest("auth-hoisted-security-api-key-first");

        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .apiKeyAuth("testApiKey")
                        .accessToken("Bearer ghp_xxxx")
                        .build())
                .build();

        var res = s.auth().hoistedSecurityApiKeyFirst().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.token().isPresent());
        assertEquals("testApiKey", res.token().get().token());
    }

    @Test
    void testHoistedSecurityBasicHttpOnly() throws Exception {
        recordTest("auth-hoisted-security-basic-http-only");

        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .apiKeyAuth("testApiKey")
                        .basicHttp(SchemeBasicHTTP.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .accessToken("Bearer ghp_xxxx")
                        .build())
                .build();

        var res = s.auth().hoistedSecurityBasicHttpOnly().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.basicAuth().isPresent());
        assertTrue(res.basicAuth().get().authenticated());
        assertEquals("testUser", res.basicAuth().get().user());
    }

    @Test
    void testHoistedSecurityInvalidField() {
        recordTest("auth-hoisted-security-invalid-field");

        // Provide only basicHttp — not valid for accessTokenFirst which expects bearer/apiKey
        SDK s = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .security(Security.builder()
                        .basicHttp(SchemeBasicHTTP.builder()
                                .username("user")
                                .password("pass")
                                .build())
                        .build())
                .build();

        APIException ex = assertThrows(APIException.class, () -> {
            s.auth().hoistedSecurityAccessTokenFirst().call();
        });
        assertEquals(401, ex.code());
    }
}
