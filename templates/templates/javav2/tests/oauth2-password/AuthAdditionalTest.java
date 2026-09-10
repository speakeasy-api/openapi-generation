package org.openapis.oauth2password.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.oauth2password.openapi.CommonHelpers.recordTest;
import static org.openapis.oauth2password.openapi.hooks.OAuth2PasswordHook.oauth2FieldValue;
import static org.openapis.oauth2password.openapi.hooks.OAuth2PasswordHook.oauth2Value;

import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublisher;
import java.net.http.HttpResponse;
import java.time.OffsetDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;

import org.junit.jupiter.api.Test;
import org.openapis.oauth2password.openapi.models.errors.APIException;
import org.openapis.oauth2password.openapi.models.errors.AuthException;
import org.openapis.oauth2password.openapi.models.operations.CreateProductResponse;
import org.openapis.oauth2password.openapi.models.operations.HealthCheckResponse;
import org.openapis.oauth2password.openapi.models.operations.ListProductsResponse;
import org.openapis.oauth2password.openapi.models.operations.UpdateProductResponse;
import org.openapis.oauth2password.openapi.models.operations.UpdateProductStockSecurity;
import org.openapis.oauth2password.openapi.models.operations.UpdateProductStockResponse;
import org.openapis.oauth2password.openapi.models.shared.NewProductForm;
import org.openapis.oauth2password.openapi.models.shared.Oauth2Credentials;
import org.openapis.oauth2password.openapi.models.shared.Oauth2Input;
import org.openapis.oauth2password.openapi.models.shared.ProductInventoryStatus;
import org.openapis.oauth2password.openapi.models.shared.ProductInventoryUpdateForm;
import org.openapis.oauth2password.openapi.models.shared.ProductUpdateForm;
import org.openapis.oauth2password.openapi.utils.HTTPClient;
import org.openapis.oauth2password.openapi.utils.HasSecurity;
import org.openapis.oauth2password.openapi.utils.Helpers;
import static org.openapis.oauth2password.openapi.Helpers.API_TEST_SERVICE_URL;
import org.openapis.oauth2password.openapi.utils.RecordingClient;
import org.openapis.oauth2password.openapi.utils.RequestBody;
import org.openapis.oauth2password.openapi.utils.Security;
import org.openapis.oauth2password.openapi.utils.SpeakeasyHTTPClient;
import org.openapis.oauth2password.openapi.utils.SpeakeasyMetadata;
import org.openapis.oauth2password.openapi.utils.TypedObject;
import org.openapis.oauth2password.openapi.utils.Utils;
import org.openapis.oauth2password.openapi.utils.Utils.JsonShape;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

public class AuthAdditionalTest {

    @Test
    public void testExtractionOfOAuth2PasswordFieldsFromSecurityObject() {
        MySecurity security = new MySecurity(Optional.of(MyOauth2Input.of(new MyOauth2Credentials())));
        Object creds = oauth2Value(security).get();
        assertEquals("fred", oauth2FieldValue(creds, "username").get());
        assertEquals("pass", oauth2FieldValue(creds, "password").get());
        assertEquals("me", oauth2FieldValue(creds, "clientID").get());
        assertEquals("secretive", oauth2FieldValue(creds, "clientSecret").get());
        assertEquals("https://the-url", oauth2FieldValue(creds, "tokenURL").get());
        assertFalse(oauth2FieldValue(creds, "notfound").isPresent());
    }

    @Test
    public void testExtractionOfOAuth2PasswordFieldsFromSecurityObjectWhenOauth2PasswordNotPresent() {
        MySecurityInvalid security = new MySecurityInvalid(Optional.of(MyOauth2Input.of(new MyOauth2Credentials())));
        assertFalse(oauth2Value(security).isPresent());
    }

    @Test
    public void testExtractionOfOAuth2PasswordFieldsFromSecurityObjectWhenOauth2PasswordAnnotationNotPresent() {
        MySecurityInvalid security = new MySecurityInvalid(Optional.of(MyOauth2Input.of(new MyOauth2Credentials())));
        assertFalse(oauth2Value(security).isPresent());
    }

    @Test
    public void testExtractionOfOAuth2PasswordFieldsFromSecurityObjectWhenOauth2PasswordFieldValueNotPresent() {
        MySecurity security = new MySecurity(Optional.empty());
        assertFalse(oauth2Value(security).isPresent());
    }

    @Test
    public void testExtractionOfOAuth2PasswordFieldsFromSecurityObjectWhenOauth2PasswordIsExplicitToken() {
        MySecurity security = new MySecurity(Optional.of(MyOauth2Input.of("theToken")));
        assertTrue(oauth2Value(security).get() instanceof String);
    }

    public static class MySecurity {

        @SpeakeasyMetadata("security:scheme=true,type=oauth2,subtype=password,name=Authorization")
        private Optional<? extends MyOauth2Input> oauth2;

        public MySecurity(Optional<? extends MyOauth2Input> oauth2) {
            Utils.checkNotNull(oauth2, "oauth2");
            this.oauth2 = oauth2;
        }
    }

    public static class MySecurityInvalid {

        @SpeakeasyMetadata("security:scheme=true,type=oauth2,subtype=passwording,name=Authorization")
        private Optional<? extends MyOauth2Input> oauth2;

        public MySecurityInvalid(Optional<? extends MyOauth2Input> oauth2) {
            Utils.checkNotNull(oauth2, "oauth2");
            this.oauth2 = oauth2;
        }
    }

    public static class MyOauth2Input {

        private TypedObject value;

        private MyOauth2Input(TypedObject value) {
            this.value = value;
        }

        public static MyOauth2Input of(MyOauth2Credentials value) {
            Utils.checkNotNull(value, "value");
            return new MyOauth2Input(TypedObject.of(value, JsonShape.DEFAULT, new TypeReference<MyOauth2Credentials>() {
            }));
        }

        public static MyOauth2Input of(String value) {
            Utils.checkNotNull(value, "value");
            return new MyOauth2Input(TypedObject.of(value, JsonShape.DEFAULT, new TypeReference<String>() {
            }));
        }

        public java.lang.Object value() {
            return value.value();
        }

    }

    public static class MyOauth2Credentials {

        @SpeakeasyMetadata("security:name=username")
        private String username = "fred";

        @SpeakeasyMetadata("security:name=password")
        private String password = "pass";

        @SpeakeasyMetadata("security:name=clientID")
        private Optional<String> clientID = Optional.of("me");

        @SpeakeasyMetadata("security:name=clientSecret")
        private Optional<String> clientSecret = Optional.of("secretive");

        @SpeakeasyMetadata("security:name=tokenURL")
        private String tokenURL = "https://the-url";
        
    }
    
    @SuppressWarnings("unused")
    public static class MyOauth2CredentialsNested {
        
        @SpeakeasyMetadata("security:scheme=true,type=oauth2,subtype=password,name=Authorization")
        private Optional<? extends Oauth2Input> oauth2;
        
        @SpeakeasyMetadata("security:scheme=true,type=http,subtype=basic")
        private Optional<? extends SchemeBasicAuth> basicAuth;

        @SpeakeasyMetadata("security:scheme=true,type=oauth2,subtype=client_credentials")
        private Optional<? extends MyOauth2Credentials> clientCredentials;
        
        private String nonAnnotatedField = "notfound";
        
        @JsonIgnore  
        private String unrelatedField = "blah";

    }
    
    public static final class SchemeBasicAuth implements HasSecurity {
        
    }
    
    @Test
    public void testFindCredentialsObjectWhenNotNested() {
        MyOauth2CredentialsNested s = new MyOauth2CredentialsNested();
        s.oauth2 = Optional.empty();
        s.basicAuth = Optional.empty();
        MyOauth2Credentials creds = new MyOauth2Credentials();
        creds.username = "fred";
        creds.password = "pass";
        creds.tokenURL = "https://the-url";
        s.clientCredentials = Optional.of(creds);
        
        assertFalse(Security.findComplexObjectWithNonEmptyAnnotatedField(creds, "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=password\\b").isPresent());
        assertFalse(Security.findComplexObjectWithNonEmptyAnnotatedField(s, "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=password\\b").isPresent());
        assertTrue(Security.findComplexObjectWithNonEmptyAnnotatedField(s, "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=client_credentials\\b").isPresent());
        s.clientCredentials = Optional.empty();
        assertFalse(Security.findComplexObjectWithNonEmptyAnnotatedField(s, "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=client_credentials\\b").isPresent());
        assertFalse(Security.findComplexObjectWithNonEmptyAnnotatedField(null, "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=password\\b").isPresent());
        assertFalse(Security.findComplexObjectWithNonEmptyAnnotatedField("hello", "\\bscheme=true\\b", "\\btype=oauth2\\b", "\\bsubtype=password\\b").isPresent());
    }

    @Test
    void testUsesOath2PasswordCredentialsToObtainTokenAndMakeRequest() throws Exception {
        recordTest("hooks-oauth2-password-with-credentials");

        RecordingClient client = new RecordingClient();
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of( //
                        Oauth2Credentials.builder() //
                                .clientID("beezy") //
                                .clientSecret("super-secret") //
                                .username("testuser") //
                                .password("testpassword") //
                                .build()))
                .client(client) //
                .build();
        ListProductsResponse res = sdk.products().list().call();

        assertEquals(200, res.statusCode());
        assertEquals("pass", res.rawResponse().headers().firstValue("x-oauth2").get());
        assertEquals(2, client.requests().size());
        // check token request
        HttpRequest r = client.requests().get(0);
        assertEquals("POST", r.method());
        assertEquals(API_TEST_SERVICE_URL + "/oauth2/token", r.uri().toString());
        assertEquals(10, res.products().get().products().size());
    }

    @Test
    public void testBypassesOAuth2PasswordHookWithOAuth2AccessToken() throws Exception {
        recordTest("hooks-oauth2-password-with-token");

        final String token;
        {
            Map<String, String> map = new HashMap<>();
            map.put("grant_type", "password");
            map.put("username", "testuser");
            map.put("password", "testpassword");
            map.put("client_id", "beezy");
            map.put("client_secret", "super-secret");
            map.put("scope", "products:read");
            BodyPublisher body = RequestBody.serializeFormData(map).body();
            HttpRequest request = HttpRequest //
                    .newBuilder(new URI(API_TEST_SERVICE_URL + "/oauth2/token")) //
                    .POST(body) //
                    .header("Content-Type", "application/x-www-form-urlencoded") //
                    .build();

            HTTPClient client = new SpeakeasyHTTPClient();
            HttpResponse<InputStream> response = client.send(request);
            String json = Utils.toUtf8AndClose(response.body());
            assertEquals(200, response.statusCode(), json);
            ObjectMapper m = new ObjectMapper();
            JsonNode node = m.readTree(json);
            token = node.get("access_token").asText();
        }

        RecordingClient client = new RecordingClient();
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of(token)) //
                .client(client) //
                .build();
        ListProductsResponse res = sdk.products().list().call();
        assertEquals(200, res.statusCode());
        assertEquals(1, client.requests().size());
        assertEquals(10, res.products().get().products().size());
    }

    @Test
    public void testOAuth2PasswordFailsToObtainTokenWithBadCredentials() throws Exception {
        recordTest("hooks-oauth2-password-bad-credentials");
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of( //
                        Oauth2Credentials.builder() //
                                .clientID("beezy") //
                                .clientSecret("super-secret") //
                                .username("testuser") //
                                .password("BAD_PASSWORD") //
                                .build()))
                .build();
        AuthException e = assertThrows(AuthException.class, () -> sdk.products().list().call());
        assertTrue(e.getMessage().contains("invalid username or password"));
    }

    @Test
    public void testOauth2PasswordAutomaticallyRenewsAccessToken() throws Exception {
        recordTest("hooks-oauth2-password-token-renewal");
        RecordingClient client = new RecordingClient();
        AtomicBoolean first = new AtomicBoolean(true);
        // we use custom client with it's own hooks because normal hooks don't get
        // picked up in token calls (will question team)
        client.beforeRequest(request -> {
            if (request.uri().toString().endsWith("/token") && first.get()) {
                first.set(false);
                return Helpers.copy(request) //
                        // ask server to return an expired token
                        .setHeader("x-oauth2-expire-at", OffsetDateTime.now().minusDays(1).toString()) //
                        .build();
            } else {
                return request;
            }
        });
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of( //
                        Oauth2Credentials.builder() //
                                .clientID("beezy") //
                                .clientSecret("super-secret") //
                                .username("testuser") //
                                .password("testpassword") //
                                .build()))
                .client(client).build();

        // Note that 401 does not indicate specifically that token needs refreshing so
        // is legit to throw normally the session cache will know that the token has or
        // is about to expire and request a new one

        // we could use retries but I think needs work so we'll just test token refresh
        APIException e = assertThrows(APIException.class, () -> sdk.products().list().call());
        assertEquals(401, e.rawResponse().statusCode());
        assertEquals(2, client.requests().size());
        // next call refreshes token automatically because the 401 cleared the session
        // cache entry
        client.reset();
        sdk.products().list().call();
        assertEquals(2, client.requests().size());
    }

    @Test
    public void testSkipsOauth2FlowForMethodsThatDoNotRequireIt() throws Exception {
        recordTest("hooks-oauth2-password-not-required");
        RecordingClient client = new RecordingClient();
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of( //
                        Oauth2Credentials.builder() //
                                .clientID("beezy") //
                                .clientSecret("super-secret") //
                                .username("testuser") //
                                .password("testpassword") //
                                .build()))
                .client(client) //
                .build();
        HealthCheckResponse res = sdk.healthCheck().call();
        assertEquals(200, res.statusCode());
        assertEquals("pong", res.res().get());
        assertFalse(res.rawResponse().headers().map().containsKey("x-oauth2"));
        assertEquals(1, client.requests().size());
    }

    @Test
    public void testOAuth2PasswordAutomaticallyObtainsTokenWithOperationScope() throws Exception {
        recordTest("hooks-oauth2-password-operation-scope");
        RecordingClient client = new RecordingClient();
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .oauth2(Oauth2Input.of( //
                        Oauth2Credentials.builder() //
                                .clientID("beezy") //
                                .clientSecret("super-secret") //
                                .username("testuser") //
                                .password("testpassword") //
                                .build()))
                .client(client) //
                .build();
        {
            // This call will request a token for the 'products:create' scope
            client.reset();
            CreateProductResponse res = sdk.products().create() //
                    .request(NewProductForm.builder() //
                    .description("Games console") //
                    .name("Playstation 5") //
                    .price(499.99f) //
                    .build()) //
                    .call();
            assertEquals("Playstation 5", res.product().get().name());
            assertEquals(2, client.requests().size());
            String body = Helpers.bodyUtf8(client.requests().get(0));
            assertTrue(body.contains("&scope=products%3Acreate&"));
        }
        {
            // This operation requires an additional 'products:read' scope
            // so we expect a new token to be requested.
            client.reset();
            UpdateProductResponse res = sdk.products().update() //
                    .id("123")
                    .productUpdateForm(ProductUpdateForm.builder() //
                        .price(21.49f) //
                        .build()) //
                    .call();
            assertEquals(21.49f, res.product().get().price());
            assertEquals(2, client.requests().size());
            String body = Helpers.bodyUtf8(client.requests().get(0));
            assertTrue(body.contains("&scope=products%3Aread+products%3Acreate&"));
        }
        {
            // This operation only requires the 'products:read' scope, for which a token has
            // already been requested in the previous call so we expect the token to be reused.
            client.reset();
            ListProductsResponse res = sdk.products().list().call();
            assertEquals(200, res.statusCode());
            assertEquals(10, res.products().get().products().size());
            assertEquals(1, client.requests().size());
            assertFalse(client.requests().get(0).uri().toString().contains("/token"));
        }
    }

    @Test
    public void testOAuth2PasswordAsOperationLevelSecurityOption() throws Exception {
        recordTest("hooks-oauth2-password-operation-security-option");
        RecordingClient client = new RecordingClient();
        SDK sdk = SDK.builder() //
                .serverURL(API_TEST_SERVICE_URL) //
                .client(client) //
                .build();

        UpdateProductStockResponse res = sdk.inventory().updateProductStock()
                .security(UpdateProductStockSecurity.builder()
                    .oauth2(Oauth2Input.of( //
                            Oauth2Credentials.builder() //
                                    .clientID("beezy") //
                                    .clientSecret("super-secret") //
                                    .username("testuser") //
                                    .password("testpassword") //
                                    .build())) //
                    .build()) //
                .productInventoryUpdateForm(ProductInventoryUpdateForm.builder() //
                            .quantityDelta(99) //
                            .build()) //
                .id("123") //
                .call();

        assertEquals(200, res.statusCode());
        assertEquals(2, client.requests().size());
        String body = Helpers.bodyUtf8(client.requests().get(0));
        assertTrue(body.contains("&scope=admin&"));

        // check token request
        HttpRequest r = client.requests().get(0);
        assertEquals("POST", r.method());
        assertEquals(API_TEST_SERVICE_URL + "/oauth2/token", r.uri().toString());

        assertTrue(res.productInventoryStatus().isPresent());
        ProductInventoryStatus status = res.productInventoryStatus().get();
        assertEquals("123", status.productId());
        assertTrue(status.quantity() >= 99);
    }

}
