using System;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using Xunit;
using No_Security.API.Models.Shared;
using No_Security.API.Utils;
using NodaTime;

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
            Date = LocalDate.FromDateTime(DateTime.Parse("2020-01-01")),
            DateTime = DateTime.Parse("2020-01-01T00:00:00.0000001Z").ToUniversalTime(),
            Enum = No_Security.API.Models.Shared.Enum.One,
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


    /// <summary>
    /// CustomHttpClient is used to test both the SendAsync and CloneAsync
    /// interface methods in the base DefaultHttpClient implementation.
    /// </summary>
    /// <remarks>
    /// CloneAsync is only used in the context of Retries and is specific to C#.
    /// </remarks>
    internal class CustomHttpClient : SDKHttpClient
    {
        public CustomHttpClient() {}

        public override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage httpRequest)
        {
            var clonedRequest = await base.CloneAsync(httpRequest);
            clonedRequest.Headers.Add("X-Custom-Header", "someValue");

            return await base.SendAsync(clonedRequest);
        }
    }

}
