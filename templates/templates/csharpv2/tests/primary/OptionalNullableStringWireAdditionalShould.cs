#nullable enable
using Openapi.Models;
using Openapi.Utils;
using System;
using System.Collections.Generic;
using System.Numerics;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Text.Json.Serialization.Metadata;
using Xunit;

public class OptionalNullableStringWireAdditionalShould
{
    private static readonly BigInteger LargeBigInteger = BigInteger.Parse("9223372036854775808");
    private static readonly decimal PreciseDecimal = 3.1415926535897932384626433833m;

    [Fact]
    public void ConvertBigIntegerAsString()
    {
        AssertTriState<BigInteger?, BigIntegerStringJsonConverterFactory>(
            LargeBigInteger,
            "\"9223372036854775808\""
        );
    }

    [Fact]
    public void ConvertDecimalAsString()
    {
        AssertTriState<decimal?, DecimalStringJsonConverterFactory>(
            PreciseDecimal,
            "\"3.1415926535897932384626433833\""
        );
    }

    [Fact]
    public void ConvertDateTimeAsIsoString()
    {
        AssertTriState<DateTime?, DateTimeJsonConverterFactory>(
            new DateTime(2024, 1, 2, 3, 4, 5, DateTimeKind.Utc),
            "\"2024-01-02T03:04:05.0000000Z\""
        );
    }

    [Fact]
    public void ConvertBigIntegerListAsStrings()
    {
        AssertTriState<List<BigInteger>?, BigIntegerStringJsonConverterFactory>(
            new List<BigInteger> { LargeBigInteger, 7 },
            "[\"9223372036854775808\",\"7\"]"
        );
    }

    [Fact]
    public void ConvertNullableDecimalListAsStrings()
    {
        AssertTriState<List<decimal?>?, DecimalStringJsonConverterFactory>(
            new List<decimal?> { PreciseDecimal, null },
            "[\"3.1415926535897932384626433833\",null]"
        );
    }

    [Fact]
    public void ConvertNullableBigIntegerMapAsStrings()
    {
        AssertTriState<Dictionary<string, BigInteger?>?, BigIntegerStringJsonConverterFactory>(
            new Dictionary<string, BigInteger?> { ["big"] = LargeBigInteger, ["none"] = null },
            "{\"big\":\"9223372036854775808\",\"none\":null}"
        );
    }

    private static void AssertTriState<T, TFactory>(T value, string expectedJson)
        where TFactory : JsonConverterFactory, new()
    {
        var typeInfo = JsonMetadataServices.CreateValueInfo<OptionalNullable<T>>(
            Openapi.Utils.Json.JsonOptions.SerializerOptions,
            new OptionalNullableJsonConverter<T, TFactory>()
        );

        var json = JsonSerializer.Serialize(OptionalNullable<T>.Set(value), typeInfo);
        Assert.Equal(expectedJson, json);
        var read = JsonSerializer.Deserialize(json, typeInfo);
        Assert.True(read.IsSet);
        Assert.False(read.IsNull);
        Assert.Equal(value, read.Value);

        var nullJson = JsonSerializer.Serialize(OptionalNullable<T>.Set(default), typeInfo);
        Assert.Equal("null", nullJson);
        var readNull = JsonSerializer.Deserialize(nullJson, typeInfo);
        Assert.True(readNull.IsSet);
        Assert.True(readNull.IsNull);

        Assert.Equal("null", JsonSerializer.Serialize(OptionalNullable<T>.Absent, typeInfo));
    }
}
