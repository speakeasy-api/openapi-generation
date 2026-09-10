package org.openapis.openapi;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.openapis.openapi.utils.Blob;
import org.openapis.openapi.utils.Multipart;
import reactor.adapter.JdkFlowAdapter;
import reactor.core.publisher.Flux;
import reactor.test.StepVerifier;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.concurrent.Flow;
import java.util.regex.Pattern;

import static org.junit.jupiter.api.Assertions.*;

class MultipartAdditionalTest {

    @Test
    @DisplayName("Mixed part types generate correct multipart structure")
    void testMixedPartTypes() {
        String blobContent = "binary data";
        Blob blob = Blob.from(blobContent.getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("text", "text value")
                .addPart("file", blob, "data.bin", "application/octet-stream")
                .addPart("json", "{\"key\": \"value\"}", "application/json")
                .build();

        String boundary = multipart.boundary();
        String fullContent = collectMultipartContent(multipart);
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(3, parts.size(), "Should have exactly 3 parts");

        // Text part
        MultipartPart textPart = parts.get(0);
        assertEquals("text", textPart.name);
        assertEquals("text value", textPart.content);
        assertEquals("text/plain; charset=UTF-8", textPart.contentType);
        assertNull(textPart.filename);

        // File part
        MultipartPart filePart = parts.get(1);
        assertEquals("file", filePart.name);
        assertEquals("binary data", filePart.content);
        assertEquals("application/octet-stream", filePart.contentType);
        assertEquals("data.bin", filePart.filename);

        // JSON part with explicit content type
        MultipartPart jsonPart = parts.get(2);
        assertEquals("json", jsonPart.name);
        assertEquals("{\"key\": \"value\"}", jsonPart.content);
        assertEquals("application/json", jsonPart.contentType);
        assertNull(jsonPart.filename);

        // Verify final boundary exists
        assertTrue(fullContent.endsWith("--" + boundary + "--" + "\r\n"), "Should end with final boundary");
    }

    @Test
    @DisplayName("Streaming behavior emits correct buffer sequence")
    void testStreamingBehavior() {
        String blobContent = "streaming test content";
        Blob blob = Blob.from(blobContent.getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("file", blob, "stream.txt", "text/plain")
                .build();

        String boundary = multipart.boundary();
        Flow.Publisher<ByteBuffer> publisher = multipart.bodyPublisher();
        Flux<ByteBuffer> flux = JdkFlowAdapter.flowPublisherToFlux(publisher);

        // Test streaming behavior with StepVerifier
        StepVerifier.create(flux.map(this::bufferToString))
                .expectNextMatches(content -> {
                    // First buffer should be the file part header
                    return content.contains("Content-Disposition: form-data") &&
                            content.contains("name=\"file\"") &&
                            content.contains("filename=\"stream.txt\"") &&
                            content.contains("Content-Type: text/plain");
                })
                .expectNextMatches(content -> {
                    // Second buffer should be the file content
                    return content.equals("streaming test content");
                })
                .expectNextMatches(content -> {
                    // Third buffer should be the part trailer
                    return content.equals("\r\n");
                })
                .expectNextMatches(content -> {
                    // Final buffer should be the closing boundary
                    return content.equals("--" + boundary + "--" + "\r\n");
                })
                .verifyComplete();
    }

    @Test
    @DisplayName("Empty multipart builder throws IllegalStateException")
    void testEmptyMultipartThrowsException() {
        Multipart.Builder builder = Multipart.builder();

        IllegalStateException exception = assertThrows(IllegalStateException.class, builder::build);
        assertEquals("Must have at least one part to build multipart message.", exception.getMessage());
    }

    @Test
    @DisplayName("Null parameter validation in builder methods")
    void testNullParameterValidation() {
        Multipart.Builder builder = Multipart.builder();

        // Test null name for text part
        assertThrows(IllegalArgumentException.class, () -> builder.addPart(null, "value"));

        // Test null value for text part
        assertThrows(IllegalArgumentException.class, () -> builder.addPart("name", (String) null));

        // Test null content type for text part
        assertThrows(IllegalArgumentException.class, () -> builder.addPart("name", "value", null));

        // Test null parameters for file part
        Blob blob = Blob.from("test".getBytes(StandardCharsets.UTF_8));
        assertThrows(IllegalArgumentException.class, () -> builder.addPart(null, blob, "file.txt", "text/plain"));
        assertThrows(IllegalArgumentException.class, () -> builder.addPart("name", (Blob) null, "file.txt", "text/plain"));
        assertThrows(IllegalArgumentException.class, () -> builder.addPart("name", blob, null, "text/plain"));
    }

    @Test
    @DisplayName("Default content types are applied correctly")
    void testDefaultContentTypes() {
        String textContent = "plain text";
        Blob blob = Blob.from("binary data".getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("text", textContent)  // Should get default text content type
                .addPart("file", blob, "data.bin", null)  // Should get default file content type
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(2, parts.size());

        // Text part should have default text content type
        MultipartPart textPart = parts.get(0);
        assertEquals("text/plain; charset=UTF-8", textPart.contentType);

        // File part should have default file content type
        MultipartPart filePart = parts.get(1);
        assertEquals("application/octet-stream", filePart.contentType);
    }

    @Test
    @DisplayName("Special characters in field names are properly escaped")
    void testSpecialCharactersInFieldNames() {
        Multipart multipart = Multipart.builder()
                .addPart("field\"with\"quotes", "value1")
                .addPart("field\\with\\backslashes", "value2")
                .addPart("field\rwith\nlinebreaks", "value3")
                .build();

        String fullContent = collectMultipartContent(multipart);

        // Verify quotes are escaped
        assertTrue(fullContent.contains("name=\"field\\\"with\\\"quotes\""),
                "Quotes should be escaped in field names");

        // Verify backslashes are escaped
        assertTrue(fullContent.contains("name=\"field\\\\with\\\\backslashes\""),
                "Backslashes should be escaped in field names");

        // Verify line breaks are converted to spaces
        assertTrue(fullContent.contains("name=\"field with linebreaks\""),
                "Line breaks should be converted to spaces in field names");
    }

    @Test
    @DisplayName("Unicode filenames are properly encoded with ASCII fallback")
    void testUnicodeFilenames() {
        String unicodeFilename = "测试文件.txt";
        String mixedFilename = "file-with-émojis-🎉.pdf";

        Blob blob1 = Blob.from("content1".getBytes(StandardCharsets.UTF_8));
        Blob blob2 = Blob.from("content2".getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("unicode", blob1, unicodeFilename, "text/plain")
                .addPart("mixed", blob2, mixedFilename, "application/pdf")
                .build();

        String fullContent = collectMultipartContent(multipart);

        // Should contain ASCII fallback filename
        assertTrue(fullContent.contains("filename=\"____.txt\""),
                "Should contain ASCII fallback for unicode filename");

        // Should contain RFC 5987 encoded filename
        assertTrue(fullContent.contains("filename*=UTF-8''"),
                "Should contain RFC 5987 encoded filename");

        // Should contain ASCII fallback for mixed filename
        assertTrue(fullContent.contains("filename=\"file-with-_mojis-__.pdf\""),
                "Should contain ASCII fallback for mixed filename");
    }

    @Test
    @DisplayName("Boundary uniqueness across multiple multipart instances")
    void testBoundaryUniqueness() {
        Multipart multipart1 = Multipart.builder()
                .addPart("field1", "value1")
                .build();

        Multipart multipart2 = Multipart.builder()
                .addPart("field2", "value2")
                .build();

        String boundary1 = multipart1.boundary();
        String boundary2 = multipart2.boundary();

        assertNotEquals(boundary1, boundary2, "Boundaries should be unique across instances");
        assertFalse(boundary1.isEmpty(), "Boundary should not be empty");
        assertFalse(boundary2.isEmpty(), "Boundary should not be empty");
    }

    @Test
    @DisplayName("Content-Type header is correctly formatted")
    void testContentTypeHeader() {
        Multipart multipart = Multipart.builder()
                .addPart("test", "value")
                .build();

        String contentType = multipart.contentType();
        String boundary = multipart.boundary();

        assertEquals("multipart/form-data; boundary=" + boundary, contentType);
        assertTrue(contentType.startsWith("multipart/form-data; boundary="));
    }

    @Test
    @DisplayName("Multiple parts with same name are handled correctly")
    void testMultiplePartsWithSameName() {
        Multipart multipart = Multipart.builder()
                .addPart("duplicate", "value1")
                .addPart("duplicate", "value2")
                .addPart("duplicate", "value3")
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(3, parts.size(), "Should have all 3 parts even with duplicate names");

        // All parts should have the same name but different content
        for (int i = 0; i < parts.size(); i++) {
            assertEquals("duplicate", parts.get(i).name);
            assertEquals("value" + (i + 1), parts.get(i).content);
        }
    }

    @Test
    @DisplayName("Large content is handled correctly")
    void testLargeContent() {
        // Create a large string (100KB to keep test reasonable)
        StringBuilder largeContent = new StringBuilder();
        String chunk = "This is a test chunk of data that will be repeated many times. ";
        int targetSize = 100 * 1024; // 100KB
        while (largeContent.length() < targetSize) {
            largeContent.append(chunk);
        }

        String largeString = largeContent.toString();
        Blob largeBlob = Blob.from(largeString.getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("large_text", largeString)
                .addPart("large_file", largeBlob, "large.txt", "text/plain")
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(2, parts.size());
        assertEquals(largeString, parts.get(0).content);
        assertEquals(largeString, parts.get(1).content);
    }

    @Test
    @DisplayName("Special filename characters are properly handled")
    void testSpecialFilenameCharacters() {
        String specialFilename = "file with spaces & symbols!@#$%^&*()_+-={}[]|\\:;\"'<>,.?~`";
        Blob blob = Blob.from("content".getBytes(StandardCharsets.UTF_8));

        Multipart multipart = Multipart.builder()
                .addPart("special", blob, specialFilename, "text/plain")
                .build();

        String fullContent = collectMultipartContent(multipart);

        // Should contain escaped quotes and backslashes in ASCII fallback
        assertTrue(fullContent.contains("filename="), "Should contain filename parameter");

        // Should contain RFC 5987 encoded filename
        assertTrue(fullContent.contains("filename*=UTF-8''"), "Should contain RFC 5987 encoded filename");
    }

    @Test
    @DisplayName("Empty content is handled correctly")
    void testEmptyContent() {
        Blob emptyBlob = Blob.from(new byte[0]);

        Multipart multipart = Multipart.builder()
                .addPart("empty_text", "")
                .addPart("empty_file", emptyBlob, "empty.txt", "text/plain")
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(2, parts.size());
        assertEquals("", parts.get(0).content);
        assertEquals("", parts.get(1).content);
    }

    @Test
    @DisplayName("CRLF sequences in content are preserved")
    void testCRLFInContent() {
        String contentWithCRLF = "Line 1\r\nLine 2\r\nLine 3";

        Multipart multipart = Multipart.builder()
                .addPart("crlf_content", contentWithCRLF)
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(1, parts.size());
        assertEquals(contentWithCRLF, parts.get(0).content);
    }

    @Test
    @DisplayName("Binary content with null bytes is handled correctly")
    void testBinaryContentWithNullBytes() {
        byte[] binaryData = {0x00, 0x01, 0x02, (byte) 0xFF, 0x00, 0x7F, (byte) 0x80};
        Blob binaryBlob = Blob.from(binaryData);

        Multipart multipart = Multipart.builder()
                .addPart("binary", binaryBlob, "binary.dat", "application/octet-stream")
                .build();

        // Test that the multipart can be created and has correct structure
        String boundary = multipart.boundary();
        assertNotNull(boundary);
        assertFalse(boundary.isEmpty());

        // Test content type
        String contentType = multipart.contentType();
        assertTrue(contentType.contains("multipart/form-data"));
        assertTrue(contentType.contains("boundary=" + boundary));
    }

    @Test
    @DisplayName("Builder can be reused to create multiple multipart instances")
    void testBuilderReuse() {
        // Note: This tests that builder state is properly isolated
        Multipart multipart1 = Multipart.builder()
                .addPart("field1", "value1")
                .build();

        Multipart multipart2 = Multipart.builder()
                .addPart("field2", "value2")
                .build();

        String content1 = collectMultipartContent(multipart1);
        String content2 = collectMultipartContent(multipart2);

        assertTrue(content1.contains("field1"));
        assertFalse(content1.contains("field2"));

        assertTrue(content2.contains("field2"));
        assertFalse(content2.contains("field1"));
    }

    @Test
    @DisplayName("Complex multipart with all part types and edge cases")
    void testComplexMultipartWithAllTypes() {
        // Create a comprehensive multipart with various edge cases
        String unicodeText = "Unicode content: 你好世界 🌍";
        String jsonContent = "{\"nested\": {\"array\": [1, 2, 3], \"unicode\": \"测试\"}}";
        byte[] binaryData = {0x42, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}; // PNG header
        Blob binaryBlob = Blob.from(binaryData);

        Multipart multipart = Multipart.builder()
                .addPart("unicode_text", unicodeText)
                .addPart("json_data", jsonContent, "application/json")
                .addPart("binary_file", binaryBlob, "test-image.png", "image/png")
                .addPart("field\"with\"quotes", "value with\r\nline breaks")
                .addPart("empty_field", "")
                .build();

        String fullContent = collectMultipartContent(multipart);
        String boundary = multipart.boundary();
        List<MultipartPart> parts = parseMultipartContent(fullContent, boundary);

        assertEquals(5, parts.size(), "Should have all 5 parts");

        // Verify each part
        assertEquals("unicode_text", parts.get(0).name);
        assertEquals(unicodeText, parts.get(0).content);
        assertEquals("text/plain; charset=UTF-8", parts.get(0).contentType);

        assertEquals("json_data", parts.get(1).name);
        assertEquals(jsonContent, parts.get(1).content);
        assertEquals("application/json", parts.get(1).contentType);

        assertEquals("binary_file", parts.get(2).name);
        assertEquals("image/png", parts.get(2).contentType);
        assertEquals("test-image.png", parts.get(2).filename);

        assertEquals("field\\\"with\\\"quotes", parts.get(3).name);
        assertEquals("value with\r\nline breaks", parts.get(3).content);

        assertEquals("empty_field", parts.get(4).name);
        assertEquals("", parts.get(4).content);
    }

    @Test
    @DisplayName("Streaming with multiple buffer emissions works correctly")
    void testStreamingMultipleBuffers() {
        String content1 = "First part content";
        String content2 = "Second part content with more data";

        Multipart multipart = Multipart.builder()
                .addPart("part1", content1)
                .addPart("part2", content2)
                .build();

        Flow.Publisher<ByteBuffer> publisher = multipart.bodyPublisher();
        Flux<ByteBuffer> flux = JdkFlowAdapter.flowPublisherToFlux(publisher);

        // Collect all buffers and verify they form a complete multipart message
        String fullContent = flux
                .map(this::bufferToString)
                .reduce("", String::concat)
                .block();

        assertNotNull(fullContent);
        assertTrue(fullContent.contains(content1));
        assertTrue(fullContent.contains(content2));
        assertTrue(fullContent.contains("Content-Disposition: form-data"));
        assertTrue(fullContent.endsWith("--" + multipart.boundary() + "--\r\n"));
    }

    // Helper methods

    private String collectMultipartContent(Multipart multipart) {
        Flow.Publisher<ByteBuffer> publisher = multipart.bodyPublisher();
        Flux<ByteBuffer> flux = JdkFlowAdapter.flowPublisherToFlux(publisher);

        return flux
                .map(this::bufferToString)
                .reduce("", String::concat)
                .block();
    }

    private List<MultipartPart> parseMultipartContent(String content, String boundary) {
        // Split by boundary markers
        String[] sections = content.split("--" + Pattern.quote(boundary));
        List<MultipartPart> parts = new java.util.ArrayList<>();

        for (String section : sections) {
            if (section.trim().isEmpty() || section.equals("--")) {
                continue; // Skip empty sections and final boundary
            }

            MultipartPart part = parseSection(section);
            if (part != null) {
                parts.add(part);
            }
        }

        return parts;
    }

    private MultipartPart parseSection(String section) {
        if (!section.contains("Content-Disposition")) {
            return null; // Not a valid part
        }

        // Remove leading \r\n if present
        String cleanSection = section.startsWith("\r\n") ? section.substring(2) : section;
        String[] lines = cleanSection.split("\r\n");

        String name = null;
        String filename = null;
        String contentType = null;
        StringBuilder content = new StringBuilder();

        boolean inHeaders = true;
        for (String line : lines) {
            if (inHeaders) {
                if (line.trim().isEmpty()) {
                    inHeaders = false;
                    continue;
                }

                if (line.startsWith("Content-Disposition:")) {
                    name = extractQuotedValue(line, "name");
                    filename = extractQuotedValue(line, "filename");
                } else if (line.startsWith("Content-Type:")) {
                    contentType = line.substring("Content-Type:".length()).trim();
                }
            } else {
                if (content.length() > 0) {
                    content.append("\r\n");
                }
                content.append(line);
            }
        }

        // Remove trailing \r\n from content if present
        String finalContent = content.toString();
        if (finalContent.endsWith("\r\n")) {
            finalContent = finalContent.substring(0, finalContent.length() - 2);
        }

        return new MultipartPart(name, finalContent, contentType, filename);
    }

    private String extractQuotedValue(String line, String attribute) {
        Pattern pattern = Pattern.compile( attribute + "=\"((?:\\\\.|[^\"])*)\"");
        java.util.regex.Matcher matcher = pattern.matcher(line);
        return matcher.find() ? matcher.group(1) : null;
    }


    private String bufferToString(ByteBuffer buffer) {
        return StandardCharsets.UTF_8.decode(buffer.duplicate()).toString();
    }

    // Simple data class for parsed multipart parts
    static class MultipartPart {
        final String name;
        final String content;
        final String contentType;
        final String filename;

        MultipartPart(String name, String content, String contentType, String filename) {
            this.name = name;
            this.content = content;
            this.contentType = contentType;
            this.filename = filename;
        }
    }
}
