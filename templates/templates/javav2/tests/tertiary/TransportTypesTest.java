package org.openapis.tertiary.openapi;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.Test;
import org.openapis.tertiary.openapi.utils.Headers;
import org.openapis.tertiary.openapi.utils.Java8Compat;
import org.openapis.tertiary.openapi.utils.transport.HttpBody;
import org.openapis.tertiary.openapi.utils.transport.HttpRequest;
import org.openapis.tertiary.openapi.utils.transport.HttpResponse;

/**
 * Unit-level contract for the SDK-owned transport value types
 * ({@link HttpRequest} / {@link HttpResponse}) that replace {@code java.net.http}
 * on the okhttp arms. These pin the documented immutability / defensive-copy
 * semantics, header case-insensitivity, method normalization, and null checks
 * that async/retry/hook code assumes. Java 8-compatible: no {@code var},
 * {@code List.of}, or {@code Map.of}.
 */
public class TransportTypesTest {

    private static final URI URI_ = URI.create("https://example.com/v1");
    private static final List<String> JDK_ONLY_BLOB_DOC_TERMS = Java8Compat.listOf(
            "java.net.http",
            "Flow",
            "BodyPublisher",
            "asPublisher",
            "toByteArray",
            "toInputStream",
            "toFile");

    // --- builder defaults / normalization --------------------------------

    @Test
    public void testBuilderDefaultsMethodToGet() {
        HttpRequest request = HttpRequest.builder().uri(URI_).build();
        assertEquals("GET", request.method());
        assertFalse(request.body().isPresent());
    }

    @Test
    public void testBuilderUppercasesMethod() {
        HttpRequest request = HttpRequest.builder().uri(URI_).method("post").build();
        assertEquals("POST", request.method());
    }

    @Test
    public void testBuildRequiresUri() {
        assertThrows(IllegalStateException.class, () -> HttpRequest.builder().build());
    }

    @Test
    public void testBuilderPreservesBody() {
        HttpBody body = HttpBody.of("payload");
        HttpRequest request = HttpRequest.builder().uri(URI_).body(body).build();
        assertTrue(request.body().isPresent());
        assertEquals(body, request.body().get());
    }

    @Test
    public void testBuilderRejectsNulls() {
        HttpRequest.Builder b = HttpRequest.builder();
        assertThrows(IllegalArgumentException.class, () -> b.method(null));
        assertThrows(IllegalArgumentException.class, () -> b.uri(null));
        assertThrows(IllegalArgumentException.class, () -> b.header(null, "v"));
        assertThrows(IllegalArgumentException.class, () -> b.header("n", null));
        assertThrows(IllegalArgumentException.class, () -> b.body(null));
    }

    // --- header manipulation ---------------------------------------------

    @Test
    public void testHeaderAppendsValues() {
        HttpRequest request = HttpRequest.builder()
                .uri(URI_)
                .header("X-Multi", "a")
                .header("x-multi", "b")
                .build();
        assertEquals(Java8Compat.listOf("a", "b"), request.headers().get("X-Multi"));
    }

    @Test
    public void testSetHeaderReplacesCaseInsensitively() {
        HttpRequest request = HttpRequest.builder()
                .uri(URI_)
                .header("Content-Type", "text/plain")
                .header("content-type", "text/html")
                .build()
                .toBuilder()
                .setHeader("CONTENT-TYPE", "application/json")
                .build();
        assertEquals(Java8Compat.listOf("application/json"), request.headers().get("content-type"));
    }

    @Test
    public void testHeadersMapCopiesMultiValue() {
        Map<String, List<String>> src = Java8Compat.mapOf("X-Multi", Java8Compat.listOf("a", "b"));
        HttpRequest request = HttpRequest.builder().uri(URI_).headers(src).build();
        assertEquals(Java8Compat.listOf("a", "b"), request.headers().get("x-multi"));
    }

    @Test
    public void testHeaderAccessIsCaseInsensitive() {
        Headers headers = new Headers()
                .add("X-Multi", "a")
                .add("x-multi", "b");
        assertEquals(Java8Compat.listOf("a", "b"), headers.get("X-MULTI"));
        assertEquals("a", headers.firstValue("x-multi").orElse(null));
    }

    @Test
    public void testBlobDocsDescribeOkHttpSurface() throws IOException {
        String docs = new String(
                Files.readAllBytes(Paths.get("docs", "utils", "Blob.md")),
                StandardCharsets.UTF_8);

        assertTrue(docs.contains("toHttpBody()"));
        assertTrue(docs.contains("contentLength()"));
        for (String term : JDK_ONLY_BLOB_DOC_TERMS) {
            assertFalse(docs.contains(term), "OkHttp Blob docs must not mention " + term);
        }

        String readme = new String(Files.readAllBytes(Paths.get("README.md")), StandardCharsets.UTF_8);
        assertFalse(readme.contains("memory-efficient streaming and reactive processing"));
    }

    // --- defensive copy semantics ----------------------------------------

    @Test
    public void testBuilderReuseDoesNotMutateBuiltRequest() {
        HttpRequest.Builder b = HttpRequest.builder().uri(URI_).header("X", "1");
        HttpRequest request = b.build();
        b.header("X", "2");
        assertEquals(Java8Compat.listOf("1"), request.headers().get("X"));
    }

    @Test
    public void testMutatingReturnedHeadersDoesNotMutateRequest() {
        HttpRequest request = HttpRequest.builder().uri(URI_).header("X", "1").build();
        request.headers().add("X", "injected");
        request.headers().add("Y", "injected");
        assertEquals(Java8Compat.listOf("1"), request.headers().get("X"));
        assertTrue(request.headers().get("Y").isEmpty());
    }

    @Test
    public void testToBuilderIsIndependentOfOriginal() {
        HttpRequest original = HttpRequest.builder().uri(URI_).header("X", "1").build();
        HttpRequest derived = original.toBuilder().header("X", "2").setHeader("Y", "y").build();
        assertEquals(Java8Compat.listOf("1"), original.headers().get("X"));
        assertTrue(original.headers().get("Y").isEmpty());
        assertEquals(Java8Compat.listOf("1", "2"), derived.headers().get("X"));
    }

    // --- HttpResponse -----------------------------------------------------

    private static HttpResponse<String> response() {
        HttpRequest request = HttpRequest.builder().uri(URI_).build();
        Headers headers = new Headers().add("Content-Type", "application/json");
        return new HttpResponse<>(request, 200, headers, "body");
    }

    @Test
    public void testResponsePreservesMetadata() {
        HttpResponse<String> response = response();
        assertEquals(200, response.statusCode());
        assertEquals("body", response.body());
        assertEquals(URI_, response.request().uri());
        assertEquals(Java8Compat.listOf("application/json"), response.headers().get("content-type"));
    }

    @Test
    public void testContentTypeIsCaseInsensitive() {
        HttpRequest request = HttpRequest.builder().uri(URI_).build();
        Headers headers = new Headers().add("CONTENT-TYPE", "text/plain");
        HttpResponse<String> response = new HttpResponse<>(request, 200, headers, "body");
        assertEquals("text/plain", response.contentType().orElse(null));
    }

    @Test
    public void testMutatingReturnedResponseHeadersDoesNotMutateResponse() {
        HttpResponse<String> response = response();
        response.headers().add("Content-Type", "injected");
        assertEquals(Java8Compat.listOf("application/json"), response.headers().get("content-type"));
    }

    @Test
    public void testResponseConstructorRejectsNulls() {
        HttpRequest request = HttpRequest.builder().uri(URI_).build();
        Headers headers = new Headers();
        assertThrows(IllegalArgumentException.class,
                () -> new HttpResponse<>(null, 200, headers, "body"));
        assertThrows(IllegalArgumentException.class,
                () -> new HttpResponse<>(request, 200, null, "body"));
        assertThrows(IllegalArgumentException.class,
                () -> new HttpResponse<>(request, 200, headers, null));
    }
}
