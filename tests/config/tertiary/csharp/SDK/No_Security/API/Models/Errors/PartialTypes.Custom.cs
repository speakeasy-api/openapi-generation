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

    public partial class SDKBaseException
    {
        public bool IsServerError => StatusCode >= 500;
    }

    public partial class SDKException
    {
        public bool IsDefault => true;
    }

    public partial class ResponseValidationException
    {
        public bool HasCause => InnerException != null;
    }
}
