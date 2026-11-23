#nullable enable
using Openapi;
using Openapi.Models;
using Openapi.Models.Shared;
using Openapi.Utils.Json;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Numerics;
using System.Text.Json;
using System.Text.Json.Serialization.Metadata;
using System.Threading.Tasks;
using Xunit;

public class AdditionalPropertiesShould
{
    [Fact]
    public void SkipReadOnlyDeclaredJsonNamesOnRead()
    {
        var obj = JsonSerialization.Deserialize<ConstKindWithIntegerExtras>(
            "{\"kind\":\"widget\",\"Kind\":\"widget\",\"count\":5}"
        )!;

        Assert.Equal(new Dictionary<string, long>() { { "count", 5 } }, obj.AdditionalProperties);
        Assert.Equal("{\"kind\":\"widget\",\"count\":5}", JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void KeepDecimalDigits()
    {
        var obj = JsonSerialization.Deserialize<DecimalExtras>("{\"x\":1.2345678901234567}")!;

        Assert.Equal(new Dictionary<string, decimal>() { { "x", 1.2345678901234567m } }, obj.AdditionalProperties);
        Assert.Equal("{\"x\":1.2345678901234567}", JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void KeepMetadataPropertyNamesInTypedExtras()
    {
        const string json = "{\"schema\":{\"$ref\":\"#/$defs/Widget\",\"type\":\"object\"}}";
        var obj = JsonSerialization.Deserialize<StringMapExtras>(json)!;

        Assert.Equal(
            new Dictionary<string, Dictionary<string, string>>()
            {
                {
                    "schema",
                    new Dictionary<string, string>() { { "$ref", "#/$defs/Widget" }, { "type", "object" } }
                },
            },
            obj.AdditionalProperties
        );
        Assert.Equal(json, JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void KeepDateStringsInTypedExtras()
    {
        var obj = JsonSerialization.Deserialize<AnyMapExtras>(
            "{\"event\":{\"when\":\"2020-01-01T00:00:00.000+02:00\"}}"
        )!;

        Assert.Equal("2020-01-01T00:00:00.000+02:00", obj.AdditionalProperties!["event"]["when"]);
    }

    [Fact]
    public void SelectUnionVariantWithAdditionalProperties()
    {
        var union = JsonSerialization.Deserialize<LabelShapeOrFreeform>("{\"anything\":1}")!;

        Assert.NotNull(union.SiblingFreeform);
        Assert.Equal(1, union.SiblingFreeform!.AdditionalProperties!["anything"].GetInt32());
        Assert.Equal("{\"anything\":1}", JsonSerialization.Serialize(union));
    }

    [Fact]
    public async Task SendAndReadMapsOfFreeFormMaps()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new AdditionalPropertiesShapes()
        {
            AnyMaps = new AnyMapExtras()
            {
                AdditionalProperties = new Dictionary<string, Dictionary<string, object>>()
                {
                    { "alpha", new Dictionary<string, object>() { { "text", "x" }, { "count", 10L } } },
                    { "beta", new Dictionary<string, object>() },
                },
            },
        };

        var json = await Helpers.GetSerializedBodyJson(req);
        Assert.Equal("{\"anyMaps\":{\"alpha\":{\"text\":\"x\",\"count\":10},\"beta\":{}}}", json);

        var res = await sdk.RequestBodies.AdditionalPropertiesShapesPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        var extras = res.Object!.Json!.AnyMaps!.AdditionalProperties!;
        Assert.Equal(new[] { "alpha", "beta" }, extras.Keys.OrderBy(k => k));
        Assert.Equal("x", extras["alpha"]["text"]);
        Assert.Equal(10L, Convert.ToInt64(extras["alpha"]["count"]));
        Assert.Empty(extras["beta"]);
    }

    [Fact]
    public async Task SendAndReadEveryShape()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = new AdditionalPropertiesShapes()
        {
            ConstKind = new ConstKindWithIntegerExtras()
            {
                AdditionalProperties = new Dictionary<string, long>() { { "count", 5 } },
            },
            Decimals = new DecimalExtras()
            {
                AdditionalProperties = new Dictionary<string, decimal>() { { "x", 1.2345678901234567m } },
            },
            BigIntLists = new BigIntStringListExtras()
            {
                AdditionalProperties = new Dictionary<string, List<BigInteger>>()
                {
                    { "ids", new List<BigInteger>() { BigInteger.Parse("123456789012345678901234567890") } },
                },
            },
            StringMaps = new StringMapExtras()
            {
                AdditionalProperties = new Dictionary<string, Dictionary<string, string>>()
                {
                    { "labels", new Dictionary<string, string>() { { "env", "prod" } } },
                },
            },
            Untyped = new UntypedExtras()
            {
                NormalField = "declared",
                AdditionalProperties = new Dictionary<string, JsonElement>()
                {
                    { "big", Element("123456789012345678901234567890") },
                    { "nothing", Element("null") },
                },
            },
            Node = new ExtrasNode()
            {
                Label = "root",
                AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "leaf", new ExtrasNode() { Label = "leaf" } } },
            },
            ListNode = new ExtrasListNode()
            {
                AdditionalProperties = new Dictionary<string, List<ExtrasListNode>>()
                {
                    { "items", new List<ExtrasListNode>() { new ExtrasListNode() { Label = "item" } } },
                },
            },
            UnionNode = new ExtrasUnionNode()
            {
                AdditionalProperties = new Dictionary<string, ExtrasUnionValue>()
                {
                    { "text", ExtrasUnionValue.CreateStr("value") },
                    { "node", ExtrasUnionValue.CreateExtrasUnionNode(new ExtrasUnionNode() { Label = "inner" }) },
                },
            },
        };
        const string expected =
            "{\"bigIntLists\":{\"ids\":[\"123456789012345678901234567890\"]},"
            + "\"constKind\":{\"kind\":\"widget\",\"count\":5},"
            + "\"decimals\":{\"x\":1.2345678901234567},"
            + "\"listNode\":{\"items\":[{\"label\":\"item\"}]},"
            + "\"node\":{\"label\":\"root\",\"leaf\":{\"label\":\"leaf\"}},"
            + "\"stringMaps\":{\"labels\":{\"env\":\"prod\"}},"
            + "\"unionNode\":{\"text\":\"value\",\"node\":{\"label\":\"inner\"}},"
            + "\"untyped\":{\"normalField\":\"declared\",\"big\":123456789012345678901234567890,\"nothing\":null}}";

        Assert.Equal(expected, await Helpers.GetSerializedBodyJson(req));

        var res = await sdk.RequestBodies.AdditionalPropertiesShapesPostAsync(req);
        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            CommonHelpers.SortJSONObjectKeys(expected),
            CommonHelpers.SortJSONObjectKeys(JsonSerialization.Serialize(res.Object!.Json))
        );
        var echoed = res.Object.Json!;
        Assert.Equal(1.2345678901234567m, echoed.Decimals!.AdditionalProperties!["x"]);
        Assert.Equal(BigInteger.Parse("123456789012345678901234567890"), echoed.BigIntLists!.AdditionalProperties!["ids"][0]);
        Assert.Equal("leaf", echoed.Node!.AdditionalProperties!["leaf"].Label);
        Assert.Equal("inner", echoed.UnionNode!.AdditionalProperties!["node"].ExtrasUnionNode!.Label);
        Assert.Equal(JsonValueKind.Null, echoed.Untyped!.AdditionalProperties!["nothing"].ValueKind);
    }

    [Fact]
    public void KeepUntypedValuesLossless()
    {
        const string json =
            "{\"normalField\":\"a\",\"big\":123456789012345678901234567890,\"d\":1.10,\"n\":null,\"o\":{\"x\":[1,\"y\"]}}";
        var obj = JsonSerialization.Deserialize<UntypedExtras>(json)!;

        Assert.Equal(new[] { "big", "d", "n", "o" }, obj.AdditionalProperties!.Keys.OrderBy(k => k));
        Assert.Equal(JsonValueKind.Null, obj.AdditionalProperties["n"].ValueKind);
        Assert.Equal(json, JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void DropUntypedCaseVariantOfDeclaredNameOnRead()
    {
        var obj = JsonSerialization.Deserialize<UntypedExtras>("{\"normalField\":\"a\",\"NormalField\":\"b\",\"x\":1}")!;

        Assert.Equal("a", obj.NormalField);
        Assert.Equal(new[] { "x" }, obj.AdditionalProperties!.Keys);
    }

    [Fact]
    public void SkipUntypedDeclaredNamesOnWrite()
    {
        var obj = new UntypedExtras()
        {
            NormalField = "a",
            AdditionalProperties = new Dictionary<string, JsonElement>()
            {
                { "normalField", Element("\"b\"") },
                { "NORMALFIELD", Element("\"c\"") },
                { "x", Element("1") },
            },
        };

        Assert.Equal("{\"normalField\":\"a\",\"x\":1}", JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void BindCaseVariantToDeclaredNameUnderCaseInsensitiveOptions()
    {
        var options = new JsonSerializerOptions(JsonOptions.SerializerOptions) { PropertyNameCaseInsensitive = true };
        var typeInfo = (JsonTypeInfo<UntypedExtras>)options.GetTypeInfo(typeof(UntypedExtras));

        var obj = JsonSerializer.Deserialize("{\"NORMALFIELD\":\"a\",\"x\":1}", typeInfo)!;

        Assert.Equal("a", obj.NormalField);
        Assert.Equal(new[] { "x" }, obj.AdditionalProperties!.Keys);
    }

    [Fact]
    public void ReadKeyNamedLikeTheDictionaryMemberUnderNamingPolicyAsExtra()
    {
        var options = new JsonSerializerOptions(JsonOptions.SerializerOptions)
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
        };
        var typeInfo = (JsonTypeInfo<ConstKindWithIntegerExtras>)options.GetTypeInfo(typeof(ConstKindWithIntegerExtras));

        var obj = JsonSerializer.Deserialize("{\"additionalProperties\":1}", typeInfo)!;

        Assert.Equal(new Dictionary<string, long>() { { "additionalProperties", 1 } }, obj.AdditionalProperties);
        Assert.Equal("{\"kind\":\"widget\",\"additionalProperties\":1}", JsonSerializer.Serialize(obj, typeInfo));
    }

    [Fact]
    public void KeepDatesInTypedExtras()
    {
        const string json = "{\"normalField\":\"n\",\"day\":\"2020-01-01\"}";
        var obj = JsonSerialization.Deserialize<ObjWithDateAdditionalProperties>(json)!;

        Assert.Equal(new DateOnly(2020, 1, 1), obj.AdditionalProperties!["day"]);
        Assert.Equal(json, JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void ReadKeyNamedLikeTheDictionaryMemberAsExtra()
    {
        var obj = JsonSerialization.Deserialize<ConstKindWithIntegerExtras>("{\"AdditionalProperties\":1}")!;

        Assert.Equal(new Dictionary<string, long>() { { "AdditionalProperties", 1 } }, obj.AdditionalProperties);
    }

    [Theory]
    [InlineData("{\"x\":\"text\"}")]
    [InlineData("{\"x\":null}")]
    [InlineData("{\"x\":1.5}")]
    public void RejectWrongTypedExtraOnStrictRead(string json)
    {
        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Deserialize<ConstKindWithIntegerExtras>(json));
    }

    [Fact]
    public void RejectWrongTypedNestedExtraOnStrictRead()
    {
        Assert.ThrowsAny<JsonException>(
            () => JsonSerialization.Deserialize<StringMapExtras>("{\"labels\":{\"env\":1}}")
        );
    }

    [Fact]
    public void ReplaceDictionaryAfterRead()
    {
        var obj = JsonSerialization.Deserialize<ConstKindWithIntegerExtras>("{\"count\":5}")!;
        obj.AdditionalProperties = new Dictionary<string, long>() { { "other", 7 } };

        Assert.Equal("{\"kind\":\"widget\",\"other\":7}", JsonSerialization.Serialize(obj));
    }

    [Fact]
    public void RejectSelfReferenceOnWrite()
    {
        var node = new ExtrasNode() { Label = "self" };
        node.AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "self", node } };

        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(node));
    }

    [Fact]
    public void RejectMutualReferenceOnWrite()
    {
        var first = new ExtrasNode() { Label = "first" };
        var second = new ExtrasNode() { Label = "second" };
        first.AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "next", second } };
        second.AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "next", first } };

        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(first));
    }

    [Fact]
    public void RejectDeclaredChildReferencingAncestorThroughExtraOnWrite()
    {
        var root = new ExtrasNode() { Label = "root" };
        var child = new ExtrasNode() { Label = "child" };
        root.Child = child;
        child.AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "up", root } };

        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(root));
    }

    [Fact]
    public void RejectListReferenceOnWrite()
    {
        var node = new ExtrasListNode() { Label = "self" };
        node.AdditionalProperties = new Dictionary<string, List<ExtrasListNode>>() { { "items", new List<ExtrasListNode>() { node } } };

        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(node));
    }

    [Fact]
    public void RejectUnionReferenceOnWrite()
    {
        var node = new ExtrasUnionNode() { Label = "self" };
        node.AdditionalProperties = new Dictionary<string, ExtrasUnionValue>()
        {
            { "self", ExtrasUnionValue.CreateExtrasUnionNode(node) },
        };

        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(node));
    }

    [Fact]
    public void BoundDeepChainsByConfiguredMaxDepth()
    {
        var options = new JsonSerializerOptions(JsonOptions.SerializerOptions) { MaxDepth = 8 };
        var typeInfo = (JsonTypeInfo<ExtrasNode>)options.GetTypeInfo(typeof(ExtrasNode));

        JsonSerializer.Serialize(Chain(6), typeInfo);
        Assert.ThrowsAny<JsonException>(() => JsonSerializer.Serialize(Chain(20), typeInfo));
    }

    [Fact]
    public void BoundDeepChainsByDefaultMaxDepth()
    {
        JsonSerialization.Serialize(Chain(32));
        Assert.ThrowsAny<JsonException>(() => JsonSerialization.Serialize(Chain(100)));
    }

    [Fact]
    public void WriteUnchangedGraphConcurrently()
    {
        var root = Chain(20);
        var expected = JsonSerialization.Serialize(root);

        var results = new string[64];
        Parallel.For(0, results.Length, i => results[i] = JsonSerialization.Serialize(root));

        Assert.All(results, result => Assert.Equal(expected, result));
    }

    [Fact]
    public void AllocateLinearlyWithChainLength()
    {
        var options = new JsonSerializerOptions(JsonOptions.SerializerOptions) { MaxDepth = 1000 };
        var typeInfo = (JsonTypeInfo<ExtrasNode>)options.GetTypeInfo(typeof(ExtrasNode));
        var shortChain = Chain(100);
        var longChain = Chain(400);
        JsonSerializer.Serialize(shortChain, typeInfo);
        JsonSerializer.Serialize(longChain, typeInfo);

        var shortBytes = AllocatedBytes(() => JsonSerializer.Serialize(shortChain, typeInfo));
        var longBytes = AllocatedBytes(() => JsonSerializer.Serialize(longChain, typeInfo));

        Assert.InRange(longBytes, shortBytes, shortBytes * 8);
    }

    [Fact]
    public void ScoreTypedExtrasAsOverflowOnly()
    {
        var options = JsonOptions.SerializerOptions;
        var typeInfo = options.GetTypeInfo(typeof(ExtrasNode));
        using var valid = JsonDocument.Parse("{\"label\":\"a\",\"x\":{\"label\":\"b\"}}");
        using var invalid = JsonDocument.Parse("{\"label\":\"a\",\"x\":{\"other\":1,\"more\":2}}");

        var validScore = UnionScoring.Score(valid.RootElement, typeInfo, options);
        var invalidScore = UnionScoring.Score(invalid.RootElement, typeInfo, options);

        Assert.Equal(1, validScore.Overflow);
        Assert.Equal(validScore.Matched, invalidScore.Matched);
        Assert.Equal(validScore.Overflow, invalidScore.Overflow);
        Assert.Equal(validScore.Missing, invalidScore.Missing);
    }

    private static ExtrasNode Chain(int length)
    {
        var head = new ExtrasNode() { Label = "0" };
        var current = head;
        for (var i = 1; i < length; i++)
        {
            var next = new ExtrasNode() { Label = i.ToString() };
            current.AdditionalProperties = new Dictionary<string, ExtrasNode>() { { "next", next } };
            current = next;
        }
        return head;
    }

    private static long AllocatedBytes(Action action)
    {
        var before = GC.GetAllocatedBytesForCurrentThread();
        action();
        return GC.GetAllocatedBytesForCurrentThread() - before;
    }

    private static JsonElement Element(string json)
    {
        using var document = JsonDocument.Parse(json);
        return document.RootElement.Clone();
    }
}
