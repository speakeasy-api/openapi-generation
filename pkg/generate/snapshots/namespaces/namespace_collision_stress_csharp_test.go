package namespaces

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapCsharpNamespaceCollisionStress stress-tests C# namespace collision detection
// by using x-speakeasy-model-namespace values that match common internal C# SDK
// namespace segments (utils, types, operations, errors, hooks).
// Each "Widget" model is placed into a namespace that would normally collide with
// an internal namespace. The C# generator handles these collisions by using fully
// qualified names (e.g., Models.Utils.Widget instead of just Widget).
// A "safe" namespace is included as a non-colliding control, and "models" tests
// the parent-segment duplication case (Models/Models) which requires fully-qualified
// namespace references to avoid ambiguity.
//
// All operations use a single "widgets" tag to avoid top-level class name collisions
// (e.g., a "utils" tag would generate a Utils.cs class that conflicts with the
// internal Utils/ directory). The focus is on model namespace path collisions.
func TestSnapCsharpNamespaceCollisionStress(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Namespace Collision Stress Test
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /ns/utils:
    post:
      operationId: createUtilsWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/utils_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/utils_Widget"
  /ns/types:
    post:
      operationId: createTypesWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/types_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/types_Widget"
  /ns/operations:
    post:
      operationId: createOperationsWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/operations_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/operations_Widget"
  /ns/errors:
    post:
      operationId: createErrorsWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/errors_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/errors_Widget"
  /ns/hooks:
    post:
      operationId: createHooksWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/hooks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/hooks_Widget"
  /ns/models:
    post:
      operationId: createModelsWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/models_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/models_Widget"
  /ns/safe:
    post:
      operationId: createSafeWidget
      tags:
        - widgets
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/safe_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/safe_Widget"
components:
  schemas:
    utils_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: utils
      type: object
      properties:
        name:
          type: string
    types_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: types
      type: object
      properties:
        name:
          type: string
    operations_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: operations
      type: object
      properties:
        name:
          type: string
    errors_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: errors
      type: object
      properties:
        name:
          type: string
    hooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: hooks
      type: object
      properties:
        name:
          type: string
    models_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: models
      type: object
      properties:
        name:
          type: string
    safe_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: safe
      type: object
      properties:
        name:
          type: string
`

	genYaml := `csharp:
  packageName: NSCollisionTest
`

	expectedSnapshotFiles := []string{
		"src/NSCollisionTest/Models/Models/Widget.cs",
		"src/NSCollisionTest/Models/Utils/Widget.cs",
		"src/NSCollisionTest/Models/Safe/Widget.cs",
		"src/NSCollisionTest/Widgets.cs",
	}

	expectedSnapshot := `--- src/NSCollisionTest/Models/Models/Widget.cs ---
//------------------------------------------------------------------------------
// <auto-generated>
// This code was generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
//
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only
// Changes to this file may cause incorrect behavior and will be lost when
// the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------
#nullable enable
namespace NSCollisionTest.Models.Models
{
    using NSCollisionTest.Utils;
    using Newtonsoft.Json;

    public class Widget
    {
        [JsonProperty("name", Required = Newtonsoft.Json.Required.DisallowNull)]
        public string? Name { get; set; }
    }
}

--- src/NSCollisionTest/Models/Safe/Widget.cs ---
//------------------------------------------------------------------------------
// <auto-generated>
// This code was generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
//
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only
// Changes to this file may cause incorrect behavior and will be lost when
// the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------
#nullable enable
namespace NSCollisionTest.Models.Safe
{
    using NSCollisionTest.Utils;
    using Newtonsoft.Json;

    public class Widget
    {
        [JsonProperty("name", Required = Newtonsoft.Json.Required.DisallowNull)]
        public string? Name { get; set; }
    }
}

--- src/NSCollisionTest/Models/Utils/Widget.cs ---
//------------------------------------------------------------------------------
// <auto-generated>
// This code was generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
//
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only
// Changes to this file may cause incorrect behavior and will be lost when
// the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------
#nullable enable
namespace NSCollisionTest.Models.Utils
{
    using NSCollisionTest.Utils;
    using Newtonsoft.Json;

    public class Widget
    {
        [JsonProperty("name", Required = Newtonsoft.Json.Required.DisallowNull)]
        public string? Name { get; set; }
    }
}

--- src/NSCollisionTest/Widgets.cs ---
//------------------------------------------------------------------------------
// <auto-generated>
// This code was generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
//
// Generated under the AGPL-3.0-only license.
// SPDX-License-Identifier: AGPL-3.0-only
// Changes to this file may cause incorrect behavior and will be lost when
// the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------
#nullable enable
namespace NSCollisionTest
{
    using NSCollisionTest.Hooks;
    using NSCollisionTest.Models.Errors;
    using NSCollisionTest.Models.Hooks;
    using NSCollisionTest.Models.Models;
    using NSCollisionTest.Models.Operations;
    using NSCollisionTest.Models.Requests;
    using NSCollisionTest.Models.Safe;
    using NSCollisionTest.Models.Types;
    using NSCollisionTest.Models.Utils;
    using NSCollisionTest.Utils;
    using NSCollisionTest.Utils.Retries;
    using Newtonsoft.Json;
    using System;
    using System.Collections.Generic;
    using System.Net.Http;
    using System.Net.Http.Headers;
    using System.Threading;
    using System.Threading.Tasks;

    public interface IWidgets
    {
        /// <param name="request">A <see cref="NSCollisionTest.Models.Utils.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateUtilsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateUtilsWidgetResponse> CreateUtilsWidgetAsync(
            NSCollisionTest.Models.Utils.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Types.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateTypesWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateTypesWidgetResponse> CreateTypesWidgetAsync(
            NSCollisionTest.Models.Types.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Operations.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateOperationsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateOperationsWidgetResponse> CreateOperationsWidgetAsync(
            NSCollisionTest.Models.Operations.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Errors.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateErrorsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateErrorsWidgetResponse> CreateErrorsWidgetAsync(
            NSCollisionTest.Models.Errors.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Hooks.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateHooksWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateHooksWidgetResponse> CreateHooksWidgetAsync(
            NSCollisionTest.Models.Hooks.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Models.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateModelsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateModelsWidgetResponse> CreateModelsWidgetAsync(
            NSCollisionTest.Models.Models.Widget request,
            CancellationToken? cancellationToken = null
        );

        /// <param name="request">A <see cref="NSCollisionTest.Models.Safe.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateSafeWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public Task<CreateSafeWidgetResponse> CreateSafeWidgetAsync(
            NSCollisionTest.Models.Safe.Widget request,
            CancellationToken? cancellationToken = null
        );
    }

    public class Widgets : IWidgets
    {
        /// <summary>
        /// SDK Configuration.
        /// <see cref="SDKConfig"/>
        /// </summary>
        public SDKConfig SDKConfiguration { get; private set; }

        public Widgets(SDKConfig config)
        {
            SDKConfiguration = config;
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Utils.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateUtilsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateUtilsWidgetResponse> CreateUtilsWidgetAsync(
            NSCollisionTest.Models.Utils.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/utils";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createUtilsWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Utils.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Utils.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Utils.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateUtilsWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Types.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateTypesWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateTypesWidgetResponse> CreateTypesWidgetAsync(
            NSCollisionTest.Models.Types.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/types";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createTypesWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Types.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Types.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Types.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateTypesWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Operations.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateOperationsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateOperationsWidgetResponse> CreateOperationsWidgetAsync(
            NSCollisionTest.Models.Operations.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/operations";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createOperationsWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Operations.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Operations.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Operations.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateOperationsWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Errors.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateErrorsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateErrorsWidgetResponse> CreateErrorsWidgetAsync(
            NSCollisionTest.Models.Errors.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/errors";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createErrorsWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Errors.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Errors.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Errors.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateErrorsWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Hooks.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateHooksWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateHooksWidgetResponse> CreateHooksWidgetAsync(
            NSCollisionTest.Models.Hooks.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/hooks";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createHooksWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Hooks.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Hooks.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Hooks.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateHooksWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Models.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateModelsWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateModelsWidgetResponse> CreateModelsWidgetAsync(
            NSCollisionTest.Models.Models.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/models";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createModelsWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Models.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Models.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Models.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateModelsWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }

        /// <param name="request">A <see cref="NSCollisionTest.Models.Safe.Widget"/> parameter.</param>
        /// <param name="cancellationToken">An optional cancellation token to signal when the operation should be aborted.</param>
        /// <returns>An awaitable task that returns a <see cref="CreateSafeWidgetResponse"/> response envelope when completed.</returns>
        /// <exception cref="ArgumentNullException">The required parameter <paramref name="request"/> is null.</exception>
        /// <exception cref="OperationCanceledException">The operation was aborted via the provided cancellation token.</exception>
        /// <exception cref="HttpRequestException">The HTTP request failed due to network issues.</exception>
        /// <exception cref="ResponseValidationException">The response body could not be deserialized.</exception>
        /// <exception cref="APIException">Default API Exception. Thrown when the API returns a 4XX or 5XX response.</exception>
        public async Task<CreateSafeWidgetResponse> CreateSafeWidgetAsync(
            NSCollisionTest.Models.Safe.Widget request,
            CancellationToken? cancellationToken = null
        )
        {
            if (request == null) throw new ArgumentNullException(nameof(request));

            string baseUrl = this.SDKConfiguration.GetTemplatedServerUrl();
            var urlString = baseUrl + "/ns/safe";

            var httpRequest = new HttpRequestMessage(HttpMethod.Post, urlString);
            httpRequest.Headers.Add("user-agent", SDKConfiguration.UserAgent);

            if (!httpRequest.Headers.Contains("Accept"))
            {
                httpRequest.Headers.Add("Accept", "application/json");
            }

            var serializedBody = RequestBodySerializer.Serialize(request, "Request", "json", false, false);
            if (serializedBody != null)
            {
                httpRequest.Content = serializedBody;
            }

            var hookCtx = new HookContext(SDKConfiguration, baseUrl, "createSafeWidget", null, null, cancellationToken);

            httpRequest = await this.SDKConfiguration.Hooks.BeforeRequestAsync(new BeforeRequestContext(hookCtx), httpRequest);

            HttpResponseMessage httpResponse;
            try
            {
                httpResponse = await SDKConfiguration.Client.SendAsync(httpRequest, cancellationToken);
                int _statusCode = (int)httpResponse.StatusCode;

                if (_statusCode >= 400 && _statusCode < 500 || _statusCode >= 500 && _statusCode < 600)
                {
                    var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), httpResponse, null);
                    if (_httpResponse != null)
                    {
                        httpResponse = _httpResponse;
                    }
                }
            }
            catch (Exception _hookError)
            {
                var _httpResponse = await this.SDKConfiguration.Hooks.AfterErrorAsync(new AfterErrorContext(hookCtx), null, _hookError);
                if (_httpResponse != null)
                {
                    httpResponse = _httpResponse;
                }
                else
                {
                    throw;
                }
            }

            httpResponse = await this.SDKConfiguration.Hooks.AfterSuccessAsync(new AfterSuccessContext(hookCtx), httpResponse);

            var contentType = httpResponse.Content.Headers.ContentType?.MediaType;
            int responseStatusCode = (int)httpResponse.StatusCode;
            if (responseStatusCode == 200)
            {
                if (Utilities.IsContentTypeMatch("application/json", contentType))
                {
                    var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
                    NSCollisionTest.Models.Safe.Widget obj;
                    try
                    {
                        obj = ResponseBodyDeserializer.DeserializeNotNull<NSCollisionTest.Models.Safe.Widget>(httpResponseBody, NullValueHandling.Ignore);
                    }
                    catch (Exception ex)
                    {
                        throw new ResponseValidationException("Failed to deserialize response body into NSCollisionTest.Models.Safe.Widget.", httpRequest, httpResponse, httpResponseBody, ex);
                    }

                    var response = new CreateSafeWidgetResponse() {
                        HttpMeta = new NSCollisionTest.Models.Components.HTTPMetadata() {
                            Response = httpResponse,
                            Request = httpRequest
                        }
                    };
                    response.Widget = obj;
                    return response;
                }

                throw new NSCollisionTest.Models.Errors.APIException("Unknown content type received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 400 && responseStatusCode < 500)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }
            else if (responseStatusCode >= 500 && responseStatusCode < 600)
            {
                throw new NSCollisionTest.Models.Errors.APIException("API error occurred", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
            }

            throw new NSCollisionTest.Models.Errors.APIException("Unknown status code received", httpRequest, httpResponse, await httpResponse.Content.ReadAsStringAsync());
        }
    }
}

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
