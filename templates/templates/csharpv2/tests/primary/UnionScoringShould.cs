#nullable enable
using Openapi.Models;
using Openapi.Models.Errors;
using Openapi.Models.Shared;
using Openapi.Utils.Json;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.Json.Serialization.Metadata;
using Xunit;

public class UnionScoringShould
{
    [Theory]
    [InlineData("{\"count\":1,\"name\":\"item\"}")]
    [InlineData("{\"count\":1,\"name\":\"item\",\"extra\":true}")]
    [InlineData("{\"count\":1,\"name\":\"item\",\"note\":\"note\",\"extra\":true}")]
    public void PreferDeclaredObjectOverFreeFormMap(string json)
    {
        var union = JsonSerialization.Deserialize<ObjectOrMap>(json)!;

        Assert.NotNull(union.CountedItem);
        Assert.Null(union.MapOfAny);
    }

    [Theory]
    [InlineData("{\"extra\":true}", "extra")]
    [InlineData("{\"NAME\":\"item\"}", "NAME")]
    public void UseFreeFormMapWithoutDeclaredMembers(string json, string key)
    {
        var union = JsonSerialization.Deserialize<ObjectOrMap>(json)!;

        Assert.Null(union.CountedItem);
        Assert.NotNull(union.MapOfAny);
        Assert.True(union.MapOfAny.ContainsKey(key));
        Assert.Equal(json, JsonSerialization.Serialize(union));
    }

    [Fact]
    public void PreferDeclaredObjectWithCaseInsensitiveNames()
    {
        var union = DeserializeCaseInsensitive<ObjectOrMap>("{\"NAME\":\"item\"}");

        Assert.NotNull(union.CountedItem);
        Assert.Equal("item", union.CountedItem.Name);
        Assert.Null(union.MapOfAny);
    }

    [Theory]
    [InlineData(
        "{\"value\":{\"count\":1,\"name\":\"old\",\"note\":\"obsolete\"},\"value\":{\"name\":\"item\"},\"alive\":true}",
        "{\"value\":{\"name\":\"item\"},\"alive\":true}",
        false
    )]
    [InlineData(
        "{\"value\":{\"count\":1,\"name\":\"old\",\"note\":\"obsolete\"},\"VALUE\":{\"name\":\"item\"},\"alive\":true}",
        "{\"value\":{\"name\":\"item\"},\"alive\":true}",
        true
    )]
    [InlineData(
        "{\"entries\":{\"key\":{\"count\":1,\"name\":\"old\",\"note\":\"obsolete\"},\"key\":{\"name\":\"item\"}},\"alive\":true}",
        "{\"entries\":{\"key\":{\"name\":\"item\"}},\"alive\":true}",
        false
    )]
    public void ScoreLastOccurrenceOfRepeatedKey(string json, string expected, bool caseInsensitive)
    {
        var union = caseInsensitive
            ? DeserializeCaseInsensitive<ItemContainerChoice>(json)
            : JsonSerialization.Deserialize<ItemContainerChoice>(json)!;

        Assert.NotNull(union.NamedItemContainer);
        Assert.Null(union.CountedItemContainer);
        Assert.True(JsonNode.DeepEquals(JsonNode.Parse(expected), JsonNode.Parse(JsonSerialization.Serialize(union))));
    }

    [Fact]
    public void KeepDistinctMapKeysWithCaseInsensitivePropertyNames()
    {
        var union = DeserializeCaseInsensitive<ItemContainerChoice>(
            "{\"entries\":{\"key\":{\"count\":1,\"name\":\"first\",\"note\":\"note\"},\"KEY\":{\"name\":\"second\"}},\"alive\":true}"
        );

        Assert.NotNull(union.CountedItemContainer);
        Assert.NotNull(union.CountedItemContainer.Entries);
        Assert.Equal("first", union.CountedItemContainer.Entries["key"].Name);
        Assert.Equal("second", union.CountedItemContainer.Entries["KEY"].Name);
    }

    [Fact]
    public void IgnoreNonSerializableMetadataInScore()
    {
        using var document = JsonDocument.Parse("{\"HttpMeta\":true}");
        var options = JsonOptions.SerializerOptions;
        var typeInfo = options.GetTypeInfo(typeof(ErrorType1Payload));
        var score = UnionScoring.Score(document.RootElement, typeInfo, options);

        Assert.Equal(0, score.Matched);
        Assert.Equal(1, score.Overflow);
        Assert.Equal(1, score.Missing);
    }

    [Fact]
    public void CountMatchesInNestedUnionVariant()
    {
        var union = JsonSerialization.Deserialize<NestedObjectChoice>(
            "{\"value\":{\"count\":1,\"name\":\"item\",\"extra\":true}}"
        )!;

        Assert.NotNull(union.ObjectOrMapWrapper);
        Assert.Null(union.NamedItemWrapper);
        Assert.NotNull(union.ObjectOrMapWrapper.Value.CountedItem);
    }

    private static T DeserializeCaseInsensitive<T>(string json)
    {
        var options = new JsonSerializerOptions(JsonOptions.SerializerOptions)
        {
            PropertyNameCaseInsensitive = true,
        };
        var typeInfo = (JsonTypeInfo<T>)options.GetTypeInfo(typeof(T));
        return JsonSerializer.Deserialize(json, typeInfo)!;
    }
}
