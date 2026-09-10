#nullable enable
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Xunit;
using Openapi.Models.Shared;
using Openapi.Utils;

public class Helpers
{
    // Test service URLs - read from environment with defaults for local development
    public static readonly string HttpBinPort = Environment.GetEnvironmentVariable("HTTPBIN_PORT") ?? "35123";
    public static readonly string ApiTestServicePort = Environment.GetEnvironmentVariable("API_TEST_SERVICE_PORT") ?? "35456";
    public static readonly string HttpBinUrl = $"http://localhost:{HttpBinPort}";
    public static readonly string ApiTestServiceUrl = $"http://localhost:{ApiTestServicePort}";

    public static SimpleObject CreateSimpleObject() =>
        new SimpleObject()
        {
            Any = "any",
            Bool = true,
            BoolOpt = true,
            Date = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
            Enum = Openapi.Models.Shared.Enum.One,
            Float32 = 1.1F,
            Int32 = 1,
            Int32Enum = Int32Enum.FiftyFive,
            IntEnum = IntEnum.Second,
            IntOptNull = null,
            Int = 1,
            Num = 1.1D,
            NumOptNull = null,
            Str = "test",
            StrOpt = "testOptional"
        };

    public static void AssertSimpleObject(SimpleObject a)
    {
        var e = CreateSimpleObject();
        Assert.Equal(e.Any, a.Any);
        Assert.Equal(e.Bool, a.Bool);
        Assert.Equal(e.BoolOpt, a.BoolOpt);
        Assert.Equal(e.Date, a.Date);
        Assert.Equal(e.DateTime.ToUniversalTime(), a.DateTime.ToUniversalTime());
        Assert.Equal(e.Enum, a.Enum);
        Assert.Equal(e.Float32, a.Float32);
        Assert.Equal(e.Int32, a.Int32);
        Assert.Equal(e.IntOptNull, a.IntOptNull);
        Assert.Equal(e.Int, a.Int);
        Assert.Equal(e.Num, a.Num);
        Assert.Equal(e.NumOptNull, a.NumOptNull);
        Assert.Equal(e.Str, a.Str);
        Assert.Equal(e.StrOpt, a.StrOpt);
    }

    public static SimpleObjectWithType CreateSimpleObjectWithType() =>
        new SimpleObjectWithType()
        {
            Any = "any",
            Bool = true,
            BoolOpt = true,
            Date = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
            Enum = Openapi.Models.Shared.Enum.One,
            Float32 = 1.1F,
            Int32 = 1,
            Int32Enum = SimpleObjectWithTypeInt32Enum.FiftyFive,
            IntEnum = SimpleObjectWithTypeIntEnum.Second,
            IntOptNull = null,
            Int = 1,
            Num = 1.1D,
            NumOptNull = null,
            Str = "test",
            StrOpt = "testOptional",
            Type = StronglyTypedOneOfObjectType.SimpleObjectWithType.ToString()
        };

    public static void AssertSimpleObjectWithType(SimpleObjectWithType a)
    {
        var e = CreateSimpleObjectWithType();
        Assert.Equal(e.Any, a.Any);
        Assert.Equal(e.Bigint, a.Bigint);
        Assert.Equal(e.BigintStr, a.BigintStr);
        Assert.Equal(e.Bool, a.Bool);
        Assert.Equal(e.BoolOpt, a.BoolOpt);
        Assert.Equal(e.Date, a.Date);
        Assert.Equal(e.DateTime.ToUniversalTime(), a.DateTime.ToUniversalTime());
        Assert.Equal(e.Decimal, a.Decimal);
        Assert.Equal(e.DecimalNullableOpt, a.DecimalNullableOpt);
        Assert.Equal(e.DecimalStr, a.DecimalStr);
        Assert.Equal(e.Enum, a.Enum);
        Assert.Equal(e.Float32, a.Float32);
        Assert.Equal(e.Float64Str, a.Float64Str);
        Assert.Equal(e.Int32, a.Int32);
        Assert.Equal(e.Int32Enum, a.Int32Enum);
        Assert.Equal(e.Int64Str, a.Int64Str);
        Assert.Equal(e.IntEnum, a.IntEnum);
        Assert.Equal(e.IntOptNull, a.IntOptNull);
        Assert.Equal(e.Int, a.Int);
        Assert.Equal(e.Num, a.Num);
        Assert.Equal(e.NumOptNull, a.NumOptNull);
        Assert.Equal(e.Str, a.Str);
        Assert.Equal(e.StrOpt, a.StrOpt);
        Assert.Equal(e.Type, a.Type);
    }

    public static SimpleObjectCamelCase CreateSimpleObjectCamelCase() =>
        new SimpleObjectCamelCase()
        {
            AnyVal = "any",
            BoolVal = true,
            BoolOptVal = true,
            DateVal = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTimeVal = DateTime.Parse("2020-01-01T00:00:00.0000001Z"),
            EnumVal = Openapi.Models.Shared.Enum.One,
            Float32Val = 1.1F,
            Int32Val = 1,
            Int32EnumVal = Int32EnumVal.FiftyFive,
            IntEnumVal = IntEnumVal.Second,
            IntOptNullVal = null,
            IntVal = 1,
            NumVal = 1.1D,
            NumOptNullVal = null,
            StrVal = "test",
            StrOptVal = "test_optional"
        };

    public static void AssertSimpleObjectCamelCase(SimpleObjectCamelCase a)
    {
        var e = CreateSimpleObjectCamelCase();
        Assert.Equal(e.AnyVal, a.AnyVal);
        Assert.Equal(e.BoolVal, a.BoolVal);
        Assert.Equal(e.BoolOptVal, a.BoolOptVal);
        Assert.Equal(e.DateVal, a.DateVal);
        Assert.Equal(e.DateTimeVal.ToUniversalTime(), a.DateTimeVal.ToUniversalTime());
        Assert.Equal(e.EnumVal, a.EnumVal);
        Assert.Equal(e.Float32Val, a.Float32Val);
        Assert.Equal(e.Int32Val, a.Int32Val);
        Assert.Null(a.IntOptNullVal);
        Assert.Equal(e.IntVal, a.IntVal);
        Assert.Equal(e.NumVal, a.NumVal);
        Assert.Null(a.NumOptNullVal);
        Assert.Equal(e.StrVal, a.StrVal);
        Assert.Equal(e.StrOptVal, a.StrOptVal);
    }

    public static byte[] GetData() =>
        Encoding.Unicode.GetBytes(
            "{\r  \"some\": \"json\",\r  \"to\": \"be\",\r  \"uploaded\": \"in\",\r  \"a\": \"file\"\r}\r"
        );

    public static async Task<string> GetSerializedBodyJson(object request, string format = "")
    {
        var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false, format);
        Assert.NotNull(serializedBody);
        return await serializedBody.ReadAsStringAsync();
    }

    public static DeepObject CreateDeepObject() =>
        new DeepObject()
        {
            Any = Any.CreateSimpleObject(CreateSimpleObject()),
            Arr = new List<SimpleObject>() { CreateSimpleObject(), CreateSimpleObject() },
            Bool = true,
            Int = 1,
            Map = new Dictionary<string, SimpleObject>() { { "key", CreateSimpleObject() } },
            Num = 1.1D,
            Obj = CreateSimpleObject(),
            Str = "test"
        };

    public static void AssertDeepObject(DeepObject a)
    {
        var anySimple = a.Any.SimpleObject;
        Assert.NotNull(anySimple);
        AssertSimpleObject(anySimple);

        Assert.Equal(2, a.Arr.Count());
        AssertSimpleObject(a.Arr.ToList().First());
        AssertSimpleObject(a.Arr.ToList().Last());

        Assert.True(a.Bool);
        Assert.Equal(1, a.Int);

        Assert.Single(a.Map);
        AssertSimpleObject(a.Map["key"]);

        Assert.Equal(1.1D, a.Num);
        AssertSimpleObject(a.Obj);
        Assert.Equal("test", a.Str);
    }


    public static DeepObjectCamelCase CreateDeepObjectCamelCase() =>
        new DeepObjectCamelCase()
        {
            AnyVal = AnyVal.CreateSimpleObjectCamelCase(CreateSimpleObjectCamelCase()),
            ArrVal = new List<SimpleObjectCamelCase>() { CreateSimpleObjectCamelCase(), CreateSimpleObjectCamelCase() },
            BoolVal = true,
            IntVal = 1,
            MapVal = new Dictionary<string, SimpleObjectCamelCase>() { { "key", CreateSimpleObjectCamelCase() } },
            NumVal = 1.1D,
            ObjVal = CreateSimpleObjectCamelCase(),
            StrVal = "test"
        };

    public static void AssertDeepObjectCamelCase(DeepObjectCamelCase a)
    {
        var anySimple = a.AnyVal.SimpleObjectCamelCase;
        Assert.NotNull(anySimple);
        AssertSimpleObjectCamelCase(anySimple);

        Assert.Equal(2, a.ArrVal.Count());
        AssertSimpleObjectCamelCase(a.ArrVal.ToList().First());
        AssertSimpleObjectCamelCase(a.ArrVal.ToList().Last());

        Assert.True(a.BoolVal);
        Assert.Equal(1, a.IntVal);

        Assert.Single(a.MapVal);
        AssertSimpleObjectCamelCase(a.MapVal["key"]);

        Assert.Equal(1.1D, a.NumVal);
        AssertSimpleObjectCamelCase(a.ObjVal);
        Assert.Equal("test", a.StrVal);
    }

    public static void AssertDictEqual(Dictionary<string, object> a, Dictionary<string, object> b)
    {
        Assert.Equal(a.Count, b.Count);

        foreach (var pair in a)
        {
            Assert.True(b.TryGetValue(pair .Key, out var value));

            if (pair.Value is Dictionary<string, object> nestedA && value is Dictionary<string, object> nestedB)
            {
                AssertDictEqual(nestedA, nestedB);
            }
            else
            {
                Assert.Equal(pair.Value, value);
            }
        }
    }

    /// <summary>
    /// Sorts query parameters in a URL string for consistent comparison
    /// </summary>
    /// <param name="url">The URL string with query parameters</param>
    /// <returns>URL with sorted query parameters</returns>
    public static string SortQueryParameters(string url)
    {
        if (string.IsNullOrEmpty(url))
            return string.Empty;

        var parts = url.Split('?');
        if (parts.Length == 1)
            return url;

        var query = parts[1];
        var parameters = query.Split('&');

        Array.Sort(parameters, (a, b) =>
        {
            var keyA = a.Split('=')[0];
            var keyB = b.Split('=')[0];
            return string.Compare(keyA, keyB, StringComparison.Ordinal);
        });

        return parts[0] + "?" + string.Join("&", parameters);
    }

    /// <summary>
    /// CustomHttpClient is used to test both the SendAsync and CloneAsync
    /// interface methods in the base DefaultHttpClient implementation.
    /// </summary>
    /// <remarks>
    /// CloneAsync is only used in the context of Retries and is specific to C#.
    /// </remarks>
    internal class CustomHttpClient : DefaultHttpClient
    {
        public CustomHttpClient() {}

        public override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage httpRequest, CancellationToken? cancellationToken = null)
        {
            var clonedRequest = await base.CloneAsync(httpRequest);
            clonedRequest.Headers.Add("X-Custom-Header", "someValue");

            return await base.SendAsync(clonedRequest, cancellationToken);
        }
    }

    // ── Tri-state assertions over every optional+nullable field of ObjectWithOptionalNullableFields.

    // Every optional+nullable field populated: IsSet && !IsNull.
    public static void AssertObjectWithOptionalNullableFieldsAllSet(
        ObjectWithOptionalNullableFields o
    )
    {
        Assert.True(o.OptNullStr.IsSet);
        Assert.False(o.OptNullStr.IsNull);
        Assert.True(o.OptNullInt.IsSet);
        Assert.False(o.OptNullInt.IsNull);
        Assert.True(o.OptNullNum.IsSet);
        Assert.False(o.OptNullNum.IsNull);
        Assert.True(o.OptNullBool.IsSet);
        Assert.False(o.OptNullBool.IsNull);
        Assert.True(o.OptNullBigInt.IsSet);
        Assert.False(o.OptNullBigInt.IsNull);
        Assert.True(o.OptNullBigIntStr.IsSet);
        Assert.False(o.OptNullBigIntStr.IsNull);
        // OptNullDecimal is plain decimal? (unwrapped for full precision)
        Helpers.AssertNoOptionalNullableWrapper(o, "OptNullDecimal");
        Assert.True(o.OptNullDecimalStr.IsSet);
        Assert.False(o.OptNullDecimalStr.IsNull);
        Assert.True(o.OptNullArr.IsSet);
        Assert.False(o.OptNullArr.IsNull);
        Assert.True(o.OptNullBigIntStrArr.IsSet);
        Assert.False(o.OptNullBigIntStrArr.IsNull);
        Assert.True(o.OptNullMap.IsSet);
        Assert.False(o.OptNullMap.IsNull);
        Assert.True(o.OptNullDecimalStrMap.IsSet);
        Assert.False(o.OptNullDecimalStrMap.IsNull);
        Assert.True(o.OptNullUnion.IsSet);
        Assert.False(o.OptNullUnion.IsNull);
        Assert.True(o.OptNullObj.IsSet);
        Assert.False(o.OptNullObj.IsNull);
        Assert.True(o.OptNullDateTime.IsSet);
        Assert.False(o.OptNullDateTime.IsNull);
        Assert.True(o.OptNullDate.IsSet);
        Assert.False(o.OptNullDate.IsNull);
        Assert.True(o.OptNullEnum.IsSet);
        Assert.False(o.OptNullEnum.IsNull);
    }

    // Every optional+nullable field explicitly null: IsSet && IsNull.
    public static void AssertObjectWithOptionalNullableFieldsAllNull(
        ObjectWithOptionalNullableFields o
    )
    {
        Assert.True(o.OptNullStr.IsSet);
        Assert.True(o.OptNullStr.IsNull);
        Assert.True(o.OptNullInt.IsSet);
        Assert.True(o.OptNullInt.IsNull);
        Assert.True(o.OptNullNum.IsSet);
        Assert.True(o.OptNullNum.IsNull);
        Assert.True(o.OptNullBool.IsSet);
        Assert.True(o.OptNullBool.IsNull);
        Assert.True(o.OptNullBigInt.IsSet);
        Assert.True(o.OptNullBigInt.IsNull);
        Assert.True(o.OptNullBigIntStr.IsSet);
        Assert.True(o.OptNullBigIntStr.IsNull);
        // OptNullDecimal is plain decimal? (unwrapped for full precision).
        Helpers.AssertNoOptionalNullableWrapper(o, "OptNullDecimal");
        Assert.True(o.OptNullDecimalStr.IsSet);
        Assert.True(o.OptNullDecimalStr.IsNull);
        Assert.True(o.OptNullArr.IsSet);
        Assert.True(o.OptNullArr.IsNull);
        Assert.True(o.OptNullBigIntStrArr.IsSet);
        Assert.True(o.OptNullBigIntStrArr.IsNull);
        Assert.True(o.OptNullMap.IsSet);
        Assert.True(o.OptNullMap.IsNull);
        Assert.True(o.OptNullDecimalStrMap.IsSet);
        Assert.True(o.OptNullDecimalStrMap.IsNull);
        Assert.True(o.OptNullUnion.IsSet);
        Assert.True(o.OptNullUnion.IsNull);
        Assert.True(o.OptNullObj.IsSet);
        Assert.True(o.OptNullObj.IsNull);
        Assert.True(o.OptNullDateTime.IsSet);
        Assert.True(o.OptNullDateTime.IsNull);
        Assert.True(o.OptNullDate.IsSet);
        Assert.True(o.OptNullDate.IsNull);
        Assert.True(o.OptNullEnum.IsSet);
        Assert.True(o.OptNullEnum.IsNull);
    }

    // Every optional+nullable field absent: neither set nor null (distinct from null).
    public static void AssertObjectWithOptionalNullableFieldsAllAbsent(
        ObjectWithOptionalNullableFields o
    )
    {
        Assert.False(o.OptNullStr.IsSet);
        Assert.False(o.OptNullStr.IsNull);
        Assert.False(o.OptNullInt.IsSet);
        Assert.False(o.OptNullInt.IsNull);
        Assert.False(o.OptNullNum.IsSet);
        Assert.False(o.OptNullNum.IsNull);
        Assert.False(o.OptNullBool.IsSet);
        Assert.False(o.OptNullBool.IsNull);
        Assert.False(o.OptNullBigInt.IsSet);
        Assert.False(o.OptNullBigInt.IsNull);
        Assert.False(o.OptNullBigIntStr.IsSet);
        Assert.False(o.OptNullBigIntStr.IsNull);
        // OptNullDecimal is plain decimal? (unwrapped for full precision).
        Helpers.AssertNoOptionalNullableWrapper(o, "OptNullDecimal");
        Assert.False(o.OptNullDecimalStr.IsSet);
        Assert.False(o.OptNullDecimalStr.IsNull);
        Assert.False(o.OptNullArr.IsSet);
        Assert.False(o.OptNullArr.IsNull);
        Assert.False(o.OptNullBigIntStrArr.IsSet);
        Assert.False(o.OptNullBigIntStrArr.IsNull);
        Assert.False(o.OptNullMap.IsSet);
        Assert.False(o.OptNullMap.IsNull);
        Assert.False(o.OptNullDecimalStrMap.IsSet);
        Assert.False(o.OptNullDecimalStrMap.IsNull);
        Assert.False(o.OptNullUnion.IsSet);
        Assert.False(o.OptNullUnion.IsNull);
        Assert.False(o.OptNullObj.IsSet);
        Assert.False(o.OptNullObj.IsNull);
        Assert.False(o.OptNullDateTime.IsSet);
        Assert.False(o.OptNullDateTime.IsNull);
        Assert.False(o.OptNullDate.IsSet);
        Assert.False(o.OptNullDate.IsNull);
        Assert.False(o.OptNullEnum.IsSet);
        Assert.False(o.OptNullEnum.IsNull);
    }

    // Primitive optional+nullable fields populated (set, not null); complex/nested
    // fields left absent (neither set nor null). Used for the form/multipart set case,
    // whose value assertions stay on primitives.
    public static void AssertObjectWithOptionalNullableFieldsPrimitivesSet(
        ObjectWithOptionalNullableFields o
    )
    {
        Assert.True(o.OptNullStr.IsSet);
        Assert.False(o.OptNullStr.IsNull);
        Assert.True(o.OptNullInt.IsSet);
        Assert.False(o.OptNullInt.IsNull);
        Assert.True(o.OptNullNum.IsSet);
        Assert.False(o.OptNullNum.IsNull);
        Assert.True(o.OptNullBool.IsSet);
        Assert.False(o.OptNullBool.IsNull);
        Assert.True(o.OptNullBigInt.IsSet);
        Assert.False(o.OptNullBigInt.IsNull);
        Assert.True(o.OptNullBigIntStr.IsSet);
        Assert.False(o.OptNullBigIntStr.IsNull);
        // OptNullDecimal is plain decimal? (unwrapped for full precision).
        Helpers.AssertNoOptionalNullableWrapper(o, "OptNullDecimal");
        Assert.True(o.OptNullDecimalStr.IsSet);
        Assert.False(o.OptNullDecimalStr.IsNull);
        Assert.False(o.OptNullArr.IsSet);
        Assert.False(o.OptNullArr.IsNull);
        Assert.False(o.OptNullBigIntStrArr.IsSet);
        Assert.False(o.OptNullBigIntStrArr.IsNull);
        Assert.False(o.OptNullMap.IsSet);
        Assert.False(o.OptNullMap.IsNull);
        Assert.False(o.OptNullDecimalStrMap.IsSet);
        Assert.False(o.OptNullDecimalStrMap.IsNull);
        Assert.False(o.OptNullUnion.IsSet);
        Assert.False(o.OptNullUnion.IsNull);
        Assert.False(o.OptNullObj.IsSet);
        Assert.False(o.OptNullObj.IsNull);
        Assert.True(o.OptNullDateTime.IsSet);
        Assert.False(o.OptNullDateTime.IsNull);
        Assert.True(o.OptNullDate.IsSet);
        Assert.False(o.OptNullDate.IsNull);
        Assert.True(o.OptNullEnum.IsSet);
        Assert.False(o.OptNullEnum.IsNull);
    }

    // Asserts a property is *not* wrapped in OptionalNullable<T>.
    public static void AssertNoOptionalNullableWrapper(object obj, string propName)
    {
        var fieldType = obj.GetType().GetProperty(propName)!.PropertyType;
        Assert.False(
            fieldType.IsGenericType
                && fieldType.GetGenericTypeDefinition() == typeof(OptionalNullable<>),
            $"{obj.GetType().Name}.{propName} must not be wrapped in OptionalNullable<T>"
        );
    }
}
