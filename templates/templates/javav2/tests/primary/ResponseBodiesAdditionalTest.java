package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.openapis.openapi.CommonHelpers.recordTest;

import java.io.UncheckedIOException;
import java.lang.reflect.Method;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

import org.apache.commons.io.IOUtils;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.SDK.Builder;
import org.openapis.openapi.models.operations.ResponseBodyAdditionalPropertiesAnyPostResponse;
import org.openapis.openapi.models.operations.ResponseBodyAdditionalPropertiesAnyPostResponseBody;
import org.openapis.openapi.models.operations.ResponseBodyBytesGetResponse;
import org.openapis.openapi.models.operations.JsonResponse;
import org.openapis.openapi.models.shared.JsonEvent;
import org.openapis.openapi.utils.EventStream;
import org.openapis.openapi.utils.ResponseTooLargeException;
import org.openapis.openapi.models.operations.ResponseBodyStringGetResponse;
import org.openapis.openapi.models.shared.ObjWithAnyAdditionalProperties;
import org.openapis.openapi.utils.Hook.AfterSuccess;
import org.openapis.openapi.utils.Hook.BeforeRequest;
import org.openapis.openapi.utils.Hook.SdkInit;
import org.openapis.openapi.utils.Hooks;
import org.openapis.openapi.utils.Utils;


public class ResponseBodiesAdditionalTest {

    @Test
    void testEventStreamsAreNotCapped() throws Exception {
        SDK s = SDK.builder().maxResponseBodyBytes(16).serverURL(Helpers.HTTPBIN_URL).build();

        JsonResponse res = s.eventstreams().json().call();
        try (EventStream<JsonEvent> events = res.events()) {
            assertEquals("Hello world!", events.stream()
                    .map(x -> x.data().content())
                    .collect(Collectors.joining()));
        }
    }

    @Test
    void testMaxResponseBodyBytesMustBePositive() {
        assertThrows(IllegalArgumentException.class, () -> SDK.builder().maxResponseBodyBytes(0));
        assertThrows(IllegalArgumentException.class, () -> SDK.builder().maxResponseBodyBytes(-1));
    }

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
    void testResponseBodyLimit() throws Exception {
        recordTest("response-body-limit");

        SDK capped = SDK.builder().serverURL(Helpers.HTTPBIN_URL).maxResponseBodyBytes(50).build();

        ResponseBodyBytesGetResponse res = capped.responseBodies().responseBodyBytesGet().call();
        assertEquals(200, res.statusCode());
        ResponseTooLargeException tooLarge = assertThrows(ResponseTooLargeException.class,
                () -> IOUtils.toByteArray(res.bytes().get()));
        assertEquals(50, tooLarge.maxBytes());

        org.openapis.openapi.models.operations.async.ResponseBodyBytesGetResponse asyncRes =
                capped.async().responseBodies().responseBodyBytesGet().call().get();
        ExecutionException asyncThrown = assertThrows(ExecutionException.class,
                () -> asyncRes.bytes().get().toByteArray().get());
        assertTrue(asyncThrown.getCause() instanceof ResponseTooLargeException, String.valueOf(asyncThrown.getCause()));

        SDK roomy = SDK.builder()
                .serverURL(Helpers.HTTPBIN_URL)
                .maxResponseBodyBytes(200)
                .enableHTTPDebugLogging(true)
                .build();
        ResponseBodyBytesGetResponse ok = roomy.responseBodies().responseBodyBytesGet().call();
        assertEquals(100, IOUtils.toByteArray(ok.bytes().get()).length);

        SDK tiny = SDK.builder().maxResponseBodyBytes(8).build();
        UncheckedIOException parsed = assertThrows(UncheckedIOException.class,
                () -> tiny.retries().retriesGet().requestId(UUID.randomUUID().toString()).call());
        assertTrue(parsed.getCause() instanceof ResponseTooLargeException, String.valueOf(parsed.getCause()));
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
                .additionalProperty("key1", Map.of("foo", "bar", "subkey1", Map.of("foo", "bar"))) //
                .additionalProperty("key4", Map.of("foo", "bar")) //
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
