#nullable enable
using Openapi.Models.Shared;
using Openapi.Utils.Json;
using System;
using System.Numerics;
using System.Text.Json;
using System.Text.Json.Nodes;
using Xunit;

public class JsonSerializationShould
{
    private const string SimpleObjectJson =
        "{\"any\":\"any\",\"bigint\":123,\"bigintStr\":\"123456789012345678901234567890\",\"bool\":true,"
        + "\"date\":\"2020-01-01\",\"dateTime\":\"2020-01-01T00:00:00.0000000Z\",\"decimal\":3.141592653589793,"
        + "\"decimalStr\":\"3.14159265358979344719667586\",\"enum\":\"one\",\"float32\":1.1,\"int\":1,\"int32\":1,"
        + "\"int32Enum\":55,\"intEnum\":2,\"num\":1.1,\"str\":\"test\"}";

    private static SimpleObject CreateSimpleObject() =>
        new SimpleObject
        {
            Any = "any",
            Bool = true,
            Date = new DateOnly(2020, 1, 1),
            DateTime = new DateTime(2020, 1, 1, 0, 0, 0, DateTimeKind.Utc),
            Enum = Openapi.Models.Shared.Enum.One,
            Float32 = 1.1f,
            Int = 1,
            Int32 = 1,
            Int32Enum = Int32Enum.FiftyFive,
            IntEnum = IntEnum.Second,
            Num = 1.1,
            Str = "test",
            Bigint = new BigInteger(123),
            BigintStr = BigInteger.Parse("123456789012345678901234567890"),
            Decimal = 3.141592653589793m,
            DecimalStr = 3.14159265358979344719667586m,
        };

    private static void AssertJsonEqual(string expected, string actual) =>
        Assert.True(
            JsonNode.DeepEquals(JsonNode.Parse(expected), JsonNode.Parse(actual)),
            $"expected {expected}, got {actual}"
        );

    private sealed class UnregisteredType
    {
        public int Value { get; set; }
    }

    [Fact]
    public void BePublic()
    {
        Assert.True(typeof(JsonSerialization).IsPublic);
    }

    [Fact]
    public void RoundTripModelWithConverters()
    {
        var json = JsonSerialization.Serialize(CreateSimpleObject());
        AssertJsonEqual(SimpleObjectJson, json);

        var model = JsonSerialization.Deserialize<SimpleObject>(json);
        Assert.NotNull(model);
        Assert.Equal(new DateOnly(2020, 1, 1), model!.Date);
        Assert.Equal(BigInteger.Parse("123456789012345678901234567890"), model.BigintStr);
        Assert.Equal(3.14159265358979344719667586m, model.DecimalStr);
        AssertJsonEqual(json, JsonSerialization.Serialize(model));
    }

    [Fact]
    public void RoundTripOpenUnion()
    {
        var json = JsonSerialization.Serialize(WeaklyTypedOneOfObject.CreateSimpleObject(CreateSimpleObject()));
        AssertJsonEqual(SimpleObjectJson, json);

        var union = JsonSerialization.Deserialize<WeaklyTypedOneOfObject>(json);
        Assert.NotNull(union);
        Assert.NotNull(union!.SimpleObject);
        AssertJsonEqual(json, JsonSerialization.Serialize(union));
    }

    [Fact]
    public void KeepUnknownOpenUnionMember()
    {
        var union = JsonSerialization.Deserialize<WeaklyTypedOneOfObject>("3");

        Assert.NotNull(union);
        Assert.True(union!.IsUnknown());
        Assert.Equal("3", JsonSerialization.Serialize(union));
    }

    [Fact]
    public void KeepUnknownOpenEnumValue()
    {
        const string json = "{\"name\":\"n\",\"priority\":\"low\",\"status\":\"not-a-known-status\",\"value\":1}";

        var model = JsonSerialization.Deserialize<ObjectWithUnknownOpenEnum>(json);
        Assert.NotNull(model);
        Assert.False(model!.Status.IsKnown());
        AssertJsonEqual(json, JsonSerialization.Serialize(model));
    }

    [Fact]
    public void ReadNullRoot()
    {
        Assert.Null(JsonSerialization.Deserialize<SimpleObject>("null"));
        Assert.Null(JsonSerialization.Deserialize<long?>("null"));
        Assert.Throws<JsonException>(() => JsonSerialization.Deserialize<long>("null"));
        Assert.Throws<InvalidOperationException>(() => JsonSerialization.Deserialize<DiscriminatedVariantBody>("null"));

        var openUnion = JsonSerialization.Deserialize<WeaklyTypedOneOfObject>("null");
        Assert.NotNull(openUnion);
        Assert.True(openUnion!.IsUnknown());

        Assert.Null(JsonSerialization.Deserialize<OneOfSameType>("null"));
        Assert.Equal("null", JsonSerialization.Serialize<OneOfSameType?>(null));
    }

    [Fact]
    public void RejectUnionWithoutMember()
    {
        Assert.Equal("null", JsonSerialization.Serialize<SimpleObject?>(null));
        Assert.Throws<InvalidOperationException>(() => JsonSerialization.Serialize<DiscriminatedVariantBody?>(null));
        Assert.Throws<InvalidOperationException>(
            () => JsonSerialization.Serialize(new DiscriminatedVariantBody(DiscriminatedVariantBodyType.Model))
        );
    }

    [Fact]
    public void RejectUnregisteredType()
    {
        Assert.Throws<NotSupportedException>(() => JsonSerialization.Serialize(new UnregisteredType()));
        Assert.Throws<NotSupportedException>(() => JsonSerialization.Deserialize<UnregisteredType>("{}"));
    }

    [Fact]
    public void ValidateStrictly()
    {
        Assert.Throws<JsonException>(() => JsonSerialization.Deserialize<SimpleObject>("{}"));
        Assert.Throws<JsonException>(
            () => JsonSerialization.Deserialize<SimpleObject>(SimpleObjectJson.Replace("\"str\":\"test\"", "\"str\":null"))
        );
    }
}
