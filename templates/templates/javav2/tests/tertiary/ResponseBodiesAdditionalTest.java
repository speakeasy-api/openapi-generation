package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.tertiary.openapi.CommonHelpers.recordTest;

import java.lang.reflect.Method;
import java.util.concurrent.atomic.AtomicInteger;

import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.openapis.tertiary.openapi.SDK.Builder;
import org.openapis.tertiary.openapi.models.operations.ResponseBodyAdditionalPropertiesAnyPostResponse;
import org.openapis.tertiary.openapi.models.operations.ResponseBodyAdditionalPropertiesAnyPostResponseBody;
import org.openapis.tertiary.openapi.models.operations.ResponseBodyBytesGetResponse;
import org.openapis.tertiary.openapi.models.operations.ResponseBodyStringGetResponse;
import org.openapis.tertiary.openapi.models.components.ObjWithAnyAdditionalProperties;
import org.openapis.tertiary.openapi.utils.Hook.AfterSuccess;
import org.openapis.tertiary.openapi.utils.Hook.BeforeRequest;
import org.openapis.tertiary.openapi.utils.Hook.SdkInit;
import org.openapis.tertiary.openapi.utils.Hooks;
import org.openapis.tertiary.openapi.utils.Java8Compat;
import org.openapis.tertiary.openapi.utils.Utils;

public class ResponseBodiesAdditionalTest {

    @Test
    void testResponseBodyBytesGet() throws Exception {
        recordTest("response-bodies-bytes-get");

        SDK s = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();
        assertNotNull(s);

        ResponseBodyBytesGetResponse res = s.responseBodies().responseBodyBytesGet().call();

        assertNotNull(res);
        assertEquals(200, res.statusCode());

        byte[] bytes = new byte[100];
        assertTrue(res.bytes().isPresent());
        IOUtils.readFully(res.bytes().get(), bytes, 0, 100); // throws if 100 bytes not  streamed.
        int bytesRead = res.bytes().get().read();
        assertEquals(-1, bytesRead);
    }
    
    @Test
    void testClientInitAndBeforeRequestAndAfterSuccessHooks() throws Exception {
        
        AtomicInteger a = new AtomicInteger();
        AtomicInteger b = new AtomicInteger();
        AtomicInteger c = new AtomicInteger();
        
        SdkInit sdkInit = data -> {a.incrementAndGet(); return data;};
        BeforeRequest beforeRequest = (context, request)  -> {b.incrementAndGet(); return request;};
        AfterSuccess afterSuccess = (context, response) -> {c.incrementAndGet(); return response;};
        
        Hooks hooks = new Hooks();
        hooks.registerSdkInit(sdkInit);
        hooks.registerBeforeRequest(beforeRequest);
        hooks.registerAfterSuccess(afterSuccess);
        
        Builder builder = SDK.builder().serverURL(Helpers.HTTPBIN_URL);
        
        // override hooks using reflection
        Method m = builder.getClass().getDeclaredMethod("_hooks", Hooks.class);
        m.setAccessible(true);
        m.invoke(builder, hooks);
        
        SDK s = builder.build();
        
        // clientInit has been called
        assertEquals(1, a.get());
        
        ResponseBodyStringGetResponse res = s.responseBodies().responseBodyStringGet().call();
        assertEquals(200, res.statusCode());
        
        assertEquals(1, b.get());
        assertEquals(1, c.get());
        
        res = s.responseBodies().responseBodyStringGet().call();
        
        assertEquals(1, a.get());
        assertEquals(2, b.get());
        assertEquals(2, c.get());
    }

    @Test
    public void testResponseBodies_ResponseBodyAdditionalPropertiesAnyPost() throws Exception {
        // TODO doesn't match behaviour of generated test (null property value)
        Utils.recordTest("response-bodies-additional-properties-any-values");

        SDK sdk = SDK.builder().serverURL(Helpers.HTTPBIN_URL).build();

        ObjWithAnyAdditionalProperties req = ObjWithAnyAdditionalProperties.builder() //
                .normalField("normal") //
                .additionalProperty("key1", "value2") //
                .additionalProperty("key2", null) //
                .additionalProperty("key3", Java8Compat.mapOf("foo", "bar", "subkey1", Java8Compat.mapOf("foo", "bar"))) //
                .additionalProperty("key4", Java8Compat.listOf("foo", "bar")) //
                .build();

        ResponseBodyAdditionalPropertiesAnyPostResponse res = sdk.responseBodies().responseBodyAdditionalPropertiesAnyPost()
                .request(req)
                .call();

        assertEquals(
            ResponseBodyAdditionalPropertiesAnyPostResponseBody.builder()
                .json(req)
                .build(),
            res.object().get());
    }
}
