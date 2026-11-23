using Xunit;
using Speakeasy.BasicHttp;
using Speakeasy.BasicHttp.Models.Components;
using System;
using System.Net;
using System.Threading.Tasks;

public class UnionScalarVariantsAdditionalShould
{
    private static SDK CreateSdk() => new SDK(
        security: new Security()
        {
            Username = "testUser",
            Password = "testPass"
        }
    );

    [Fact]
    public async Task RoundTripScalarVariants()
    {
        var sent = new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc);

        var res = await CreateSdk().Unions.UnionScalarVariantsPostAsync(new ScalarVariantUnions()
        {
            Integer = IntegerOrObject.CreateInteger(5),
            Number = NumberOrObject.CreateNumber(1.5),
            DateTime = DateTimeOrObject.CreateDateTime(sent),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var json = res.Res!.Json;
        Assert.Equal(5, json.Integer.Integer);
        Assert.Equal(1.5, json.Number.Number);
        Assert.Equal(sent, json.DateTime.DateTime!.Value.ToUniversalTime());
    }

    [Fact]
    public async Task RoundTripObjectVariants()
    {
        var res = await CreateSdk().Unions.UnionScalarVariantsPostAsync(new ScalarVariantUnions()
        {
            Integer = IntegerOrObject.CreateScalarVariantObject(new ScalarVariantObject() { Name = "a" }),
            Number = NumberOrObject.CreateScalarVariantObject(new ScalarVariantObject() { Name = "b" }),
            DateTime = DateTimeOrObject.CreateScalarVariantObject(new ScalarVariantObject() { Name = "c" }),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var json = res.Res!.Json;
        Assert.Equal("a", json.Integer.ScalarVariantObject!.Name);
        Assert.Equal("b", json.Number.ScalarVariantObject!.Name);
        Assert.Equal("c", json.DateTime.ScalarVariantObject!.Name);
    }
}
