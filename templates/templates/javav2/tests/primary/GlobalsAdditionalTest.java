package org.openapis.openapi;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.models.operations.*;

import static org.junit.jupiter.api.Assertions.*;
import static org.openapis.openapi.Helpers.HTTPBIN_URL;

public class GlobalsAdditionalTest {

    @Test
    void testGlobalsQueryParameterGetUsesGlobal() throws Exception {
        CommonHelpers.recordTest("globals-query-parameter-get-uses-global");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalQueryParam("test").build();
        assertNotNull(s);

        GlobalsQueryParameterGetResponse res = s.globals().globalsQueryParameterGet().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals("test", res.res().get().args().globalQueryParam());
    }

    @Test
    void testGlobalsQueryParameterGetUsesLocal() throws Exception {
        CommonHelpers.recordTest("globals-query-parameter-get-uses-local");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalQueryParam("test").build();
        assertNotNull(s);

        GlobalsQueryParameterGetResponse res =
                s.globals().globalsQueryParameterGet().globalQueryParam("local").call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals("local", res.res().get().args().globalQueryParam());
    }

    @Test
    void testGlobalPathParameterGetUsesGlobal() throws Exception {
        CommonHelpers.recordTest("globals-path-parameter-get-uses-global");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalPathParam(1L).build();
        assertNotNull(s);

        GlobalPathParameterGetResponse res = s.globals().globalPathParameterGet().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(HTTPBIN_URL + "/anything/globals/pathParameter/1",
                res.res().get().url());
    }

    @Test
    void testGlobalPathParameterGetUsesLocal() throws Exception {
        CommonHelpers.recordTest("globals-path-parameter-get-uses-local");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalPathParam(1L).build();
        assertNotNull(s);

        GlobalPathParameterGetResponse res =
                s.globals().globalPathParameterGet().globalPathParam(2L).call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals(HTTPBIN_URL + "/anything/globals/pathParameter/2",
                res.res().get().url());
    }

    @Test
    void testGlobalHeaderGetUsesGlobal() throws Exception {
        CommonHelpers.recordTest("globals-header-get-uses-global");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalHeaderParam(true).build();
        assertNotNull(s);

        GlobalsHeaderGetResponse res = s.globals().globalsHeaderGet().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals("true", res.res().get().headers().get().get("Globalheaderparam"));
    }

    @Test
    void testGlobalHeaderGetUsesLocal() throws Exception {
        CommonHelpers.recordTest("globals-header-get-uses-local");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).globalHeaderParam(true).build();
        assertNotNull(s);

        GlobalsHeaderGetResponse res =
                s.globals().globalsHeaderGet().globalHeaderParam(false).call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals("false", res.res().get().headers().get().get("Globalheaderparam"));
    }

    @Test
    void testGlobalsHiddenPost() throws Exception {
        CommonHelpers.recordTest("globals-hidden-post");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
                .globalHiddenQueryParam("hello")
                .globalHiddenHeaderParam("world")
                .globalHiddenPathParam("test")
                .build();

        var body = GlobalsHiddenPostRequestBody.builder()
                .test("friend")
                .other(37)
                .build();

        var res = s.globals().globalsHiddenPost(body);

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertEquals("hello", res.res().get().args().globalHiddenQueryParam());
        assertEquals("friend", res.res().get().json().test());
        assertEquals(37, res.res().get().json().other());
        assertEquals("world", res.res().get().headers().get("Globalhiddenheaderparam"));
        assertEquals(HTTPBIN_URL + "/anything/globals/hidden/test?globalHiddenQueryParam=hello",
                res.res().get().url());
    }

    @Test
    void testGlobalsOperationParamsOnly() throws Exception {
        CommonHelpers.recordTest("globals-operation-params-only");

        // Initialize SDK with ALL global parameters
        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL)
                .globalQueryParam("globalQueryValue")
                .globalPathParam(999L)
                .globalHeaderParam(true)
                .globalHiddenQueryParam("hiddenQueryValue")
                .globalHiddenHeaderParam("hiddenHeaderValue")
                .globalHiddenPathParam("hiddenPathValue")
                .build();

        assertNotNull(s);

        // Call operation with operation-specific parameters
        GlobalsOperationScopedExclusiveResponse res = s.globals()
                .globalsOperationScopedExclusive()
                .operationQueryParam("operationQuery")
                .operationPathParam("operationPath")
                .operationHeaderParam("operationHeader")
                .call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.res().isPresent());
        var payload = res.res().get();

        // Verify that ONLY operation-specific parameters are present in the request
        // Query params should NOT contain global params
        var args = payload.args();
        assertFalse(args.containsKey("globalQueryParam"),
                "Global query param should NOT be included when operation defines its own params");
        assertFalse(args.containsKey("globalHiddenQueryParam"),
                "Global hidden query param should NOT be included when operation defines its own params");

        // URL should contain operation path param, NOT global path param
        var url = payload.url();
        assertEquals(HTTPBIN_URL + "/anything/globals/operationScopedExclusive/operationPath?operationQueryParam=operationQuery", url,
                "URL should use operation-specific path parameter, not global");

        // Headers should NOT contain global headers
        var headers = payload.headers();
        assertFalse(headers.containsKey("Globalheaderparam"),
                "Global header param should NOT be included when operation defines its own params");
        assertFalse(headers.containsKey("Globalhiddenheaderparam"),
                "Global hidden header param should NOT be included when operation defines its own params");
    }

    @Test
    void testGlobalsKebabCaseParamGet() throws Exception {
        CommonHelpers.recordTest("globals-kebab-case-param-get");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).kebabCaseParam("kebab-case-value").build();
        assertNotNull(s);

        GlobalsKebabCaseParamGetResponse res = s.globals().globalsKebabCaseParamGet().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());
        assertTrue(res.res().isPresent());
        assertEquals("kebab-case-value", res.res().get().args().get("kebab-case-param"));
    }
}
