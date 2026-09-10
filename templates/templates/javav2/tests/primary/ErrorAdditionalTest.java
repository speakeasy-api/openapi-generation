package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;

import java.io.InputStream;
import java.io.UncheckedIOException;
import java.net.http.HttpResponse;
import java.net.http.HttpRequest;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.errors.APIException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.JvmParamLimitError;
import org.openapis.openapi.models.operations.Shape;
import org.openapis.openapi.models.operations.StatusGetErrorResponse;
import org.openapis.openapi.models.operations.StatusGetXSpeakeasyErrorsResponse;
import org.openapis.openapi.utils.Utils;
import org.openapis.openapi.utils.HTTPClient;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonMappingException;

public class ErrorAdditionalTest {
    
    @SuppressWarnings("unused")
    @Test
    void testStatusGetErrorDefaultErrorCodes() throws Exception {
        CommonHelpers.recordTest("errors-status-get-error-default-error-codes");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // When `clientServerStatusCodesAsErrors: true` is set (default), 4XX and 5XX ranges
        // are automatically treated as errors.

        // 400 and 500 responses are explicitly defined
        APIException thrown = assertThrows(APIException.class, () -> {
            StatusGetErrorResponse res = s.errors().statusGetError(400l);
        });
        assertEquals(400, thrown.code());
        assertEquals(400, thrown.rawResponse().statusCode());

        thrown = assertThrows(APIException.class, () -> {
            StatusGetErrorResponse res = s.errors().statusGetError(500l);
        });
        assertEquals(500, thrown.code());
        assertEquals(500, thrown.rawResponse().statusCode());

        // 404 and 503 responses are undefined but still treated as errors by default
        thrown = assertThrows(APIException.class, () -> {
            StatusGetErrorResponse res = s.errors().statusGetError(404l);
        });
        assertEquals(404, thrown.code());

        thrown = assertThrows(APIException.class, () -> {
            StatusGetErrorResponse res = s.errors().statusGetError(503l);
        });
        assertEquals(503, thrown.code());
    }

    @Test
    void testStatusGetError300NonError() throws Exception {
        CommonHelpers.recordTest("errors-status-get-error300-non-error");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        StatusGetErrorResponse res = s.errors().statusGetError(300l);
        assertEquals(300, res.statusCode());
    }

    @SuppressWarnings("unused")
    @Test
    void testStatusGetXSpeakeasyErrors() throws Exception {
        CommonHelpers.recordTest("errors-status-get-error-x-speakeasy-errors");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        // 400 response is explicitly defined and is marked as an error in `x-speakeasy-errors`
        APIException thrown = assertThrows(APIException.class, () -> {
            StatusGetXSpeakeasyErrorsResponse res =
                    s.errors().statusGetXSpeakeasyErrors(400l, null);
        });
        assertTrue(thrown.message().contains("API error occurred"));
        assertEquals(400, thrown.code());
        assertEquals(400, thrown.rawResponse().statusCode());

        // 401 response is undefined but it is marked as an error in `x-speakeasy-errors`
        thrown = assertThrows(APIException.class, () -> {
            StatusGetXSpeakeasyErrorsResponse res =
                    s.errors().statusGetXSpeakeasyErrors(401l, null);
        });
        assertTrue(thrown.message().contains("API error occurred"));
        assertEquals(401, thrown.code());
        assertEquals(401, thrown.rawResponse().statusCode());

        // 402 response is undefined and is not treated as an API error since it's not listed in `x-speakeasy-errors`.
        // Instead we raise a "Unexpected status code" exception.
        thrown = assertThrows(APIException.class, () -> {
            StatusGetXSpeakeasyErrorsResponse res =
                    s.errors().statusGetXSpeakeasyErrors(402l, null);
        });
        assertTrue(thrown.message().contains("Unexpected status code received"));
        assertEquals(402, thrown.code());
        assertEquals(402, thrown.rawResponse().statusCode());

        // Both 500 and 501 responses are marked as errors since `5XX` is listed `x-speakeasy-errors`.
        org.openapis.openapi.models.errors.Error eThrown =
                assertThrows(org.openapis.openapi.models.errors.Error.class, () -> {
                    StatusGetXSpeakeasyErrorsResponse res =
                            s.errors().statusGetXSpeakeasyErrors(500l, null);
                });
        assertEquals("an error occurred", eThrown.getMessage());
        assertEquals("an error occurred", eThrown.message());
        assertTrue(eThrown.data().get().message().get().contains("an error occurred"));
        assertEquals("500", eThrown.data().get().code().get());
        assertEquals(500, eThrown.code());
        assertEquals("application/json", eThrown.headers().first("content-type").get());

        org.openapis.openapi.models.errors.StatusGetXSpeakeasyErrorsResponseBody oThrown =
                assertThrows(
                        org.openapis.openapi.models.errors.StatusGetXSpeakeasyErrorsResponseBody.class,
                        () -> {
                            StatusGetXSpeakeasyErrorsResponse res =
                                    s.errors().statusGetXSpeakeasyErrors(501l, null);
                        });
        assertTrue(oThrown.data().get().message().get().contains("an error occurred"));
        assertEquals("501", oThrown.data().get().code().get());
        assertEquals(501, oThrown.rawResponse().statusCode());
    }

    @Test
    void testStatusGetSuccessXSpeakeasyErrorsEmptyStatusCodeList() throws Exception {
        CommonHelpers.recordTest("errors-status-get-error-x-speakeasy-errors-none");

        SDK s = SDK.builder().build();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // with a dummy list, meaning all responses are treated as non-errors.

        var res200 = s.errors().statusGetNonError(200l);
        assertEquals(200, res200.statusCode());

        var res400 = s.errors().statusGetNonError(400l);
        assertEquals(400, res400.statusCode());

        var res500 = s.errors().statusGetNonError(500l);
        assertEquals(500, res500.statusCode());
    }

    @SuppressWarnings("unused")
    @Test
    void testStatusGetSuccessXSpeakeasyErrorsUnspecifiedResponses() throws Exception {
        CommonHelpers.recordTest("errors-status-get-error-x-speakeasy-errors-default");

        SDK s = SDK.builder().build();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // by marking all *unspecified* responses as errors.

        // 200 and 400 responses are explicitly defined, so they are treated as non-errors.
        var res200 = s.errors().statusGetDefaultError(200l);
        assertEquals(200, res200.statusCode());

        var res400 = s.errors().statusGetDefaultError(400l);
        assertEquals(400, res400.statusCode());

        // 404 and 500 responses are undefined, so they are treated as errors.
        APIException thrown = assertThrows(APIException.class, () -> {
            s.errors().statusGetDefaultError(404l);
        });
        assertEquals(404, thrown.code());

        thrown = assertThrows(APIException.class, () -> {
            s.errors().statusGetDefaultError(500l);
        });
        assertEquals(500, thrown.code());

        // To make sure the catch-all "default" code gets properly applied in `templateErrorStatusCodesCheck`,
        // an AfterError hook was added to Hooks/TestHook.java to recover from 418 error.
        var res418 = s.errors().statusGetDefaultError(418l);
        assertEquals(200, res418.statusCode());
    }

    @Test
    void testConnectionErrorGet() throws Exception {
        CommonHelpers.recordTest("errors-connection-error");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        UncheckedIOException thrown = assertThrows(UncheckedIOException.class, () -> {
            s.errors().connectionErrorGet().call();
        });
        assertTrue(thrown.getCause() instanceof java.net.ConnectException);
    }
    
    @Test
    public void testErrorAdditionalProperties() throws Exception {
        CommonHelpers.recordTest("errors-additional-properties");
        // TODO add test server endpoint for this test, in the meantime we just mock it
        // (which is completely fine but the test server helps standardization across
        // langs
        String json = "{\"message\":\"something happened\", \"category\": \"happening\", \"code\": \"500\"}";
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).client(new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                return CommonHelpers.createJsonResponse(request, 500, json);
            }
        }).build();

        var err = assertThrows(org.openapis.openapi.models.errors.Error.class, () -> {
            s.errors().statusGetXSpeakeasyErrors(400);
        });
        assertEquals("500", err.data().get().code().get());
        assertEquals("something happened", err.data().get().message().get());
        assertEquals("happening", err.data().get().additionalProperties().get("category"));
    }
    
    @Test
    public void testErrorMessageExtension() throws JsonMappingException, JsonProcessingException {
        String json = "{\"message\":\"something happened\", \"code\": \"500\"}";
        Error.Data err = Utils.mapper().readValue(json, Error.Data.class);
        assertEquals("something happened", err.message().get());
        assertEquals("500", err.code().get());
    }
    
    @Test
    public void testErrorResponseBodyDeserializationFails() throws Exception {
        CommonHelpers.recordTest("errors-error-body-validation-lenient");
        SDK s = SDK.builder().build();

        // The service returns a schema-invalid error body for wrong-type-error:
        // the typed string field `code` is an object, so strict deserialization
        // of the typed Error fails. Lenient best-effort still throws the typed
        // Error but leaves data()/type() absent and records the strict failure
        // on deserializationException().
        var err = assertThrows(org.openapis.openapi.models.errors.Error.class, () -> {
            s.errors().getMalformedErrorResponse()
                    .statusCode(500L)
                    .shape(Shape.WRONG_TYPE_ERROR)
                    .call();
        });
        assertTrue(err.deserializationException().isPresent());
        assertInstanceOf(JsonMappingException.class, err.deserializationException().get());
        assertFalse(err.data().isPresent());
        assertFalse(err.type().isPresent());
        // The raw payload must survive the deserialization failure so callers
        // can inspect the body the schema could not represent.
        assertEquals("{\"code\":{\"invalid\":true},\"message\":\"degraded error\"}",
                err.bodyAsString().orElse(""));
    }

    @Test
    void testJvmParamLimitError_nestedDataDeserializes() throws Exception {
        // Error component with >250 properties (see overlays/primary/javav2/overlay.yaml)
        // would trip the JVM 255 method-parameter limit on the all-args constructor, so the
        // generator falls back to a no-args ctor plus Builder-routed Jackson deserialization.
        SDK s = SDK.builder()
            .serverURL(Helpers.API_TEST_SERVICE_URL)
            .build();

        java.util.Map<String, String> reqMap = new java.util.LinkedHashMap<>();
        reqMap.put("k1", "v1");
        org.openapis.openapi.models.shared.JvmParamLimitRequestBody req =
            org.openapis.openapi.models.shared.JvmParamLimitRequestBody.builder()
            .requiredStr("req")
            .requiredNullableStr("notNullYet")
            .requiredStrMap(reqMap)
            .myCasedField("kebabSnakeVal")
            .class_("reservedKeywordVal")
            .build();

        JvmParamLimitError thrown =
                assertThrows(JvmParamLimitError.class,
                        () -> s.errors().jvmParamLimitForcedError(req),
                        "4XX response must surface as JvmParamLimitError");

        assertEquals(400, thrown.code());
        assertFalse(thrown.deserializationException().isPresent(),
                "nested Data must deserialize cleanly via builder-deser — any "
                        + "Jackson failure would land in deserializationException()");
        assertTrue(thrown.data().isPresent(),
                "data() must be populated when deser succeeds");

        JvmParamLimitError.Data data = thrown.data().get();
        assertEquals("req", data.requiredStr());
        assertEquals("constVal", data.requiredConstStr());
        assertEquals("kebabSnakeVal", data.myCasedField(),
                "sanitized field must roundtrip in the error class too");
        assertEquals("reservedKeywordVal", data.class_().orElse(null));
        assertEquals(java.util.Map.of("k1", "v1"), data.requiredStrMap(),
                "map field must roundtrip via builder-deser");
        assertTrue(data.requiredNullableStr().isPresent(),
                "nullable-but-required field must roundtrip when client sent non-null");
        assertEquals("notNullYet", data.requiredNullableStr().get());
        assertFalse(data.errorCode().isPresent(),
                "errorCode is optional in the error envelope and absent from echoed body");
    }

    @Test
    void testJvmParamLimitError_deserializationFails() throws Exception {
        // Tests payload deserialization failure for the the no-args ctor + builder-deser
        // fallback path: must surface the failure via deserializationException().
        String malformed = "{\"requiredStr\": [1,2,3]}"; // array where String required
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).client(new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) {
                return CommonHelpers.createJsonResponse(request, 400, malformed);
            }
        }).build();

        java.util.Map<String, String> reqMap = new java.util.LinkedHashMap<>();
        reqMap.put("k1", "v1");
        org.openapis.openapi.models.shared.JvmParamLimitRequestBody req =
            org.openapis.openapi.models.shared.JvmParamLimitRequestBody.builder()
            .requiredStr("req")
            .requiredNullableStr("notNullYet")
            .requiredStrMap(reqMap)
            .myCasedField("kebabSnakeVal")
            .class_("reservedKeywordVal")
            .build();

        JvmParamLimitError thrown = assertThrows(JvmParamLimitError.class,
                () -> s.errors().jvmParamLimitForcedError(req));

        assertEquals(400, thrown.code());
        assertTrue(thrown.deserializationException().isPresent(),
                "builder-deser must surface Jackson failure via deserializationException()");
        assertInstanceOf(JsonMappingException.class, thrown.deserializationException().get());
        assertFalse(thrown.data().isPresent(),
                "data() must be empty when nested deser failed");
    }
}
