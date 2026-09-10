using HoistedSecurity;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Xunit;

public class HeadersShould
{
    [Fact]
    public async Task HeadersResponseBodyWithHeadersFlat()
    {
        CommonHelpers.RecordTest("headers-response-body-with-headers-flat");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.ResponseHeaders.ResponseHeadersAsync(true, 200);
        Assert.NotNull(res);
        Assert.NotNull(res.Result);
        Assert.Equal("test-token", res.Result.Token);
        Assert.NotNull(res.Headers);
        Assert.Equal(new string[] { "required" }, res.Headers["X-Required-Header"]);
    }

    [Fact]
    public async Task HeadersEmptyResponseBodyWithHeadersFlat()
    {
        CommonHelpers.RecordTest("headers-empty-response-body-with-headers-flat");

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl);

        var res = await sdk.ResponseHeaders.ResponseBodyEmptyWithHeadersAsync(1.1, "hello");
        Assert.NotNull(res);
        Assert.NotNull(res.Headers);
        Assert.Equal(new string[] { "hello" }, res.Headers["X-String-Header"]);
        Assert.Equal(new string[] { "1.1" }, res.Headers["X-Number-Header"]);
    }
}
