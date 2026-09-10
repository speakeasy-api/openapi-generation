using System;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;

namespace Openapi.Hooks
{
    public class TestCancellationHook : IBeforeRequestHook
    {
        public static bool WasCompleted { get; private set; } = false;
        public static bool WasStarted { get; private set; } = false;

        public static void Reset()
        {
            WasCompleted = false;
            WasStarted = false;
        }

        public async Task<HttpRequestMessage> BeforeRequestAsync(BeforeRequestContext hookCtx, HttpRequestMessage request)
        {

            if (hookCtx.OperationID == "cancellationHook")
            {
                WasStarted = true;
                // Delay long enough to allow external cancellation during hook execution
                await Task.Delay(5000, hookCtx.CancellationToken ?? CancellationToken.None);
                WasCompleted = true;
            }

            return request;
        }
    }
}