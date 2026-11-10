#nullable enable
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using Newtonsoft.Json.Linq;
using Openapi;
using Openapi.Models.Shared;
using Xunit;

public class FormUnionFieldsAdditionalShould
{
    private const string FormExplodedArrayKeySuffix = "[]";

    [Fact]
    public async Task SendStringEnumAndJsonVariants()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateStr("abc"),
            StringOrArrayUnion = FormUnionStringOrArray.CreateStr("solo"),
            EnumOrBooleanUnion = FormUnionEnumOrBoolean.CreateFormUnionEnum(FormUnionEnum.Second),
            NullableUnion = FormUnionNullable.CreateNull(),
            JsonUnion = FormUnionObjectOrString.CreateFormUnionObject(new FormUnionObject()
            {
                Label = "x",
                Count = 1,
            }),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.Object!.Form!;

        Assert.Equal("abc", form["scalarUnion"]);
        Assert.Equal("solo", form["stringOrArrayUnion"]);
        Assert.Equal("second", form["enumOrBooleanUnion"]);
        Assert.True(JToken.DeepEquals(
            JObject.Parse("{\"label\":\"x\",\"count\":1}"),
            JObject.Parse((string)form["jsonUnion"])
        ));
        Assert.False(form.ContainsKey("nullableUnion"));
        Assert.False(form.ContainsKey("optionalUnion"));
        Assert.Equal(4, form.Count);
    }

    [Fact]
    public async Task SendIntegerArrayAndBooleanVariants()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateInteger(42),
            StringOrArrayUnion = FormUnionStringOrArray.CreateArrayOfStr(new List<string>() { "a", "b" }),
            EnumOrBooleanUnion = FormUnionEnumOrBoolean.CreateBoolean(true),
            NullableUnion = FormUnionNullable.CreateInteger(7),
            OptionalUnion = FormUnionStringOrInteger.CreateStr("set"),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.Object!.Form!;

        Assert.Equal("42", form["scalarUnion"]);
        Assert.Equal(
            new List<object>() { "a", "b" },
            form[$"stringOrArrayUnion{FormExplodedArrayKeySuffix}"]
        );
        Assert.Equal("true", form["enumOrBooleanUnion"]);
        Assert.Equal("7", form["nullableUnion"]);
        Assert.Equal("set", form["optionalUnion"]);
        Assert.False(form.ContainsKey("jsonUnion"));
        Assert.Equal(5, form.Count);
    }
}
