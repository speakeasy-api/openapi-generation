package org.openapis.openapi;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.*;
import org.openapis.openapi.models.shared.DeepObject;
import org.openapis.openapi.models.shared.OpenEnum;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.JSON;
import org.openapis.openapi.utils.Utils;

import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.HashMap;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.openapis.openapi.Helpers.HTTPBIN_URL;

public class ParameterAdditionalTest {

    @Test
    public void testPathEncoding() throws Exception {
        CommonHelpers.recordTest("parameters-path-encoding");

        // Note because our test server currently returns unencoded urls in a response it is of no use
        // to confirm url encoding. For this reason we specify a dummy client (that does nothing over
        // the network and always returns an empty 200 response).
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL) //
                .client(new HTTPClient() {
                    @Override
                    public HttpResponse<InputStream> send(HttpRequest request) {
                        return CommonHelpers.createJsonResponse(request, 200, "{\"url\":\"\"}");
                    }
                }) //
                .build();

        // a mix of unreserved, reserved, and must-be-encoded characters (space)
        String param1 = "abc 123:/?#[]@!$&'()*+,;=";

        // param2 has `x-speakeasy-param-encoding: allowReserved`
        // a mix of unreserved, reserved, and must-be-encoded characters (space)
        String param2 = "def ghi:/?#[]@!$&'()*+,;=";
        PathEncodingResponse res = s.parameters().pathEncoding(param1, param2);

        // inspect and test the request uri (param1 will be encoded, param2 will not have encoded reserved chars)
        URI uri = res.rawResponse().uri();
        assertEquals(HTTPBIN_URL + "/anything/pathencoding/abc%20123%3A%2F%3F%23%5B%5D%40%21%24%26%27%28%29*%2B%2C%3B%3D/def%20ghi:/?#[]@!$&'()*+,;=", uri.toString());
    }

    @Test
    public void testQueryEncoding() throws Exception {
        CommonHelpers.recordTest("parameters-query-encoding");

        // Note because our test server currently returns unencoded urls in a response it is of no use
        // to confirm url encoding. For this reason we specify a dummy client (that does nothing over
        // the network and always returns an empty 200 response).
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL) //
                .client(new HTTPClient() {
                    @Override
                    public HttpResponse<InputStream> send(HttpRequest request) {
                        return CommonHelpers.createJsonResponse(request, 200, "{\"url\":\"\"}");
                    }
                }) //
                .build();

        // a mix of unreserved, reserved, and must-be-encoded characters (space)
        String param1 = "abc 123:/?#[]@!$&'()*+,;=";

        QueryEncodingResponse res = s.parameters().queryEncoding(param1);

        // inspect and test the request uri (param1 will be encoded, param2 will not have encoded reserved chars)

        // note that space in a query parameter could be + or %20. Speakeasy is standardizing on %20.
        URI uri = res.rawResponse().uri();
        assertEquals(HTTPBIN_URL + "/anything/queryencoding?param1=abc%20123:/?#[]@!$&'()*+,;=", uri.toString());
    }

    @Test
    void testPathParameterJson() throws Exception {
        CommonHelpers.recordTest("parameters-path-parameter-json");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        PathParameterJsonResponse res = s.parameters()
                .pathParameterJson(Helpers.createSimpleObject());

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(
                HTTPBIN_URL + "/anything/pathParams/json/{\"any\":\"any\",\"bool\":true,\"boolOpt\":true,\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.000000001Z\",\"enum\":\"one\",\"float32\":1.1,\"int\":1,\"int32\":1,\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\",\"strOpt\":\"testOptional\"}",
                res.res().get().url());
    }

    @Test
    void testJsonQueryParamsObject() throws Exception {
        CommonHelpers.recordTest("parameters-json-query-params-object");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        SimpleObject simpleObj = Helpers.createSimpleObject();
        DeepObject deepObj = Helpers.createDeepObject();

        JsonQueryParamsObjectResponse res = s.parameters()
                .jsonQueryParamsObject(deepObj, simpleObj);

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(
                HTTPBIN_URL + "/anything/queryParams/json/obj?deepObjParam={\"any\"%3A{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}%2C\"arr\"%3A[{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}%2C{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}]%2C\"bool\"%3Atrue%2C\"int\"%3A1%2C\"map\"%3A{\"key\"%3A{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}}%2C\"num\"%3A1.1%2C\"obj\"%3A{\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}%2C\"str\"%3A\"test\"}&simpleObjParam={\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}",
                res.res().get().url());
        assertEquals(JSON.getMapper().writeValueAsString(simpleObj), res.res().get().args().simpleObjParam());
        assertEquals(JSON.getMapper().writeValueAsString(deepObj), res.res().get().args().deepObjParam());
    }

    @Test
    void testMixedQueryParams() throws Exception {
        CommonHelpers.recordTest("parameters-mixed-query-params");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        SimpleObject obj = Helpers.createSimpleObject();

        MixedQueryParamsResponse res = s.parameters()
                .mixedQueryParams(obj, obj, obj);

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertNotNull(res.res().get());
        assertEquals(
                HTTPBIN_URL + "/anything/queryParams/mixed?deepObjectParam[any]=any&deepObjectParam[bool]=true&deepObjectParam[boolOpt]=true&deepObjectParam[date]=2020-01-01&deepObjectParam[dateTime]=2020-01-01T00%3A00%3A00.000000001Z&deepObjectParam[enum]=one&deepObjectParam[float32]=1.1&deepObjectParam[int]=1&deepObjectParam[int32]=1&deepObjectParam[int32Enum]=55&deepObjectParam[intEnum]=2&deepObjectParam[num]=1.1&deepObjectParam[str]=test&deepObjectParam[strOpt]=testOptional&any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.000000001Z&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&num=1.1&str=test&strOpt=testOptional&jsonParam={\"any\"%3A\"any\"%2C\"bool\"%3Atrue%2C\"boolOpt\"%3Atrue%2C\"date\"%3A\"2020-01-01\"%2C\"dateTime\"%3A\"2020-01-01T00%3A00%3A00.000000001Z\"%2C\"enum\"%3A\"one\"%2C\"float32\"%3A1.1%2C\"int\"%3A1%2C\"int32\"%3A1%2C\"int32Enum\"%3A55%2C\"intEnum\"%3A2%2C\"num\"%3A1.1%2C\"str\"%3A\"test\"%2C\"strOpt\"%3A\"testOptional\"}",
                res.res().get().url());
        assertEquals(new HashMap<String, String>() {
            {
                put("any", "any");
                put("bool", "true");
                put("boolOpt", "true");
                put("date", "2020-01-01");
                put("dateTime", "2020-01-01T00:00:00.000000001Z");
                put("deepObjectParam[any]", "any");
                put("deepObjectParam[bool]", "true");
                put("deepObjectParam[boolOpt]", "true");
                put("deepObjectParam[date]", "2020-01-01");
                put("deepObjectParam[dateTime]", "2020-01-01T00:00:00.000000001Z");
                put("deepObjectParam[enum]", "one");
                put("deepObjectParam[float32]", "1.1");
                put("deepObjectParam[int]", "1");
                put("deepObjectParam[int32]", "1");
                put("deepObjectParam[int32Enum]", "55");
                put("deepObjectParam[intEnum]", "2");
                put("deepObjectParam[num]", "1.1");
                put("deepObjectParam[str]", "test");
                put("deepObjectParam[strOpt]", "testOptional");
                put("enum", "one");
                put("float32", "1.1");
                put("int", "1");
                put("int32", "1");
                put("int32Enum", "55");
                put("intEnum", "2");
                put("jsonParam", JSON.getMapper().writeValueAsString(obj));
                put("num", "1.1");
                put("str", "test");
                put("strOpt", "testOptional");
            }
        }, res.res().get().args());
    }

    @Test
    public void testParameters_ParameterOpenEnum() throws Exception {
        Utils.recordTest("parameters-open-enum");

        var testHttpClient = Utils.createTestHTTPClient("parameterOpenEnum");
        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
                .client(testHttpClient)
                .build();

        ParameterOpenEnumResponse res = sdk.parameters().parameterOpenEnum()
                .paramH(OpenEnum.ONE_HUNDRED_AND_ONE)
                .paramP(OpenEnum.ONE_HUNDRED_AND_ONE)
                .paramQ(OpenEnum.FOUR_HUNDRED_AND_FOUR)
                .call();
        assertEquals(200, res.statusCode());
        assertThat(res.res())
                .isPresent()
                .get()
                .satisfies(r -> {
                    assertEquals(HTTPBIN_URL + "/anything/openEnum/101/suffix?param-q=404", r.url());
                    assertThat(r.headers()).containsEntry("Param-H", "101");
                });
    }
}

