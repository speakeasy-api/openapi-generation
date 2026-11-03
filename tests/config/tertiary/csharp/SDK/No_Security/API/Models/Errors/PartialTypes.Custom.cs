#nullable enable
namespace No_Security.API.Models.Errors
{
    public partial class RequestTimeoutErrorPayload
    {
        public string Summary => $"{Code}: {Message}";
    }

    public partial class RequestTimeoutError
    {
        public bool IsRetryable => true;
    }
}
