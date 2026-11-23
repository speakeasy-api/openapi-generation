#nullable enable
namespace No_Security.API
{
    public partial interface ISDK
    {
        string PartialTypesMarker();
    }

    public partial class SDK
    {
        public string PartialTypesMarker() => "sdk";
    }

    public partial interface IUnions
    {
        string PartialTypesMarker();
    }

    public partial class Unions
    {
        public string PartialTypesMarker() => "unions";
    }
}
