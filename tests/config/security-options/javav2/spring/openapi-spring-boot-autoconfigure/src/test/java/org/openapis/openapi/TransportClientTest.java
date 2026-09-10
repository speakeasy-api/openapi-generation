package org.openapis.openapi;

import java.io.IOException;
import java.io.InputStream;
import java.util.concurrent.atomic.AtomicInteger;

import org.junit.jupiter.api.Test;
import org.springframework.boot.autoconfigure.AutoConfigurations;
import org.springframework.boot.test.context.runner.ApplicationContextRunner;

import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.transport.HttpRequest;
import org.openapis.openapi.utils.transport.HttpResponse;

import static org.assertj.core.api.Assertions.assertThat;

public class TransportClientTest {

    @Test
    public void testSpringContextClosesHttpClientExactlyOnce() {
        AtomicInteger closeCount = new AtomicInteger();
        HTTPClient client = new HTTPClient() {
            @Override
            public HttpResponse<InputStream> send(HttpRequest request) throws IOException {
                throw new UnsupportedOperationException("No request expected");
            }

            @Override
            public void close() {
                closeCount.incrementAndGet();
            }
        };

        new ApplicationContextRunner()
            .withConfiguration(AutoConfigurations.of(OpenapiAutoConfig.class))
            .withBean(HTTPClient.class, () -> client)
            .withBean(SecuritySource.class, () -> () -> null)
            .run(context -> {
                assertThat(context).hasSingleBean(SDK.class);
                assertThat(closeCount.get()).isZero();
            });

        assertThat(closeCount.get()).isEqualTo(1);
    }
}
