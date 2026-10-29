#nullable enable
namespace Openapi.Models.Shared
{
    public partial class ImplicitConversionCircle
    {
        public double Diameter => Radius * 2;
    }

    public partial class ImplicitConversionShape
    {
        public static implicit operator ImplicitConversionShape(double radius) =>
            CreateCircle(new ImplicitConversionCircle { Kind = "circle", Radius = radius });
    }
}
