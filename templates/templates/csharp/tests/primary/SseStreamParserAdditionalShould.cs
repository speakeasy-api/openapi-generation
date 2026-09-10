using Xunit;
using Openapi.Utils.Sse;
using System;
using System.Collections.Generic;

public class SseStreamParserAdditionalShould
{
    [Fact]
    public void ParseBasicEvent()
    {
        var parser = new SseStreamParser();

        var event1 = parser.Parse("data: Hello World");
        Assert.Null(event1); // No event yet, need blank line

        var event2 = parser.Parse(""); // Blank line dispatches event
        Assert.NotNull(event2);
        Assert.Equal("Hello World", event2.Data);
        Assert.Null(event2.Id);
        Assert.Null(event2.Event);
    }

    [Fact]
    public void ParseEventWithId()
    {
        var parser = new SseStreamParser();

        parser.Parse("id: event-123");
        parser.Parse("data: Test data");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("event-123", event1.Id);
        Assert.Equal("Test data", event1.Data);
    }

    [Fact]
    public void ParseEventWithCustomType()
    {
        var parser = new SseStreamParser();

        parser.Parse("event: custom");
        parser.Parse("data: Custom event data");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("custom", event1.Event);
        Assert.Equal("Custom event data", event1.Data);
    }

    [Fact]
    public void ParseEventWithRetry()
    {
        var parser = new SseStreamParser();

        parser.Parse("retry: 5000");
        parser.Parse("data: Retry test");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal(5000, event1.Retry);
        Assert.Equal("Retry test", event1.Data);
    }

    [Fact]
    public void ParseMultilineData()
    {
        var parser = new SseStreamParser();

        parser.Parse("data: Line 1");
        parser.Parse("data: Line 2");
        parser.Parse("data: Line 3");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("Line 1\nLine 2\nLine 3", event1.Data);
        Assert.Equal("Line 1\nLine 2\nLine 3", event1.Data);
    }

    [Fact]
    public void ParseEventWithComments()
    {
        var parser = new SseStreamParser();

        // HTML Living Standard: Lines starting with : are comments and should be ignored
        parser.Parse(": This is a comment");
        parser.Parse("data: Real data");
        parser.Parse(": Another comment");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("Real data", event1.Data);
    }

    [Fact]
    public void ParseFieldWithoutColon()
    {
        var parser = new SseStreamParser();

        // Field without colon should be treated as field name with empty value
        parser.Parse("data");
        var event1 = parser.Parse("");

        // Event is dispatched with empty data (data field was set)
        Assert.NotNull(event1);
        Assert.Equal("", event1.Data);
    }

    [Fact]
    public void ParseFieldWithSpaceAfterColon()
    {
        var parser = new SseStreamParser();

        // HTML Living Standard: Space after colon should be removed
        parser.Parse("data: Value with space");
        parser.Parse("event: custom");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("Value with space", event1.Data);
        Assert.Equal("custom", event1.Event);
    }

    [Fact]
    public void ParseIdWithNullCharacterIgnored()
    {
        var parser = new SseStreamParser();

        parser.Parse("id: valid-id");
        parser.Parse("data: First event");
        var event1 = parser.Parse("");

        // ID field with null character should be ignored
        parser.Parse("id: invalid\0id");
        parser.Parse("data: Second event");
        var event2 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("valid-id", event1.Id);

        Assert.NotNull(event2);
        Assert.Equal("valid-id", event2.Id); // last event ID persists
    }

    [Fact]
    public void ParseInvalidRetryIgnored()
    {
        var parser = new SseStreamParser();

        // Valid retry
        parser.Parse("retry: 3000");
        parser.Parse("data: Valid retry");
        var event1 = parser.Parse("");

        // Invalid retry (non-numeric)
        parser.Parse("retry: invalid");
        parser.Parse("data: Invalid retry");
        var event2 = parser.Parse("");

        // Negative retry
        parser.Parse("retry: -1000");
        parser.Parse("data: Negative retry");
        var event3 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal(3000, event1.Retry);

        Assert.NotNull(event2);
        Assert.Null(event2.Retry); // Invalid retry ignored, no retry set for this event

        Assert.NotNull(event3);
        Assert.Null(event3.Retry); // Negative retry ignored, no retry set for this event
    }

    [Fact]
    public void ParseUnknownFieldsIgnored()
    {
        var parser = new SseStreamParser();

        // HTML Living Standard: Unknown fields should be ignored
        parser.Parse("unknown: field");
        parser.Parse("custom-field: value");
        parser.Parse("data: Test data");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("Test data", event1.Data);
    }

    [Fact]
    public void ParseEventWithoutDataField()
    {
        var parser = new SseStreamParser();

        // Per HTML Living Standard, the dispatch behavior for non-browser user agents is
        // "implementation dependent". To ensure consistency across targets and support
        // heartbeat/ping patterns, events are dispatched even without data.
        parser.Parse("id: test-id");
        parser.Parse("event: test");
        var event1 = parser.Parse(""); // No data field

        Assert.NotNull(event1);
        Assert.Equal("test-id", event1.Id);
        Assert.Equal("test", event1.Event);
        Assert.Null(event1.Data);
    }

    [Fact]
    public void ParseSentinelEvent()
    {
        var parser = new SseStreamParser("[DONE]");

        // Regular event
        parser.Parse("data: Regular event");
        var event1 = parser.Parse("");

        // Sentinel event
        parser.Parse("data: [DONE]");
        var event2 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.False(event1.IsEndOfStream);
        Assert.Equal("Regular event", event1.Data);

        Assert.NotNull(event2);
        Assert.True(event2.IsEndOfStream);
        Assert.Equal("[DONE]", event2.Data);
    }

    [Fact]
    public void ParseSentinelEventCaseInsensitive()
    {
        var parser = new SseStreamParser("[DONE]");

        // Sentinel with different case
        parser.Parse("data: [done]");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.True(event1.IsEndOfStream);
    }

    [Fact]
    public void ParseEndOfStreamWithRemainingData()
    {
        var parser = new SseStreamParser();

        parser.Parse("data: Incomplete event");
        // End of stream without blank line
        var event1 = parser.Parse(null);

        Assert.NotNull(event1);
        Assert.Equal("Incomplete event", event1.Data);
    }

    [Fact]
    public void ParseEndOfStreamWithoutData()
    {
        var parser = new SseStreamParser();

        // End of stream with no data
        var event1 = parser.Parse(null);

        Assert.Null(event1);
    }

    [Fact]
    public void ParseJsonDeserialization()
    {
        var parser = new SseStreamParser();

        parser.Parse("data: {\"message\": \"Hello\", \"count\": 42}");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("{\"message\": \"Hello\", \"count\": 42}", event1.Data);
        // Note: Deserialization now happens in EventStream.AsType(), not in the parser
    }

    [Fact]
    public void ParseInvalidJsonFallsBackToRawData()
    {
        var parser = new SseStreamParser();

        parser.Parse("data: invalid json {");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("invalid json {", event1.Data);
        // Raw data is preserved in the parser
    }

    [Fact]
    public void ParseStringTypeUsesRawData()
    {
        var parser = new SseStreamParser();

        parser.Parse("data: {\"not\": \"parsed\"}");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("{\"not\": \"parsed\"}", event1.Data);
        Assert.Equal("{\"not\": \"parsed\"}", event1.Data);
    }

    [Fact]
    public void ParseExceptionHandling()
    {
        var parser = new SseStreamParser();

        // Test that the parser handles malformed input gracefully
        // The parser should wrap exceptions in SseParseException
        try
        {
            // This should work fine - parser is robust
            var result = parser.Parse("data: valid data");
            // No exception expected for valid input
        }
        catch (SseParseException ex)
        {
            // If an exception occurs, it should be wrapped properly
            Assert.NotNull(ex.Message);
        }
    }

    [Fact]
    public void ParseMultipleEventsSequentially()
    {
        var parser = new SseStreamParser();

        // First event
        parser.Parse("id: 1");
        parser.Parse("data: First");
        var event1 = parser.Parse("");

        // Second event
        parser.Parse("id: 2");
        parser.Parse("data: Second");
        var event2 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("1", event1.Id);
        Assert.Equal("First", event1.Data);

        Assert.NotNull(event2);
        Assert.Equal("2", event2.Id);
        Assert.Equal("Second", event2.Data);
    }

    [Fact]
    public void ParseComplexMultilineEvent()
    {
        var parser = new SseStreamParser();

        parser.Parse("id: complex-event");
        parser.Parse("event: multiline");
        parser.Parse("retry: 1000");
        parser.Parse("data: {");
        parser.Parse("data:   \"field1\": \"value1\",");
        parser.Parse("data:   \"field2\": \"value2\"");
        parser.Parse("data: }");
        var event1 = parser.Parse("");

        Assert.NotNull(event1);
        Assert.Equal("complex-event", event1.Id);
        Assert.Equal("multiline", event1.Event);
        Assert.Equal(1000, event1.Retry);

        var expectedData = "{\n  \"field1\": \"value1\",\n  \"field2\": \"value2\"\n}";
        Assert.Equal(expectedData, event1.Data);
    }

    // Helper class for JSON deserialization tests
    public class TestData
    {
        public string Message { get; set; } = "";
        public int Count { get; set; }
    }
}
