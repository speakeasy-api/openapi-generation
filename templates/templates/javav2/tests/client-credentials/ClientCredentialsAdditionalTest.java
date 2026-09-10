package org.openapis.clientcredentials.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.net.HttpURLConnection;
import java.net.http.HttpRequest;
import java.util.Arrays;
import java.util.Optional;

import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.openapis.clientcredentials.openapi.hooks.ClientCredentialsHook;
import org.openapis.clientcredentials.openapi.models.errors.AuthException;
import org.openapis.clientcredentials.openapi.models.operations.AuthenticatedRequestResponse;
import org.openapis.clientcredentials.openapi.models.operations.AuthenticatedRequestGlobalServerResponse;
import org.openapis.clientcredentials.openapi.models.operations.AuthenticatedRequestNoScopesResponse;
import org.openapis.clientcredentials.openapi.models.shared.Security;
import org.openapis.clientcredentials.openapi.utils.HTTPClient;
import org.openapis.clientcredentials.openapi.utils.Hook.BeforeRequestContext;
import org.openapis.clientcredentials.openapi.utils.Hook.BeforeRequestContextImpl;
import org.openapis.clientcredentials.openapi.utils.Hook.SdkInitData;
import org.openapis.clientcredentials.openapi.utils.RecordingClient;
import org.openapis.clientcredentials.openapi.utils.Helpers;
import org.openapis.clientcredentials.openapi.hooks.OAuth2Scopes;
import static org.openapis.clientcredentials.openapi.Helpers.API_TEST_SERVICE_URL;

public final class ClientCredentialsAdditionalTest {

    @Test
    public void testClientCredentialsHookMissingScopes() throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-no-scopes");

        String clientId = "speakeasy-sdks";
        String clientSecret = "supersecret-" + CommonHelpers.randomId();

        // Scenario 1: expected to fail since the token endpoint requires 'read' and 'write' scopes
        // but the authenticatedRequestNoScopes operation does not specify any.
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID(clientId) //
                        .clientSecret(clientSecret) //
                        .build()) //
                .build();

        AuthException thrown = assertThrows(AuthException.class, () -> {
            sdk.hooks().authenticatedRequestNoScopes().call();
        });
        assertEquals(Optional.of(400), thrown.statusCode());
        assertTrue(thrown.getMessage().contains("Unexpected status code 400: empty_scopes"));

        // Scenario 2: same check but this time we override the default scopes with an empty list
        SDK sdk2 = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID(clientId) //
                        .clientSecret(clientSecret) //
                        .scopes(Arrays.asList()) // overrides global scopes
                        .build()) //
                .build();

        AuthException thrown2 = assertThrows(AuthException.class, () -> {
            sdk2.hooks().authenticatedRequest().call();
        });
        assertEquals(Optional.of(400), thrown2.statusCode());
        assertTrue(thrown2.getMessage().contains("Unexpected status code 400: empty_scopes"));

        // Scenario 3: now use a different tokenURL that will allow no scopes to be requested
        String tokenUrl = "/clientcredentials/token?expires_in=90&skip_scopes=true";
        RecordingClient client3 = new RecordingClient();

        SDK sdk3 = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .client(client3) //
                .security(Security.builder() //
                        .clientID(clientId) //
                        .clientSecret(clientSecret) //
                        .tokenURL(tokenUrl) //
                        .build()) //
                .build();

        {
            AuthenticatedRequestNoScopesResponse res3 = sdk3.hooks().authenticatedRequestNoScopes().call();
            assertEquals(HttpURLConnection.HTTP_OK, res3.statusCode());
        }

        // since the token is not expired, it should be reused on subsequent call
        {
            AuthenticatedRequestNoScopesResponse res4 = sdk3.hooks().authenticatedRequestNoScopes().call();
            assertEquals(HttpURLConnection.HTTP_OK, res4.statusCode());
        }

        long tokenRequests = client3.requests().stream()
            .filter(request -> request.uri().toString().contains(tokenUrl))
            .count();
        assertEquals(1, tokenRequests);

        // The request body should not contain "scope=" parameter
        boolean hasNoScopes = client3.requests().stream()
            .filter(request -> request.uri().toString().contains(tokenUrl))
            .allMatch(request -> {
                String body = Helpers.bodyUtf8(request);
                return !body.contains("scope=");
            });
        assertTrue(hasNoScopes);
    }

    @Test
    public void testWithClientCredentialsThatClientSecretOnlyUsedInRefreshTokenCall() throws Exception {
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("supersecret-" + CommonHelpers.randomId()) //
                        .build()) //
                .build();
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
            assertTrue(res.rawResponse().request().headers().firstValue("clientId").isEmpty());
            assertTrue(res.rawResponse().request().headers().firstValue("clientSecret").isEmpty());
        }
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
            assertTrue(res.rawResponse().request().headers().firstValue("clientId").isEmpty());
            assertTrue(res.rawResponse().request().headers().firstValue("clientSecret").isEmpty());
        }
    }

    @Test
    public void testClientCredentialsHookSuccessfullyAuthenticates() throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-success");

        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("supersecret-" + CommonHelpers.randomId()) //
                        .build()) //
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
        CommonHelpers.recordTest("hooks-client-credentials-success-global-server");

        SDK sdk = SDK.builder() //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("supersecret-" + CommonHelpers.randomId()) //
                        .build()) //
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
        CommonHelpers.recordTest("hooks-client-credentials-success-alt-token-url");

        RecordingClient client = new RecordingClient();
        String tokenUrl = "/clientcredentials/alt/token";

        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .client(client) //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("supersecret-" + CommonHelpers.randomId()) //
                        .tokenURL(tokenUrl) //
                        .scopes(Arrays.asList("alt:one", OAuth2Scopes.ClientCredentials.AltTwo.value())) //
                        .build()) //
                .build();
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }
        // since the token is already expired, a new one should be requested
        {
            AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
            assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
        }

        long tokenRequests = client.requests().stream()
            .filter(request -> request.uri().toString().contains(tokenUrl))
            .filter(request -> {
                String body = Helpers.bodyUtf8(request);
                return body.contains("scope=alt%3Aone+alt%3Atwo");
            })
            .count();
        assertEquals(2, tokenRequests);
    }

    @Test
    public void testClientCredentialsHookAuthenticationFails() throws Exception {
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("badsecret") //
                        .build()) //
                .build();
        assertThrows(AuthException.class, () -> sdk.hooks().authenticatedRequest().call());
    }

    @Test
    public void testRequestUntouchedIfOauthScopesNotPresent() throws Exception {
        ClientCredentialsHook hook = new ClientCredentialsHook();
        HTTPClient client = Mockito.mock(HTTPClient.class);
        hook.sdkInit(new SdkInitData("http://localhost", client));
        HttpRequest request = Mockito.mock(HttpRequest.class);

        BeforeRequestContext c =
            new BeforeRequestContextImpl(new SDKConfiguration(), "http://localhost", "operationId", Optional.empty(), Optional.empty());
        assertEquals(request, hook.beforeRequest(c, request));
    }

    @Test
    public void testRequestUntouchedIfSecuritySourceNotPresent() throws Exception {
        ClientCredentialsHook hook = new ClientCredentialsHook();
        HTTPClient client = Mockito.mock(HTTPClient.class);
        hook.sdkInit(new SdkInitData("http://localhost", client));
        HttpRequest request = Mockito.mock(HttpRequest.class);

        BeforeRequestContext c = new BeforeRequestContextImpl(new SDKConfiguration(), "http://localhost", "operationId",
                Optional.of(Arrays.asList("read")), Optional.empty());
        assertEquals(request, hook.beforeRequest(c, request));
    }

    @Test
    public void testClientCredentialsHookLowercaseBearerToken() throws Exception {
        CommonHelpers.recordTest("hooks-client-credentials-lowercase-bearer");

        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .security(Security.builder() //
                        .clientID("speakeasy-sdks") //
                        .clientSecret("supersecret-" + CommonHelpers.randomId()) //
                        .tokenURL("/clientcredentials/token?token_type=bearer") //
                        .build()) //
                .build();

        AuthenticatedRequestResponse res = sdk.hooks().authenticatedRequest().call();
        assertEquals(HttpURLConnection.HTTP_OK, res.statusCode());
    }

}
