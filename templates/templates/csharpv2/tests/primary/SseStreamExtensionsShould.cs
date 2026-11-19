#nullable enable
using Openapi;
using Openapi.Utils.Sse;
using System;
using System.Collections.Generic;
using System.IO;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;
using Xunit;

public partial class SseStreamExtensionsShould
{
    [Fact]
    public async Task AsyncEnumerableBreakClosesResponseStream()
    {
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\ndata: second\n\n"));
        using var response = EventStreamResponse(underlying);

        await foreach (var e in response.AsAsyncEnumerable(SseTestJsonContext.Default.String))
        {
            Assert.Equal("first", e);
            break;
        }

        Assert.False(underlying.CanRead, "The response stream should be closed when enumeration exits early");
    }

    [Fact]
    public async Task AsyncEnumerableCompletionClosesResponseStream()
    {
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\n"));
        using var response = EventStreamResponse(underlying);

        var events = new List<string>();
        await foreach (var e in response.AsAsyncEnumerable(SseTestJsonContext.Default.String))
        {
            events.Add(e);
        }

        Assert.Equal(new[] { "first" }, events);
        Assert.False(underlying.CanRead, "The response stream should be closed when enumeration completes");
    }

    [Fact]
    public async Task AsyncEnumerableSecondEnumerationYieldsNothing()
    {
        using var response = EventStreamResponse(new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\ndata: second\n\n")));
        var events = response.AsAsyncEnumerable(SseTestJsonContext.Default.String);

        await foreach (var e in events)
        {
            Assert.Equal("first", e);
            break;
        }

        var repeated = new List<string>();
        await foreach (var e in events)
        {
            repeated.Add(e);
        }

        Assert.Empty(repeated);
    }

    [Fact]
    public async Task AsyncEnumerableDeserializationFailureClosesResponseStream()
    {
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: {\"data\":\"x\"}\n\n"));
        using var response = EventStreamResponse(underlying);

        await Assert.ThrowsAsync<InvalidOperationException>(async () =>
        {
            await foreach (var _ in response.AsAsyncEnumerable(SseTestJsonContext.Default.ValueTypeEvent))
            {
            }
        });

        Assert.False(underlying.CanRead, "The response stream should be closed when deserialization fails");
    }

    [Fact]
    public async Task AsyncEnumerableDeserializesDefaultValueTypeEvent()
    {
        using var response = EventStreamResponse(new MemoryStream(Encoding.UTF8.GetBytes("data: 0\n\n")));

        var events = new List<ValueTypeEvent>();
        await foreach (var e in response.AsAsyncEnumerable(SseTestJsonContext.Default.ValueTypeEvent))
        {
            events.Add(e);
        }

        var value = Assert.Single(events);
        Assert.Equal(0, value.Data);
    }

    [Fact]
    public async Task MoveNextPropagatesCancellationDuringRead()
    {
        using var cts = new CancellationTokenSource();
        var underlying = new CancellableReadStream("data: Hello\n\n");
        using var response = EventStreamResponse(underlying);
        await using var enumerator = response
            .AsAsyncEnumerable(SseTestJsonContext.Default.String, cancellationToken: cts.Token)
            .GetAsyncEnumerator();

        Assert.True(await enumerator.MoveNextAsync());
        Assert.Equal("Hello", enumerator.Current);

        var pendingRead = enumerator.MoveNextAsync().AsTask();
        cts.Cancel();

        await Assert.ThrowsAnyAsync<OperationCanceledException>(() => pendingRead.WaitAsync(TimeSpan.FromSeconds(5)));
        Assert.False(underlying.CanRead, "The response stream should be closed when a read is cancelled");
    }

    [Fact]
    public async Task AsyncEnumerablePreCancelledTokenClosesResponseStream()
    {
        using var cts = new CancellationTokenSource();
        cts.Cancel();
        var underlying = new MemoryStream(Encoding.UTF8.GetBytes("data: first\n\n"));
        using var response = EventStreamResponse(underlying);

        await Assert.ThrowsAnyAsync<OperationCanceledException>(async () =>
        {
            await foreach (var _ in response.AsAsyncEnumerable(SseTestJsonContext.Default.String, cancellationToken: cts.Token))
            {
            }
        });

        Assert.False(underlying.CanRead, "The response stream should be closed when enumeration is cancelled");
    }

    [Fact]
    public async Task OperationCancellationTokenCancelsEnumeration()
    {
        using var cts = new CancellationTokenSource();
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Eventstreams.TextAsync(cancellationToken: cts.Token);
        Assert.NotNull(res.TextEvent);
        cts.Cancel();

        await Assert.ThrowsAnyAsync<OperationCanceledException>(async () =>
        {
            await foreach (var _ in res.TextEvent!)
            {
            }
        });
    }

    private static HttpResponseMessage EventStreamResponse(Stream body)
    {
        var response = new HttpResponseMessage { Content = new StreamContent(body) };
        response.Content.Headers.ContentType = new MediaTypeHeaderValue("text/event-stream");
        return response;
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

    internal struct ValueTypeEvent
    {
        public int Data { get; set; }
    }

    [JsonSourceGenerationOptions(PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase)]
    [JsonSerializable(typeof(string))]
    [JsonSerializable(typeof(ValueTypeEvent))]
    internal partial class SseTestJsonContext : JsonSerializerContext
    {
    }
}
