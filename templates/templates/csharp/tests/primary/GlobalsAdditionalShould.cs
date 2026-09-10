using System.Threading.Tasks;
using Openapi;
using Openapi.Utils;
using Xunit;
using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;


public class GlobalsAdditionalShould
{
    [Fact]
    public async Task GlobalsQueryParameterGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-query-parameter-get-uses-global");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalQueryParam: "test");

        var res = await sdk.Globals.GlobalsQueryParameterGetAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test", res.Res.Args.GlobalQueryParam);
    }

    [Fact]
    public async Task GlobalsQueryParameterGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-query-parameter-get-uses-local");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalQueryParam: "test");

        var res = await sdk.Globals.GlobalsQueryParameterGetAsync("local");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("local", res.Res.Args.GlobalQueryParam);
    }

    [Fact]
    public async Task GlobalPathParameterGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-path-parameter-get-uses-global");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalPathParam: 1);

        var res = await sdk.Globals.GlobalPathParameterGetAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal($"{Helpers.HttpBinUrl}/anything/globals/pathParameter/1", res.Res.Url);
    }

    [Fact]
    public async Task GlobalPathParameterGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-path-parameter-get-uses-local");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalPathParam: 1);

        var res = await sdk.Globals.GlobalPathParameterGetAsync(2);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal($"{Helpers.HttpBinUrl}/anything/globals/pathParameter/2", res.Res.Url);
    }

    [Fact]
    public async Task GlobalHeaderGetUsesGlobal()
    {
        CommonHelpers.RecordTest("globals-header-get-uses-global");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalHeaderParam: true);

        var res = await sdk.Globals.GlobalsHeaderGetAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("true", res.Res.Headers["Globalheaderparam"]);
    }

    [Fact]
    public async Task GlobalHeaderGetUsesLocal()
    {
        CommonHelpers.RecordTest("globals-header-get-uses-local");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalHeaderParam: true);

        var res = await sdk.Globals.GlobalsHeaderGetAsync(globalHeaderParam: false);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("false", res.Res.Headers["Globalheaderparam"]);
    }

    [Fact]
    public async Task GlobalHeaderKeepsCustomClientHeaders()
    {
        CommonHelpers.RecordTest("globals-header-keeps-custom-client-headers");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, globalHeaderParam: true, client: new Helpers.CustomHttpClient());

        var res = await sdk.Globals.GlobalsHeaderGetAsync();

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("true", res.Res.Headers["Globalheaderparam"]);
        Assert.Equal("someValue", res.Res.Headers["X-Custom-Header"]);
    }
}

