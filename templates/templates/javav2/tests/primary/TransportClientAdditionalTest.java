package org.openapis.openapi;

import static org.junit.jupiter.api.Assertions.assertFalse;

import org.junit.jupiter.api.Test;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.SpeakeasyHTTPClient;

/**
 * Control for the okhttp AutoCloseable lifecycle feature: the JDK transport
 * has nothing to close (java.net.http.HttpClient uses daemon threads and,
 * pre-Java-21, exposes no close()), so the generator must not emit
 * AutoCloseable on the transport or SDK surface for JDK-transport variants.
 * Stream utilities (EventStream, JsonLStream) implement AutoCloseable
 * unconditionally and are intentionally out of scope here.
 */
public class TransportClientAdditionalTest {

    @Test
    public void testJdkTransportSurfaceIsNotAutoCloseable() {
        assertFalse(AutoCloseable.class.isAssignableFrom(HTTPClient.class),
                "JDK-transport HTTPClient must not extend AutoCloseable");
        assertFalse(AutoCloseable.class.isAssignableFrom(SpeakeasyHTTPClient.class),
                "JDK-transport SpeakeasyHTTPClient must not implement AutoCloseable");
        assertFalse(AutoCloseable.class.isAssignableFrom(SDK.class),
                "JDK-transport sync SDK must not implement AutoCloseable");
        assertFalse(AutoCloseable.class.isAssignableFrom(AsyncSDK.class),
                "JDK-transport async SDK must not implement AutoCloseable");
    }
}
