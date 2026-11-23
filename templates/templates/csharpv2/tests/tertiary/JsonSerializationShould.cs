#nullable enable
using No_Security.API.Models.Shared;
using No_Security.API.Utils.Json;
using System.Text.Json;
using Xunit;

public class JsonSerializationShould
{
    [Fact]
    public void ValidateStrictlyWhenResponseValidationIsLenient()
    {
        Assert.Throws<JsonException>(() => JsonSerialization.Deserialize<AirShipment>("{}"));
        Assert.Throws<JsonException>(() => JsonSerialization.Deserialize<AirShipment>("{\"flightNumber\":null}"));
    }
}
