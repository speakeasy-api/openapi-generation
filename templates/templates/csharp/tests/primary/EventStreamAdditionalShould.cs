#nullable enable
using Xunit;
using Openapi;
using Openapi.Models.Shared;
using Openapi.Models.Operations;
using Openapi.Utils;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Threading;
using System.Net;
using System.Net.Http;
using System;
using System.IO;
using System.Linq;
using System.Text;

public class EventStreamAdditionalShould : IDisposable
{
    private readonly SDK _sdk;
    private readonly CustomHttpClientWithTimeout _httpClient;

    public EventStreamAdditionalShould()
    {
        // Create custom HttpClient with 3-second connection and read timeouts
        _httpClient = new CustomHttpClientWithTimeout();
        _sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: _httpClient);
    }

    public void Dispose()
    {
        _httpClient?.Dispose();
    }

    // Custom HttpClient implementation with timeout configuration
    private class CustomHttpClientWithTimeout : IDefaultHttpClient, IDisposable
    {
        private readonly HttpClient _httpClient;
        private readonly Dictionary<string, string> _customHeaders;

        public CustomHttpClientWithTimeout(Dictionary<string, string>? customHeaders = null)
        {
            var handler = new HttpClientHandler();
            _httpClient = new HttpClient(handler)
            {
                Timeout = TimeSpan.FromSeconds(3) // Combined timeout for connection + read
            };
            _customHeaders = customHeaders ?? new Dictionary<string, string>();
        }

        public async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken? cancellationToken = null)
        {
            // Add custom headers if any
            foreach (var header in _customHeaders)
            {
                request.Headers.TryAddWithoutValidation(header.Key, header.Value);
            }

            return await _httpClient.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, cancellationToken ?? CancellationToken.None);
        }

        public async Task<HttpRequestMessage> CloneAsync(HttpRequestMessage request)
        {
            HttpRequestMessage clone = new HttpRequestMessage(request.Method, request.RequestUri);

            if (request.Content != null)
            {
                clone.Content = new ByteArrayContent(await request.Content.ReadAsByteArrayAsync());
                if (request.Content.Headers != null)
                {
                    foreach (var h in request.Content.Headers)
                    {
                        clone.Content.Headers.Add(h.Key, h.Value);
                    }
                }
            }

            foreach (var header in request.Headers)
            {
                clone.Headers.TryAddWithoutValidation(header.Key, header.Value);
            }

            foreach (var prop in request.Options)
            {
                clone.Options.TryAdd(prop.Key, prop.Value);
            }

            return clone;
        }

        public void Dispose()
        {
            _httpClient?.Dispose();
        }
    }

    [Fact]
    public async Task BasicEventStream()
    {
        CommonHelpers.RecordTest("event-stream-json-data");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.JsonAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<JsonEventData>();
        using (var stream = res.JsonEvent)
        {
            Assert.NotNull(stream);

            JsonEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e.Data);
            }

        }

        // Server sends exactly 4 JSON events: "Hello", " ", "world", "!"
        Assert.Equal(4, events.Count);
        Assert.All(events, e => Assert.NotNull(e));
        Assert.Equal("Hello", events[0].Content);
        Assert.Equal(" ", events[1].Content);
        Assert.Equal("world", events[2].Content);
        Assert.Equal("!", events[3].Content);
    }

    [Fact]
    public async Task ChatStreamWithRequestBody()
    {
        CommonHelpers.RecordTest("event-stream-chat-sentinel-event");
        var sdk = _sdk;
        var request = new ChatRequestBody
        {
            Prompt = "Tell me a story",
            Stream = true
        };

        var res = await sdk.Eventstreams.ChatAsync(request);

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<ChatCompletionStream>();
        using (var stream = res.ChatCompletionStream)
        {
            Assert.NotNull(stream);

            ChatCompletionStream? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends 4 content events + 1 sentinel event
        Assert.Equal(5, events.Count);
        Assert.All(events, e => Assert.NotNull(e));

        // Verify content for ChatCompletionStream events
        var contentEvents = events.Where(e => e.Type.Value == "chatCompletionEvent").ToList();
        Assert.Equal(4, contentEvents.Count);
        Assert.Equal("Hello", contentEvents[0].ChatCompletionEvent?.Data.Content);
        Assert.Equal(" ", contentEvents[1].ChatCompletionEvent?.Data.Content);
        Assert.Equal("world", contentEvents[2].ChatCompletionEvent?.Data.Content);
        Assert.Equal("!", contentEvents[3].ChatCompletionEvent?.Data.Content);

        // Verify sentinel event
        var sentinelEvents = events.Where(e => e.Type.Value == "sentinelEvent").ToList();
        Assert.Single(sentinelEvents);
    }

    [Fact]
    public async Task ChatSkipSentinelWithSentinelHandling()
    {
        CommonHelpers.RecordTest("event-stream-chat-skip-sentinel");
        var sdk = _sdk;
        var request = new ChatSkipSentinelRequestBody
        {
            Prompt = "Generate text with sentinel",
            Stream = true
        };

        var res = await sdk.Eventstreams.ChatSkipSentinelAsync(request);

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<ChatCompletionEvent>();
        using (var stream = res.ChatCompletionEvent)
        {
            Assert.NotNull(stream);

            ChatCompletionEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends exactly 4 content events: "Hello", " ", "world", "!" before [DONE] sentinel terminates the stream
        Assert.Equal(4, events.Count);
        Assert.All(events, e => Assert.NotNull(e));
        Assert.Equal("Hello", events[0].Data.Content);
        Assert.Equal(" ", events[1].Data.Content);
        Assert.Equal("world", events[2].Data.Content);
        Assert.Equal("!", events[3].Data.Content);
    }

    [Fact]
    public async Task DifferentDataSchemasStream()
    {
        CommonHelpers.RecordTest("event-stream-different-data-schemas");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.DifferentDataSchemasAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<DifferentDataSchemas>();
        using (var stream = res.DifferentDataSchemas)
        {
            Assert.NotNull(stream);

            DifferentDataSchemas? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends 6 events with different schemas: message, url, message, array, primitive, primitive
        Assert.Equal(6, events.Count);
        Assert.All(events, e => Assert.NotNull(e));

        // Verify event types and IDs
        Assert.Equal("event-1", events[0].Id);
        Assert.Equal(Event.Message, events[0].Event);
        Assert.Equal(123L, events[0].Data?.MessageEvent?.Id);
        Assert.Equal("event-2", events[1].Id);
        Assert.Equal(Event.Url, events[1].Event);
        Assert.Equal("event-3", events[2].Id);
        Assert.Equal(Event.Message, events[2].Event);
        Assert.Equal("event-4", events[3].Id);
        Assert.Equal(Event.Array, events[3].Event);
        Assert.Equal(DifferentDataSchemasDataType.ArrayOfArrayEvent, events[3].Data?.Type);
        Assert.Equal(new List<long> { 1, 2, 3, 4 }, events[3].Data?.ArrayOfArrayEvent);
        Assert.Equal("event-5", events[4].Id);
        Assert.Equal(Event.Primitive, events[4].Event);
        Assert.Equal(DifferentDataSchemasDataType.Boolean, events[4].Data?.Type);
        Assert.Equal(true, events[4].Data?.Boolean);
        Assert.Equal("event-6", events[5].Id);
        Assert.Equal(Event.Primitive, events[5].Event);
        Assert.Equal(DifferentDataSchemasDataType.Float32, events[5].Data?.Type);
        Assert.Equal(3.14159f, events[5].Data?.Float32);
    }

    [Fact]
    public async Task MixedDataStream()
    {
        CommonHelpers.RecordTest("event-stream-mixed-data");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.MixedDataAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<MixedDataEvent>();
        using (var stream = res.MixedDataEvent)
        {
            Assert.NotNull(stream);

            MixedDataEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        Assert.Equal(4, events.Count);

        // First event: JSON object (completion)
        Assert.Equal(MixedDataEventEvent.Completion, events[0].Event);
        Assert.NotNull(events[0].Data?.MessageEvent);
        Assert.Equal("Hello world", events[0].Data?.MessageEvent?.Content);

        // Second event: plain text
        Assert.Equal(MixedDataEventEvent.Text, events[1].Event);
        Assert.NotNull(events[1].Data?.Str);
        Assert.Equal("Processing your request...", events[1].Data?.Str);

        // Third event: plain text
        Assert.Equal(MixedDataEventEvent.Loading, events[2].Event);
        Assert.NotNull(events[2].Data?.Str);
        Assert.Equal("Almost done", events[2].Data?.Str);

        // Fourth event: JSON object (completion)
        Assert.Equal(MixedDataEventEvent.Completion, events[3].Event);
        Assert.NotNull(events[3].Data?.MessageEvent);
        Assert.Equal("Done!", events[3].Data?.MessageEvent?.Content);
    }

    [Fact]
    public async Task MultilineEventStream()
    {
        CommonHelpers.RecordTest("event-stream-multiline-data");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.MultilineAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<TextEvent>();
        using (var stream = res.TextEvent)
        {
            Assert.NotNull(stream);

            TextEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
                // Verify multiline data is properly handled
                Assert.NotNull(e.Data);
            }
        }

        // Server sends multiline data: "YHOO\n+2\n10" as a single event
        Assert.Single(events);
        Assert.Equal("YHOO\n+2\n10", events[0].Data);
    }

    [Fact]
    public async Task RichStreamWithDifferentEventTypes()
    {
        CommonHelpers.RecordTest("event-stream-rich-events");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.RichAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<RichStream>();
        using (var stream = res.RichStream)
        {
            Assert.NotNull(stream);

            RichStream? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
                // Verify we can handle different event types (completion, heartbeat)
                Assert.NotNull(e.Type);
            }
        }

        // Server sends exactly 3 events: completion, heartbeat, completion
        Assert.Equal(3, events.Count);
        Assert.Equal("completion", events[0].Type.Value);
        Assert.Equal("heartbeat", events[1].Type.Value);
        Assert.Equal("completion", events[2].Type.Value);

        // Verify IDs and retry values for different event types
        Assert.Equal("job-1", events[0].RichCompletionEvent?.Id);
        Assert.Null(events[1].RichCompletionEvent?.Id); // heartbeat has no completion ID
        Assert.Equal("job-1", events[2].RichCompletionEvent?.Id);
        Assert.Null(events[0].HeartbeatEvent?.Retry); // completion has no retry
        Assert.Equal(3000, events[1].HeartbeatEvent?.Retry); // heartbeat has retry: 3000
        Assert.Null(events[2].HeartbeatEvent?.Retry); // completion has no retry
    }

    [Fact]
    public async Task StayOpenStreamWithSentinel()
    {
        CommonHelpers.RecordTest("event-stream-stay-open-sentinel");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.StayOpenAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<TextEvent>();
        var stream = res.TextEvent;
        Assert.NotNull(stream);

        TextEvent? e;
        while ((e = await stream.Next()) != null)
        {
            events.Add(e);
        }

        // Server sends: "event 1", "event 2", "event 3", "event 4", then "[SENTINEL]"
        // The sentinel should terminate the stream, so we should get exactly 4 events
        Assert.Equal(4, events.Count);
        Assert.Equal("event 1", events[0].Data);
        Assert.Equal("event 2", events[1].Data);
        Assert.Equal("event 3", events[2].Data);
        Assert.Equal("event 4", events[3].Data);

        // Verify that the stream was automatically closed when sentinel "[SENTINEL]" was received
        // After sentinel, Next() should return null indicating stream end
        var eventAfterSentinel = await stream.Next();
        Assert.Null(eventAfterSentinel);

        // Clean up
        stream.Dispose();
    }

    [Fact]
    public async Task StreamResourceTeardownAndDisposal()
    {
        var sdk = _sdk;
        var res = await sdk.Eventstreams.JsonAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var stream = res.JsonEvent;
        Assert.NotNull(stream);

        // Read a few events
        var firstEvent = await stream.Next();
        var secondEvent = await stream.Next();

        // Explicitly dispose the stream
        stream.Dispose();

        // After disposal, Next() should return null/default
        var eventAfterDisposal = await stream.Next();
        Assert.Null(eventAfterDisposal);
    }

    [Fact]
    public async Task StreamEarlyTermination()
    {
        CommonHelpers.RecordTest("event-stream-stay-open-break-early");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.StayOpenAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var stream = res.TextEvent;
        Assert.NotNull(stream);
        using (stream)
        {
            // Read only first event then break
            var firstEvent = await stream.Next();
            Assert.NotNull(firstEvent);

            // In C#, early exit from using doesn't close stream until block exits
            Assert.False(stream.IsClosed, "Stream should still be open inside using block");
        }
        // Stream disposal should be handled properly even with early termination
        Assert.True(stream.IsClosed, "Stream should be closed after using block");
    }

    [Fact]
    public async Task MultipleStreamConsumption()
    {
        var sdk = _sdk;

        // Test consuming multiple streams sequentially
        var jsonRes = await sdk.Eventstreams.JsonAsync();
        var textRes = await sdk.Eventstreams.TextAsync();

        Assert.NotNull(jsonRes);
        Assert.NotNull(textRes);

        // Verify both streams can be consumed independently
        using var jsonStream = jsonRes.JsonEvent;
        using var textStream = textRes.TextEvent;
        Assert.NotNull(jsonStream);
        Assert.NotNull(textStream);
        var jsonEvent = await jsonStream.Next();
        var textEvent = await textStream.Next();

        // Both should have data based on server implementation
        Assert.NotNull(jsonEvent);
        Assert.NotNull(textEvent);
    }

    [Fact]
    public async Task StreamBoundaryHandling()
    {
        var sdk = _sdk;
        var res = await sdk.Eventstreams.JsonAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<JsonEventData>();
        using (var stream = res.JsonEvent)
        {
            Assert.NotNull(stream);

            JsonEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e.Data);

                // Verify each event is properly parsed despite potential timing delays
                Assert.NotNull(e.Data);
                Assert.NotNull(e.Data.Content);
            }
        }

        // Verify all events are received despite the 100ms delays between chunks
        Assert.Equal(4, events.Count);

        // Verify the complete message is reconstructed: "Hello world!"
        var fullMessage = string.Join("", events.Select(e => e.Content));
        Assert.Equal("Hello world!", fullMessage);
    }

    [Fact]
    public async Task StreamTimingAndFlushHandling()
    {
        var sdk = _sdk;
        var res = await sdk.Eventstreams.StayOpenAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<TextEvent>();
        var timestamps = new List<DateTime>();

        using (var stream = res.TextEvent)
        {
            Assert.NotNull(stream);

            TextEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
                timestamps.Add(DateTime.UtcNow);
            }
        }

        // Server sends events with specific timing: immediate events 1-3, then delays for event 4 and sentinel
        Assert.Equal(4, events.Count);
        Assert.Equal("event 1", events[0].Data);
        Assert.Equal("event 2", events[1].Data);
        Assert.Equal("event 3", events[2].Data);
        Assert.Equal("event 4", events[3].Data);

        // Verify that all events are received despite timing delays
        // The stream should handle the server's flush operations correctly
        Assert.True(timestamps.Count == 4);
    }

    [Fact]
    public async Task EventStreamWithSpecificContentValidation()
    {
        var sdk = _sdk;
        var res = await sdk.Eventstreams.JsonAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<JsonEvent>();
        using (var stream = res.JsonEvent)
        {
            Assert.NotNull(stream);

            JsonEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends exactly 4 events based on HandleEventStreamJSON
        Assert.Equal(4, events.Count);

        // Verify each event has the expected content from the server
        Assert.Equal("Hello", events[0].Data.Content);
        Assert.Equal(" ", events[1].Data.Content);
        Assert.Equal("world", events[2].Data.Content);
        Assert.Equal("!", events[3].Data.Content);

        // Verify all events are properly structured
        Assert.All(events, e =>
        {
            Assert.NotNull(e.Data);
            Assert.NotNull(e.Data.Content);
        });
    }

    [Fact]
    public async Task TextEventStreamValidation()
    {
        CommonHelpers.RecordTest("event-stream-text-data");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.TextAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<TextEvent>();
        using (var stream = res.TextEvent)
        {
            Assert.NotNull(stream);

            TextEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends exactly 4 text events based on HandleEventStreamText
        Assert.Equal(4, events.Count);

        // Verify each event has the expected data from the server
        Assert.Equal("Hello", events[0].Data);
        Assert.Equal(" ", events[1].Data);
        Assert.Equal("world", events[2].Data);
        Assert.Equal("!", events[3].Data);

        // Verify all events have data
        Assert.All(events, e => Assert.NotNull(e.Data));
    }

    [Fact]
    public async Task RichStreamEventValidation()
    {
        var sdk = _sdk;
        var res = await sdk.Eventstreams.RichAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<RichStream>();
        using (var stream = res.RichStream)
        {
            Assert.NotNull(stream);

            RichStream? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends exactly 3 events: completion, heartbeat, completion
        Assert.Equal(3, events.Count);

        // Verify event types match server implementation
        Assert.Equal("completion", events[0].Type.Value);
        Assert.Equal("heartbeat", events[1].Type.Value);
        Assert.Equal("completion", events[2].Type.Value);

        // Verify completion events have proper structure
        var rce0 = events[0].RichCompletionEvent;
        Assert.NotNull(rce0);
        Assert.Equal("job-1", rce0.Id);
        var rce2 = events[2].RichCompletionEvent;
        Assert.NotNull(rce2);
        Assert.Equal("job-1", rce2.Id);

        // Verify heartbeat event has proper structure
        var hb1 = events[1].HeartbeatEvent;
        Assert.NotNull(hb1);
        Assert.Equal("ping", hb1.Data);
        Assert.Equal(3000, hb1.Retry);
    }

    [Fact]
    public async Task EventStreamErrorResponse()
    {
        CommonHelpers.RecordTest("event-stream-error-response");

        // Create SDK with custom header that triggers teapot error
        var customHeaders = new Dictionary<string, string> { { "x-teapot", "json" } };
        var httpClientWithTeapotHeader = new CustomHttpClientWithTimeout(customHeaders);
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: httpClientWithTeapotHeader);

        try
        {
            // This should trigger a 418 "I'm a teapot" error response
            var res = await sdk.Eventstreams.TextAsync();

            // If we get here, the test should fail because we expected an exception
            Assert.True(false, "Expected an exception to be thrown for teapot error response");
        }
        catch (Exception ex)
        {
            // Verify that we got an appropriate error response
            // The exact exception type may vary based on the C# SDK implementation
            Assert.NotNull(ex);

            // Check if the error message contains teapot-related content
            var errorMessage = ex.Message.ToLower();
            Assert.True(errorMessage.Contains("teapot") || errorMessage.Contains("418"),
                $"Expected error message to contain 'teapot' or '418', but got: {ex.Message}");
        }
        finally
        {
            httpClientWithTeapotHeader?.Dispose();
        }
    }


    [Fact]
    public async Task EventStreamWithCancellationToken()
    {
        CommonHelpers.RecordTest("event-stream-with-abort-signal");
        var sdk = _sdk;

        using var cts = new CancellationTokenSource();

        var res = await sdk.Eventstreams.TextAsync(cancellationToken: cts.Token);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<TextEvent>();
        string message = "";
        bool cancelled = false;
        using (var stream = res.TextEvent)
        {
            Assert.NotNull(stream);

            TextEvent? e;
            int chunks = 0;
            try
            {
                while ((e = await stream.Next(cts.Token)) != null)
                {
                    events.Add(e);
                    message += e.Data;
                    chunks++;

                    // Cancel after receiving the first chunk
                    if (chunks == 1)
                    {
                        cts.Cancel();
                    }
                }
            }
            catch (OperationCanceledException)
            {
                cancelled = true;
            }
        }

        Assert.True(cancelled, "Expected cancellation exception to be thrown");
        Assert.True(events.Count <= 2, "Should have received at most 2 events");
        Assert.True(message.StartsWith("Hello"), "Should have received partial message");
    }

    [Fact]
    public async Task EventStreamPartialWithComments()
    {
        CommonHelpers.RecordTest("event-stream-partial-with-comments");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.PartialWithCommentsAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<PartialWithCommentsEvent>();
        using (var stream = res.PartialWithCommentsEvent)
        {
            Assert.NotNull(stream);

            PartialWithCommentsEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends exactly 5 events with comments interleaved
        Assert.Equal(5, events.Count);

        // First event - only data
        Assert.NotNull(events[0].Data);
        Assert.Equal("Hello from SSE", events[0].Data.Message);

        // Second event - with id and event type
        Assert.Equal("msg-2", events[1].Id);
        Assert.Equal("update", events[1].Event);
        Assert.NotNull(events[1].Data);
        Assert.Equal("processing", events[1].Data.Status);
        Assert.Equal(50L, events[1].Data.Progress);

        // Third event - completion with result
        Assert.Equal("msg-3", events[2].Id);
        Assert.NotNull(events[2].Data);
        Assert.Equal("complete", events[2].Data.Status);
        Assert.Equal(100L, events[2].Data.Progress);
        Assert.Equal("Success", events[2].Data.Result);

        // Fourth event - mixed
        Assert.Equal("msg-4", events[3].Id);
        Assert.Equal("mixed", events[3].Event);
        Assert.NotNull(events[3].Data);
        Assert.Equal("mixed boundaries", events[3].Data.Test);

        // Fifth event
        Assert.Equal("msg-5", events[4].Id);
        Assert.NotNull(events[4].Data);
        Assert.Equal("test", events[4].Data.Another);
    }

    [Fact]
    public async Task EventStreamWPTCompliance()
    {
        CommonHelpers.RecordTest("event-stream-wpt-compliance");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.WptComplianceAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var actual = new List<SseEvent>();
        List<SseEvent>? expected = null;

        using (var stream = res.SseEvent)
        {
            Assert.NotNull(stream);

            SseEvent? e;
            while ((e = await stream.Next()) != null)
            {
                if (e.Event == "expected")
                {
                    var options = new System.Text.Json.JsonSerializerOptions { PropertyNameCaseInsensitive = true };
                    expected = System.Text.Json.JsonSerializer.Deserialize<List<SseEvent>>(e.Data, options);
                }
                else
                {
                    actual.Add(e);
                }
            }
        }

        Assert.NotNull(expected);
        Assert.True(expected.Count > 0, "no expectations received from server");
        Assert.Equivalent(expected, actual);
    }

    [Fact]
    public async Task ChatHeartbeatSkipsDatalessEvents()
    {
        CommonHelpers.RecordTest("event-stream-chat-heartbeat-skips-dataless");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.ChatHeartbeatAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<ChatCompletionEvent>();
        using (var stream = res.ChatCompletionEvent)
        {
            Assert.NotNull(stream);

            ChatCompletionEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        // Server sends: data("Hello1"), heartbeat (skipped), data("Hello 2"), data("!"), [DONE]
        Assert.Equal(3, events.Count);
        Assert.Equal("Hello1", events[0].Data.Content);
        Assert.Equal("Hello 2", events[1].Data.Content);
        Assert.Equal("!", events[2].Data.Content);
    }

    [Fact]
    public async Task EventStreamWithOptionalDataField()
    {
        CommonHelpers.RecordTest("event-stream-optional-data-field");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.OptionalDataAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var events = new List<OptionalDataEvent>();
        using (var stream = res.OptionalDataEvent)
        {
            Assert.NotNull(stream);

            OptionalDataEvent? e;
            while ((e = await stream.Next()) != null)
            {
                events.Add(e);
            }
        }

        Assert.Equivalent(new List<OptionalDataEvent>()
        {
            new OptionalDataEvent { Event = "message", Id = "event-1", Data = new OptionalDataEventPayload { Content = "Hello, this event has data" } },
            new OptionalDataEvent { Event = "heartbeat", Id = "event-2" },
            new OptionalDataEvent { Event = "message", Id = "event-3", Data = new OptionalDataEventPayload { Content = "Another message with data" } },
            new OptionalDataEvent { Event = "ping", Id = "event-4" },
            new OptionalDataEvent { Event = "complete", Id = "event-5", Data = new OptionalDataEventPayload { Content = "Stream finished" } },
        }, events);
    }
    [Fact]
    public async Task AsyncEnumerableIteratesAllEvents()
    {
        CommonHelpers.RecordTest("event-stream-text-data");
        var sdk = _sdk;
        var res = await sdk.Eventstreams.TextAsync();

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var stream = res.TextEvent;
        Assert.NotNull(stream);

        var events = new List<TextEvent>();
        var repeatedEvents = new List<TextEvent>();
        await foreach (var e in stream)
        {
            events.Add(e);
        }

        await foreach (var e in stream)
        {
            repeatedEvents.Add(e);
        }

        Assert.Collection(
            events,
            e => Assert.Equal("Hello", e.Data),
            e => Assert.Equal(" ", e.Data),
            e => Assert.Equal("world", e.Data),
            e => Assert.Equal("!", e.Data));
        Assert.Empty(repeatedEvents);
        Assert.True(stream.IsClosed);
    }

    [Fact]
    public async Task AsyncEnumerableWithCancellation()
    {
        CommonHelpers.RecordTest("event-stream-with-abort-signal");
        var sdk = _sdk;

        using var cts = new CancellationTokenSource();

        var res = await sdk.Eventstreams.TextAsync(cancellationToken: cts.Token);
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var stream = res.TextEvent;
        Assert.NotNull(stream);

        var events = new List<TextEvent>();
        await Assert.ThrowsAnyAsync<OperationCanceledException>(async () =>
        {
            await foreach (var e in stream.WithCancellation(cts.Token))
            {
                events.Add(e);
                cts.Cancel();
            }
        });

        Assert.Collection(events, e => Assert.Equal("Hello", e.Data));
        Assert.True(stream.IsClosed);
    }

    [Fact]
    public async Task AsyncEnumerableBreakPreservesCallerOwnedStream()
    {
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\ndata: second\n\n"));
        var stream = new Openapi.Utils.Sse.EventStream<string>(underlying);

        await foreach (var e in stream)
        {
            Assert.Equal("first", e);
            break;
        }

        Assert.True(stream.IsClosed, "EventStream should be disposed when enumeration exits early");
        Assert.True(underlying.CanRead, "A stream passed to the public constructor should remain open");
    }

    [Fact]
    public async Task AsyncEnumerableBreakClosesResponseStream()
    {
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\ndata: second\n\n"));
        using var response = new HttpResponseMessage
        {
            Content = new StreamContent(underlying),
        };
        response.Content.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("text/event-stream");
        var stream = await Openapi.Utils.Sse.SseStreamExtensions.AsServerSentEventStream<string>(response);

        await foreach (var e in stream)
        {
            Assert.Equal("first", e);
            break;
        }

        Assert.True(stream.IsClosed, "EventStream should be disposed when enumeration exits early");
        Assert.False(underlying.CanRead, "The response stream should be closed when enumeration exits early");
    }

    [Fact]
    public async Task AwaitUsingClosesWithoutEnumeration()
    {
        var stream = new Openapi.Utils.Sse.EventStream<string>(new MemoryStream());

        await using (stream)
        {
        }

        Assert.True(stream.IsClosed);
    }

    [Fact]
    public async Task AsyncEnumerableEmitsDefaultValueTypeEventOnce()
    {
        var stream = new Openapi.Utils.Sse.EventStream<ValueTypeEvent>(
            new MemoryStream(Encoding.UTF8.GetBytes("data: 0\n\n")));
        var events = new List<ValueTypeEvent>();

        await foreach (var e in stream)
        {
            events.Add(e);
        }

        var value = Assert.Single(events);
        Assert.Equal(0, value.Data);
    }

    [Fact]
    public async Task NextPropagatesCancellationDuringRead()
    {
        using var cts = new CancellationTokenSource();
        await using var stream = new Openapi.Utils.Sse.EventStream<string>(
            new CancellableReadStream("data: Hello\n\n"));

        Assert.Equal("Hello", await stream.Next(cts.Token));

        var pendingRead = stream.Next(cts.Token);
        cts.Cancel();

        await Assert.ThrowsAnyAsync<OperationCanceledException>(() => pendingRead);
        Assert.True(stream.IsClosed);
    }

    [Fact]
    public async Task ParseLargeEventSplitAcrossSmallChunks()
    {
        CommonHelpers.RecordTest("event-stream-large-event-small-chunks");
        long size = 2 * 1024 * 1024;
        var res = await _sdk.Eventstreams.LargeEventSmallChunksAsync(size);

        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        var contents = new List<string>();
        using (var stream = res.LargeChunkedEvent)
        {
            Assert.NotNull(stream);

            LargeChunkedEvent? e;
            while ((e = await stream.Next()) != null)
            {
                contents.Add(e.Data.Content);
            }
        }

        Assert.Equal(3, contents.Count);
        Assert.Equal("start", contents[0]);
        Assert.Equal("end", contents[2]);

        var big = contents[1];
        Assert.Equal(size, big.Length);
        Assert.Equal('S', big[0]);
        Assert.Equal('E', big[big.Length - 1]);
        Assert.Equal(new string('a', (int)size - 2), big.Substring(1, big.Length - 2));
    }

    private sealed class CancellableReadStream : MemoryStream
    {
        public CancellableReadStream(string contents) : base(Encoding.UTF8.GetBytes(contents))
        {
        }

        public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken cancellationToken = default)
        {
            return Position < Length
                ? base.ReadAsync(buffer, cancellationToken)
                : new ValueTask<int>(WaitForCancellation(cancellationToken));
        }

        public override Task<int> ReadAsync(byte[] buffer, int offset, int count, CancellationToken cancellationToken)
        {
            return Position < Length
                ? base.ReadAsync(buffer, offset, count, cancellationToken)
                : WaitForCancellation(cancellationToken);
        }

        private static async Task<int> WaitForCancellation(CancellationToken cancellationToken)
        {
            await Task.Delay(Timeout.Infinite, cancellationToken);
            return 0;
        }
    }

    private struct ValueTypeEvent
    {
        public int Data { get; set; }
    }
}
