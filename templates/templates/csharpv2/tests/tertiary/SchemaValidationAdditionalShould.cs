#nullable enable
using Xunit;
using No_Security.API.Models.Shared;

public class SchemaValidationAdditionalShould
{
    [Fact]
    public void DropWrongTypedElementsOfPresenceAwareStringEncodedCollectionsDuringLenientConstruct()
    {
        var obj = Compat.ConstructUnvalidated<ObjectWithOptionalNullableFields>(
            "{\"optNullDecimalStrMap\":{\"a\":\"1.5\",\"b\":{\"x\":1}},\"optNullBigIntStrArr\":[\"1\",{\"x\":1},\"2\"]}"
        );

        Assert.NotNull(obj);
        var entry = Assert.Single(obj!.OptNullDecimalStrMap.Value!);
        Assert.Equal("a", entry.Key);
        Assert.Equal(1.5m, entry.Value);
        Assert.Equal(new System.Collections.Generic.List<System.Numerics.BigInteger?> { 1, 2 }, obj.OptNullBigIntStrArr.Value);
    }
}
