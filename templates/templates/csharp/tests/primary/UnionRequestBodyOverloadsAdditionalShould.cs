#nullable enable
using System;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;
using Xunit;

public class UnionRequestBodyOverloadsAdditionalShould
{
    private static int OverloadCount(System.Type owner, string method) =>
        owner.GetMethods().Count(m => m.Name == method);

    [Fact]
    public void EmitOneOverloadPerVariantOnClassAndInterface()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-discriminated");

        foreach (var owner in new[] { typeof(IUnions), typeof(Unions) })
        {
            Assert.Equal(3, OverloadCount(owner, "BodyVariantOverloadsRenamedParamsAsync"));
            Assert.Equal(1, OverloadCount(owner, "BodyVariantOverloadsOptionalAsync"));
        }
    }

    [Fact]
    public void SkipNullableBodiesAndAmbiguousOrObjectMembers()
    {
        foreach (var owner in new[] { typeof(IUnions), typeof(Unions) })
        {
            Assert.Equal(1, OverloadCount(owner, "NullableOneOfSchemaPostAsync"));
            Assert.Equal(1, OverloadCount(owner, "DiscriminatedOpenEnumAsync"));
            Assert.Equal(3, OverloadCount(owner, "WeaklyTypedOneOfNullEnumPostAsync"));
            Assert.DoesNotContain(
                owner.GetMethods(),
                m => m.Name == "WeaklyTypedOneOfNullEnumPostAsync"
                    && m.GetParameters()[0].ParameterType == typeof(object)
            );
        }
    }

    [Fact]
    public void RenameVariantParameterThatClashesWithAnotherParameter()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-renamed-argument");

        var variantAOverload = typeof(IUnions)
            .GetMethods()
            .Single(m =>
                m.Name == "BodyVariantOverloadsParameterNameClashAsync"
                && m.GetParameters()[0].ParameterType == typeof(NonDiscriminatedVariantA)
            );
        Assert.Equal(
            new[] { "nonDiscriminatedVariantA2", "nonDiscriminatedVariantA", "cancellationToken" },
            variantAOverload.GetParameters().Select(p => p.Name)
        );
    }

    [Fact]
    public void EmitOverloadsForNonDiscriminatedUnion()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-non-discriminated");

        var parameterTypes = typeof(IUnions)
            .GetMethods()
            .Where(m => m.Name == "BodyVariantOverloadsNonDiscriminatedAsync")
            .Select(m => m.GetParameters()[0].ParameterType)
            .ToHashSet();

        Assert.Equal(
            new[]
            {
                typeof(NonDiscriminatedVariantBody),
                typeof(NonDiscriminatedVariantA),
                typeof(NonDiscriminatedVariantB),
                typeof(NonDiscriminatedVariantC),
            }.ToHashSet(),
            parameterTypes
        );
    }

    [Fact]
    public async Task ForwardToCanonicalMethod()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-request-body");
        var client = new CapturingHttpClient();
        var s = new SDK(client: client);

        var model = new BodyVariantModel
        {
            Kind = (BodyVariantModelKind)(-1),
            Model = "m",
            Input = "i",
        };
        await Assert.ThrowsAsync<CapturedRequestException>(() =>
            s.Unions.BodyVariantOverloadsRenamedParamsAsync(model, 1, "p", "ref", "q")
        );
        Assert.Equal(BodyVariantModelKind.Model, model.Kind);
        Assert.Equal("ref", client.LastRequestRef);
        var viaOverload = (client.LastRequestUri, client.LastRequestBody, client.LastRequestRef);

        await Assert.ThrowsAsync<CapturedRequestException>(() =>
            s.Unions.BodyVariantOverloadsRenamedParamsAsync(
                DiscriminatedVariantBody.CreateModel(new BodyVariantModel { Model = "m", Input = "i" }),
                1,
                "p",
                "ref",
                "q"
            )
        );

        Assert.Equal((client.LastRequestUri, client.LastRequestBody, client.LastRequestRef), viaOverload);
    }

    [Fact]
    public async Task RejectNullVariantWithVariantParameterName()
    {
        var s = new SDK(client: new CapturingHttpClient());

        var task = s.Unions.BodyVariantOverloadsRenamedParamsAsync((BodyVariantModel)null!, 1, "p");
        var ex = await Assert.ThrowsAsync<ArgumentNullException>(() => task);
        Assert.Equal("model", ex.ParamName);
    }

    [Fact]
    public async Task ForwardOpenEnumMember()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-open-enum");
        var s = new SDK();

        var res = await s.Unions.BodyVariantOverloadsOpenEnumAsync(BodyVariantOpenStatus.Of("paused"));

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("\"paused\"", await res.HttpMeta.Request!.Content!.ReadAsStringAsync());

        var task = s.Unions.BodyVariantOverloadsOpenEnumAsync((BodyVariantOpenStatus)null!);
        var ex = await Assert.ThrowsAsync<ArgumentNullException>(() => task);
        Assert.Equal("bodyVariantOpenStatus", ex.ParamName);
    }

    [Fact]
    public async Task ForwardValueTypeMember()
    {
        var s = new SDK();

        Func<long, CancellationToken?, Task<PrimitiveTypeOneOfPostResponse>> overload =
            s.Unions.PrimitiveTypeOneOfPostAsync;
        var res = await overload(111, null);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(PrimitiveTypeOneOfPostJsonType.Integer, res.Res!.Json.Type.ToString());
        Assert.Equal(111, res.Res.Json.Integer);
    }

    [Fact]
    public async Task BindNestedUnionMemberToNestedUnionOverload()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-nested-union");
        var s = new SDK(client: CommonHelpers.CreateTestHTTPClient("smartUnionUnionVsUnion"));

        var viaOverload = await s.Unions.SmartUnionUnionVsUnionAsync(
            new SmartUnionVsUnionA1 { A = "x" }
        );
        var viaWrapper = await s.Unions.SmartUnionUnionVsUnionAsync(
            SmartUnionUnionVsUnionRequestBody.CreateOne(
                Openapi.Models.Operations.One.CreateSmartUnionVsUnionA1(new SmartUnionVsUnionA1 { A = "x" })
            )
        );
        Assert.Equal(
            await viaWrapper.HttpMeta.Request!.Content!.ReadAsStringAsync(),
            await viaOverload.HttpMeta.Request!.Content!.ReadAsStringAsync()
        );

        var task = s.Unions.SmartUnionUnionVsUnionAsync((SmartUnionVsUnionA1)null!);
        var ex = await Assert.ThrowsAsync<ArgumentNullException>(() => task);
        Assert.Equal("one", ex.ParamName);
    }

    [Fact]
    public async Task ForwardServerUrl()
    {
        CommonHelpers.RecordTest("unions-body-variant-overloads-non-body-arguments");
        var client = new CapturingHttpClient();
        var s = new SDK(client: client);

        Func<BodyVariantModel, string?, CancellationToken?, Task<BodyVariantOverloadsServerUrlResponse>> overload =
            s.Unions.BodyVariantOverloadsServerUrlAsync;
        await Assert.ThrowsAsync<CapturedRequestException>(() =>
            overload(new BodyVariantModel { Model = "m", Input = "i" }, "http://overload.invalid", null)
        );

        Assert.Equal("overload.invalid", client.LastRequestUri!.Host);
    }

    private sealed class CapturedRequestException : Exception { }

    private sealed class CapturingHttpClient : IDefaultHttpClient
    {
        public Uri? LastRequestUri { get; private set; }

        public string? LastRequestBody { get; private set; }

        public string? LastRequestRef { get; private set; }

        public async Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken? cancellationToken = null
        )
        {
            LastRequestUri = request.RequestUri;
            LastRequestBody = request.Content == null ? null : await request.Content.ReadAsStringAsync();
            LastRequestRef = request.Headers.TryGetValues("X-Request-Ref", out var values)
                ? string.Join(",", values)
                : null;
            throw new CapturedRequestException();
        }

        public Task<HttpRequestMessage> CloneAsync(HttpRequestMessage request) =>
            Task.FromResult(request);
    }
}
