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
        Assert.True((bool)res.Res.Json["bool"]);
        Assert.Equal(1.1, (double)res.Res.Json["num"], 0.0001);
        Assert.Equal("test", res.Res.Json["str"]);
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
        Assert.True((bool)res.Res.Json["bool"]);
        Assert.Equal(1.1, (double)res.Res.Json["num"]);
        Assert.Equal("test body", res.Res.Json["str"].ToString());
        Assert.Equal("test param", res.Res.Args["paramStr"]);
    }

    [Fact]
    public async Task PostJsonOptionalNullable_Legacy()
    {
        // Legacy-mode equivalent of "request-bodies-post-json-optional-nullable" (presenceAwareJsonSerialization: false):
        // every field is a plain T? with no OptionalNullable<T> wrapper and no tri-state, so JSON cannot distinguish
        // absent from explicit null — both are omitted on the wire and round-trip back to null.

        var sdk = new SDK(serverUrl: CommonHelpers.HttpBinUrl, apiKeyAuth: "Token YOUR_API_KEY");

        var setJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(
                new ObjectWithOptionalNullableFields()
                {
                    ReqStr = "base",
                    OptStr = "opt",
                    OptNullStr = "hello",
                    OptNullInt = 7L,
                    OptNullNum = 1.5,
                    OptNullBool = true,
                    OptNullBigInt = new BigInteger(8821239038968084),
                    OptNullBigIntStr = BigInteger.Parse("9223372036854775808"),
                    OptNullDecimal = 3.141592653589793m,
                    OptNullDecimalStr = 3.14159265358979344719667586m,
                    OptNullArr = new List<string> { "a", "b" },
                    OptNullMap = new Dictionary<string, string> { ["k"] = "v" },
                    OptNullUnion = OptNullUnion.CreateStr("u"),
                    OptNullObj = new OptNullObj() { InnerReqStr = "in", InnerOptNullStr = "innerval" },
                }
            )
        ).Res.Json;
        Assert.Equal("base", setJson.ReqStr);
        Assert.Equal("opt", setJson.OptStr);
        Assert.Equal("hello", setJson.OptNullStr);
        Assert.Equal(7L, setJson.OptNullInt);
        Assert.Equal(1.5, setJson.OptNullNum);
        Assert.Equal(true, setJson.OptNullBool);
        Assert.Equal(new BigInteger(8821239038968084), setJson.OptNullBigInt);
        Assert.Equal(BigInteger.Parse("9223372036854775808"), setJson.OptNullBigIntStr);
        Assert.Equal(3.141592653589793m, setJson.OptNullDecimal);
        Assert.Equal(3.14159265358979344719667586m, setJson.OptNullDecimalStr);
        Assert.Equal(new List<string> { "a", "b" }, setJson.OptNullArr);
        Assert.Equal("v", setJson.OptNullMap!["k"]);
        Assert.Equal("u", setJson.OptNullUnion!.Str);
        Assert.Equal("in", setJson.OptNullObj!.InnerReqStr);
        Assert.Equal("innerval", setJson.OptNullObj.InnerOptNullStr);

        // Explicit null: legacy omits null fields, so they round-trip back to null.
        var nullJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(
                new ObjectWithOptionalNullableFields()
                {
                    ReqStr = "base",
                    OptStr = null,
                    OptNullStr = null,
                    OptNullInt = null,
                    OptNullNum = null,
                    OptNullBool = null,
                    OptNullBigInt = null,
                    OptNullBigIntStr = null,
                    OptNullDecimal = null,
                    OptNullDecimalStr = null,
                    OptNullArr = null,
                    OptNullMap = null,
                    OptNullUnion = null,
                    OptNullObj = null,
                }
            )
        ).Res.Json;
        Assert.Equal("base", nullJson.ReqStr);
        Assert.Null(nullJson.OptStr);
        Assert.Null(nullJson.OptNullStr);
        Assert.Null(nullJson.OptNullInt);
        Assert.Null(nullJson.OptNullNum);
        Assert.Null(nullJson.OptNullBool);
        Assert.Null(nullJson.OptNullBigInt);
        Assert.Null(nullJson.OptNullBigIntStr);
        Assert.Null(nullJson.OptNullDecimal);
        Assert.Null(nullJson.OptNullDecimalStr);
        Assert.Null(nullJson.OptNullArr);
        Assert.Null(nullJson.OptNullMap);
        Assert.Null(nullJson.OptNullUnion);
        Assert.Null(nullJson.OptNullObj);

        // Absent: unset fields are omitted too — indistinguishable from explicit null.
        var absentJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(
                new ObjectWithOptionalNullableFields() { ReqStr = "base" }
            )
        ).Res.Json;
        Assert.Equal("base", absentJson.ReqStr);
        Assert.Null(absentJson.OptStr);
        Assert.Null(absentJson.OptNullStr);
        Assert.Null(absentJson.OptNullInt);
        Assert.Null(absentJson.OptNullNum);
        Assert.Null(absentJson.OptNullBool);
        Assert.Null(absentJson.OptNullBigInt);
        Assert.Null(absentJson.OptNullBigIntStr);
        Assert.Null(absentJson.OptNullDecimal);
        Assert.Null(absentJson.OptNullDecimalStr);
        Assert.Null(absentJson.OptNullArr);
        Assert.Null(absentJson.OptNullMap);
        Assert.Null(absentJson.OptNullUnion);
        Assert.Null(absentJson.OptNullObj);
    }
}
