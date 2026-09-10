package org.openapis.openapi;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.*;
import org.openapis.openapi.models.shared.Base64InputFileModeRequest;
import org.openapis.openapi.models.shared.Base64InputFileModeResponse;
import org.openapis.openapi.models.shared.BinaryString;
import org.openapis.openapi.models.shared.NullableObject;
import org.openapis.openapi.models.shared.JvmParamLimitRequestBody;
import org.openapis.openapi.models.shared.JvmParamLimitResponseBody;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.utils.Blob;
import org.openapis.openapi.utils.JSON;

import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.Base64;
import java.util.HashMap;

import static org.junit.jupiter.api.Assertions.*;

public class RequestBodiesAdditionalTest {

    @Test
    void testRequestBodyPostMultipleContentTypesComponentFiltered() throws Exception {
        CommonHelpers.recordTest(
                "request-bodies-post-multiple-content-types-component-filtered-application-json");
        CommonHelpers.recordTest(
                "request-bodies-post-multiple-content-types-component-filtered-multipart-form-data");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        SimpleObject obj = Helpers.createSimpleObject();

        RequestBodyPostMultipleContentTypesComponentFilteredResponse res =
                s.requestBodies().requestBodyPostMultipleContentTypesComponentFiltered(obj);

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        Helpers.assertSimpleObject(res.res().get().json());
    }

    @Test
    void testRequestBodyPostMultipleContentTypesSplitMultipartWithParam() throws Exception {
        CommonHelpers.recordTest(
                "request-bodies-post-multiple-content-types-split-multipart-with-param");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        RequestBodyPostMultipleContentTypesSplitParamMultipartRequestBody formData =
                new RequestBodyPostMultipleContentTypesSplitParamMultipartRequestBody(true, 1.1,
                        "test body");
        RequestBodyPostMultipleContentTypesSplitParamMultipartResponse res = s.requestBodies()
                .requestBodyPostMultipleContentTypesSplitParamMultipart(formData, "test param");

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(new HashMap<>() {
            {
                put("bool2", "true");
                put("num2", "1.1");
                put("str2", "test body");
            }
        }, res.res().get().form().get());
        assertEquals(new HashMap<>() {
            {
                put("paramStr", "test param");
            }
        }, res.res().get().args().get());
    }

    @Test
    void testRequestBodyPutMultipartFile() throws Exception {
        CommonHelpers.recordTest("request-bodies-put-multipart-file");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        byte[] data = Helpers.getData();

        RequestBodyPutMultipartFileResponse res = s.requestBodies()
                .requestBodyPutMultipartFile(RequestBodyPutMultipartFileRequestBody.builder()
                        .file(RequestBodyPutMultipartFileFile.builder().content(Blob.from(data)).fileName("testUpload.json").build())
                        .build());

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(new HashMap<String, String>() {
            {
                put("file", new String(data, StandardCharsets.UTF_8));
            }
        }, res.res().get().files());
    }

    @Test
    void testRequestBodyPutMultipartFileRef() throws Exception {
        CommonHelpers.recordTest("request-bodies-put-multipart-file-ref");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        byte[] data = Helpers.getData();

        RequestBodyPutMultipartFileRefResponse res = s.requestBodies()
                .requestBodyPutMultipartFileRef(RequestBodyPutMultipartFileRefRequestBody.builder()
                        .file(BinaryString.builder().content(Blob.from(data)).fileName("testUpload.json").build())
                        .build());

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(new HashMap<String, String>() {
            {
                put("file", new String(data, StandardCharsets.UTF_8));
            }
        }, res.res().get().files());
    }

    @Test
    void testRequestBodyPutBytes() throws Exception {
        CommonHelpers.recordTest("request-bodies-put-bytes");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        byte[] data = Helpers.getData();
        RequestBodyPutBytesResponse res = s.requestBodies().requestBodyPutBytes(Blob.from(data));

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(new String(data, StandardCharsets.UTF_8), res.res().get().data());
    }

    @Test
    void testRequestBodiesBase64FileInputIdempotent() throws Exception {
        CommonHelpers.recordTest("request-bodies-base64-file-input-idempotent");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        String preEncoded = Base64.getEncoder().encodeToString(new byte[] {
                (byte) 0xff, (byte) 0xfe, 0x00, 0x01,
                'b', 'i', 'n', 'a', 'r', 'y',
                (byte) 0xc3, 0x28});

        Base64InputFileModeRequest request = Base64InputFileModeRequest.builder()
                .dataByte(preEncoded)
                .dataContentEncoding(preEncoded)
                .dataPlain("plain-idempotent")
                .build();

        PostBase64InputModeResponse first = s.requestBodies().postBase64InputMode(request);
        assertNotNull(first);
        assertEquals(200, first.statusCode());
        Base64InputFileModeResponse firstJson = first.res().get().json();
        assertEquals(preEncoded, firstJson.dataByte());
        assertEquals(preEncoded, firstJson.dataContentEncoding());
        assertEquals("plain-idempotent", firstJson.dataPlain());

        PostBase64InputModeResponse second = s.requestBodies().postBase64InputMode(request);
        assertNotNull(second);
        assertEquals(200, second.statusCode());
        assertEquals(firstJson, second.res().get().json());
    }

    @Test
    void testRequestBodyPostNullableRequiredProperty_AllNull() throws Exception {
        // relates to "request-bodies-post-nullable-required-property-all-null" test

        // Note this test is spread across three tests:
        // testRequestBodyPostNullableRequiredProperty_AllNull
        // testRequestBodyPostNullableRequiredProperty_NotRequiredAbsent
        // testRequestBodyPostNullableRequiredProperty_AllPresent

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        NullableRequiredPropertyPostRequestBody req = NullableRequiredPropertyPostRequestBody
                .builder().nullableOptionalInt(null)
                .nullableRequiredArray(null).nullableRequiredEnum(null)
                .nullableRequiredInt(null).nullableRequiredDateTime(null)
                .nullableRequiredBigIntStr(null)
                .nullableRequiredDecimalStr(null).build();
        NullableRequiredPropertyPostResponse res =
                s.requestBodies().nullableRequiredPropertyPost().request(req).call();
        assertEquals(200, res.statusCode());

        assertEquals(
                "{\"NullableOptionalInt\":null,\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                res.object().get().data().get());
    }

    @Test
    void testRequestBodyPostNullableRequiredProperty_NotRequiredAbsent() throws Exception {
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        NullableRequiredPropertyPostRequestBody req = NullableRequiredPropertyPostRequestBody
                .builder()
                .nullableRequiredArray(null).nullableRequiredEnum(null)
                .nullableRequiredInt(null).nullableRequiredDateTime(null)
                .nullableRequiredBigIntStr(null)
                .nullableRequiredDecimalStr(null).build();
        NullableRequiredPropertyPostResponse res =
                s.requestBodies().nullableRequiredPropertyPost().request(req).call();
        assertEquals(200, res.statusCode());

        assertEquals(
                "{\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                res.object().get().data().get());
    }

    @Test
    void testRequestBodyPostNullableRequiredSharedObject_AllNull() throws Exception {
        // Note: this test is spread across three tests:
        // testRequestBodyPostNullableRequiredSharedObject_AllNull
        // testRequestBodyPostNullableRequiredSharedObject_NotRequiredAbsent
        // testRequestBodyPostNullableRequiredSharedObject_AllPresent
        CommonHelpers.recordTest("request-bodies-post-nullable-required-shared-object-all-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        NullableRequiredSharedObjectPostRequestBody req =
                NullableRequiredSharedObjectPostRequestBody.builder()
                        .nullableOptionalObj(null)
                        .nullableRequiredObj(null).build();
        NullableRequiredSharedObjectPostResponse res =
                s.requestBodies().nullableRequiredSharedObjectPost().request(req).call();
        assertEquals(200, res.statusCode());
        ObjectMapper m = JSON.getMapper();

        assertEquals("{\"NullableOptionalObj\":null,\"NullableRequiredObj\":null}",
                res.object().get().data().get());
    }

    @Test
    void testRequestBodyPostNullableRequiredSharedObject_NotRequiredAbsent() throws Exception {
        CommonHelpers
                .recordTest("request-bodies-post-nullable-required-shared-object-required-null");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        NullableRequiredSharedObjectPostRequestBody req =
                NullableRequiredSharedObjectPostRequestBody.builder()
                        .nullableRequiredObj(null).build();
        NullableRequiredSharedObjectPostResponse res =
                s.requestBodies().nullableRequiredSharedObjectPost().request(req).call();
        assertEquals(200, res.statusCode());
        ObjectMapper m = JSON.getMapper();

        assertEquals("{\"NullableRequiredObj\":null}", res.object().get().data().get());
    }

    @Test
    void testRequestBodyPostNullableRequiredSharedObject_AllPresent() throws Exception {
        // CommonHelpers.recordTest("request-bodies-post-nullable-required-shared-object");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        NullableRequiredSharedObjectPostRequestBody req =
                NullableRequiredSharedObjectPostRequestBody.builder()
                        .nullableRequiredObj(
                                NullableObject.builder().optional("hello").required(123L).build())
                        .build();
        NullableRequiredSharedObjectPostResponse res =
                s.requestBodies().nullableRequiredSharedObjectPost().request(req).call();
        assertEquals(200, res.statusCode());

        assertEquals("{\"NullableRequiredObj\":{\"optional\":\"hello\",\"required\":123}}",
                res.object().get().data().get());
    }

    @Test
    @Disabled
    void testRequestBodiesFileUploadExtractContentType() throws Exception {
        CommonHelpers.recordTest("request-bodies-file-upload-extract-content-type");
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        var req = "a,b,c".getBytes(StandardCharsets.UTF_8);
        // TODO allow content-type of request to be overridden from default
        // application/octet-stream to text/csv
        var res = s.requestBodies()
                .requestBodyPostWildcard()
                .request(Blob.from(req))
                .call();
        assertEquals("text/csv", res.rawResponse().request().headers().firstValue("content-type").get());
        assertEquals("text/csv", res.rawResponse().headers().firstValue("content-type").get());
    }

    /**
     * Populate every field on the wide base JvmParamLimitObject section of the
     * builder. Caller still sets variant-specific fields (missingStr,
     * responseOnlyConst) separately.
     */
    private static void populateWideBase(JvmParamLimitRequestBody.Builder b,
            java.util.Map<String, String> reqMap) throws Exception {
        b.requiredStr("req")
                .requiredNullableStr("notNullYet")
                .requiredStrMap(reqMap)
                .decimal(new java.math.BigDecimal("2.71828"))
                .decimalStr(new java.math.BigDecimal("3.14159"))
                .bigint(new java.math.BigInteger("88888888888888888888"))
                .bigIntStr(new java.math.BigInteger("99999999999999999999"))
                .stringArray(java.util.Arrays.asList("alpha", "beta", "gamma"))
                .defaultStr("custom")
                .myCasedField("kebabSnakeVal")
                .class_("reservedKeywordVal")
                .closedEnum(org.openapis.openapi.models.shared.Enum.TWO)
                .openEnum(org.openapis.openapi.models.shared.EnumUsedInRequestExplicitlyOpen.BETA)
                .nestedObj(Helpers.createSimpleObject())
                .discUnion(org.openapis.openapi.models.shared.TaggedObject1.builder()
                        .tag(org.openapis.openapi.models.shared.TaggedObject1Tag.TAG1)
                        .imageURL("https://example.com/img.png")
                        .build())
                .nonDiscUnion(org.openapis.openapi.models.shared.WeaklyTypedOneOfObject
                        .of(Helpers.createSimpleObject()))
                .additionalProperty("extra1", "x1")
                .additionalProperty("extra2", "x2");
        for (int i = 1; i <= 240; i++) {
            String name = String.format("field%03d", i);
            JvmParamLimitRequestBody.Builder.class
                    .getMethod(name, String.class)
                    .invoke(b, "v" + i);
        }
    }

    @Test
    void testJvmParamLimitRoundtrip() throws Exception {
        // Tests schema with >250 properties (see overlays/primary/javav2/overlay.yaml).
        // Would trip the JVM 255 method-parameter limit on the all-args constructor
        // so generator falls back to a no-args ctor plus field-injection deserialization.
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        java.util.Map<String, String> reqMap = new java.util.LinkedHashMap<>();
        reqMap.put("k1", "v1");
        reqMap.put("k2", "v2");
        JvmParamLimitRequestBody.Builder b = JvmParamLimitRequestBody.builder()
                .missingStr("present")
                .responseOnlyConst("fromResponse");
        populateWideBase(b, reqMap);

        JvmParamLimitResponse res = s.requestBodies().jvmParamLimit(b.build());

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res());
        JvmParamLimitResponseBody json = res.res().get().json();
        assertNotNull(json);
        for (int i = 1; i <= 240; i++) {
            String name = String.format("field%03d", i);
            String expected = "v" + i;
            @SuppressWarnings("unchecked")
            java.util.Optional<String> actual = (java.util.Optional<String>)
                    JvmParamLimitResponseBody.class.getMethod(name).invoke(json);
            assertEquals(expected, actual.orElse(null),
                    "roundtrip mismatch for " + name);
        }
        assertEquals("req", json.requiredStr());
        assertEquals("constVal", json.requiredConstStr());
        assertEquals("notNullYet", json.requiredNullableStr().orElse(null));
        assertEquals(reqMap, json.requiredStrMap());
        assertEquals(0, new java.math.BigDecimal("2.71828").compareTo(
                json.decimal().orElse(null)),
                "decimal (JSON number) BigDecimal must roundtrip");
        assertEquals(0, new java.math.BigDecimal("3.14159").compareTo(
                json.decimalStr().orElse(null)),
                "decimalStr (JSON string) BigDecimal must roundtrip");
        assertEquals(new java.math.BigInteger("88888888888888888888"),
                json.bigint().orElse(null),
                "bigint (JSON number) BigInteger > Long.MAX_VALUE must roundtrip");
        assertEquals(new java.math.BigInteger("99999999999999999999"),
                json.bigIntStr().orElse(null),
                "bigIntStr (JSON string) BigInteger > Long.MAX_VALUE must roundtrip");
        assertEquals(java.util.Arrays.asList("alpha", "beta", "gamma"),
                json.stringArray().orElse(null));
        assertEquals("custom", json.defaultStr().orElse(null));
        assertEquals("x1", json.additionalProperties().get("extra1"));
        assertEquals("x2", json.additionalProperties().get("extra2"));
        assertEquals("present", json.missingStr());
        assertEquals("fromResponse", json.responseOnlyConst().orElse(null));
        assertEquals("kebabSnakeVal", json.myCasedField(),
                "wire 'my-cased_field' must bind to sanitized myCasedField setter");
        assertEquals("reservedKeywordVal", json.class_().orElse(null),
                "wire 'class' must bind to sanitized class_ setter");
        assertEquals(org.openapis.openapi.models.shared.Enum.TWO,
                json.closedEnum().orElse(null),
                "closed enum should roundtrip");
        assertEquals(org.openapis.openapi.models.shared.EnumUsedInRequestExplicitlyOpen.BETA,
                json.openEnum().orElse(null),
                "open enum should roundtrip");
        assertNotNull(json.nestedObj().orElse(null),
                "nested SimpleObject should deserialize via Builder setter");
        Helpers.assertSimpleObject(json.nestedObj().get());
        assertNotNull(json.discUnion().orElse(null),
                "discriminated union variant should deserialize");
        assertTrue(json.discUnion().get() instanceof org.openapis.openapi.models.shared.TaggedObject1,
                "discriminator 'tag1' should resolve to TaggedObject1");
        assertNotNull(json.nonDiscUnion().orElse(null),
                "non-discriminated union variant should deserialize");
    }

    @Test
    void testJvmParamLimit_missingRequiredAtDeser_throws() throws Exception {
        // Asymmetric schema: missingStr is OPTIONAL in request but REQUIRED in
        // response. Client omits it; httpbin echoes JSON without it; SDK must
        // throw at response deserialization.
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        java.util.Map<String, String> reqMap = new java.util.LinkedHashMap<>();
        reqMap.put("k1", "v1");
        JvmParamLimitRequestBody.Builder b = JvmParamLimitRequestBody.builder();
        populateWideBase(b, reqMap);
        // missingStr intentionally NOT set — request schema allows omission.

        assertThrows(Exception.class,
                () -> s.requestBodies().jvmParamLimit(b.build()),
                "deserialization of response with missing required field must fail");
    }

    @Test
    void testJvmParamLimit_constOverrideRejected() throws Exception {
        // Asymmetric schema: responseOnlyConst is a plain string in request but
        // CONST "fromResponse" in response. Client sends a non-const value;
        // httpbin echoes it; SDK must keep the const singleton, NOT overwrite
        // with the wire value — Jackson must not be able to mutate the field.
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        java.util.Map<String, String> reqMap = new java.util.LinkedHashMap<>();
        reqMap.put("k1", "v1");
        JvmParamLimitRequestBody.Builder b = JvmParamLimitRequestBody.builder()
                .missingStr("present")
                .responseOnlyConst("MUTATED_BY_SERVER");
        populateWideBase(b, reqMap);

        JvmParamLimitResponseBody json = s.requestBodies()
                .jvmParamLimit(b.build())
                .res().get().json();
        assertEquals("fromResponse", json.responseOnlyConst().orElse(null),
                "const field on response must not be overridable from wire value");
    }

}
