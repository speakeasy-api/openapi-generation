#nullable enable
using Openapi;
using Openapi.Models.Shared;
using Openapi.Utils;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using Xunit;

public class WrappedRequestBodyAdditionalShould
{
    [Fact]
    public async Task SendNoContentWhenOptionalMultipartBodyIsOmitted()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        var res = await sdk.RequestBodies.OptionalWrappedMultipartBodyAsync("w1");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var sent = Assert.Single(client.Requests);
        Assert.Equal(HttpMethod.Put, sent.Method);
        Assert.Equal("/anything/widgets/w1/optionalMultipart", sent.Path);
        Assert.Null(sent.ContentType);
        Assert.Empty(sent.Parts);
    }

    [Fact]
    public async Task SendMultipartPartsWhenOptionalBodyIsSet()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);
        var image = new byte[] { 0x89, 0x50, 0x4E, 0x47 };

        await sdk.RequestBodies.OptionalWrappedMultipartBodyAsync(
            "w1",
            new WidgetUpload()
            {
                Label = "front",
                WidgetImage = new WidgetImage() { FileName = "widget.png", Content = image },
            }
        );

        var sent = Assert.Single(client.Requests);
        Assert.Equal("multipart/form-data", sent.ContentType);
        Assert.Collection(
            sent.Parts,
            part =>
            {
                Assert.Equal("label", part.Name);
                Assert.Null(part.FileName);
                Assert.Equal("front", System.Text.Encoding.UTF8.GetString(part.Bytes));
            },
            part =>
            {
                Assert.Equal("widgetImage", part.Name);
                Assert.Equal("widget.png", part.FileName);
                Assert.Equal("image/png", part.ContentType);
                Assert.Equal(image, part.Bytes);
            }
        );
    }

    [Fact]
    public async Task SendNoContentWhenOptionalFormBodyIsOmitted()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        await sdk.RequestBodies.OptionalWrappedFormBodyAsync("w1");

        var sent = Assert.Single(client.Requests);
        Assert.Equal(HttpMethod.Put, sent.Method);
        Assert.Equal("/anything/widgets/w1/optionalForm", sent.Path);
        Assert.Null(sent.ContentType);
        Assert.Null(sent.Body);
    }

    [Fact]
    public async Task SendFormFieldsWhenOptionalBodyIsSet()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        await sdk.RequestBodies.OptionalWrappedFormBodyAsync(
            "w1",
            new WidgetForm() { Label = "front side", Count = 3 }
        );

        var sent = Assert.Single(client.Requests);
        Assert.Equal("application/x-www-form-urlencoded", sent.ContentType);
        Assert.Equal("count=3&label=front+side", sent.Body);
    }

    [Fact]
    public async Task RejectNullRequiredMultipartBodyWithoutSending()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        await Assert.ThrowsAsync<ArgumentNullException>(() =>
            sdk.RequestBodies.RequiredWrappedMultipartBodyAsync(null!, "w1")
        );

        Assert.Empty(client.Requests);
    }

    [Fact]
    public async Task RejectNullRequiredFormBodyWithoutSending()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        await Assert.ThrowsAsync<ArgumentNullException>(() =>
            sdk.RequestBodies.RequiredWrappedFormBodyAsync(null!, "w1")
        );

        Assert.Empty(client.Requests);
    }

    [Fact]
    public async Task SendRequiredFormBodyFields()
    {
        var client = new CapturingHttpClient();
        var sdk = new SDK(client: client);

        await sdk.RequestBodies.RequiredWrappedFormBodyAsync(new WidgetForm() { Label = "back" }, "w1");

        var sent = Assert.Single(client.Requests);
        Assert.Equal("/anything/widgets/w1/requiredForm", sent.Path);
        Assert.Equal("label=back", sent.Body);
    }

    private sealed record SentPart(string? Name, string? FileName, string? ContentType, byte[] Bytes);

    private sealed record SentRequest(
        HttpMethod Method,
        string Path,
        string? ContentType,
        string? Body,
        IReadOnlyList<SentPart> Parts
    );

    private sealed class CapturingHttpClient : IDefaultHttpClient
    {
        public List<SentRequest> Requests { get; } = new();

        public async Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken? cancellationToken = null
        )
        {
            var parts = new List<SentPart>();
            string? body = null;
            if (request.Content is MultipartFormDataContent multipart)
            {
                foreach (var part in multipart)
                {
                    parts.Add(
                        new SentPart(
                            part.Headers.ContentDisposition?.Name?.Trim('"'),
                            part.Headers.ContentDisposition?.FileName?.Trim('"'),
                            part.Headers.ContentType?.MediaType,
                            await part.ReadAsByteArrayAsync()
                        )
                    );
                }
            }
            else if (request.Content != null)
            {
                body = await request.Content.ReadAsStringAsync();
            }

            Requests.Add(
                new SentRequest(
                    request.Method,
                    request.RequestUri!.AbsolutePath,
                    request.Content?.Headers.ContentType?.MediaType,
                    body,
                    parts
                )
            );
            return new HttpResponseMessage(HttpStatusCode.OK) { RequestMessage = request };
        }

        public Task<HttpRequestMessage> CloneAsync(HttpRequestMessage request) =>
            Task.FromResult(request);
    }
}
