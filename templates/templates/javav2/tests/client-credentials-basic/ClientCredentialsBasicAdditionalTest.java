package org.openapis.clientcredentialsbasic.openapi;

import org.junit.jupiter.api.Test;
import org.openapis.clientcredentialsbasic.openapi.hooks.OAuth2Scopes;
import org.openapis.clientcredentialsbasic.openapi.models.errors.AuthException;
import org.openapis.clientcredentialsbasic.openapi.models.operations.AuthenticatedRequestGlobalServerResponse;
import org.openapis.clientcredentialsbasic.openapi.models.operations.AuthenticatedRequestResponse;
import org.openapis.clientcredentialsbasic.openapi.models.shared.SchemeClientCredentials;
import org.openapis.clientcredentialsbasic.openapi.models.shared.Security;

import java.net.HttpURLConnection;
import java.util.Random;
import java.util.stream.Collectors;
import java.util.stream.IntStream;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.openapis.clientcredentialsbasic.openapi.Helpers.API_TEST_SERVICE_URL;

public final class ClientCredentialsBasicAdditionalTest {

    private static final Random RANDOM = new Random();

    @Test
    public void testClientCredentialsHookSuccessfullyAuthenticates() throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-basic-success");

        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientCredentials(SchemeClientCredentials.builder()
                                .clientID("speakeasy-sdks") //
                                .clientSecret("supersecret-" + nextId()) //
                                .audience("") //
                                .build()) //
                        .build())
                .build();
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
    }

    @Test
    public void testClientCredentialsHookSuccessfullyAuthenticatesGlobalServer() throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-basic-success-global-server");

        SDK sdk = SDK.builder() //
                .security(Security.builder() //
                        .clientCredentials(SchemeClientCredentials.builder()
                                .clientID("speakeasy-sdks") //
                                .clientSecret("supersecret-" + nextId()) //
                                .audience("") //
                                .build()) //
                        .build())
                .build();
        {
            AuthenticatedRequestGlobalServerResponse res = sdk.hooks().authenticatedRequestGlobalServer().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
        {
            AuthenticatedRequestGlobalServerResponse res = sdk.hooks().authenticatedRequestGlobalServer().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
    }

    @Test
    public void testClientCredentialsHookSuccessfullyAuthenticatesWithAltTokenURL()
            throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-basic-success-alt-token-url");

        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientCredentials(SchemeClientCredentials.builder()
                                .clientID("speakeasy-sdks") //
                                .clientSecret("supersecret-" + nextId()) //
                                .tokenURL("/clientcredentials/alt/token") //
                                .scopes(java.util.Arrays.asList("alt:one", OAuth2Scopes.ClientCredentials.AltTwo.value())) //
                                .audience("") //
                                .build()) //
                        .build())
                .build();
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
    }

    @Test
    public void testClientCredentialsHookAuthenticationFails() throws Exception {
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientCredentials(SchemeClientCredentials.builder()
                                .clientID("speakeasy-sdks") //
                                .clientSecret("badsecret") //
                                .audience("") //
                                .build()) //
                        .build())
                .build();

        assertThrows(AuthException.class, () -> sdk.hooks().authenticatedRequest().call());
    }

    private static String nextId() {
        String s = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
        int idLength = 10;
        return IntStream //
                .range(0, idLength) //
                .mapToObj(i -> String.valueOf(s.charAt(RANDOM.nextInt(s.length())))) //
                .collect(Collectors.joining());
    }
}
