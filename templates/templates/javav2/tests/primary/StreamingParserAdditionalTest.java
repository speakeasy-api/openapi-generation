package org.openapis.openapi;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;
import org.openapis.openapi.utils.EventStreamMessage;
import org.openapis.openapi.utils.StreamingParser;

import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.nio.charset.StandardCharsets;
import java.util.Optional;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;
import static org.openapis.openapi.Helpers.stringToBuffer;

public class StreamingParserAdditionalTest {

    // ===== SSE PARSING TESTS =====

    static Stream<Arguments> basicFieldsProvider() {
        return Stream.of(
                Arguments.of("event: user-message\ndata: Hello", "user-message", "Hello"),
                Arguments.of("id: 123\ndata: World", null, "World"),
                Arguments.of("retry: 5000\ndata: Test", null, "Test"),
                Arguments.of("data: Simple message", null, "Simple message")
        );
    }

    @ParameterizedTest
    @MethodSource("basicFieldsProvider")
    void testSSEParseBasicFields(String input, String expectedEvent, String expectedData) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input + "\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage message = result.get();
        
        if (expectedEvent != null) {
            assertTrue(message.event().isPresent());
            assertEquals(expectedEvent, message.event().get());
        } else {
            assertFalse(message.event().isPresent());
        }
        assertEquals(expectedData, message.data().get());
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "retry: 5000\ndata: test",
            "retry: 0\ndata: zero",
            "retry: 999999\ndata: large"
    })
    void testSSEParseValidRetry(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input + "\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage message = result.get();
        assertTrue(message.retryMs().isPresent());
        assertTrue(message.retryMs().get() >= 0);
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "retry: not-a-number\ndata: test",
            "retry: 12.5\ndata: decimal",
            "retry: \ndata: empty"
    })
    void testSSEParseInvalidRetry(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input + "\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage message = result.get();
        assertFalse(message.retryMs().isPresent());
    }

    @ParameterizedTest
    @ValueSource(strings = {
            ": This is a comment\ndata: actual data",
            "data: line1\n: comment\ndata: line2",
            ": comment1\n: comment2\ndata: after comments"
    })
    void testSSEParseWithComments(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input + "\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage message = result.get();
        assertTrue(message.data().isPresent());
        assertFalse(message.data().get().contains(": "));
    }

    static Stream<Arguments> multiLineDataProvider() {
        return Stream.of(
                Arguments.of("data: Line 1\ndata: Line 2", "Line 1\nLine 2"),
                Arguments.of("data: First\ndata: \ndata: Third", "First\n\nThird"),
                Arguments.of("data: Single", "Single"),
                Arguments.of("data: \ndata: Empty first", "\nEmpty first")
        );
    }

    @ParameterizedTest
    @MethodSource("multiLineDataProvider")
    void testSSEParseMultiLineData(String input, String expected) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input + "\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage message = result.get();
        assertEquals(expected, message.data().get());
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "data: Hello\n\n",
            "event: test\ndata: Hello\n\n",
            "id: 123\ndata: Hello\n\n"
    })
    void testSSEStreamingSingleMessage(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input));
        assertTrue(result.isPresent());
        assertEquals("Hello", result.get().data().get());
    }

    static Stream<Arguments> multipleMessagesProvider() {
        return Stream.of(
                Arguments.of("data: First\n\ndata: Second\n\n", "First", "Second"),
                Arguments.of("data: A\n\n\n\ndata: B\n\n", "A", "B"),
                Arguments.of("event: e1\ndata: D1\n\nevent: e2\ndata: D2\n\n", "D1", "D2")
        );
    }

    @ParameterizedTest
    @MethodSource("multipleMessagesProvider")
    void testSSEStreamingMultipleMessages(String input, String first, String second) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> msg1 = parser.add(stringToBuffer(input));
        assertTrue(msg1.isPresent());
        assertEquals(first, msg1.get().data().get());
        Optional<EventStreamMessage> msg2 = parser.next();
        assertTrue(msg2.isPresent());
        assertEquals(second, msg2.get().data().get());
        assertFalse(parser.next().isPresent());
    }

    @Test
    void testSSEStreamingChunkedInput() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Build message piece by piece
        assertFalse(parser.add(stringToBuffer("da")).isPresent());
        assertFalse(parser.add(stringToBuffer("ta: ")).isPresent());
        assertFalse(parser.add(stringToBuffer("Chunked")).isPresent());
        assertFalse(parser.add(stringToBuffer("\n")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("\n"));
        assertTrue(result.isPresent());
        assertEquals("Chunked", result.get().data().get());
    }

    @Test
    void testSSEStreamingVerySmallChunks() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        String message = "event: test\ndata: Small chunks\n\n";
        // Feed one character at a time
        Optional<EventStreamMessage> result = Optional.empty();
        for (int i = 0; i < message.length(); i++) {
            result = parser.add(stringToBuffer(String.valueOf(message.charAt(i))));
            if (result.isPresent()) {
                break;
            }
        }
        assertTrue(result.isPresent());
        assertEquals("Small chunks", result.get().data().get());
        assertEquals("test", result.get().event().get());
    }

    @Test
    void testSSEStreamingMixedChunkSizes() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Mix different chunk sizes
        assertFalse(parser.add(stringToBuffer("event: mixed")).isPresent());
        assertFalse(parser.add(stringToBuffer("\ndata: Different ")).isPresent());
        assertFalse(parser.add(stringToBuffer("sizes")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("\n\n"));
        assertTrue(result.isPresent());
        assertEquals("Different sizes", result.get().data().get());
        assertEquals("mixed", result.get().event().get());
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "\uFEFFdata: With BOM\n\n",
            "\uFEFFevent: test\ndata: BOM event\n\n"
    })
    void testSSEStreamingBOMHandling(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input));
        assertTrue(result.isPresent());
        assertTrue(result.get().data().isPresent());
        assertFalse(result.get().data().get().contains("\uFEFF"));
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "data: Test\r\n\r\n", // CRLF_CRLF
            "data: Test\r\n\n",   // CRLF_LF
            "data: Test\r\n\r",   // CRLF_CR
            "data: Test\n\r\n",   // LF_CRLF
            "data: Test\n\r",     // LF_CR
            "data: Test\n\n",     // LF_LF
            "data: Test\r\r\n",   // CR_CRLF
            "data: Test\r\r",     // CR_CR
    })
    void testSSEStreamingLineEndingNormalization(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input));
        assertTrue(result.isPresent());
        assertEquals("Test", result.get().data().get());
    }

    @ParameterizedTest
    @ValueSource(strings = {
            "\r\ndata: Test\r\n\r\n", // CRLF
            "\ndata: Test\n\n",       // LF
            "\rdata: Test\r\r",       // CR
    })
    void testSSEStreamingEmptyFirstLine(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(input));
        assertTrue(result.isPresent());
        assertEquals("Test", result.get().data().get());
    }

    @Test
    void testSSEStreamingInterleavedEmptyMessages() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Empty messages should be skipped
        assertFalse(parser.add(stringToBuffer("\n\n")).isPresent());
        assertFalse(parser.add(stringToBuffer("   \n\n")).isPresent());
        assertFalse(parser.add(stringToBuffer("\t\n\n")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("data: Real\n\n"));
        assertTrue(result.isPresent());
        assertEquals("Real", result.get().data().get());
    }

    @Test
    void testSSEStreamingPartialMessageAtFinish() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        parser.add(stringToBuffer("data: Incomplete"));
        Optional<EventStreamMessage> result = parser.finish();
        assertTrue(result.isPresent());
        assertEquals("Incomplete", result.get().data().get());
    }

    @Test
    void testSSEStreamingBufferStateTracking() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        assertFalse(parser.hasBufferedData());
        parser.add(stringToBuffer("data: Partial"));
        assertTrue(parser.hasBufferedData());
        parser.add(stringToBuffer("\n\n"));
        assertFalse(parser.hasBufferedData());
    }

    @Test
    void testSSEStreamingLargeMessage() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        String largeData = "data: " + "X".repeat(50000) + "\n\n";
        Optional<EventStreamMessage> result = parser.add(stringToBuffer(largeData));
        assertTrue(result.isPresent());
        assertTrue(result.get().data().isPresent());
        assertEquals(50000, result.get().data().get().length());
    }

    @Test
    void testSSEStreamingBoundaryConditions() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Test message boundary exactly at chunk boundary
        assertFalse(parser.add(stringToBuffer("data: Test\n")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("\n"));
        assertTrue(result.isPresent());
        assertEquals("Test", result.get().data().get());
        // Test double newline split across chunks
        assertFalse(parser.add(stringToBuffer("data: Split\n")).isPresent());
        result = parser.add(stringToBuffer("\n"));
        assertTrue(result.isPresent());
        assertEquals("Split", result.get().data().get());
    }

    @Test
    void testSSEStreamingMultipleMessageBoundaries() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Multiple complete messages in fragments
        Optional<EventStreamMessage> msg1 = parser.add(stringToBuffer("data: Msg1\n\ndata: "));
        assertTrue(msg1.isPresent());
        assertEquals("Msg1", msg1.get().data().get());
        Optional<EventStreamMessage> msg2 = parser.add(stringToBuffer("Msg2\n\ndata: Msg3\n\n"));
        assertTrue(msg2.isPresent());
        assertEquals("Msg2", msg2.get().data().get());
        Optional<EventStreamMessage> msg3 = parser.next();
        assertTrue(msg3.isPresent());
        assertEquals("Msg3", msg3.get().data().get());
    }

    @Test
    void testSSEComplexStreamingScenario() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Simulate realistic streaming with mixed chunk boundaries
        assertFalse(parser.add(stringToBuffer(": Comment\nevent: notif")).isPresent());
        assertFalse(parser.add(stringToBuffer("ication\nid: msg")).isPresent());
        assertFalse(parser.add(stringToBuffer("-123\nretry: 3000\ndata: {\"user\"")).isPresent());
        assertFalse(parser.add(stringToBuffer(": \"Alice\"}\ndata: {\"time")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("stamp\": \"2023\"}\n\n"));
        assertTrue(result.isPresent());
        EventStreamMessage msg = result.get();
        assertEquals("notification", msg.event().get());
        assertEquals("msg-123", msg.id().get());
        assertEquals(3000, msg.retryMs().get());
        assertEquals("{\"user\": \"Alice\"}\n{\"timestamp\": \"2023\"}", msg.data().get());
    }

    @ParameterizedTest
    @ValueSource(strings = {"", "   ", "\t\n", "\n\n\n"})
    void testSSEStreamingEmptyInputs(String input) {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        assertFalse(parser.add(stringToBuffer(input)).isPresent());
    }

    @Test
    void testSSEStreamingNullInput() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        assertFalse(parser.add((ByteBuffer) null).isPresent());
    }

    @Test
    void testSSEBOMOnlyAppliedToFirstMessage() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // First message with BOM
        Optional<EventStreamMessage> msg1 = parser.add(stringToBuffer("\uFEFFdata: First\n\n"));
        assertTrue(msg1.isPresent());
        assertEquals("First", msg1.get().data().get());
        // Second message - BOM in data should be preserved
        Optional<EventStreamMessage> msg2 = parser.add(stringToBuffer("data: \uFEFFSecond\n\n"));
        assertTrue(msg2.isPresent());
        assertEquals("\uFEFFSecond", msg2.get().data().get());
    }

    @Test
    void testSSEStreamingExtremeFragmentation() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        String fullMessage = "event: fragmented\nid: extreme\ndata: Very fragmented message\n\n";
        // Add one character at a time
        Optional<EventStreamMessage> result = Optional.empty();
        for (char c : fullMessage.toCharArray()) {
            result = parser.add(stringToBuffer(String.valueOf(c)));
            if (result.isPresent()) {
                break;
            }
        }
        assertTrue(result.isPresent());
        assertEquals("Very fragmented message", result.get().data().get());
        assertEquals("fragmented", result.get().event().get());
        assertEquals("extreme", result.get().id().get());
    }

    @Test
    void testSSEStreamingMixedBoundaries() {
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        // Test various boundary splitting scenarios - build up a multi-line message
        assertFalse(parser.add(stringToBuffer("data: Boundary test")).isPresent());
        assertFalse(parser.add(stringToBuffer("\n")).isPresent());
        assertFalse(parser.add(stringToBuffer("data: Second line")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("\n\n"));
        assertTrue(result.isPresent());
        assertEquals("Boundary test\nSecond line", result.get().data().get());
    }

    // ===== JSON LINES PARSING TESTS =====

    @Test
    void testJsonLinesBasicParsing() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        String input = "{\"message\":\"hello\"}\n{\"message\":\"world\"}\n";
        
        Optional<String> result1 = parser.add(stringToBuffer(input));
        assertTrue(result1.isPresent());
        assertEquals("{\"message\":\"hello\"}", result1.get());
        
        Optional<String> result2 = parser.next();
        assertTrue(result2.isPresent());
        assertEquals("{\"message\":\"world\"}", result2.get());
        
        assertFalse(parser.next().isPresent());
    }

    @Test
    void testJsonLinesChunkedInput() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        
        // Send partial JSON line
        assertFalse(parser.add(stringToBuffer("{\"message\":")).isPresent());
        assertFalse(parser.add(stringToBuffer("\"hello\"}")).isPresent());
        
        // Complete the line
        Optional<String> result = parser.add(stringToBuffer("\n"));
        assertTrue(result.isPresent());
        assertEquals("{\"message\":\"hello\"}", result.get());
    }

    @Test
    void testJsonLinesEmptyLines() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        String input = "{\"message\":\"hello\"}\n\n{\"message\":\"world\"}\n";
        
        Optional<String> result1 = parser.add(stringToBuffer(input));
        assertTrue(result1.isPresent());
        assertEquals("{\"message\":\"hello\"}", result1.get());
        
        Optional<String> result2 = parser.next();
        assertTrue(result2.isPresent());
        assertEquals("{\"message\":\"world\"}", result2.get());
    }

    @Test
    void testJsonLinesWithCRLF() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        String input = "{\"message\":\"hello\"}\r\n{\"message\":\"world\"}\r\n";
        
        Optional<String> result1 = parser.add(stringToBuffer(input));
        assertTrue(result1.isPresent());
        assertEquals("{\"message\":\"hello\"}", result1.get());
        
        Optional<String> result2 = parser.next();
        assertTrue(result2.isPresent());
        assertEquals("{\"message\":\"world\"}", result2.get());
    }

    @Test
    void testJsonLinesPartialAtFinish() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        parser.add(stringToBuffer("{\"message\":\"incomplete\"}"));
        
        Optional<String> result = parser.finish();
        assertTrue(result.isPresent());
        assertEquals("{\"message\":\"incomplete\"}", result.get());
    }

    @Test
    void testJsonLinesLargeMessage() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        String largeJson = "{\"data\":\"" + "X".repeat(10000) + "\"}\n";
        
        Optional<String> result = parser.add(stringToBuffer(largeJson));
        assertTrue(result.isPresent());
        assertTrue(result.get().contains("X".repeat(10000)));
    }

    @Test
    void testJsonLinesBufferStateTracking() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        assertFalse(parser.hasBufferedData());
        
        parser.add(stringToBuffer("{\"partial\":"));
        assertTrue(parser.hasBufferedData());
        
        parser.add(stringToBuffer("\"data\"}\n"));
        assertFalse(parser.hasBufferedData());
    }

    @ParameterizedTest
    @ValueSource(strings = {"", "   ", "\t\n", "\n\n\n"})
    void testJsonLinesEmptyInputs(String input) {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        assertFalse(parser.add(stringToBuffer(input)).isPresent());
    }

    @Test
    void testJsonLinesNullInput() {
        StreamingParser<String> parser = StreamingParser.forJsonLines();
        assertFalse(parser.add((ByteBuffer) null).isPresent());
    }

    // ===== BOUNDARY SEAM & LINEARITY TESTS =====

    @Test
    void testSSEStreamingCRLFSplitAcrossChunks() {
        // A CRLF line ending split across reads must not be misread as a CR
        // line ending followed by an LF empty line, which would split one
        // event in two.
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        assertFalse(parser.add(stringToBuffer("data: x\r")).isPresent());
        assertFalse(parser.add(stringToBuffer("\ndata: y\r")).isPresent());
        Optional<EventStreamMessage> result = parser.add(stringToBuffer("\n\r\n"));
        assertTrue(result.isPresent());
        assertEquals("x\ny", result.get().data().get());
    }

    static Stream<Arguments> boundarySeamProvider() {
        return Stream.of(
                Arguments.of((Object) new String[]{"data: x\n", "\n"}),
                Arguments.of((Object) new String[]{"data: x\r", "\n\r\n"}),
                Arguments.of((Object) new String[]{"data: x\r\n", "\r\n"}),
                Arguments.of((Object) new String[]{"data: x\r\n\r", "\n"}),
                Arguments.of((Object) new String[]{"data: x", "\r", "\n", "\r", "\n"})
        );
    }

    @ParameterizedTest
    @MethodSource("boundarySeamProvider")
    void testSSEStreamingBoundarySplitAcrossChunkSeams(String[] chunks) {
        // The event may fire before the final chunk (a trailing bare CR ends
        // an event eagerly and a dangling LF is then skipped), so collect
        // across every add() and assert exactly one event total.
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        List<EventStreamMessage> received = new ArrayList<>();
        for (String chunk : chunks) {
            parser.add(stringToBuffer(chunk)).ifPresent(received::add);
            Optional<EventStreamMessage> more;
            while ((more = parser.next()).isPresent()) {
                received.add(more.get());
            }
        }
        assertEquals(1, received.size());
        assertEquals("x", received.get(0).data().get());
    }

    @Test
    void testSSEStreamingLargeEventSmallChunksIsLinear() {
        // An implementation that rescans the whole accumulated buffer on
        // every add() degrades quadratically with event size; this payload
        // would take tens of seconds. The resumable scan parses it in well
        // under the timeout.
        StreamingParser<EventStreamMessage> parser = StreamingParser.forSSE();
        int size = 8 * 1024 * 1024;
        byte[] message = ("data: " + "a".repeat(size) + "\n\n").getBytes(StandardCharsets.UTF_8);
        assertTimeoutPreemptively(java.time.Duration.ofSeconds(10), () -> {
            Optional<EventStreamMessage> result = Optional.empty();
            int chunkSize = 1024;
            for (int off = 0; off < message.length; off += chunkSize) {
                int len = Math.min(chunkSize, message.length - off);
                result = parser.add(ByteBuffer.wrap(message, off, len));
            }
            assertTrue(result.isPresent());
            assertEquals(size, result.get().data().get().length());
        });
    }
}
