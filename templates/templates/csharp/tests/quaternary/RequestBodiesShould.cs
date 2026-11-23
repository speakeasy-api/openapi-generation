using System.Collections.Generic;
using System.Net;
using System.Numerics;
using System.Threading.Tasks;
using Company.Product.Feature.Subnamespace;
using Company.Product.Feature.Subnamespace.Models.Operations;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Xunit;

public class RequestBodiesShould
{
    [Fact]
    public async Task PostMultipleContentTypeSplitJson()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-json");
        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitJsonAsync(
            new RequestBodyPostMultipleContentTypesSplitJsonRequestBody()
            {
                Bool = true,
                Num = 1.1F,
                Str = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        var json = Compat.Unwrap(res.Res!.Json)!;
        Assert.True((bool)json["bool"]);
        Assert.Equal(1.1, (double)json["num"], 0.0001);
        Assert.Equal("test", json["str"]);
    }

    [Fact]
    public async Task PostMultipleContentTypesSplitJsonWithParam()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-json-with-param");
        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var requestBody = new RequestBodyPostMultipleContentTypesSplitParamJsonRequestBody()
        {
            Bool = true,
            Num = 1.1D,
            Str = "test body"
        };

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamJsonAsync(
            requestBody,
            "test param"
        );

        Assert.Equal(HttpStatusCode.OK, res.RawResponse.StatusCode);
        var json = Compat.Unwrap(res.Res!.Json)!;
        Assert.True((bool)json["bool"]);
        Assert.Equal(1.1, (double)json["num"]);
        Assert.Equal("test body", json["str"].ToString());
        Assert.Equal("test param", res.Res.Args["paramStr"]);
    }
}
