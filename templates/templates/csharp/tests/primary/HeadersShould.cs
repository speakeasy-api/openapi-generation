using Xunit;
using Openapi;
using Openapi.Models.Shared;
using Openapi.Models.Operations;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Net;
using Openapi.Models.Errors;

public class HeadersShould
{
    [Fact]
    public async Task EmptyResponseBodyWithHeaders()
    {
        CommonHelpers.RecordTest("headers-empty-response-body-with-headers");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.ResponseHeaders.ResponseBodyEmptyWithHeadersAsync(1.1, "hello");
        Assert.NotNull(res);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.HttpMeta.Response.Headers);
        Assert.Equal(new string[] { "hello" }, res.HttpMeta.Response.Headers.GetValues("X-String-Header"));
        Assert.Equal(new string[] { "1.1" }, res.HttpMeta.Response.Headers.GetValues("X-Number-Header"));

        // In C#, response headers are always collected in a `res.Headers` map, regardless of `responseFormat`.
        Assert.True(res.Headers.ContainsKey("X-String-Header"));
        Assert.Equal(new string[] { "hello" }, res.Headers["X-String-Header"]);
        Assert.True(res.Headers.ContainsKey("X-Number-Header"));
        Assert.Equal(new string[] { "1.1" }, res.Headers["X-Number-Header"]);
    }

    [Fact]
    public async Task ResponseWithoutHeaders()
    {
        CommonHelpers.RecordTest("headers-response-headers-none");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // The 200 response in `errorResponseHeaders` has a response body but no headers.
        // Some other responses in the same operation do include headers however.
        var res200 = await sdk.ResponseHeaders.ErrorResponseHeadersAsync(true, 200);
        Assert.NotNull(res200);
        Assert.Equal(HttpStatusCode.OK, res200.HttpMeta.Response.StatusCode);
        Assert.NotNull(res200.AuthToken);
        Assert.Equal("test-token", res200.AuthToken.Token);
    }

    [Fact]
    public async Task ErrorResponsesWithoutHeaders()
    {
        CommonHelpers.RecordTest("headers-error-response-headers-none");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // 1. Basic error response (no body nor custom headers) whose
        // operation includes another response that has headers.
        var ex1 = await Assert.ThrowsAsync<APIException>(async () =>
            await sdk.ResponseHeaders.ResponseHeadersAsync(true, 400));
        Assert.Equal(HttpStatusCode.BadRequest, ex1.Response.StatusCode);
        Assert.Equal("", ex1.Body);

        // 2. Error response with a body but no custom headers whose
        // operation includes another response that has headers.
        var ex2 = await Assert.ThrowsAsync<APIException>(async () =>
            await sdk.ResponseHeaders.ResponseHeadersAsync(true, 500));
        Assert.Equal(HttpStatusCode.InternalServerError, ex2.Response.StatusCode);
        Assert.Equal("\"Internal server error.\"\n", ex2.Body);
    }

    [Fact]
    public async Task ResponseHeadersOptional()
    {
        CommonHelpers.RecordTest("headers-response-headers-optional");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // 1. Success response with required header included
        var res1 = await sdk.ResponseHeaders.ResponseHeadersAsync(true, 200);
        Assert.NotNull(res1);
        Assert.Equal(HttpStatusCode.OK, res1.HttpMeta.Response.StatusCode);
        Assert.NotNull(res1.AuthToken);
        Assert.Equal("test-token", res1.AuthToken.Token);
        Assert.NotNull(res1.HttpMeta.Response.Headers);
        Assert.Equal(new string[] { "required" }, res1.HttpMeta.Response.Headers.GetValues("X-Required-Header"));
        Assert.False(res1.HttpMeta.Response.Headers.Contains("X-Optional-Header"));
        Assert.NotNull(res1.Headers);
        Assert.True(res1.Headers.ContainsKey("X-Required-Header"));
        Assert.Equal(new string[] { "required" }, res1.Headers["X-Required-Header"]);
        Assert.False(res1.Headers.ContainsKey("X-Optional-Header"));

        // 2. Success response with required header omitted - SDK should not throw
        var res2 = await sdk.ResponseHeaders.ResponseHeadersAsync(false, 200);
        Assert.NotNull(res2);
        Assert.Equal(HttpStatusCode.OK, res2.HttpMeta.Response.StatusCode);
        Assert.NotNull(res2.AuthToken);
        Assert.Equal("test-token", res2.AuthToken.Token);
        Assert.False(res2.HttpMeta.Response.Headers.Contains("X-Required-Header"));
        Assert.False(res2.HttpMeta.Response.Headers.Contains("X-Optional-Header"));
    }

    [Fact]
    public async Task ErrorResponseHeadersOptional()
    {
        CommonHelpers.RecordTest("headers-error-response-headers-optional");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // 1. Error response with Retry-After header included
        var ex1 = await Assert.ThrowsAsync<APIException>(async () =>
            await sdk.ResponseHeaders.ErrorResponseHeadersAsync(true, 429));
        Assert.Equal(HttpStatusCode.TooManyRequests, ex1.Response.StatusCode);
        Assert.Equal(new string[] { "60" }, ex1.Response.Headers.GetValues("retry-after"));
        Assert.Equal("\"Too many attempts. Please try again later.\"\n", ex1.Body);

        // 2. Error response with Retry-After header omitted - SDK should not throw
        var ex2 = await Assert.ThrowsAsync<APIException>(async () =>
            await sdk.ResponseHeaders.ErrorResponseHeadersAsync(false, 429));
        Assert.Equal(HttpStatusCode.TooManyRequests, ex2.Response.StatusCode);
        Assert.False(ex2.Response.Headers.Contains("retry-after"));
        Assert.Equal("\"Too many attempts. Please try again later.\"\n", ex2.Body);
    }
}
