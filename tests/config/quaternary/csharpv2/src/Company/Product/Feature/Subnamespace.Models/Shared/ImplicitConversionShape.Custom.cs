#nullable enable
namespace Company.Product.Feature.Subnamespace.Models.Shared
{
    public partial class ImplicitConversionShape
    {
        public static implicit operator ImplicitConversionShape(double radius) =>
            CreateCircle(new ImplicitConversionCircle { Kind = "circle", Radius = radius });
    }
}
