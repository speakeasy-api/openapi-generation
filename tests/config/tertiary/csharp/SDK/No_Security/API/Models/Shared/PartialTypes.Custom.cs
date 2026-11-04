#nullable enable
namespace No_Security.API.Models.Shared
{
    public partial class ImplicitConversionCircle
    {
        public double Diameter => Radius * 2;
    }

    public partial class ImplicitConversionShapeType
    {
        public bool IsCircle => Value == Circle.Value;
    }

    public partial class ImplicitConversionShape
    {
        public static ImplicitConversionShape FromRadius(double radius) =>
            CreateCircle(new ImplicitConversionCircle { Kind = "circle", Radius = radius });
    }

    public partial class Color
    {
        public bool IsPrimary => Equals(Red) || Equals(Green) || Equals(Blue);
    }
}
