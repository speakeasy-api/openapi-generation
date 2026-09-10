#nullable enable
using System.Collections.Generic;
using System.Threading.Tasks;
using No_Security.API;
using No_Security.API.Models.Shared;
using Xunit;

public class EventStreamAdditionalShould
{
    [Fact]
    public async Task AsyncEnumerableIteratesAllEvents()
    {
        CommonHelpers.RecordTest("event-stream-text-data");
        var sdk = new SDK(serverUrl: Helpers.ApiTestServiceUrl);

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
}
