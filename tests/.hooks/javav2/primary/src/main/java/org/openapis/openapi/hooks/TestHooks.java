package org.openapis.openapi.hooks;

import org.apache.hc.core5.net.URIBuilder;
import org.openapis.openapi.SDKConfiguration;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.Helpers;
import org.openapis.openapi.utils.Hook.*;
import org.openapis.openapi.utils.Blob;

import java.io.IOException;
import java.io.InputStream;
import java.net.URISyntaxException;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.Builder;
import java.net.http.HttpResponse;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;

final class TestHooks implements BeforeRequest, AfterError, AfterSuccess, SdkInit {

    private volatile String initSdkVersion = "";

    @Override
    public SDKConfiguration sdkInit(SDKConfiguration config) {
        initSdkVersion = SDKConfiguration.SDK_VERSION;
        config.setClient(new TestClient(config.client()));
        return config;
    }

    @Override
    public HttpRequest beforeRequest(BeforeRequestContext context, HttpRequest request) throws Exception {
        Builder b = Helpers.copy(request);
        b.header("Idempotency-Key", "some-key");
        if ("testHooks".equals(context.operationId())) {
            URIBuilder ub = new URIBuilder(request.uri());
            ub.setParameter("someParam", "overriddenParam");
            b.uri(ub.build());
        } else if ("authorizationHeaderModification".equals(context.operationId())) {
            // overwrite existing header by using setHeader
            request
                    .headers()
                    .firstValue("Authorization")
                    .ifPresent(value -> b.setHeader("Authorization", value + " modified"));
        } else if ("hooksCustomUserAgent".equals(context.operationId())) {
            b.setHeader("User-Agent", "acme-corp/" + SDKConfiguration.SDK_VERSION + " acme-corp/" + System.getProperty("java.version"));
            b.setHeader("X-Test-Gen-Version", SDKConfiguration.GEN_VERSION);
            b.setHeader("X-Test-Doc-Version", SDKConfiguration.OPENAPI_DOC_VERSION);
            b.setHeader("X-Test-Init-Sdk-Version", initSdkVersion);
        }
        return b.build();
    }

    @Override
    public HttpResponse<InputStream> afterSuccess(AfterSuccessContext context, HttpResponse<InputStream> response)
            throws Exception {
        if ("testHooksAfterResponse".equals(context.operationId())) {
            throw new RuntimeException("validation failed");
        } else {
            return response;
        }
    }

    @Override
    public HttpResponse<InputStream> afterError(AfterErrorContext context,
                                                Optional<HttpResponse<InputStream>> response, Optional<Exception> error) throws Exception {
        if ("testHooksError".equals(context.operationId())) {
            if (response.isPresent() && response.get().statusCode() != 400) {
                throw new IllegalStateException("expected status code 400");
            } else {
                throw new IllegalStateException("special test error case");
            }
        }
        if ("statusGetDefaultError".equals(context.operationId())) {
            if (response.isPresent() && response.get().statusCode() == 418) {
                return new HttpResponse<InputStream>() {
                    @Override
                    public int statusCode() {
                        return 200;
                    }

                    @Override
                    public HttpRequest request() {
                        return response.get().request();
                    }

                    @Override
                    public Optional<HttpResponse<InputStream>> previousResponse() {
                        return Optional.empty();
                    }

                    @Override
                    public java.net.http.HttpHeaders headers() {
                        return response.get().headers();
                    }

                    @Override
                    public InputStream body() {
                        return InputStream.nullInputStream();
                    }

                    @Override
                    public Optional<javax.net.ssl.SSLSession> sslSession() {
                        return Optional.empty();
                    }

                    @Override
                    public java.net.URI uri() {
                        return response.get().uri();
                    }

                    @Override
                    public java.net.http.HttpClient.Version version() {
                        return response.get().version();
                    }
                };
            }
        }
        if (error.isPresent()) {
            throw error.get();
        } else {
            return response.get();
        }
    }

    static final class TestClient implements HTTPClient {

        private final HTTPClient client;

        TestClient(HTTPClient client) {
            this.client = client;
        }

        @Override
        public HttpResponse<InputStream> send(HttpRequest request)
                throws IOException, InterruptedException, URISyntaxException {
            Builder b = Helpers.copy(request);
            b.header("Client-Level-Header", "added by client");
            return client.send(b.build());
        }

        @Override
        public CompletableFuture<HttpResponse<Blob>> sendAsync(HttpRequest request) {
            Builder b = Helpers.copy(request);
            b.header("Client-Level-Header", "added by client");
            return client.sendAsync(b.build());
        }
    }
}
