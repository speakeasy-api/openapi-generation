#nullable enable
namespace Openapi.Hooks
{
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Net.Http.Headers;
    using System.Threading;
    using System.Threading.Tasks;
    using System.Text;
    using System.Web;
    using Newtonsoft.Json;
    using Openapi.Models.Errors;
    using Openapi.Models.Shared;
    using Openapi.Utils;

    public class TestClient : DefaultHttpClient
    {
        private IDefaultHttpClient _baseClient;

        public TestClient(IDefaultHttpClient baseClient)
        {
            _baseClient = baseClient;
        }

        public override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken? cancellationToken = null)
        {
            request.Headers.Add("Client-Level-Header", "added by client");

            return await _baseClient.SendAsync(request, cancellationToken);
        }
    }

    public class TestHook : ISDKInitHook, IBeforeRequestHook, IAfterSuccessHook, IAfterErrorHook
    {
        private string _initSdkVersion = "";

        public SDKConfig SDKInit(SDKConfig config)
        {
            _initSdkVersion = Constants.SdkVersion;
            config.Client = new TestClient(config.Client);
            return config;
        }

        public async Task<HttpRequestMessage> BeforeRequestAsync(
            BeforeRequestContext hookCtx,
            HttpRequestMessage request
        )
        {
            await Task.CompletedTask;

            request.Headers.Add("Idempotency-Key", "some-key");

            switch (hookCtx.OperationID)
            {
                case "testHooks":
                    var uri = request.RequestUri;
                    if (uri == null)
                    {
                        throw new Exception("Request URI is null.");
                    }
                    var requestParams = HttpUtility.ParseQueryString(uri.Query);
                    requestParams["someParam"] = "overriddenParam";
                    request.RequestUri = new UriBuilder(uri)
                    {
                        Query = requestParams.ToString(),
                    }.Uri;
                    break;
                case "authorizationHeaderModification":
                    if (hookCtx.SecuritySource == null)
                    {
                        throw new Exception("Security source is null.");
                    }
                    var token = ((Security)hookCtx.SecuritySource()).ApiKeyAuth;
                    request.Headers.Remove("Authorization");
                    request.Headers.Add("Authorization", $"{token} modified");
                    break;
                case "hooksCustomUserAgent":
                    request.Headers.Remove("User-Agent");
                    request.Headers.Add("User-Agent", $"acme-corp/{Constants.SdkVersion} acme-corp/{Environment.Version}");
                    request.Headers.Add("X-Test-Gen-Version", Constants.SdkGenVersion);
                    request.Headers.Add("X-Test-Doc-Version", Constants.OpenApiDocVersion);
                    request.Headers.Add("X-Test-Init-Sdk-Version", _initSdkVersion);
                    break;
            }

            return request;
        }

        public async Task<HttpResponseMessage> AfterSuccessAsync(
            AfterSuccessContext hookCtx,
            HttpResponseMessage response
        )
        {
            await Task.CompletedTask;

            if (hookCtx.OperationID == "testHooksAfterResponse")
            {
                throw new Exception("validation failed");
            }

            // Testing ResponseValidationException by passing a malformed json to ResponseBodyDeserializer
            if (hookCtx.OperationID == "statusGetXSpeakeasyErrors" && (int)response.StatusCode == 418)
            {
                response.Content = new StringContent("{,}", Encoding.UTF8, "application/json");
            }

            return response;
        }

        public async Task<(HttpResponseMessage?, Exception?)> AfterErrorAsync(
            AfterErrorContext hookCtx,
            HttpResponseMessage? response,
            Exception? error
        )
        {
            await Task.CompletedTask;

            switch (hookCtx.OperationID)
            {
                case "testHooksError":
                    if (response != null && (int)response.StatusCode != 400)
                    {
                        return (null, new Exception("expected status code 400"));
                    }

                    return (null, new Exception("special test error case"));
                case "statusGetDefaultError":
                    if (response != null && (int)response.StatusCode == 418)
                    {
                        return (new HttpResponseMessage(System.Net.HttpStatusCode.OK), null);
                    }
                    break;
            }

            return (response, error);
        }
    }
}
