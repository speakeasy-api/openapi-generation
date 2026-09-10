#nullable enable
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Numerics;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Openapi;
using Openapi.Models.Operations;
using Openapi.Models.Shared;
using Openapi.Utils;
using Xunit;

public class RequestBodiesShould
{
    [Fact]
    public async Task PostApplicationJsonSimple()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-simple");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonSimpleAsync(
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertSimpleObject(res.Res!.Json!);
    }

    [Fact]
    public async Task PostApplicationJsonArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayAsync(
            new List<SimpleObject>() { Helpers.CreateSimpleObject(), Helpers.CreateSimpleObject() }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        foreach (var obj in res.Res!)
        {
            Helpers.AssertSimpleObject(obj);
        }
    }

    [Fact]
    public async Task PostApplicationJsonArrayCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayCamelCaseAsync(
            new List<SimpleObjectCamelCase>()
            {
                Helpers.CreateSimpleObjectCamelCase(),
                Helpers.CreateSimpleObjectCamelCase()
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        foreach (var obj in res.Res!)
        {
            Helpers.AssertSimpleObjectCamelCase(obj);
        }
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-array");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfArrayAsync(
            new List<List<SimpleObject>>
            {
                new List<SimpleObject>() { obj, obj },
                new List<SimpleObject>() { obj, obj }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());

        for (var i = 0; i < 2; i++)
        {
            Assert.Equal(2, res.Res!.ToList()[i].Count());
            for (var j = 0; j < 2; j++)
            {
                Helpers.AssertSimpleObject(res.Res!.ToList()[i].ToList()[j]);
            }
        }
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfArrayCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-array-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfArrayCamelCaseAsync(
            new List<List<SimpleObjectCamelCase>>
            {
                new List<SimpleObjectCamelCase>() { obj, obj },
                new List<SimpleObjectCamelCase>() { obj, obj }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());

        for (var i = 0; i < 2; i++)
        {
            Assert.Equal(2, res.Res!.ToList()[i].Count());
            for (var j = 0; j < 2; j++)
            {
                Helpers.AssertSimpleObjectCamelCase(res.Res!.ToList()[i].ToList()[j]);
            }
        }
    }

    [Fact]
    public async Task PostApplicationJsonMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapAsync(
            new Dictionary<string, SimpleObject>() { { "mapElem1", obj }, { "mapElem2", obj } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Helpers.AssertSimpleObject(res.Res!["mapElem1"]);
        Helpers.AssertSimpleObject(res.Res!["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapCamelCaseAsync(
            new Dictionary<string, SimpleObjectCamelCase>()
            {
                { "mapElem1", obj },
                { "mapElem2", obj }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfMapAsync(
            new Dictionary<string, Dictionary<string, SimpleObject>>()
            {
                {
                    "mapElem1",
                    new Dictionary<string, SimpleObject>()
                    {
                        { "subMapElem1", obj },
                        { "subMapElem2", obj }
                    }
                },
                {
                    "mapElem2",
                    new Dictionary<string, SimpleObject>()
                    {
                        { "subMapElem1", obj },
                        { "subMapElem2", obj }
                    }
                },
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!["mapElem1"].Count());
        Assert.Equal(2, res.Res!["mapElem2"].Count());
        Helpers.AssertSimpleObject(res.Res!["mapElem1"]["subMapElem1"]);
        Helpers.AssertSimpleObject(res.Res!["mapElem1"]["subMapElem2"]);
        Helpers.AssertSimpleObject(res.Res!["mapElem2"]["subMapElem1"]);
        Helpers.AssertSimpleObject(res.Res!["mapElem2"]["subMapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfMapCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-map-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfMapCamelCaseAsync(
            new Dictionary<string, Dictionary<string, SimpleObjectCamelCase>>()
            {
                {
                    "mapElem1",
                    new Dictionary<string, SimpleObjectCamelCase>()
                    {
                        { "subMapElem1", obj },
                        { "subMapElem2", obj }
                    }
                },
                {
                    "mapElem2",
                    new Dictionary<string, SimpleObjectCamelCase>()
                    {
                        { "subMapElem1", obj },
                        { "subMapElem2", obj }
                    }
                },
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!["mapElem1"].Count());
        Assert.Equal(2, res.Res!["mapElem2"].Count());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem1"]["subMapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem1"]["subMapElem2"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem2"]["subMapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem2"]["subMapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfAny()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-any");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfAnyAsync(
            new Dictionary<string, dynamic>()
            {
                {
                    "mapElemArray",
                    new List<string>() { "array-test" }
                },
                {
                    "mapElemBool",
                    true
                },
                {
                    "mapElemNumber",
                    123
                },
                {
                    "mapElemObject",
                    new { property = "property-test" }
                },
                {
                    "mapElemString",
                    "string-test"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(5, res.Res!.Count());
        Assert.Collection(
            ((IEnumerable<object>)res.Res!["mapElemArray"]).Cast<string>(),
            item => Assert.Equal("array-test", item)
        );
        Assert.Equal(true, res.Res!["mapElemBool"]);
        Assert.Equal(123L, res.Res!["mapElemNumber"]);
        Assert.Equal("property-test", ((Dictionary<string, object>)res.Res!["mapElemObject"])["property"]);
        Assert.Equal("string-test", res.Res!["mapElemString"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-array");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfArrayAsync(
            new Dictionary<string, List<SimpleObject>>()
            {
                {
                    "mapElem1",
                    new List<SimpleObject>() { obj, obj }
                },
                {
                    "mapElem2",
                    new List<SimpleObject>() { obj, obj }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!["mapElem1"].Count());
        Assert.Equal(2, res.Res!["mapElem2"].Count());
        Helpers.AssertSimpleObject(res.Res!["mapElem1"].First());
        Helpers.AssertSimpleObject(res.Res!["mapElem1"].Last());
        Helpers.AssertSimpleObject(res.Res!["mapElem2"].First());
        Helpers.AssertSimpleObject(res.Res!["mapElem2"].Last());
    }

    [Fact]
    public async Task PostApplicationJsonMapOfArrayCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-array-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfArrayCamelCaseAsync(
            new Dictionary<string, List<SimpleObjectCamelCase>>()
            {
                {
                    "mapElem1",
                    new List<SimpleObjectCamelCase>() { obj, obj }
                },
                {
                    "mapElem2",
                    new List<SimpleObjectCamelCase>() { obj, obj }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!["mapElem1"].Count());
        Assert.Equal(2, res.Res!["mapElem2"].Count());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem1"].First());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem1"].Last());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem2"].First());
        Helpers.AssertSimpleObjectCamelCase(res.Res!["mapElem2"].Last());
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfMap()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-map");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var maps = new List<Dictionary<string, SimpleObject>>();
        for (int i = 0; i < 2; i++)
        {
            maps.Add(
                new Dictionary<string, SimpleObject>()
                {
                    { "mapElem1", Helpers.CreateSimpleObject() },
                    { "mapElem2", Helpers.CreateSimpleObject() }
                }
            );
        }

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfMapAsync(maps);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!.ToList()[0].Count());
        Assert.Equal(2, res.Res!.ToList()[1].Count());
        Helpers.AssertSimpleObject(res.Res!.ToList()[0]["mapElem1"]);
        Helpers.AssertSimpleObject(res.Res!.ToList()[0]["mapElem2"]);
        Helpers.AssertSimpleObject(res.Res!.ToList()[1]["mapElem1"]);
        Helpers.AssertSimpleObject(res.Res!.ToList()[1]["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfMapCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-map-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var maps = new List<Dictionary<string, SimpleObjectCamelCase>>();
        for (int i = 0; i < 2; i++)
        {
            maps.Add(
                new Dictionary<string, SimpleObjectCamelCase>()
                {
                    { "mapElem1", Helpers.CreateSimpleObjectCamelCase() },
                    { "mapElem2", Helpers.CreateSimpleObjectCamelCase() }
                }
            );
        }

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfMapCamelCaseAsync(
            maps
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!.ToList()[0].Count());
        Assert.Equal(2, res.Res!.ToList()[1].Count());
        Helpers.AssertSimpleObjectCamelCase(res.Res!.ToList()[0]["mapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!.ToList()[0]["mapElem2"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!.ToList()[1]["mapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.Res!.ToList()[1]["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-primitive");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfPrimitiveAsync(
            new Dictionary<string, string>() { { "mapElem1", "hello" }, { "mapElem2", "world" } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal("hello", res.Res!["mapElem1"]);
        Assert.Equal("world", res.Res!["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-of-primitive");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfPrimitiveAsync(
            new List<string>() { "hello", "world" }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal("hello", res.Res!.ToList()[0]);
        Assert.Equal("world", res.Res!.ToList()[1]);
    }

    [Fact]
    public async Task PostApplicationJsonMapOfMapOfPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-of-map-of-primitive");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapOfMapOfPrimitiveAsync(
            new Dictionary<string, Dictionary<string, string>>()
            {
                {
                    "mapElem1",
                    new Dictionary<string, string>()
                    {
                        { "subMapElem1", "foo" },
                        { "subMapElem2", "bar" }
                    }
                },
                {
                    "mapElem2",
                    new Dictionary<string, string>()
                    {
                        { "subMapElem1", "buzz" },
                        { "subMapElem2", "bazz" }
                    }
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!["mapElem1"].Count());
        Assert.Equal(2, res.Res!["mapElem2"].Count());
        Assert.Equal("foo", res.Res!["mapElem1"]["subMapElem1"]);
        Assert.Equal("bar", res.Res!["mapElem1"]["subMapElem2"]);
        Assert.Equal("buzz", res.Res!["mapElem2"]["subMapElem1"]);
        Assert.Equal("bazz", res.Res!["mapElem2"]["subMapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonArrayOfArrayOfPrimitive()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-application-json-array-of-array-of-primitive"
        );
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res =
            await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayOfArrayOfPrimitiveAsync(
                new List<List<string>>()
                {
                    new List<string>() { "foo", "bar" },
                    new List<string>() { "buzz", "bazz" }
                }
            );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.Res!.Count());
        Assert.Equal(2, res.Res!.First().Count());
        Assert.Equal(2, res.Res!.Last().Count());
        Assert.Equal("foo", res.Res!.First().First());
        Assert.Equal("bar", res.Res!.First().Last());
        Assert.Equal("buzz", res.Res!.Last().First());
        Assert.Equal("bazz", res.Res!.Last().Last());
    }

    [Fact]
    public async Task PostApplicationJsonArrayObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayObjAsync(
            new List<SimpleObject>() { obj, obj }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.ArrObjValue!.Json.Count());
        Helpers.AssertSimpleObject(res.ArrObjValue!.Json.ToList()[0]);
        Helpers.AssertSimpleObject(res.ArrObjValue!.Json.ToList()[1]);
    }

    [Fact]
    public async Task PostApplicationJsonArrayObjectCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-array-object-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonArrayObjCamelCaseAsync(
            new List<SimpleObjectCamelCase>() { obj, obj }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.ArrObjValueCamelCase!.Json.Count());
        Helpers.AssertSimpleObjectCamelCase(res.ArrObjValueCamelCase!.Json.ToList()[0]);
        Helpers.AssertSimpleObjectCamelCase(res.ArrObjValueCamelCase!.Json.ToList()[1]);
    }

    [Fact]
    public async Task PostApplicationJsonMapObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-object");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObject();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapObjAsync(
            new Dictionary<string, SimpleObject>() { { "mapElem1", obj }, { "mapElem2", obj } }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.MapObjValue!.Json.Count());
        Helpers.AssertSimpleObject(res.MapObjValue!.Json["mapElem1"]);
        Helpers.AssertSimpleObject(res.MapObjValue!.Json["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonMapObjectCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-map-object-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateSimpleObjectCamelCase();

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMapObjCamelCaseAsync(
            new Dictionary<string, SimpleObjectCamelCase>()
            {
                { "mapElem1", obj },
                { "mapElem2", obj }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(2, res.MapObjValueCamelCase!.Json.Count());
        Helpers.AssertSimpleObjectCamelCase(res.MapObjValueCamelCase!.Json["mapElem1"]);
        Helpers.AssertSimpleObjectCamelCase(res.MapObjValueCamelCase!.Json["mapElem2"]);
    }

    [Fact]
    public async Task PostApplicationJsonDeep()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-deep");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonDeepAsync(
            Helpers.CreateDeepObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertDeepObject(res.Res!.Json!);
    }

    [Fact]
    public async Task PostApplicationJsonDeepCamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-deep-camel-case");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonDeepCamelCaseAsync(
            Helpers.CreateDeepObjectCamelCase()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertDeepObjectCamelCase(res.Res!.Json!);
    }

    [Fact]
    public async Task PostApplicationJsonMultipleJsonFiltered()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-multiple-json-filtered");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonMultipleJsonFilteredAsync(
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertSimpleObject(res.Res!.Json!);
    }

    [Fact]
    public async Task PostMultipleContentTypesComponentFiltered()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-component-filtered-application-json"
        );
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-component-filtered-multipart-form-data"
        );
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesComponentFilteredAsync(
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertSimpleObject(res.Res!.Json!);
    }

    [Fact]
    public async Task PostMultipleContentTypesInlineFiltered()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-inline-filtered");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesInlineFilteredAsync(
            new RequestBodyPostMultipleContentTypesInlineFilteredRequestBody()
            {
                Bool = true,
                Num = 1.1F,
                Str = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(3, res.Res!.Json!.Count());
        Assert.True((bool)res.Res!.Json!["bool"]);
        Assert.Equal(1.1, (double)res.Res!.Json!["num"], 0.0001);
        Assert.Equal("test", res.Res!.Json!["str"]);
    }

    [Fact]
    public async Task PostMultipleContentTypeSplitJson()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-json");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitJsonAsync(
            new RequestBodyPostMultipleContentTypesSplitJsonRequestBody()
            {
                Bool = true,
                Num = 1.1F,
                Str = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True((bool)res.Res!.Json!.Value!["bool"]);
        Assert.Equal(1.1, (double)res.Res!.Json!.Value!["num"], 0.0001);
        Assert.Equal("test", res.Res!.Json!.Value!["str"]);
    }

    [Fact]
    public async Task PostMutlipleContentTypesSplitMultipart()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-multipart");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitMultipartAsync(
            new RequestBodyPostMultipleContentTypesSplitMultipartRequestBody()
            {
                Bool2 = true,
                Num2 = 1.1D,
                Str2 = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);

        Assert.Equal("true", res.Res!.Form!["bool2"]);
        Assert.Equal("1.1", res.Res!.Form!["num2"]);
        Assert.Equal("test", res.Res!.Form!["str2"]);
    }

    [Fact]
    public async Task PostMultipleContentTypesSplitForm()
    {
        CommonHelpers.RecordTest("request-bodies-post-multiple-content-types-split-form");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitFormAsync(
            new RequestBodyPostMultipleContentTypesSplitFormRequestBody()
            {
                Bool3 = true,
                Num3 = 1.1D,
                Str3 = "test"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("true", res.Res!.Form!["bool3"]);
        Assert.Equal("1.1", res.Res!.Form!["num3"]);
        Assert.Equal("test", res.Res!.Form!["str3"]);
    }

    [Fact]
    public async Task PostMultipleContentTypesSplitJsonWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-json-with-param"
        );

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

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

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.True((bool)res.Res!.Json!.Value!["bool"]);
        Assert.Equal(1.1, (double)res.Res!.Json!.Value!["num"]);
        Assert.Equal("test body", res.Res!.Json!.Value!["str"].ToString());
        Assert.Equal("test param", res.Res!.Args!["paramStr"]);
    }

    [Fact]
    public async Task PostMultipleContentTypesSplitMultiplartWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-multipart-with-param"
        );

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var formData = new RequestBodyPostMultipleContentTypesSplitParamMultipartRequestBody()
        {
            Bool2 = true,
            Num2 = 1.1D,
            Str2 = "test body"
        };

        var res =
            await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamMultipartAsync(
                formData,
                "test param"
            );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal("true", res.Res!.Form!["bool2"]);
        Assert.Equal("1.1", res.Res!.Form!["num2"]);
        Assert.Equal("test body", res.Res!.Form!["str2"]);
        Assert.Equal("test param", res.Res!.Args!["paramStr"]);
    }

    [Fact]
    public async Task PostMultipleContentTypesSplitFormWithParam()
    {
        CommonHelpers.RecordTest(
            "request-bodies-post-multiple-content-types-split-form-with-param"
        );

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var requestBody = new RequestBodyPostMultipleContentTypesSplitParamFormRequestBody()
        {
            Bool3 = true,
            Num3 = 1.1D,
            Str3 = "test body"
        };

        var res = await sdk.RequestBodies.RequestBodyPostMultipleContentTypesSplitParamFormAsync(
            requestBody,
            "test param"
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal("true", res.Res!.Form!["bool3"]);
        Assert.Equal("1.1", res.Res!.Form!["num3"]);
        Assert.Equal("test body", res.Res!.Form!["str3"]);
        Assert.Equal("test param", res.Res!.Args!["paramStr"]);
    }

    [Fact]
    public async Task PutMultipartSimple()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-simple");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPutMultipartSimpleAsync(
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("any", res.Res!.Form.Any);
        Assert.Equal("true", res.Res!.Form.Bool);
        Assert.Equal("true", res.Res!.Form.BoolOpt);
        Assert.Equal("2020-01-01", res.Res!.Form.Date);
        Assert.Equal("2020-01-01T00:00:00.0000001Z", res.Res!.Form.DateTime);
        Assert.Equal("one", res.Res!.Form.Enum);
        Assert.Equal("1.1", res.Res!.Form.Float32);
        Assert.Equal("1", res.Res!.Form.Int);
        Assert.Equal("1", res.Res!.Form.Int32);
        Assert.Equal("1.1", res.Res!.Form.Num);
        Assert.Equal("test", res.Res!.Form.Str);
        Assert.Equal("testOptional", res.Res!.Form.StrOpt);
    }

    // ── Optional + nullable request bodies: one shared component (ObjectWithOptionalNullableFields)
    // round-tripped through the json, form and multipart serializers ─────────────
    //
    // The single generated class carries the union of content-type annotations, so
    // each optional+nullable field is an OptionalNullable<T> — across primitives, an
    // array, a map, a union and a nested object that itself has an optional+nullable
    // field. JSON must preserve the tri-state (absent / explicit null / set) at every
    // level; form and multipart, which have no native null, must OMIT both absent and
    // explicit-null fields and serialize a set field as its bare value (never leaking
    // the "absent" / "null" / "\"value\"" sentinels from
    // OptionalNullable.ToString()).

    // JSON round trip covers the full field matrix, including complex and nested types.
    private static ObjectWithOptionalNullableFields OptionalNullableJsonAllSet() =>
        new ObjectWithOptionalNullableFields()
        {
            ReqStr = "base",
            OptStr = "opt",
            // Passing ISO-date-time-shaped value to ensure OptionalNullable<string>
            // doesn't reformat date-shaped strings on read through DateParseHandling
            OptNullStr = "2024-01-02T03:04:05Z",
            OptNullInt = 7L,
            OptNullNum = 1.5,
            OptNullBool = true,
            OptNullBigInt = new BigInteger(8821239038968084),
            OptNullBigIntStr = BigInteger.Parse("9223372036854775808"),
            OptNullDecimal = 3.141592653589793m,
            OptNullDecimalStr = 3.14159265358979344719667586m,
            OptNullArr = new List<string> { "a", "b" },
            OptNullBigIntStrArr = new List<BigInteger?>
            {
                BigInteger.Parse("9223372036854775808"),
                new BigInteger(7),
                null,
            },
            OptNullMap = new Dictionary<string, string> { ["k"] = "v" },
            OptNullDecimalStrMap = new Dictionary<string, decimal?>
            {
                ["pi"] = 3.14159265358979344719667586m,
                ["none"] = null,
            },
            OptNullUnion = OptNullUnion.CreateStr("u"),
            OptNullObj = new OptNullObj() { InnerReqStr = "in", InnerOptNullStr = "innerval" },
            OptNullDateTime = new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc),
            OptNullDate = new DateOnly(2024, 1, 2),
            OptNullEnum = OptNullEnum.First,
        };

    // Form/multipart set-value assertions stay on primitives: complex-type form
    // encoding is format-idiosyncratic to assert exactly. Complex/nested fields are
    // still exercised here by the null/absent cases below (proving the serializer
    // unwraps and omits List, Dictionary, union and nested-class values), and their
    // set values are covered by the JSON round trip. Building here also asserts the
    // shared component keeps the tri-state on each optional+nullable field: IsSet and
    // not IsNull for a populated value.
    private static ObjectWithOptionalNullableFields OptionalNullableFieldsSet()
    {
        var o = new ObjectWithOptionalNullableFields()
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
            OptNullDateTime = new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc),
            OptNullDate = new DateOnly(2024, 1, 2),
            OptNullEnum = OptNullEnum.First,
        };
        Helpers.AssertObjectWithOptionalNullableFieldsPrimitivesSet(o);
        return o;
    }

    // Every optional field explicitly null, including complex and nested types.
    // Asserts each optional+nullable field is present with a null value: IsSet && IsNull.
    private static ObjectWithOptionalNullableFields OptionalNullableFieldsNull()
    {
        var o = new ObjectWithOptionalNullableFields()
        {
            ReqStr = "base",
            OptStr = null,
            OptNullStr = (string?)null,
            OptNullInt = (long?)null,
            OptNullNum = (double?)null,
            OptNullBool = (bool?)null,
            OptNullBigInt = (BigInteger?)null,
            OptNullBigIntStr = (BigInteger?)null,
            OptNullDecimal = (decimal?)null,
            OptNullDecimalStr = (decimal?)null,
            OptNullArr = (List<string>?)null,
            OptNullBigIntStrArr = (List<BigInteger?>?)null,
            OptNullMap = (Dictionary<string, string>?)null,
            OptNullDecimalStrMap = (Dictionary<string, decimal?>?)null,
            OptNullUnion = (OptNullUnion?)null,
            OptNullObj = (OptNullObj?)null,
            OptNullDateTime = (DateTime?)null,
            OptNullDate = (DateOnly?)null,
            OptNullEnum = (OptNullEnum?)null,
        };
        // Explicit null: every optional+nullable field is present with a null value.
        Helpers.AssertObjectWithOptionalNullableFieldsAllNull(o);
        return o;
    }

    // Every optional field absent. Absent is neither set nor null — distinct from
    // explicit null (IsSet && IsNull).
    private static ObjectWithOptionalNullableFields OptionalNullableFieldsAbsent()
    {
        var o = new ObjectWithOptionalNullableFields() { ReqStr = "base" };
        Helpers.AssertObjectWithOptionalNullableFieldsAllAbsent(o);
        return o;
    }

    [Fact]
    public async Task PostJsonOptionalNullable()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-optional-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Set: every field present, tri-state survives the round trip (incl. nested).
        var setJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(OptionalNullableJsonAllSet())
        ).Res!.Json;
        Assert.Equal("base", setJson.ReqStr);
        Assert.Equal("opt", setJson.OptStr);
        Helpers.AssertObjectWithOptionalNullableFieldsAllSet(setJson);
        // OptNullStr carries an ISO-date-time-shaped value: it must survive the round
        // trip verbatim (no DateParseHandling coercion / tz-normalization on read).
        Assert.Equal("2024-01-02T03:04:05Z", setJson.OptNullStr.Value);
        Assert.Equal(new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc), setJson.OptNullDateTime.Value!.Value.ToUniversalTime());
        Assert.Equal(new DateOnly(2024, 1, 2), setJson.OptNullDate.Value);
        Assert.Equal(7L, setJson.OptNullInt.Value);
        Assert.Equal(1.5, setJson.OptNullNum.Value);
        Assert.Equal(true, setJson.OptNullBool.Value);
        Assert.Equal(new BigInteger(8821239038968084), setJson.OptNullBigInt.Value);
        Assert.Equal(BigInteger.Parse("9223372036854775808"), setJson.OptNullBigIntStr.Value);
        // OptNullDecimal is emitted as a plain optional decimal? (NOT OptionalNullable<>).
        // A scalar bare-number `format: decimal` cannot survive the OptionalNullable wrapper
        // in C#: Json.NET's reader pre-parses a scalar float token as double before the
        // wrapper's converter runs, truncating to ~15 significant digits (sent
        // 3.141592653589793, would read back 3.14159265358979). The generator honors the
        // author's explicit `format: decimal` precision contract over the tri-state feature,
        // unwraps this field to plain decimal? (full precision, no absent/null/set), and
        // warns the author (see the CsharpOptionalNullableDecimal validation rule). Plain decimal?
        // deserializes via ReadAsDecimal and stays exact. Authors wanting BOTH full precision
        // and tri-state should use `type: string, format: decimal` — that decimalStr escape
        // hatch (OptNullDecimalStr below) travels as a string on the wire and stays wrapped.
        Assert.Equal(3.141592653589793m, setJson.OptNullDecimal);
        Assert.Equal(3.14159265358979344719667586m, setJson.OptNullDecimalStr.Value);
        Assert.Equal(new List<string> { "a", "b" }, setJson.OptNullArr.Value);
        Assert.Equal(
            new List<BigInteger?>
            {
                BigInteger.Parse("9223372036854775808"),
                new BigInteger(7),
                null,
            },
            setJson.OptNullBigIntStrArr.Value
        );
        Assert.Equal("v", setJson.OptNullMap.Value!["k"]);
        Assert.Equal(3.14159265358979344719667586m, setJson.OptNullDecimalStrMap.Value!["pi"]);
        Assert.Null(setJson.OptNullDecimalStrMap.Value!["none"]);
        Assert.Equal("u", setJson.OptNullUnion.Value!.Str);
        Assert.Equal("in", setJson.OptNullObj.Value!.InnerReqStr);
        Assert.True(setJson.OptNullObj.Value.InnerOptNullStr.IsSet);
        Assert.Equal("innerval", setJson.OptNullObj.Value.InnerOptNullStr.Value);
        Assert.Equal(OptNullEnum.First, setJson.OptNullEnum.Value);

        // Explicit null: JSON keeps every key with a null value (IsSet && IsNull).
        var nullJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(OptionalNullableFieldsNull())
        ).Res!.Json;
        Assert.Equal("base", nullJson.ReqStr);
        Assert.Null(nullJson.OptStr);
        Helpers.AssertObjectWithOptionalNullableFieldsAllNull(nullJson);

        // Absent: every key is omitted entirely (neither set nor null).
        var absentJson = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableAsync(OptionalNullableFieldsAbsent())
        ).Res!.Json;
        Assert.Equal("base", absentJson.ReqStr);
        Assert.Null(absentJson.OptStr);
        Helpers.AssertObjectWithOptionalNullableFieldsAllAbsent(absentJson);
    }

    [Fact]
    public async Task PostFormOptionalNullableJsonShared()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-optional-nullable-json-shared");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Set: each primitive serializes as its bare value (no OptionalNullable quoting).
        var setForm = (
            await sdk.RequestBodies.RequestBodyPostFormOptionalNullableJsonSharedAsync(OptionalNullableFieldsSet())
        ).Res!.Form;
        Assert.Equal("base", setForm.ReqStr);
        Assert.Equal("opt", setForm.OptStr);
        Assert.Equal("hello", setForm.OptNullStr);
        Assert.Equal("7", setForm.OptNullInt);
        Assert.Equal("1.5", setForm.OptNullNum);
        Assert.Equal("true", setForm.OptNullBool);
        Assert.Equal("8821239038968084", setForm.OptNullBigInt);
        Assert.Equal("9223372036854775808", setForm.OptNullBigIntStr);
        Assert.Equal("3.141592653589793", setForm.OptNullDecimal);
        Assert.Equal("3.14159265358979344719667586", setForm.OptNullDecimalStr);
        Assert.Equal("first", setForm.OptNullEnum);

        // Explicit null and absent both omit every field — primitives, array, map,
        // union and nested object (form has no native null).
        AssertFormAllOptionalOmitted(
            (
                await sdk.RequestBodies.RequestBodyPostFormOptionalNullableJsonSharedAsync(OptionalNullableFieldsNull())
            ).Res!.Form
        );
        AssertFormAllOptionalOmitted(
            (
                await sdk.RequestBodies.RequestBodyPostFormOptionalNullableJsonSharedAsync(OptionalNullableFieldsAbsent())
            ).Res!.Form
        );
    }

    [Fact]
    public async Task PutMultipartOptionalNullableJsonShared()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-optional-nullable-json-shared");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // Set: each primitive serializes as its bare value (no OptionalNullable quoting).
        var setForm = (
            await sdk.RequestBodies.RequestBodyPutMultipartOptionalNullableJsonSharedAsync(OptionalNullableFieldsSet())
        ).Res!.Form;
        Assert.Equal("base", setForm.ReqStr);
        Assert.Equal("opt", setForm.OptStr);
        Assert.Equal("hello", setForm.OptNullStr);
        Assert.Equal("7", setForm.OptNullInt);
        Assert.Equal("1.5", setForm.OptNullNum);
        Assert.Equal("true", setForm.OptNullBool);
        Assert.Equal("8821239038968084", setForm.OptNullBigInt);
        Assert.Equal("9223372036854775808", setForm.OptNullBigIntStr);
        Assert.Equal("3.141592653589793", setForm.OptNullDecimal);
        Assert.Equal("3.14159265358979344719667586", setForm.OptNullDecimalStr);
        Assert.Equal("first", setForm.OptNullEnum);

        // Explicit null and absent both omit every field — primitives, array, map,
        // union and nested object (multipart has no native null).
        AssertMultipartAllOptionalOmitted(
            (
                await sdk.RequestBodies.RequestBodyPutMultipartOptionalNullableJsonSharedAsync(OptionalNullableFieldsNull())
            ).Res!.Form
        );
        AssertMultipartAllOptionalOmitted(
            (
                await sdk.RequestBodies.RequestBodyPutMultipartOptionalNullableJsonSharedAsync(OptionalNullableFieldsAbsent())
            ).Res!.Form
        );
    }

    private static void AssertFormAllOptionalOmitted(
        RequestBodyPostFormOptionalNullableJsonSharedForm form
    )
    {
        Assert.Equal("base", form.ReqStr);
        Assert.Null(form.OptStr);
        Assert.Null(form.OptNullStr);
        Assert.Null(form.OptNullInt);
        Assert.Null(form.OptNullNum);
        Assert.Null(form.OptNullBool);
        Assert.Null(form.OptNullBigInt);
        Assert.Null(form.OptNullBigIntStr);
        Assert.Null(form.OptNullDecimal);
        Assert.Null(form.OptNullDecimalStr);
        Assert.Null(form.OptNullArr);
        Assert.Null(form.OptNullBigIntStrArr);
        Assert.Null(form.OptNullMap);
        Assert.Null(form.OptNullDecimalStrMap);
        Assert.Null(form.OptNullUnion);
        Assert.Null(form.OptNullObj);
        Assert.Null(form.OptNullEnum);
    }

    private static void AssertMultipartAllOptionalOmitted(
        RequestBodyPutMultipartOptionalNullableJsonSharedForm form
    )
    {
        Assert.Equal("base", form.ReqStr);
        Assert.Null(form.OptStr);
        Assert.Null(form.OptNullStr);
        Assert.Null(form.OptNullInt);
        Assert.Null(form.OptNullNum);
        Assert.Null(form.OptNullBool);
        Assert.Null(form.OptNullBigInt);
        Assert.Null(form.OptNullBigIntStr);
        Assert.Null(form.OptNullDecimal);
        Assert.Null(form.OptNullDecimalStr);
        Assert.Null(form.OptNullArr);
        Assert.Null(form.OptNullBigIntStrArr);
        Assert.Null(form.OptNullMap);
        Assert.Null(form.OptNullDecimalStrMap);
        Assert.Null(form.OptNullUnion);
        Assert.Null(form.OptNullObj);
        Assert.Null(form.OptNullEnum);
    }

    // ── Form/multipart-only optional+nullable: NO OptionalNullable wrapper ──────────
    //
    // ObjectWithOptionalNullableField is referenced only by form and multipart request
    // bodies — never a JSON body or response. Form/multipart encoding cannot express
    // null, so there is no absent/null/set tri-state to be preserved; presence-aware
    // serialization must emit a plain nullable field, not OptionalNullable<T>.

    [Fact]
    public async Task PostFormOptionalNullableField()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-optional-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var body = new ObjectWithOptionalNullableField() { ReqStr = "base", OptNullStr = "hello" };
        // The field on this form/multipart-only body carries no wrapper.
        Helpers.AssertNoOptionalNullableWrapper(body, nameof(body.OptNullStr));

        // Set: the plain nullable field serializes as its bare form value.
        var setForm = (
            await sdk.RequestBodies.RequestBodyPostFormOptionalNullableAsync(body)
        ).Res!.Form;
        Assert.Equal("base", setForm.ReqStr);
        Assert.Equal("hello", setForm.OptNullStr);

        // Absent: field omitted (form has no native null).
        var absentForm = (
            await sdk.RequestBodies.RequestBodyPostFormOptionalNullableAsync(
                new ObjectWithOptionalNullableField() { ReqStr = "base" }
            )
        ).Res!.Form;
        Assert.Equal("base", absentForm.ReqStr);
        Assert.Null(absentForm.OptNullStr);
    }

    [Fact]
    public async Task PutMultipartOptionalNullableField()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-optional-nullable");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var body = new ObjectWithOptionalNullableField() { ReqStr = "base", OptNullStr = "hello" };
        Helpers.AssertNoOptionalNullableWrapper(body, nameof(body.OptNullStr));

        var setForm = (
            await sdk.RequestBodies.RequestBodyPutMultipartOptionalNullableAsync(body)
        ).Res!.Form;
        Assert.Equal("base", setForm.ReqStr);
        Assert.Equal("hello", setForm.OptNullStr);

        var absentForm = (
            await sdk.RequestBodies.RequestBodyPutMultipartOptionalNullableAsync(
                new ObjectWithOptionalNullableField() { ReqStr = "base" }
            )
        ).Res!.Form;
        Assert.Equal("base", absentForm.ReqStr);
        Assert.Null(absentForm.OptNullStr);
    }

    [Fact]
    public async Task PostJsonOptionalNullableBodyAndParam()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-optional-nullable-body-and-param");
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        // The same component is used as both a JSON body and a query parameter, so its
        // optional+nullable field is OptionalNullable<T> — the JSON body keeps the
        // absent/null/set tri-state, while URL serialization unwraps to the bare value.
        var set = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableBodyAndParamAsync(
                optionalNullableBodyAndParam: new OptionalNullableBodyAndParam() { OptNullStr = "b" },
                filter: new OptionalNullableBodyAndParam() { OptNullStr = "q" }
            )
        ).Res!;
        // JSON body: field present with a value (IsSet, not IsNull).
        Assert.True(set.Json!.OptNullStr!.IsSet);
        Assert.False(set.Json!.OptNullStr!.IsNull);
        Assert.Equal("b", set.Json!.OptNullStr!.Value);
        // Query param: unwrapped to the bare value on the wire (deepObject key), no
        // OptionalNullable sentinel leaked.
        Assert.Equal("q", set.Args["filter[optNullStr]"]);

        // Explicit null in the body survives as IsSet && IsNull; the null query value
        // is unwrapped to null and omitted from the URL.
        var nullBody = (
            await sdk.RequestBodies.RequestBodyPostJsonOptionalNullableBodyAndParamAsync(
                optionalNullableBodyAndParam: new OptionalNullableBodyAndParam() { OptNullStr = (string?)null },
                filter: new OptionalNullableBodyAndParam() { OptNullStr = (string?)null }
            )
        ).Res!;
        Assert.True(nullBody.Json!.OptNullStr!.IsSet);
        Assert.True(nullBody.Json!.OptNullStr!.IsNull);
        Assert.False(nullBody.Args.ContainsKey("filter[optNullStr]"));
    }

    [Fact]
    public async Task PutMultipartDeep()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-deep");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateDeepObject();
        var res = await sdk.RequestBodies.RequestBodyPutMultipartDeepAsync(obj);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(Utilities.ToString(obj.Arr), res.Res!.Form.Arr);
        Assert.Equal("true", res.Res!.Form.Bool);
        Assert.Equal("1", res.Res!.Form.Int);
        Assert.Equal(Utilities.ToString(obj.Map), res.Res!.Form.Map);
        Assert.Equal("1.1", res.Res!.Form.Num);
        Assert.Equal(Utilities.ToString(obj.Obj), res.Res!.Form.Obj);
        Assert.Equal("test", res.Res!.Form.Str);
    }

    [Fact]
    public async Task PutMultipartFile()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-file");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var data = Helpers.GetData();

        var res = await sdk.RequestBodies.RequestBodyPutMultipartFileAsync(
            new RequestBodyPutMultipartFileRequestBody()
            {
                File = new Openapi.Models.Operations.RequestBodyPutMultipartFileFile()
                {
                    Content = data,
                    FileName = "testUpload.json"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal(Encoding.UTF8.GetString(data, 0, data.Length), res.Res!.Files["file"]);
    }

    [Fact]
    public async Task PutMultipartFileRef()
    {
        CommonHelpers.RecordTest("request-bodies-put-multipart-file-ref");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var data = Helpers.GetData();

        var res = await sdk.RequestBodies.RequestBodyPutMultipartFileRefAsync(
            new RequestBodyPutMultipartFileRefRequestBody()
            {
                File = new Openapi.Models.Shared.BinaryString()
                {
                    Content = data,
                    FileName = "testUpload.json"
                }
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal(Encoding.UTF8.GetString(data, 0, data.Length), res.Res!.Files["file"]);
    }

    [Fact]
    public async Task RequestBodyPutMultipartDifferentFilename()
    {
        CommonHelpers.RecordTest("request-bodies-put-different-file-name");
        // This relative path is necessary, because the CWD while under test is `Test/bin/Debug/net5.0`
        // The C# canonical solution to using files in test is to embed them in the Assembly.
        // I chose not to pursue that route for simplicity
        byte[] data = System.IO.File.ReadAllBytes("../../../testUpload.json");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var differentFileName = new DifferentFileName()
        {
            Content = data,
            FileName = "testUpload.json"
        };
        var res = await sdk.RequestBodies.RequestBodyPutMultipartDifferentFileNameAsync(
            new RequestBodyPutMultipartDifferentFileNameRequestBody()
            {
                DifferentFileName = differentFileName
            }
        );
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal(
            System.Text.Encoding.Default.GetString(data),
            res.Res!.Files["differentFileName"]
        );
    }

    [Fact]
    public async Task PostFormSimple()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-simple");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostFormSimpleAsync(
            Helpers.CreateSimpleObject()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal("any", res.Res!.Form.Any);
        Assert.Equal("true", res.Res!.Form.Bool);
        Assert.Equal("true", res.Res!.Form.BoolOpt);
        Assert.Equal("2020-01-01", res.Res!.Form.Date);
        Assert.Equal("2020-01-01T00:00:00.0000001Z", res.Res!.Form.DateTime);
        Assert.Equal("one", res.Res!.Form.Enum);
        Assert.Equal("1.1", res.Res!.Form.Float32);
        Assert.Equal("1", res.Res!.Form.Int);
        Assert.Equal("1", res.Res!.Form.Int32);
        Assert.Equal("1.1", res.Res!.Form.Num);
        Assert.Equal("test", res.Res!.Form.Str);
        Assert.Equal("testOptional", res.Res!.Form.StrOpt);
    }

    [Fact]
    public async Task PostFormDeep()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-deep");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var obj = Helpers.CreateDeepObject();

        var res = await sdk.RequestBodies.RequestBodyPostFormDeepAsync(obj);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal(Utilities.ToString(obj.Arr), res.Res!.Form.Arr);
        Assert.Equal("true", res.Res!.Form.Bool);
        Assert.Equal("1", res.Res!.Form.Int);
        Assert.Equal(Utilities.ToString(obj.Map), res.Res!.Form.Map);
        Assert.Equal("1.1", res.Res!.Form.Num);
        Assert.Equal(Utilities.ToString(obj.Obj), res.Res!.Form.Obj);
        Assert.Equal("test", res.Res!.Form.Str);
    }

    [Fact]
    public async Task PostFormMapPrimitive()
    {
        CommonHelpers.RecordTest("request-bodies-post-form-map-primitive");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var map = new Dictionary<string, string>()
        {
            { "key1", "value1" },
            { "key2", "value2" },
            { "key3", "value3" }
        };

        var res = await sdk.RequestBodies.RequestBodyPostFormMapPrimitiveAsync(map);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(map, res.Res!.Form);
    }

    [Fact]
    public async Task PutString()
    {
        CommonHelpers.RecordTest("request-bodies-put-string");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var str = "Hello world";

        var res = await sdk.RequestBodies.RequestBodyPutStringAsync(str);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(str, res.Res!.Data);
    }

    [Fact]
    public async Task PutBytes()
    {
        CommonHelpers.RecordTest("request-bodies-put-bytes");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var data = Helpers.GetData();

        var res = await sdk.RequestBodies.RequestBodyPutBytesAsync(data);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(Encoding.UTF8.GetString(data, 0, data.Length), res.Res!.Data);
    }

    [Fact]
    public async Task PutStringWithParams()
    {
        CommonHelpers.RecordTest("request-bodies-put-string-with-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPutStringWithParamsAsync(
            "Hello world",
            "test param"
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("Hello world", res.Res!.Data);
        Assert.Equal("test param", res.Res!.Args.QueryStringParam);
    }

    [Fact]
    public async Task PutBytesWithParams()
    {
        CommonHelpers.RecordTest("request-bodies-put-bytes-with-params");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var data = Helpers.GetData();

        var res = await sdk.RequestBodies.RequestBodyPutBytesWithParamsAsync(data, "test param");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(Encoding.UTF8.GetString(data, 0, data.Length), res.Res!.Data);
        Assert.Equal("test param", res.Res!.Args.QueryStringParam);
    }

    [Fact]
    public async Task EmptyObject()
    {
        CommonHelpers.RecordTest("request-bodies-post-empty-object");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostEmptyObjectAsync(
            new RequestBodyPostEmptyObjectRequestBody()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task CamelCase()
    {
        CommonHelpers.RecordTest("request-bodies-post-application-json-simple-camel-case");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostApplicationJsonSimpleCamelCaseAsync(
            Helpers.CreateSimpleObjectCamelCase()
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Helpers.AssertSimpleObjectCamelCase(res.Res!.Json!);

        var rawResponseString = await res.HttpMeta.Response.Content.ReadAsStringAsync();
        Assert.Equal(28, Regex.Matches(rawResponseString, "_val").Count);
    }

    [Fact]
    public async Task RequestBodyReadOnlyInput()
    {
        CommonHelpers.RecordTest("request-bodies-read-only-input");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyReadOnlyInputAsync(new ReadOnlyObjectInput());

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True(res.ReadOnlyObject!.Bool);
        Assert.Equal(1.0, res.ReadOnlyObject!.Num);
        Assert.Equal("hello", res.ReadOnlyObject!.String);
    }

    [Fact]
    public async Task RequestBodyWriteOnlyOutput()
    {
        CommonHelpers.RecordTest("request-bodies-write-only-output");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyWriteOnlyOutputAsync(
            new WriteOnlyObject()
            {
                Bool = true,
                Num = 1.0F,
                String = "hello"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
    }

    [Fact]
    public async Task RequestBodyWriteOnly()
    {
        CommonHelpers.RecordTest("request-bodies-write-only");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyWriteOnlyAsync(
            new WriteOnlyObject()
            {
                Bool = true,
                Num = 1.0F,
                String = "hello"
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True(res.ReadOnlyObject!.Bool);
        Assert.Equal(1.0, res.ReadOnlyObject!.Num);
        Assert.Equal("hello", res.ReadOnlyObject!.String);
    }

    [Fact]
    public async Task RequestBodyReadAndWrite()
    {
        CommonHelpers.RecordTest("request-bodies-read-and-write");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyReadAndWriteAsync(
            new ReadWriteObject()
            {
                Num1 = 1,
                Num2 = 2,
                Num3 = 4,
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(4, res.ReadWriteObject!.Num3);
        Assert.Equal(7, res.ReadWriteObject!.Sum);
    }

    [Fact]
    public async Task RequestBodyPostComplexNumberTypesAsync()
    {
        CommonHelpers.RecordTest("request-bodies-complex-number-types");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var req = new RequestBodyPostComplexNumberTypesRequest()
        {
            ComplexNumberTypes = new ComplexNumberTypes()
            {
                Bigint = BigInteger.Parse("9007199254740991"),
                BigintStr = BigInteger.Parse("9223372036854775807"),
                Decimal = 3.141592653589793M,
                DecimalStr = 3.141592653589793238462643383279M,
                Float64Str = "1.1",
                Int64Str = "100"
            },
            PathBigInt = BigInteger.Parse("9007199254740991"),
            PathBigIntStr = BigInteger.Parse("9223372036854775807"),
            PathDecimal = 3.141592653589793M,
            PathDecimalStr = 3.141592653589793238462643383279M,
            PathFloat64Str = "1.1",
            PathInt64Str = "100",
            QueryBigInt = BigInteger.Parse("9007199254740991"),
            QueryBigIntStr = BigInteger.Parse("9223372036854775807"),
            QueryDecimal = 3.141592653589793M,
            QueryDecimalStr = 3.141592653589793238462643383279M,
            QueryFloat64Str = "1.1",
            QueryInt64Str = "100"
        };

        var res = await sdk.RequestBodies.RequestBodyPostComplexNumberTypesAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req.ComplexNumberTypes.Bigint, res.Object!.Json.Bigint);
        Assert.Equal(req.ComplexNumberTypes.BigintStr, res.Object!.Json.BigintStr);
        Assert.Equal(req.ComplexNumberTypes.Decimal, res.Object!.Json.Decimal);
        Assert.Equal(req.ComplexNumberTypes.DecimalStr, res.Object!.Json.DecimalStr);
        Assert.Equal(req.ComplexNumberTypes.Int64Str, res.Object!.Json.Int64Str);
        Assert.Equal(req.ComplexNumberTypes.Float64Str, res.Object!.Json.Float64Str);
        Assert.Equal(
            $"{Helpers.HttpBinUrl}/anything/requestBodies/post/9007199254740991/9223372036854775807/3.141592653589793/3.1415926535897932384626433833/100/1.1/complex-number-types?queryBigInt=9007199254740991&queryBigIntStr=9223372036854775807&queryDecimal=3.141592653589793&queryDecimalStr=3.1415926535897932384626433833&queryFloat64Str=1.1&queryInt64Str=100",
            res.Object!.Url
        );
    }

    [Fact]
    public async Task RequestBodyPostComplexNumberTypesOptionalUnpopulatedAsync()
    {
        CommonHelpers.RecordTest("request-bodies-complex-number-types-optional-unpopulated");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new OptionalComplexNumberTypes();
        var res = await sdk.RequestBodies.RequestBodyPostComplexNumberTypesOptionalAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Object);
        Assert.NotNull(res.Object!.Json);
        Assert.Null(res.Object!.Json.Bigint);
        Assert.Null(res.Object!.Json.BigintStr);
        Assert.Null(res.Object!.Json.Decimal);
        Assert.Null(res.Object!.Json.DecimalStr);
        Assert.Null(res.Object!.Json.Float64Str);
        Assert.Null(res.Object!.Json.Int64Str);
    }

    [Fact]
    public void RequestBodyPostComplexNumberTypesRequiredMissingThrows()
    {
        CommonHelpers.RecordTest("request-bodies-complex-number-types-required-missing");

        var incompleteJson = "{}";
        var ex = Assert.Throws<JsonSerializationException>(
            () => JsonConvert.DeserializeObject<ComplexNumberTypes>(
                incompleteJson,
                Utilities.GetDefaultJsonDeserializerSettings()
            )
        );
        Assert.Contains("Required property", ex.Message);
    }

    [Fact]
    public async Task RequestBodyDefaultsAndConstsAsync()
    {
        CommonHelpers.RecordTest("request-bodies-defaults-and-consts");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new DefaultsAndConsts() { NormalField = "normal", DefaultStr = "not default" };
        var res = await sdk.RequestBodies.RequestBodyPostDefaultsAndConstsAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("normal", res.Object!.Json.NormalField);
        Assert.Equal(9007199254740991, res.Object!.Json.ConstBigInt);
        Assert.Equal(9223372036854775807, res.Object!.Json.ConstBigIntStr);
        Assert.True(res.Object!.Json.ConstBool);
        Assert.Equal(DateOnly.FromDateTime(new DateTime(2020, 1, 1)), res.Object!.Json.ConstDate);
        Assert.Equal(new DateTime(2020, 1, 1), res.Object!.Json.ConstDateTime);
        Assert.Equal(3.141592653589793M, res.Object!.Json.ConstDecimal);
        Assert.Equal(3.141592653589793238462643383279M, res.Object!.Json.ConstDecimalStr);
        Assert.Equal(ConstEnumInt.Two, res.Object!.Json.ConstEnumInt);
        Assert.Equal(ConstEnumStr.Two, res.Object!.Json.ConstEnumStr);
        Assert.Equal(123, res.Object!.Json.ConstInt);
        Assert.Equal(123.456, res.Object!.Json.ConstNum);
        Assert.Null(res.Object!.Json.ConstStrNull);

        Assert.Equal(9007199254740991, res.Object!.Json.DefaultBigInt);
        Assert.Equal(9223372036854775807, res.Object!.Json.DefaultBigIntStr);
        Assert.True(res.Object!.Json.DefaultBool);
        Assert.Equal(DateOnly.FromDateTime(new DateTime(2020, 1, 1)), res.Object!.Json.DefaultDate);
        Assert.Equal(new DateTime(2020, 1, 1), res.Object!.Json.DefaultDateTime);
        Assert.Equal(3.141592653589793M, res.Object!.Json.DefaultDecimal);
        Assert.Equal(3.141592653589793238462643383279M, res.Object!.Json.DefaultDecimalStr);
        Assert.Equal(DefaultEnumInt.Two, res.Object!.Json.DefaultEnumInt);
        Assert.Equal(DefaultEnumStr.Two, res.Object!.Json.DefaultEnumStr);
        Assert.Equal(123, res.Object!.Json.DefaultInt);
        Assert.Equal(123.456, res.Object!.Json.DefaultNum);
        Assert.Equal("not default", res.Object!.Json.DefaultStr);
        Assert.True(res.Object!.Json.DefaultStrNullable.IsNull);
        Assert.Equal("default", res.Object!.Json.DefaultStrOptional);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesStringAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-string");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesStringAsync("test");

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("test", res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesIntegerAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-integer");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesIntegerAsync(1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(1, res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesLongAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-int32");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesInt32Async(1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(1, res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesBigIntAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-bigint");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBigIntAsync(
            new BigInteger(1)
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(typeof(BigInteger), res.Object!.Json.GetType());
        Assert.Equal(new BigInteger(1), res.Object!.Json);
        Assert.Equal("1", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesBigIntStrAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-bigint-str");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = BigInteger.Parse("1");
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBigIntStrAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(typeof(BigInteger), res.Object!.Json.GetType());
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("\"1\"", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesNumberAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-number");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesNumberAsync(1.1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(1.1, res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesFloatAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-float32");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesFloat32Async(1.1);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(1.1, res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesDecimalAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-decimal");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDecimalAsync(1.1M);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(typeof(decimal), res.Object!.Json.GetType());
        Assert.Equal(1.1M, res.Object!.Json);
        Assert.Equal("1.1", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesDecimalStrAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-decimal-str");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDecimalStrAsync(1.1M);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(typeof(decimal), res.Object!.Json.GetType());
        Assert.Equal(1.1M, res.Object!.Json);
        Assert.Equal("\"1.1\"", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesBooleanAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-boolean");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesBooleanAsync(true);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.True(res.Object!.Json);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesDate()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-date");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var date = DateOnly.FromDateTime(new DateTime(2020, 1, 1));
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDateAsync(date);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(date, res.Object!.Json);
        Assert.Equal("\"2020-01-01\"", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesDateTimeAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-date-time");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var dateTime = new DateTime(2020, 1, 1, 0, 0, 0, DateTimeKind.Utc);
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesDateTimeAsync(dateTime);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(dateTime, res.Object!.Json);
        Assert.Equal("\"2020-01-01T00:00:00.0000000Z\"", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesMapDateTimeAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-date-time");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new Dictionary<string, DateTime>()
        {
            { "test", new DateTime(2020, 1, 1, 0, 0, 0, DateTimeKind.Utc) }
        };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapDateTimeAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("{\"test\":\"2020-01-01T00:00:00.0000000Z\"}", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesMapBigIntStrAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-bigint-str");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new Dictionary<string, BigInteger>() { { "test", new BigInteger(1) } };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapBigIntStrAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("{\"test\":\"1\"}", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesMapDecimalAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-map-decimal");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new Dictionary<string, Decimal>() { { "test", 3.141592653589793M } };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesMapDecimalAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("{\"test\":3.141592653589793}", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesArrayDateAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-date");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new List<DateOnly> { DateOnly.FromDateTime(new DateTime(2020, 1, 1)) };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayDateAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("[\"2020-01-01\"]", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesArrayBigIntAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-bigint");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new List<BigInteger> { new BigInteger(1) };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayBigIntAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("[1]", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesArrayDecimalStr()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-array-decimal-str");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new List<Decimal> { 3.141592653589793438462643383279M };
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesArrayDecimalStrAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req, res.Object!.Json);
        Assert.Equal("[\"3.1415926535897934384626433833\"]", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesComplexNumberArrays()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-complex-number-arrays");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new ComplexNumberArrays()
        {
            BigIntArray = new List<BigInteger> { BigInteger.Parse("9007199254740991") },
            BigIntStrArray = new List<BigInteger> { BigInteger.Parse("9223372036854775807") },
            DecimalArray = new List<Decimal> { 3.141592653589793M },
            DecimalStrArray = new List<Decimal> { 3.141592653589793238462643383279M }
        };
        var json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal(
            "{\"bigIntArray\":[9007199254740991],\"bigIntStrArray\":[\"9223372036854775807\"],\"decimalArray\":[3.141592653589793],\"decimalStrArray\":[\"3.1415926535897932384626433833\"]}",
            json
        );
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesComplexNumberArraysAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req.BigIntArray, res.Res!.Json!.BigIntArray);
        Assert.Equal(req.BigIntStrArray, res.Res!.Json!.BigIntStrArray);
        Assert.Equal(req.DecimalArray, res.Res!.Json!.DecimalArray);
        Assert.Equal(req.DecimalStrArray, res.Res!.Json!.DecimalStrArray);
    }

    [Fact]
    public async Task RequestBodyPostJsonDataTypesComplexNumberMaps()
    {
        CommonHelpers.RecordTest("request-bodies-post-json-data-types-complex-number-maps");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new ComplexNumberMaps()
        {
            BigintMap = new Dictionary<string, BigInteger>
            {
                { "bigint", BigInteger.Parse("9007199254740991") }
            },
            BigintStrMap = new Dictionary<string, BigInteger>
            {
                { "bigintStr", BigInteger.Parse("9223372036854775807") }
            },
            DecimalMap = new Dictionary<string, Decimal> { { "decimal", 3.141592653589793M } },
            DecimalStrMap = new Dictionary<string, Decimal>
            {
                { "decimalStr", 3.141592653589793238462643383279M }
            }
        };

        var json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal(
            "{\"bigintMap\":{\"bigint\":9007199254740991},\"bigintStrMap\":{\"bigintStr\":\"9223372036854775807\"},\"decimalMap\":{\"decimal\":3.141592653589793},\"decimalStrMap\":{\"decimalStr\":\"3.1415926535897932384626433833\"}}",
            json
        );
        var res = await sdk.RequestBodies.RequestBodyPostJsonDataTypesComplexNumberMapsAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(req.BigintMap, res.Res!.Json!.BigintMap);
        Assert.Equal(req.BigintStrMap, res.Res!.Json!.BigintStrMap);
        Assert.Equal(req.DecimalMap, res.Res!.Json!.DecimalMap);
        Assert.Equal(req.DecimalStrMap, res.Res!.Json!.DecimalStrMap);
    }

    [Fact]
    public async Task RequestBodyPostNullableRequiredStringBodyAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-nullable-required-string-body");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostNullableRequiredStringBodyAsync(null);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("null", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostNullableNotRequiredStringBodyAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-nullable-not-required-string-body");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostNullableNotRequiredStringBodyAsync(null);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("null", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostNotNullableNotRequiredStringBodyAsync()
    {
        CommonHelpers.RecordTest("request-bodies-post-not-nullable-not-required-string-body");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.RequestBodies.RequestBodyPostNotNullableNotRequiredStringBodyAsync(
            null
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("", res.Object!.Data);
    }

    [Fact]
    public async Task RequestBodyPostNullableRequiredPropertyAsync()
    {
        var tests = new CommonHelpers.TestTableEntry[]
        {
            new CommonHelpers.TestTableEntry()
            {
                name = "Empty initializer",
                arg = new NullableRequiredPropertyPostRequestBody { },
                want = "{\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                testId = "request-bodies-post-nullable-optional-fields"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "All fields set to null",
                arg = new NullableRequiredPropertyPostRequestBody
                {
                    NullableOptionalInt = null,
                    NullableRequiredArray = null,
                    NullableRequiredEnum = null,
                    NullableRequiredInt = null,
                    NullableRequiredDateTime = null,
                    NullableRequiredBigIntStr = null,
                    NullableRequiredDecimalStr = null
                },
                want =
                    "{\"NullableOptionalInt\":null,\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}",
                testId = "request-bodies-post-nullable-required-property-all-null"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Optional field initialized",
                arg = new NullableRequiredPropertyPostRequestBody { NullableOptionalInt = 0 },
                want = "{\"NullableOptionalInt\":0,\"NullableRequiredArray\":null,\"NullableRequiredBigIntStr\":null,\"NullableRequiredDateTime\":null,\"NullableRequiredDecimalStr\":null,\"NullableRequiredEnum\":null,\"NullableRequiredInt\":null}"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "All fields set to non-null value",
                arg = new NullableRequiredPropertyPostRequestBody
                {
                    NullableOptionalInt = 0,
                    NullableRequiredArray = new List<double> { 1.1, 2.2, 3.3 },
                    NullableRequiredEnum = NullableRequiredEnum.Second,
                    NullableRequiredInt = 1,
                    NullableRequiredDateTime = System.DateTime.Parse("2020-01-01T00:00:00Z"),
                    NullableRequiredBigIntStr = BigInteger.Parse("9223372036854775807"),
                    NullableRequiredDecimalStr = 3.141592653589793238462643383279M
                },
                want = "{\"NullableOptionalInt\":0,\"NullableRequiredArray\":[1.1,2.2,3.3],\"NullableRequiredBigIntStr\":\"9223372036854775807\",\"NullableRequiredDateTime\":\"2020-01-01T00:00:00.0000000Z\",\"NullableRequiredDecimalStr\":\"3.1415926535897932384626433833\",\"NullableRequiredEnum\":\"second\",\"NullableRequiredInt\":1}",
                testId = "request-bodies-post-nullable-required-property-all-set"
            }
        };

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        foreach (var test in tests)
        {
            if (!string.IsNullOrEmpty(test.testId))
            {
                CommonHelpers.RecordTest(test.testId);
            }

            var req = (NullableRequiredPropertyPostRequestBody)test.arg;

            var json = await Helpers.GetSerializedBodyJson(req);
            Assert.Equal(test.want, json);

            var res = await sdk.RequestBodies.NullableRequiredPropertyPostAsync(req);
            Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        }
    }

    [Fact]
    public async Task RequestBodyPostNullableRequiredSharedObject()
    {
        var tests = new CommonHelpers.TestTableEntry[]
        {
            new CommonHelpers.TestTableEntry()
            {
                name = "Required field set to null",
                arg = new NullableRequiredSharedObjectPostRequestBody { },
                want = "{\"NullableRequiredObj\":null}",
                testId = "request-bodies-post-nullable-required-shared-object-required-null"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Both fields set to null",
                arg = new NullableRequiredSharedObjectPostRequestBody
                {
                    NullableOptionalObj = null,
                    NullableRequiredObj = null
                },
                want = "{\"NullableOptionalObj\":null,\"NullableRequiredObj\":null}",
                testId = "request-bodies-post-nullable-required-shared-object-all-null"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Optional field set to non-null value",
                arg = new NullableRequiredSharedObjectPostRequestBody
                {
                    NullableOptionalObj = new NullableOptionalObject { Required = 1 }
                },
                want = "{\"NullableOptionalObj\":{\"required\":1},\"NullableRequiredObj\":null}",
                testId = "request-bodies-post-nullable-required-shared-object-optional-non-null"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Both fields set to non-null value",
                arg = new NullableRequiredSharedObjectPostRequestBody
                {
                    NullableOptionalObj = new NullableOptionalObject
                    {
                        Required = 1,
                        Optional = "test"
                    },
                    NullableRequiredObj = new NullableObject { Required = 2 }
                },
                want = "{\"NullableOptionalObj\":{\"optional\":\"test\",\"required\":1},\"NullableRequiredObj\":{\"required\":2}}",
                testId = "request-bodies-post-nullable-required-shared-object-all-set"
            }
        };

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        foreach (var test in tests)
        {
            if (!string.IsNullOrEmpty(test.testId))
            {
                CommonHelpers.RecordTest(test.testId);
            }

            var req = (NullableRequiredSharedObjectPostRequestBody)test.arg;

            var json = await Helpers.GetSerializedBodyJson(req);
            Assert.Equal(test.want, json);

            var res = await sdk.RequestBodies.NullableRequiredSharedObjectPostAsync(req);
            Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        }
    }

    [Fact]
    public async Task RequestBodyPostNullableRequiredEmptyObjectAsync()
    {
        var tests = new CommonHelpers.TestTableEntry[]
        {
            new CommonHelpers.TestTableEntry()
            {
                name = "Empty initializer",
                arg = new NullableRequiredEmptyObjectPostRequestBody { },
                want = "{\"NullableRequiredObj\":null,\"RequiredObj\":{}}"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Required field initialized only",
                arg = new NullableRequiredEmptyObjectPostRequestBody
                {
                    RequiredObj = new RequiredObj { }
                },
                want = "{\"NullableRequiredObj\":null,\"RequiredObj\":{}}",
                testId = "request-bodies-post-nullable-required-empty-object-nullable-set"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "Optional field initialized only",
                arg = new NullableRequiredEmptyObjectPostRequestBody
                {
                    NullableOptionalObj = new NullableOptionalObj { }
                },
                want = "{\"NullableOptionalObj\":{},\"NullableRequiredObj\":null,\"RequiredObj\":{}}",
                testId = "request-bodies-post-nullable-required-empty-object-optional-set"
            },
            new CommonHelpers.TestTableEntry()
            {
                name = "All fields initialized",
                arg = new NullableRequiredEmptyObjectPostRequestBody
                {
                    RequiredObj = new RequiredObj { },
                    NullableOptionalObj = new NullableOptionalObj { },
                    NullableRequiredObj = new NullableRequiredObj { }
                },
                want = "{\"NullableOptionalObj\":{},\"NullableRequiredObj\":{},\"RequiredObj\":{}}",
                testId = "request-bodies-post-nullable-required-empty-object-all-set"
            }
        };

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        foreach (var test in tests)
        {
            if (!string.IsNullOrEmpty(test.testId))
            {
                CommonHelpers.RecordTest(test.testId);
            }

            var req = (NullableRequiredEmptyObjectPostRequestBody)test.arg;

            var json = await Helpers.GetSerializedBodyJson(req);
            Assert.Equal(test.want, json);

            var res = await sdk.RequestBodies.NullableRequiredEmptyObjectPostAsync(req);
            Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        }
    }

    [Fact]
    public async Task RequestBodyNoBodyNoContentType()
    {
        CommonHelpers.RecordTest("request-bodies-no-body-no-content-type");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var res = await sdk.Methods.MethodGetAsync(Helpers.ApiTestServiceUrl);

        Assert.Null(res.HttpMeta.Request.Content);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal("OK", res.Object!.Status);
    }

    [Fact]
    public async Task PostBase64InputModeIdempotent()
    {
        CommonHelpers.RecordTest("request-bodies-base64-file-input-idempotent");

        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);

        var preEncoded = Convert.ToBase64String(
            new byte[] { 0xff, 0xfe, 0x00, 0x01, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0xc3, 0x28 }
        );

        var request = new Base64InputFileModeRequest()
        {
            DataByte = preEncoded,
            DataContentEncoding = preEncoded,
            DataPlain = "plain-idempotent"
        };

        var first = await sdk.RequestBodies.PostBase64InputModeAsync(request);

        Assert.Equal(HttpStatusCode.OK, first.HttpMeta.Response.StatusCode);
        Assert.NotNull(first.Res);
        Assert.Equal(preEncoded, first.Res.Json.DataByte);
        Assert.Equal(preEncoded, first.Res.Json.DataContentEncoding);
        Assert.Equal("plain-idempotent", first.Res.Json.DataPlain);

        var second = await sdk.RequestBodies.PostBase64InputModeAsync(request);

        Assert.Equal(HttpStatusCode.OK, second.HttpMeta.Response.StatusCode);
        Assert.NotNull(second.Res);
        Assert.Equal(first.Res.Json.DataByte, second.Res.Json.DataByte);
        Assert.Equal(first.Res.Json.DataContentEncoding, second.Res.Json.DataContentEncoding);
        Assert.Equal(first.Res.Json.DataPlain, second.Res.Json.DataPlain);
    }

    [Fact]
    public async Task RequestBodyPutMultipartFilesArray()
    {
        CommonHelpers.RecordTest("request-bodies-post-multipart-files-array");

        var sdk = new SDK(serverUrl: Helpers.ApiTestServiceUrl);

        var data = System.IO.File.ReadAllBytes("../../../testUpload.json");

        var res = await sdk.RequestBodies.RequestBodyPutMultipartFilesArrayAsync(
            new RequestBodyPutMultipartFilesArrayRequestBody()
            {
                Files = new List<RequestBodyPutMultipartFilesArrayFiles>()
                {
                    new RequestBodyPutMultipartFilesArrayFiles() { Content = data, FileName = "testUpload.json" },
                    new RequestBodyPutMultipartFilesArrayFiles() { Content = data, FileName = "some-other-name.json" },
                },
            }
        );

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.NotNull(res.Res!);
        Assert.Equal(2, res.Res!.Files.Count);

        Assert.Equal(
            "ewogICAgInNvbWUiOiAianNvbiIsCiAgICAidG8iOiAiYmUiLAogICAgInVwbG9hZGVkIjogImluIiwKICAgICJhIjogImZpbGUiCiAgfQogIA==",
            res.Res!.Files[0].Content
        );
        Assert.Equal("application/json", res.Res!.Files[0].ContentType);
        Assert.Equal("files", res.Res!.Files[0].FieldName);
        Assert.Equal("testUpload.json", res.Res!.Files[0].Filename);
        Assert.Equal(82, res.Res!.Files[0].Size);

        Assert.Equal(
            "ewogICAgInNvbWUiOiAianNvbiIsCiAgICAidG8iOiAiYmUiLAogICAgInVwbG9hZGVkIjogImluIiwKICAgICJhIjogImZpbGUiCiAgfQogIA==",
            res.Res!.Files[1].Content
        );
        Assert.Equal("application/json", res.Res!.Files[1].ContentType);
        Assert.Equal("files", res.Res!.Files[1].FieldName);
        Assert.Equal("some-other-name.json", res.Res!.Files[1].Filename);
        Assert.Equal(82, res.Res!.Files[1].Size);

        Assert.NotNull(res.Res!.FormFields);
        Assert.Empty(res.Res!.FormFields);
    }
}
