#nullable enable
    namespace HoistedSecurity.Hooks
{
    using System;
    using System.Net.Http;
    using System.Net.Http.Headers;
    using System.Threading;
    using System.Threading.Tasks;
    using HoistedSecurity.Models.Operations;


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

                    CustomSchemeAppIdSecurity? customSecurity = hookCtx.SecuritySource() as CustomSchemeAppIdSecurity;
                    if (customSecurity == null)
                    {
                        throw new Exception("Security source is not of type CustomSchemeAppIdSecurity");
                    }

                    request.Headers.Add("X-Security-App-Id", customSecurity.AppId);
                    request.Headers.Add("X-Security-Secret", customSecurity.Secret);
                    break;
            }

            return request;
        }
    }
}
