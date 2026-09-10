#nullable enable
using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Newtonsoft.Json;
using NodaTime;
using Company.Product.Feature.Subnamespace.Models.Shared;
using Company.Product.Feature.Subnamespace.Utils;
using Xunit;

public static class Helpers
{
    public static readonly string HttpBinUrl = CommonHelpers.HttpBinUrl;

    public static SimpleObject CreateSimpleObject() =>
        new SimpleObject()
        {
            Any = "any",
            Bool = true,
            BoolOpt = true,
            Date = LocalDate.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
            Enum = Company.Product.Feature.Subnamespace.Models.Shared.Enum.One,
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

    public static async Task<string> GetSerializedBodyJson(object request, string format = "")
    {
        var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false, format);
        Assert.NotNull(serializedBody);
        return await serializedBody!.ReadAsStringAsync();
    }
}
