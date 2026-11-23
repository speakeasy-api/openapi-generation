#nullable enable
using System.Collections.Generic;
using System.Net;
using System.Threading.Tasks;
using Openapi;
using Openapi.Models.Shared;
using Xunit;

public class FormUnionFieldsAdditionalShould
{
    [Fact]
    public async Task SendStringEnumAndJsonVariants()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateStr("abc"),
            StringOrArrayUnion = FormUnionStringOrArray.CreateStr("solo"),
            EnumOrBooleanUnion = FormUnionEnumOrBoolean.CreateFormUnionEnum(FormUnionEnum.Second),
            ArrayOfUnion = new List<FormUnionStringOrInteger>()
            {
                FormUnionStringOrInteger.CreateStr("x"),
                FormUnionStringOrInteger.CreateInteger(3),
            },
            StringOrArrayOfUnion = FormUnionStringOrArrayOfUnion.CreateStr("plain"),
            NullableUnion = FormUnionNullable.CreateNull(),
            JsonUnion = FormUnionObjectOrString.CreateFormUnionObject(new FormUnionObject()
            {
                Label = "x",
                Count = 1,
            }),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("abc", form["scalarUnion"]);
        Assert.Equal("solo", form["stringOrArrayUnion"]);
        Assert.Equal("second", form["enumOrBooleanUnion"]);
        Assert.Equal(
            new List<object>() { "x", "3" },
            form[$"arrayOfUnion{Compat.FormExplodedArrayKeySuffix}"]
        );
        Assert.Equal("plain", form["stringOrArrayOfUnion"]);
        Assert.True(Compat.RawEquals(Compat.ParseRaw((string)form["jsonUnion"]), "{\"label\":\"x\",\"count\":1}"));
        Assert.False(form.ContainsKey("nullableUnion"));
        Assert.False(form.ContainsKey("optionalUnion"));
        Assert.Equal(6, form.Count);
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
            StringOrArrayOfUnion = FormUnionStringOrArrayOfUnion.CreateArrayOfFormUnionIntegerOrBoolean(
                new List<FormUnionIntegerOrBoolean>()
                {
                    FormUnionIntegerOrBoolean.CreateInteger(1),
                    FormUnionIntegerOrBoolean.CreateBoolean(true),
                }
            ),
            NullableUnion = FormUnionNullable.CreateInteger(7),
            OptionalUnion = FormUnionStringOrInteger.CreateStr("set"),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("42", form["scalarUnion"]);
        Assert.Equal(
            new List<object>() { "a", "b" },
            form[$"stringOrArrayUnion{Compat.FormExplodedArrayKeySuffix}"]
        );
        Assert.Equal("true", form["enumOrBooleanUnion"]);
        Assert.Equal(
            new List<object>() { "1", "true" },
            form[$"stringOrArrayOfUnion{Compat.FormExplodedArrayKeySuffix}"]
        );
        Assert.Equal("7", form["nullableUnion"]);
        Assert.Equal("set", form["optionalUnion"]);
        Assert.False(form.ContainsKey("jsonUnion"));
        Assert.Equal(6, form.Count);
    }

    [Fact]
    public async Task SendDefaultValuesAndSingleElementArrays()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateInteger(0),
            StringOrArrayUnion = FormUnionStringOrArray.CreateArrayOfStr(new List<string>() { "only" }),
            EnumOrBooleanUnion = FormUnionEnumOrBoolean.CreateBoolean(false),
            ArrayOfUnion = new List<FormUnionStringOrInteger>()
            {
                FormUnionStringOrInteger.CreateInteger(0),
            },
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("0", form["scalarUnion"]);
        Assert.Equal("only", form["stringOrArrayUnion"]);
        Assert.Equal("false", form["enumOrBooleanUnion"]);
        Assert.Equal("0", form["arrayOfUnion"]);
        Assert.Equal(4, form.Count);
    }

    [Theory]
    [InlineData("\"raw\"", "raw")]
    [InlineData("7", "7")]
    [InlineData("true", "true")]
    public async Task SendPrimitiveUnknownValueOfOpenUnion(string rawJson, string expected)
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateStr("abc"),
            OpenUnion = FormUnionOpenStringOrInteger.CreateUnknown(Compat.ParseRaw(rawJson)),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal(expected, form["openUnion"]);
        Assert.Equal(2, form.Count);
    }

    [Theory]
    [InlineData("{\"label\":\"x\"}")]
    [InlineData("[1,2]")]
    [InlineData("null")]
    public async Task OmitNonPrimitiveUnknownValueOfOpenUnion(string rawJson)
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionFieldsAsync(new FormUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateStr("abc"),
            OpenUnion = FormUnionOpenStringOrInteger.CreateUnknown(Compat.ParseRaw(rawJson)),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.False(form.ContainsKey("openUnion"));
        Assert.Single(form);
    }

    [Fact]
    public async Task SendUnionPropertyOfWholeBodyUnionVariant()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionBodyAsync(
            FormUnionBody.CreateFormUnionBodyLabeled(new FormUnionBodyLabeled()
            {
                Label = "x",
                LabelValue = FormUnionStringOrInteger.CreateInteger(5),
                LabelJson = FormUnionObjectOrString.CreateFormUnionObject(new FormUnionObject()
                {
                    Label = "y",
                    Count = 2,
                }),
            })
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("x", form["label"]);
        Assert.Equal("5", form["labelValue"]);
        Assert.True(Compat.RawEquals(Compat.ParseRaw((string)form["labelJson"]), "{\"label\":\"y\",\"count\":2}"));
        Assert.Equal(3, form.Count);
    }

    [Fact]
    public async Task SendJsonStringVariantOfWholeBodyUnionProperty()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionBodyAsync(
            FormUnionBody.CreateFormUnionBodyLabeled(new FormUnionBodyLabeled()
            {
                Label = "x",
                LabelValue = FormUnionStringOrInteger.CreateStr("v"),
                LabelJson = FormUnionObjectOrString.CreateStr("s"),
            })
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("v", form["labelValue"]);
        Assert.Equal("\"s\"", form["labelJson"]);
        Assert.Equal(3, form.Count);
    }

    [Fact]
    public async Task SendUnionValuesOfWholeBodyMapVariant()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormUnionBodyAsync(
            FormUnionBody.CreateMapOfFormUnionStringOrInteger(new Dictionary<string, FormUnionStringOrInteger>()
            {
                { "first", FormUnionStringOrInteger.CreateStr("x") },
                { "second", FormUnionStringOrInteger.CreateInteger(2) },
            })
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("x", form["first"]);
        Assert.Equal("2", form["second"]);
        Assert.Equal(2, form.Count);
    }

    [Fact]
    public async Task SendMultipartUnionParts()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipartUnionFieldsAsync(new MultipartUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateStr("abc"),
            StringOrArrayUnion = FormUnionStringOrArray.CreateArrayOfStr(new List<string>() { "a", "b" }),
            ArrayOfUnion = new List<FormUnionStringOrInteger>()
            {
                FormUnionStringOrInteger.CreateStr("x"),
                FormUnionStringOrInteger.CreateInteger(2),
            },
            NullableUnion = FormUnionNullable.CreateNull(),
            JsonUnion = FormUnionObjectOrString.CreateFormUnionObject(new FormUnionObject()
            {
                Label = "z",
            }),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("abc", form["scalarUnion"]);
        Assert.Equal(new List<object>() { "a", "b" }, form["stringOrArrayUnion"]);
        Assert.Equal(new List<object>() { "x", "2" }, form["arrayOfUnion"]);
        Assert.True(Compat.RawEquals(Compat.ParseRaw((string)form["jsonUnion"]), "{\"label\":\"z\"}"));
        Assert.False(form.ContainsKey("nullableUnion"));
        Assert.Equal(4, form.Count);
    }

    [Fact]
    public async Task SendMultipartIntegerVariants()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipartUnionFieldsAsync(new MultipartUnionFields()
        {
            ScalarUnion = FormUnionStringOrInteger.CreateInteger(0),
            NullableUnion = FormUnionNullable.CreateInteger(0),
        });

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var form = res.FormUnionEcho!.Form!;

        Assert.Equal("0", form["scalarUnion"]);
        Assert.Equal("0", form["nullableUnion"]);
        Assert.Equal(2, form.Count);
    }
}
