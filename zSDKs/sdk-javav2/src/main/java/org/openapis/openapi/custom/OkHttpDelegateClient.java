package org.openapis.openapi.custom;

import okhttp3.Headers;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import okhttp3.ResponseBody;
import org.openapis.openapi.utils.Blob;
import org.openapis.openapi.utils.HTTPClient;
import org.openapis.openapi.utils.Helpers;
import org.openapis.openapi.utils.ResponseWithBody;

import javax.net.ssl.SSLSession;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpHeaders;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

/**
 * An HTTPClient implementation backed by OkHttp. Lets you set
 * connect / read / write / call timeouts independently.
 */
public class OkHttpDelegateClient implements HTTPClient {

    private final OkHttpClient delegate;
    private boolean debugEnabled;

    public OkHttpDelegateClient(Duration connectTimeout,
                                Duration readTimeout,
                                Duration writeTimeout,
                                Duration callTimeout) {
        this.delegate = new OkHttpClient.Builder()
                .connectTimeout(connectTimeout.toMillis(), TimeUnit.MILLISECONDS)
                .readTimeout(readTimeout.toMillis(), TimeUnit.MILLISECONDS)
                .writeTimeout(writeTimeout.toMillis(), TimeUnit.MILLISECONDS)
                .callTimeout(callTimeout.toMillis(), TimeUnit.MILLISECONDS)
                .build();
    }

    @Override
    public void enableDebugLogging(boolean enabled) {
        this.debugEnabled = enabled;
    }

    @Override
    public boolean isDebugLoggingEnabled() {
        return debugEnabled;
    }

    @Override
    public HttpResponse<InputStream> send(HttpRequest request) throws IOException {
        Request okRequest = toOkHttpRequest(request);
        Response okResponse = delegate.newCall(okRequest).execute();
        return toJdkResponse(request, okResponse);
    }

    @Override
    public CompletableFuture<HttpResponse<Blob>> sendAsync(HttpRequest request) {
        Request okRequest = toOkHttpRequest(request);
        CompletableFuture<HttpResponse<Blob>> future = new CompletableFuture<>();
        delegate.newCall(okRequest).enqueue(new okhttp3.Callback() {
            @Override
            public void onFailure(okhttp3.Call call, IOException e) {
                future.completeExceptionally(e);
            }

            @Override
            public void onResponse(okhttp3.Call call, Response okResponse) {
                try {
                    HttpResponse<InputStream> jdkResponse = toJdkResponse(request, okResponse);
                    InputStream body = jdkResponse.body();
                    Blob blob = body == null ? Blob.from(new byte[]{}) : Blob.from(body);
                    future.complete(new ResponseWithBody<>(jdkResponse, blob));
                } catch (Throwable t) {
                    future.completeExceptionally(t);
                }
            }
        });
        return future;
    }

    private Request toOkHttpRequest(HttpRequest request) {
        Request.Builder builder = new Request.Builder().url(request.uri().toString());

        // Determine content type for body
        String contentType = request.headers().firstValue("Content-Type").orElse(null);

        // Build request body if applicable
        RequestBody body = null;
        boolean hasBody = request.bodyPublisher().isPresent()
                && request.bodyPublisher().get().contentLength() != 0;
        if (hasBody) {
            byte[] bytes = Helpers.bodyBytes(request);
            MediaType mediaType = contentType == null ? null : MediaType.parse(contentType);
            body = RequestBody.create(bytes, mediaType);
        } else if (methodRequiresBody(request.method())) {
            body = RequestBody.create(new byte[]{}, null);
        }

        builder.method(request.method(), body);

        // Copy headers (skip JDK-managed restricted ones if any slipped in)
        request.headers().map().forEach((name, values) -> {
            if (name.equalsIgnoreCase("Content-Length")) {
                return; // OkHttp computes this
            }
            for (String v : values) {
                builder.addHeader(name, v);
            }
        });

        return builder.build();
    }

    private static boolean methodRequiresBody(String method) {
        return method.equals("POST") || method.equals("PUT") || method.equals("PATCH");
    }

    private HttpResponse<InputStream> toJdkResponse(HttpRequest originalRequest, Response okResponse) {
        int status = okResponse.code();
        URI uri = originalRequest.uri();

        // Build JDK HttpHeaders from OkHttp headers
        Map<String, List<String>> headerMap = new LinkedHashMap<>();
        Headers okHeaders = okResponse.headers();
        for (String name : okHeaders.names()) {
            headerMap.put(name, okHeaders.values(name));
        }
        HttpHeaders jdkHeaders = HttpHeaders.of(headerMap, (k, v) -> true);

        // Body: take the InputStream from OkHttp; closing the stream closes the response.
        ResponseBody respBody = okResponse.body();
        InputStream bodyStream = respBody == null ? InputStream.nullInputStream() : respBody.byteStream();

        return new SimpleHttpResponse<>(status, originalRequest, jdkHeaders, bodyStream, uri);
    }

    /** Minimal HttpResponse implementation backed by an InputStream. */
    private static final class SimpleHttpResponse<T> implements HttpResponse<T> {
        private final int statusCode;
        private final HttpRequest request;
        private final HttpHeaders headers;
        private final T body;
        private final URI uri;

        SimpleHttpResponse(int statusCode, HttpRequest request, HttpHeaders headers, T body, URI uri) {
            this.statusCode = statusCode;
            this.request = request;
            this.headers = headers;
            this.body = body;
            this.uri = uri;
        }

        @Override public int statusCode() { return statusCode; }
        @Override public HttpRequest request() { return request; }
        @Override public Optional<HttpResponse<T>> previousResponse() { return Optional.empty(); }
        @Override public HttpHeaders headers() { return headers; }
        @Override public T body() { return body; }
        @Override public Optional<SSLSession> sslSession() { return Optional.empty(); }
        @Override public URI uri() { return uri; }
        @Override public HttpClient.Version version() { return HttpClient.Version.HTTP_1_1; }
    }
}
