#nullable enable
using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Xunit;

public class EventStreamAdditionalShould
{
    [Fact]
    public async Task AsyncEnumerableIteratesAllEvents()
    {
        CommonHelpers.RecordTest("event-stream-text-data");
        var sdk = new SDK(apiKeyAuth: "token", serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.Eventstreams.TextAsync();
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);

        var stream = res.TextEvent;
        Assert.NotNull(stream);
        var events = new List<TextEvent>();
        await foreach (var e in stream)
        {
            events.Add(e);
        }

        Assert.Collection(
            events,
            e => Assert.Equal("Hello", e.Data),
            e => Assert.Equal(" ", e.Data),
            e => Assert.Equal("world", e.Data),
            e => Assert.Equal("!", e.Data));
        Assert.True(stream.IsClosed);
    }

    [Fact]
    public async Task AsyncEnumerableWithCancellation()
    {
        CommonHelpers.RecordTest("event-stream-with-abort-signal");
        var sdk = new SDK(apiKeyAuth: "token", serverUrl: CommonHelpers.HttpBinUrl);

        using var cts = new CancellationTokenSource();

        var res = await sdk.Eventstreams.TextAsync(cancellationToken: cts.Token);
        Assert.NotNull(res);
        Assert.Equal(200, res.StatusCode);

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
}
