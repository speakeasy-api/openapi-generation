#nullable enable
using System;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using HoistedSecurity;
using HoistedSecurity.Models.Operations;
using HoistedSecurity.Utils;
using Xunit;

public class UnionRequestBodyOverloadsAdditionalShould
{
    [Fact]
    public async Task ForwardToCanonicalMethodReturningTask()
    {
        var client = new CapturingHttpClient();
        var s = new SDK(client: client);

        Func<ErrorType1RequestBody, string?, CancellationToken?, Task> overload = s.Errors.ErrorUnionPostAsync;
        await Assert.ThrowsAsync<CapturedRequestException>(() =>
            overload(new ErrorType1RequestBody { Error = "e" }, null, null)
        );
        var viaOverload = (client.LastRequestUri, client.LastRequestBody);

        await Assert.ThrowsAsync<CapturedRequestException>(() =>
            s.Errors.ErrorUnionPostAsync(
                ErrorUnionPostRequestBody.CreateErrorType1RequestBody(new ErrorType1RequestBody { Error = "e" })
            )
        );

        Assert.Equal((client.LastRequestUri, client.LastRequestBody), viaOverload);
    }

    [Fact]
    public async Task RejectNullVariantFromReturnedTask()
    {
        var s = new SDK(client: new CapturingHttpClient());

        Func<ErrorType1RequestBody, string?, CancellationToken?, Task> overload = s.Errors.ErrorUnionPostAsync;
        var task = overload(null!, null, null);
        var ex = await Assert.ThrowsAsync<ArgumentNullException>(() => task);
        Assert.Equal("errorType1RequestBody", ex.ParamName);
    }

    private sealed class CapturedRequestException : Exception { }

    private sealed class CapturingHttpClient : ISDKHttpClient
    {
        public Uri? LastRequestUri { get; private set; }

        public string? LastRequestBody { get; private set; }

        public async Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken? cancellationToken = null
        )
        {
            LastRequestUri = request.RequestUri;
            LastRequestBody = request.Content == null ? null : await request.Content.ReadAsStringAsync();
            throw new CapturedRequestException();
        }

        public Task<HttpRequestMessage> CloneAsync(HttpRequestMessage request) =>
            Task.FromResult(request);
    }
}
