#nullable enable
namespace Openapi
{
    public partial interface ISDK
    {
        string PartialClassMarker();
    }

    public partial class SDK
    {
        public string PartialClassMarker() => "partial";
    }
}
