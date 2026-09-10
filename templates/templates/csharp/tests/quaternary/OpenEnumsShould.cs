using Newtonsoft.Json;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Xunit;

public class OpenEnumsShould
{
    [Fact]
    public void PreserveDateShapedUnknownValuesAsIso8601()
    {
        // Non-presence-aware SDKs read without DateParseHandling.None, so
        // Newtonsoft materializes ISO-shaped strings as DateTime before
        // OpenEnumConverter runs; the converter re-emits ISO 8601 instead of
        // the invariant general format ("MM/dd/yyyy HH:mm:ss").
        var parsed = JsonConvert.DeserializeObject<OpenEnumSelfNamed>("\"2024-01-02T03:04:05Z\"");
        Assert.NotNull(parsed);
        Assert.False(parsed!.IsKnown());
        Assert.Equal("2024-01-02T03:04:05Z", parsed.Value);
    }
}
