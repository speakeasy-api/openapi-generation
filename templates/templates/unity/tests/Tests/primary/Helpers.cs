using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using NUnit.Framework;
using Newtonsoft.Json;
using Openapi.Models.Shared;
using Openapi.Utils;

public static class Helpers
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
            Enum = Openapi.Models.Shared.Enum.FourAndMore,
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
        Assert.AreEqual(e.Any, a.Any);
        Assert.AreEqual(e.Bool, a.Bool);
        Assert.AreEqual(e.BoolOpt, a.BoolOpt);
        Assert.AreEqual(e.Date, a.Date);
        Assert.AreEqual(e.DateTime.ToUniversalTime(), a.DateTime.ToUniversalTime());
        Assert.AreEqual(e.Enum, a.Enum);
        Assert.AreEqual(e.Float32, a.Float32);
        Assert.AreEqual(e.Int32, a.Int32);
        Assert.AreEqual(e.IntOptNull, a.IntOptNull);
        Assert.AreEqual(e.Int, a.Int);
        Assert.AreEqual(e.Num, a.Num);
        Assert.AreEqual(e.NumOptNull, a.NumOptNull);
        Assert.AreEqual(e.Str, a.Str);
        Assert.AreEqual(e.StrOpt, a.StrOpt);
    }

    public static SimpleObjectCamelCase CreateSimpleObjectCamelCase() =>
        new SimpleObjectCamelCase()
        {
            AnyVal = "any",
            BoolVal = true,
            BoolOptVal = true,
            DateVal = DateOnly.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTimeVal = DateTime.Parse("2020-01-01T00:00:00.0000001Z"),
            EnumVal = Openapi.Models.Shared.Enum.FourAndMore,
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
        Assert.AreEqual(e.AnyVal, a.AnyVal);
        Assert.AreEqual(e.BoolVal, a.BoolVal);
        Assert.AreEqual(e.BoolOptVal, a.BoolOptVal);
        Assert.AreEqual(e.DateVal, a.DateVal);
        Assert.AreEqual(e.DateTimeVal.ToUniversalTime(), a.DateTimeVal.ToUniversalTime());
        Assert.AreEqual(e.EnumVal, a.EnumVal);
        Assert.AreEqual(e.Float32Val, a.Float32Val);
        Assert.AreEqual(e.Int32Val, a.Int32Val);
        Assert.Null(a.IntOptNullVal);
        Assert.AreEqual(e.IntVal, a.IntVal);
        Assert.AreEqual(e.NumVal, a.NumVal);
        Assert.Null(a.NumOptNullVal);
        Assert.AreEqual(e.StrVal, a.StrVal);
        Assert.AreEqual(e.StrOptVal, a.StrOptVal);
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

    public static byte[] GetData() =>
        Encoding.Unicode.GetBytes(
            "{\n  \"some\": \"json\",\r\n  \"to\": \"be\",\r\n  \"uploaded\": \"in\",\r\n  \"a\": \"file\"\r\n}\r\n"
        );

    public static string GetSerializedBodyJson(object request, string format = "")
    {
        var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false, format);
        return System.Text.Encoding.Default.GetString(serializedBody.Body);
    }

    public static void AssertDeepObject(DeepObject a)
    {
        AssertSimpleObject(a.Any.SimpleObject);

        Assert.AreEqual(2, a.Arr.Count());
        AssertSimpleObject(a.Arr.ToList().First());
        AssertSimpleObject(a.Arr.ToList().Last());

        Assert.True(a.Bool);
        Assert.AreEqual(1, a.Int);

        Assert.AreEqual(1, a.Map.Count());
        AssertSimpleObject(a.Map["key"]);

        Assert.AreEqual(1.1D, a.Num);
        AssertSimpleObject(a.Obj);
        Assert.AreEqual("test", a.Str);
    }

    public static void AssertDictEqual(Dictionary<string, object> a, Dictionary<string, object> b)
    {
        Assert.AreEqual(a.Count, b.Count);

        foreach (var pair in a)
        {
            Assert.True(b.TryGetValue(pair .Key, out var value));

            if (pair.Value is Dictionary<string, object> nestedA && value is Dictionary<string, object> nestedB)
            {
                AssertDictEqual(nestedA, nestedB);
            }
            else
            {
                Assert.AreEqual(pair.Value, value);
            }
        }
    }
}
