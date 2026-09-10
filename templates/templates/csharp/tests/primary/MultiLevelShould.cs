using System.Threading.Tasks;
using Xunit;
using Openapi;
using System.Net;

public class MultiLevelShould
{
    [Fact]
    public async Task MultiLevelTest()
    {
        CommonHelpers.RecordTest("multi-level-grouping");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Nested.First.GetAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }
}
