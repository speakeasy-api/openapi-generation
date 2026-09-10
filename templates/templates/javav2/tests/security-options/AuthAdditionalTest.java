package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

import org.openapis.openapi.models.errors.APIException;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.BasicHttp;
import org.openapis.openapi.models.shared.ApiKeyAuth;
import org.openapis.openapi.models.shared.AccessToken;

public class AuthAdditionalTest {
    @Test
    void testGlobalSecurityBasicHttpSuccess() throws Exception {
        CommonHelpers.recordTest("auth-basic-http-global-option");

        // Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
        SDK s = SDK.builder()
                .security(Security.builder()
                        .basicHttp(BasicHttp.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .accessToken(AccessToken.builder()
                                .accessToken("Bearer ignored")
                                .build())
                        .build())
                .build();

        var res = s.auth().globalSecurityOptionBasicHttp().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.basicAuth().isPresent());
        assertTrue(res.basicAuth().get().authenticated());
        assertEquals("testUser", res.basicAuth().get().user());
    }

    @Test
    void testGlobalSecurityFieldsOrdering() {
        CommonHelpers.recordTest("auth-global-security-option-fields-ordering");

        // Expected to fail since APIKeyAuth takes priority over BasicHTTP in global security definition
        SDK s = SDK.builder()
                .security(Security.builder()
                        .apiKeyAuth(ApiKeyAuth.builder()
                                .apiKeyAuth("Bearer test_api_key")
                                .build())
                        .basicHttp(BasicHttp.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .build())
                .build();

        APIException ex = assertThrows(APIException.class, () -> {
            s.auth().globalSecurityOptionBasicHttp().call();
        });
        assertEquals(401, ex.code());
    }

    @Test
    void testHoistedSecurityAccessTokenFirst() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-security-option-access-token-first");

        SDK s = SDK.builder()
                .security(Security.builder()
                        .apiKeyAuth(ApiKeyAuth.builder()
                                .apiKeyAuth("testApiKey")
                                .build())
                        .accessToken(AccessToken.builder()
                                .accessToken("Bearer ghp_xxxx")
                                .build())
                        .build())
                .build();

        var res = s.auth().hoistedSecurityOptionAccessTokenFirst().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.tokenAuthResponse().isPresent());
        assertEquals("Bearer ghp_xxxx", res.tokenAuthResponse().get().token());
    }

    @Test
    void testHoistedSecurityApiKeyFirst() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-security-option-api-key-first");

        SDK s = SDK.builder()
                .security(Security.builder()
                        .apiKeyAuth(ApiKeyAuth.builder()
                                .apiKeyAuth("testApiKey")
                                .build())
                        .accessToken(AccessToken.builder()
                                .accessToken("Bearer ghp_xxxx")
                                .build())
                        .build())
                .build();

        var res = s.auth().hoistedSecurityOptionApiKeyFirst().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.tokenAuthResponse().isPresent());
        assertEquals("testApiKey", res.tokenAuthResponse().get().token());
    }

    @Test
    void testHoistedSecurityBasicHttpOnly() throws Exception {
        CommonHelpers.recordTest("auth-hoisted-security-option-basic-http-only");

        SDK s = SDK.builder()
                .security(Security.builder()
                        .apiKeyAuth(ApiKeyAuth.builder()
                                .apiKeyAuth("testApiKey")
                                .build())
                        .basicHttp(BasicHttp.builder()
                                .username("testUser")
                                .password("testPass")
                                .build())
                        .accessToken(AccessToken.builder()
                                .accessToken("Bearer ghp_xxxx")
                                .build())
                        .build())
                .build();

        var res = s.auth().hoistedSecurityOptionBasicHttpOnly().call();
        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.basicAuth().isPresent());
        assertTrue(res.basicAuth().get().authenticated());
        assertEquals("testUser", res.basicAuth().get().user());
    }

    @Test
    void testHoistedSecurityInvalidOption() {
        CommonHelpers.recordTest("auth-hoisted-security-invalid-option");

        SDK s = SDK.builder()
                .security(Security.builder()
                        .basicHttp(BasicHttp.builder()
                                .username("username")
                                .password("password")
                                .build())
                        .build())
                .build();

        APIException ex = assertThrows(APIException.class, () -> {
            s.auth().hoistedSecurityOptionAccessTokenFirst().call();
        });
        assertEquals(401, ex.code());
    }
}
