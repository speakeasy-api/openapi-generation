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
    public void ConvertNullableItemCollectionsWithBigIntegerStringFactory()
    {
        var factory = new BigIntegerStringJsonConverterFactory();
        AssertRoundTrip(
            factory,
            new List<BigInteger?> { null, LargeBigInteger },
            "[null,\"123456789012345678901234567890\"]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger?> { ["a"] = null, ["b"] = LargeBigInteger },
            "{\"a\":null,\"b\":\"123456789012345678901234567890\"}"
        );
    }

    [Fact]
    public void ConvertNullableItemCollectionsWithBigIntegerNumberFactory()
    {
        var factory = new BigIntegerJsonConverterFactory();
        AssertRoundTrip(
            factory,
            new List<BigInteger?> { null, LargeBigInteger },
            "[null,123456789012345678901234567890]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, BigInteger?> { ["a"] = null, ["b"] = LargeBigInteger },
            "{\"a\":null,\"b\":123456789012345678901234567890}"
        );
    }

    [Fact]
    public void ConvertNullableItemCollectionsWithDecimalStringFactory()
    {
        var factory = new DecimalStringJsonConverterFactory();
        AssertRoundTrip(
            factory,
            new List<decimal?> { null, PreciseDecimal },
            "[null,\"3.1415926535897932384626433833\"]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal?> { ["a"] = null, ["b"] = PreciseDecimal },
            "{\"a\":null,\"b\":\"3.1415926535897932384626433833\"}"
        );
    }

    [Fact]
    public void ConvertNullableItemCollectionsWithDecimalNumberFactory()
    {
        var factory = new DecimalJsonConverterFactory();
        AssertRoundTrip(
            factory,
            new List<decimal?> { null, PreciseDecimal },
            "[null,3.1415926535897932384626433833]"
        );
        AssertRoundTrip(
            factory,
            new Dictionary<string, decimal?> { ["a"] = null, ["b"] = PreciseDecimal },
            "{\"a\":null,\"b\":3.1415926535897932384626433833}"
        );
    }

    [Fact]
    public void ConvertNullableItemCollectionsWithDateTimeFactory()
    {
        var factory = new DateTimeJsonConverterFactory();
        AssertRoundTrip(
            factory,
            new List<DateTime?> { null, Timestamp },
            "[null,\"2024-01-02T03:04:05.0000000Z\"]"
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
        AssertRejected(new BigIntegerStringJsonConverterFactory());
        AssertRejected(new BigIntegerJsonConverterFactory());
    }

    [Fact]
    public void RejectUnsupportedCollectionsInDecimalFactory()
    {
        AssertRejected(new DecimalStringJsonConverterFactory());
        AssertRejected(new DecimalJsonConverterFactory());
    }

    [Fact]
    public void RejectUnsupportedCollectionsInDateTimeFactory()
    {
        AssertRejected(new DateTimeJsonConverterFactory());
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
    }

    private static void AssertRejected(JsonConverterFactory factory)
    {
        Assert.False(factory.CanConvert(typeof(List<string>)));
        Assert.Throws<NotSupportedException>(() => factory.CreateConverter(typeof(List<string>), Options));
    }

    [JsonSerializable(typeof(string))]
    internal partial class NullableItemFormatCollectionsTestJsonContext : JsonSerializerContext { }
}
