#nullable enable
using Openapi;
using Openapi.Models.Shared;
using Openapi.Utils;
using System;
using System.Collections.Generic;
using System.Net;
using System.Numerics;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Text.Json.Serialization.Metadata;
using System.Threading.Tasks;
using Xunit;

public partial class NullableItemFormatCollectionsAdditionalShould
{
    private static readonly BigInteger LargeBigInteger = BigInteger.Parse("123456789012345678901234567890");
    private static readonly decimal PreciseDecimal = 3.1415926535897932384626433833m;
    private static readonly DateTime Timestamp = new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc);

    private static readonly JsonSerializerOptions Options = new JsonSerializerOptions
    {
        TypeInfoResolver = NullableItemFormatCollectionsTestJsonContext.Default,
    };

    [Fact]
    public async Task SendAndReadNullableItemsWithBigIntegerArrayUnionMember()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = Request(CollectionOrString.CreateArrayOfBigint(new List<BigInteger?> { LargeBigInteger, null }));

        var res = await sdk.RequestBodies.NullableItemFormatCollectionsPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            "{\"bigintItems\":[123456789012345678901234567890,null],"
                + "\"bigintStrItems\":[null,\"123456789012345678901234567890\"],"
                + "\"collectionOrString\":[\"123456789012345678901234567890\",null],"
                + "\"dateTimeItems\":[null,\"2024-01-02T03:04:05.0000000Z\"],"
                + "\"decimalStrValues\":{\"a\":null,\"b\":\"3.1415926535897932384626433833\"}}",
            res.Object!.Data
        );
        AssertEchoed(req, res.Object.Json);
        Assert.Equal(CollectionOrStringType.ArrayOfBigint, res.Object.Json.CollectionOrString.Type);
        Assert.Equal(req.CollectionOrString.ArrayOfBigint, res.Object.Json.CollectionOrString.ArrayOfBigint);
    }

    [Fact]
    public async Task SendAndReadNullableItemsWithDecimalMapUnionMember()
    {
        var sdk = new SDK(serverUrl: Helpers.HttpBinUrl);
        var req = Request(
            CollectionOrString.CreateMapOfDecimal(new Dictionary<string, decimal?> { ["x"] = null, ["y"] = PreciseDecimal })
        );

        var res = await sdk.RequestBodies.NullableItemFormatCollectionsPostAsync(req);

        Assert.Equal(HttpStatusCode.OK, res.HttpMeta.Response.StatusCode);
        Assert.Equal(
            "{\"bigintItems\":[123456789012345678901234567890,null],"
                + "\"bigintStrItems\":[null,\"123456789012345678901234567890\"],"
                + "\"collectionOrString\":{\"x\":null,\"y\":\"3.1415926535897932384626433833\"},"
                + "\"dateTimeItems\":[null,\"2024-01-02T03:04:05.0000000Z\"],"
                + "\"decimalStrValues\":{\"a\":null,\"b\":\"3.1415926535897932384626433833\"}}",
            res.Object!.Data
        );
        AssertEchoed(req, res.Object.Json);
        Assert.Equal(CollectionOrStringType.MapOfDecimal, res.Object.Json.CollectionOrString.Type);
        Assert.Equal(req.CollectionOrString.MapOfDecimal, res.Object.Json.CollectionOrString.MapOfDecimal);
    }

    [Fact]
    public void ConvertCollectionsWithBigIntegerStringFactory()
    {
        var factory = new BigIntegerStringJsonConverterFactory();
        AssertRoundTrip(factory, LargeBigInteger, "\"123456789012345678901234567890\"");
        AssertRoundTrip<BigInteger?>(factory, LargeBigInteger, "\"123456789012345678901234567890\"");
        AssertRoundTrip<BigInteger?>(factory, null, "null");
        AssertRoundTrip(
            factory,
            new List<BigInteger> { LargeBigInteger, BigInteger.MinusOne },
            "[\"123456789012345678901234567890\",\"-1\"]"
        );
        AssertRoundTrip(
            factory,
            new List<BigInteger?> { null, LargeBigInteger },
            "[null,\"123456789012345678901234567890\"]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger> { ["a"] = LargeBigInteger, ["b"] = BigInteger.MinusOne },
            "{\"a\":\"123456789012345678901234567890\",\"b\":\"-1\"}"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger?> { ["a"] = null, ["b"] = LargeBigInteger },
            "{\"a\":null,\"b\":\"123456789012345678901234567890\"}"
        );
    }

    [Fact]
    public void ConvertCollectionsWithBigIntegerNumberFactory()
    {
        var factory = new BigIntegerJsonConverterFactory();
        AssertRoundTrip(factory, LargeBigInteger, "123456789012345678901234567890");
        AssertRoundTrip<BigInteger?>(factory, LargeBigInteger, "123456789012345678901234567890");
        AssertRoundTrip<BigInteger?>(factory, null, "null");
        AssertRoundTrip(
            factory,
            new List<BigInteger> { LargeBigInteger, BigInteger.MinusOne },
            "[123456789012345678901234567890,-1]"
        );
        AssertRoundTrip(
            factory,
            new List<BigInteger?> { null, LargeBigInteger },
            "[null,123456789012345678901234567890]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger> { ["a"] = LargeBigInteger, ["b"] = BigInteger.MinusOne },
            "{\"a\":123456789012345678901234567890,\"b\":-1}"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger?> { ["a"] = null, ["b"] = LargeBigInteger },
            "{\"a\":null,\"b\":123456789012345678901234567890}"
        );
    }

    [Fact]
    public void ConvertCollectionsWithDecimalStringFactory()
    {
        var factory = new DecimalStringJsonConverterFactory();
        AssertRoundTrip(factory, PreciseDecimal, "\"3.1415926535897932384626433833\"");
        AssertRoundTrip<decimal?>(factory, PreciseDecimal, "\"3.1415926535897932384626433833\"");
        AssertRoundTrip<decimal?>(factory, null, "null");
        AssertRoundTrip(
            factory,
            new List<decimal> { PreciseDecimal, -0.5m },
            "[\"3.1415926535897932384626433833\",\"-0.5\"]"
        );
        AssertRoundTrip(
            factory,
            new List<decimal?> { null, PreciseDecimal },
            "[null,\"3.1415926535897932384626433833\"]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal> { ["a"] = PreciseDecimal, ["b"] = -0.5m },
            "{\"a\":\"3.1415926535897932384626433833\",\"b\":\"-0.5\"}"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal?> { ["a"] = null, ["b"] = PreciseDecimal },
            "{\"a\":null,\"b\":\"3.1415926535897932384626433833\"}"
        );
    }

    [Fact]
    public void ConvertCollectionsWithDecimalNumberFactory()
    {
        var factory = new DecimalJsonConverterFactory();
        AssertRoundTrip(factory, PreciseDecimal, "3.1415926535897932384626433833");
        AssertRoundTrip<decimal?>(factory, PreciseDecimal, "3.1415926535897932384626433833");
        AssertRoundTrip<decimal?>(factory, null, "null");
        AssertRoundTrip(
            factory,
            new List<decimal> { PreciseDecimal, -0.5m },
            "[3.1415926535897932384626433833,-0.5]"
        );
        AssertRoundTrip(
            factory,
            new List<decimal?> { null, PreciseDecimal },
            "[null,3.1415926535897932384626433833]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal> { ["a"] = PreciseDecimal, ["b"] = -0.5m },
            "{\"a\":3.1415926535897932384626433833,\"b\":-0.5}"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal?> { ["a"] = null, ["b"] = PreciseDecimal },
            "{\"a\":null,\"b\":3.1415926535897932384626433833}"
        );
    }

    [Fact]
    public void ConvertCollectionsWithDateTimeFactory()
    {
        var factory = new DateTimeJsonConverterFactory();
        AssertRoundTrip(factory, Timestamp, "\"2024-01-02T03:04:05.0000000Z\"");
        AssertRoundTrip<DateTime?>(factory, Timestamp, "\"2024-01-02T03:04:05.0000000Z\"");
        AssertRoundTrip<DateTime?>(factory, null, "null");
        AssertRoundTrip(
            factory,
            new List<DateTime> { Timestamp, Timestamp.AddDays(1) },
            "[\"2024-01-02T03:04:05.0000000Z\",\"2024-01-03T03:04:05.0000000Z\"]"
        );
        AssertRoundTrip(
            factory,
            new List<DateTime?> { null, Timestamp },
            "[null,\"2024-01-02T03:04:05.0000000Z\"]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, DateTime> { ["a"] = Timestamp, ["b"] = Timestamp.AddDays(1) },
            "{\"a\":\"2024-01-02T03:04:05.0000000Z\",\"b\":\"2024-01-03T03:04:05.0000000Z\"}"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, DateTime?> { ["a"] = null, ["b"] = Timestamp },
            "{\"a\":null,\"b\":\"2024-01-02T03:04:05.0000000Z\"}"
        );
    }

    [Fact]
    public void RejectUnsupportedCollectionsInBigIntegerFactory()
    {
        AssertRejected<BigInteger>(new BigIntegerStringJsonConverterFactory());
        AssertRejected<BigInteger>(new BigIntegerJsonConverterFactory());
    }

    [Fact]
    public void RejectUnsupportedCollectionsInDecimalFactory()
    {
        AssertRejected<decimal>(new DecimalStringJsonConverterFactory());
        AssertRejected<decimal>(new DecimalJsonConverterFactory());
    }

    [Fact]
    public void RejectUnsupportedCollectionsInDateTimeFactory()
    {
        AssertRejected<DateTime>(new DateTimeJsonConverterFactory());
    }

    private static NullableItemFormatCollections Request(CollectionOrString collectionOrString) =>
        new NullableItemFormatCollections
        {
            BigintStrItems = new List<BigInteger?> { null, LargeBigInteger },
            BigintItems = new List<BigInteger?> { LargeBigInteger, null },
            DecimalStrValues = new Dictionary<string, decimal?> { ["a"] = null, ["b"] = PreciseDecimal },
            DateTimeItems = new List<DateTime?> { null, Timestamp },
            CollectionOrString = collectionOrString,
        };

    private static void AssertEchoed(NullableItemFormatCollections expected, NullableItemFormatCollections actual)
    {
        Assert.Equal(expected.BigintStrItems, actual.BigintStrItems);
        Assert.Equal(expected.BigintItems, actual.BigintItems);
        Assert.Equal(expected.DecimalStrValues, actual.DecimalStrValues);
        Assert.Equal(expected.DateTimeItems, actual.DateTimeItems);
    }

    private static void AssertRoundTrip<T>(JsonConverterFactory factory, T value, string expectedJson)
    {
        Assert.True(factory.CanConvert(typeof(T)));
        var converter = factory.CreateConverter(typeof(T), Options);
        Assert.IsAssignableFrom<JsonConverter<T>>(converter);
        var typeInfo = JsonMetadataServices.CreateValueInfo<T>(Options, converter!);

        var json = JsonSerializer.Serialize(value, typeInfo);

        Assert.Equal(expectedJson, json);
        Assert.Equal(value, JsonSerializer.Deserialize(json, typeInfo));
        var wrongContainer = expectedJson.StartsWith("[") ? "{}" : "[]";
        Assert.Throws<JsonException>(() => JsonSerializer.Deserialize(wrongContainer, typeInfo));
    }

    private static void AssertRejected<T>(JsonConverterFactory factory)
        where T : struct
    {
        foreach (var type in new[] { typeof(List<string>), typeof(Dictionary<int, T>), typeof(Dictionary<int, T?>) })
        {
            Assert.False(factory.CanConvert(type));
            Assert.Throws<NotSupportedException>(() => factory.CreateConverter(type, Options));
        }
        AssertNullHandling<List<T>>(factory, "[null]");
        AssertNullHandling<Dictionary<string, T>>(factory, "{\"a\":null}");
    }

    private static void AssertNullHandling<T>(JsonConverterFactory factory, string jsonWithNullItem)
        where T : class
    {
        var typeInfo = JsonMetadataServices.CreateValueInfo<T>(Options, factory.CreateConverter(typeof(T), Options)!);

        Assert.Null(JsonSerializer.Deserialize("null", typeInfo));
        Assert.Throws<JsonException>(() => JsonSerializer.Deserialize(jsonWithNullItem, typeInfo));
    }

    [JsonSerializable(typeof(string))]
    internal partial class NullableItemFormatCollectionsTestJsonContext : JsonSerializerContext { }
}
