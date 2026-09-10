#nullable enable
namespace Speakeasy.OpenAPI
{
    using System;
    using System.Net.Http;
    using System.Threading.Tasks;
    using Speakeasy.OpenAPI.Utils;

    public class IdempotencyHook : IBeforeRequestHook
    {
        public async Task<HttpRequestMessage> BeforeRequestAsync(BeforeRequestContext hookCtx, HttpRequestMessage request)
        {
            await Task.CompletedTask;

            request.Headers.Add("Idempotency-Key", Guid.NewGuid().ToString());
            return request;
        }
    }
}
