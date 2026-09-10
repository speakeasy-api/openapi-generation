using System.Net;
using System.Threading.Tasks;
using Xunit;
using Openapi;

public class CustomClientShould
{
    [Fact]
    public async Task CustomClientPost()
    {
        CommonHelpers.RecordTest("customclient-request-parameters-retained");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new Helpers.CustomHttpClient());

        var res = await sdk.CustomClient.CustomClientPostAsync(
            "headerValue",
            "pathValue",
            "queryValue",
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal($"{Helpers.HttpBinUrl}/anything/customClient/pathValue?queryStringParam=queryValue", res.Res.Url);
        Assert.Equal("queryValue", res.Res.Args.QueryStringParam);
        Assert.Equal("headerValue", res.Res.Headers["Headerparam"]);
        Assert.Equal("someValue", res.Res.Headers["X-Custom-Header"]);
        Helpers.AssertSimpleObject(res.Res.Json);
    }
}
