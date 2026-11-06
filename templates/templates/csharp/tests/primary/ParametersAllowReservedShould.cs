using System;
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using Openapi;
using Openapi.Utils;
using Xunit;

public partial class ParametersAllowReservedShould
{
    // Reserved characters with no structural meaning inside a path segment: they
    // survive URL parsing, so the request reaches the server intact.
    private const string NonStructuralReserved = ":@!$&'()*+,;=";

    // The same reserved set prefixed with an unreserved chunk that contains a
    // space, so the wire assertions also prove an unsafe unreserved character
    // still percent-encodes (space -> %20) while reserved characters pass through.
    private const string ReservedWithSpace = "abc 123" + NonStructuralReserved;
    private const string ReservedWithSpaceEncoded = "abc%20123" + NonStructuralReserved;

    [Fact]
    public async Task PathEncoding()
    {
        CommonHelpers.RecordTest("parameters-path-encoding");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new CommonHelpers.RequestRecorderClient(log));

        var res = await sdk.Parameters.PathEncodingAsync(
            ReservedWithSpace,
            ReservedWithSpace
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Single(log);

        var segments = log[0]
            .RequestUri!.AbsoluteUri.Split("/anything/pathencoding/")[1]
            .Split('/');
        // param1 percent-encodes every reserved character; param2 carries them raw
        // while the space still becomes %20.
        Assert.Equal(ReservedWithSpace, Uri.UnescapeDataString(segments[0]));
        Assert.NotEqual(segments[0], segments[1]);
        Assert.Equal(ReservedWithSpaceEncoded, segments[1]);
    }

    [Fact]
    public async Task QueryEncoding()
    {
        CommonHelpers.RecordTest("parameters-query-encoding");

        var log = new List<CommonHelpers.RequestLogEntry>();
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl, client: new CommonHelpers.RequestRecorderClient(log));

        var res = await sdk.Parameters.QueryEncodingAsync(ReservedWithSpace);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Single(log);
        Assert.Equal(
            Helpers.HttpBinUrl + "/anything/queryencoding?param1=" + ReservedWithSpaceEncoded,
            log[0].RequestUri!.AbsoluteUri
        );
    }

    [Theory]
    [InlineData(":/?#[]@!$&'()*+,;=", ":/?#[]@!$&'()*+,;=")]
    [InlineData("a b", "a%20b")]
    [InlineData("100%", "100%25")]
    [InlineData("a%2Fb", "a%252Fb")]
    [InlineData("\u00e9\u2603", "%C3%A9%E2%98%83")]
    [InlineData("plain-_.~", "plain-_.~")]
    public void EscapeExceptReservedHandlesEachCharacterClass(string input, string expected)
    {
        Assert.Equal(expected, URLBuilder.EscapeExceptReserved(input));
    }
}
