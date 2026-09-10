#nullable enable
namespace Speakeasy.CustomHttp.Hooks
{
    using System;
    using System.Net.Http;
    using System.Text.Json;
    using System.Threading.Tasks;
    using Speakeasy.CustomHttp.Models.Components;

    public class CustomSecurityHook : IBeforeRequestHook
    {
        public async Task<HttpRequestMessage> BeforeRequestAsync(BeforeRequestContext hookCtx, HttpRequestMessage request)
        {
            await Task.CompletedTask;

            switch (hookCtx.OperationID)
            {
                case "customHttpOnly":
                    if (hookCtx.SecuritySource == null)
                    {
                        throw new Exception("Security source is null");
                    }

                    Security? security = hookCtx.SecuritySource() as Security;
                    if (security == null)
                    {
                        throw new Exception("Security source is not of type Security");
                    }

                    if (security.CustomHttp == null)
                    {
                        throw new Exception("CustomHttp security is not defined");
                    }

                    var customHttp = security.CustomHttp;

                    request.Headers.Add("X-Security-UserID", customHttp.UserID.ToString());
                    request.Headers.Add("X-Security-Role", customHttp.Role.ToString().ToLower());
                    request.Headers.Add("X-Security-Passphrase", customHttp.Passphrase);
                    request.Headers.Add("X-Security-AccessCode", customHttp.AccessCode.ToString());

                    if (customHttp.Scopes != null && customHttp.Scopes.Count > 0)
                    {
                        var scopesJson = JsonSerializer.Serialize(customHttp.Scopes);
                        request.Headers.Add("X-Security-Scopes", scopesJson);
                    }

                    break;
            }

            return request;
        }
    }
}
