#nullable enable
using Company.Product.Feature.Subnamespace.Models.Shared;
using Xunit;

public class UnionImplicitConversionsAdditionalShould
{
    [Fact]
    public void CombineGeneratedAndPartialDeclaredConversions()
    {
        ImplicitConversionShape fromMember = new ImplicitConversionSquare { Kind = "", Side = 2 };
        ImplicitConversionShape fromPartial = 3.0;

        Assert.Equal(ImplicitConversionShapeType.Square, fromMember.Type);
        Assert.Equal(2.0, fromMember.ImplicitConversionSquare!.Side);
        Assert.Equal(ImplicitConversionShapeType.Circle, fromPartial.Type);
        Assert.Equal(3.0, fromPartial.ImplicitConversionCircle!.Radius);
    }
}
