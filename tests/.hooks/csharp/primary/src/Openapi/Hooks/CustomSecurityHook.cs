#nullable enable
    namespace Openapi.Hooks
{
    using System;
    using System.Net.Http;
    using System.Net.Http.Headers;
    using System.Threading;
    using System.Threading.Tasks;
    using Openapi.Models.Shared;


    public class CustomSecurityHook : IBeforeRequestHook
    {

        public async Task<HttpRequestMessage> BeforeRequestAsync(BeforeRequestContext hookCtx, HttpRequestMessage request)
        {
            await Task.CompletedTask;

            switch (hookCtx.OperationID)
            {
                case "customSchemeAppId":
                    if (hookCtx.SecuritySource == null)
                    {
                        throw new Exception("Security source is null");
                    }


                    Security? security = hookCtx.SecuritySource() as Security;
                    if (security == null)
                    {
                        throw new Exception("Security source is not of type Security");
                    }

                    if (security.CustomSchemeAppId == null)
                    {
                        throw new Exception("CustomSchemeAppID security is not defined");
                    }

                    request.Headers.Add("X-Security-App-Id", security.CustomSchemeAppId.AppId);
                    request.Headers.Add("X-Security-Secret", security.CustomSchemeAppId.Secret);
                    break;
            }

            return request;
        }
    }
}
