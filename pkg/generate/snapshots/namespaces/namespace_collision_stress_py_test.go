package namespaces

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapPyNamespaceCollisionStress stress-tests Python namespace collision detection
// by using x-speakeasy-model-namespace values that match common internal package/module
// names (utils, models, errors, types, hooks, etc.). Each "Widget" model is placed into
// a namespace that is likely to collide with an internal Python package, and we verify
// the generator produces valid, non-colliding output.
func TestSnapPyNamespaceCollisionStress(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Namespace Collision Stress Test
  version: 1.0.0
  description: >
    Stress-tests namespace collision detection by using x-speakeasy-model-namespace
    values that match common internal SDK package/module names (utils, models, errors,
    types, hooks, sdk, etc.). Each Widget model is placed into a namespace that is
    likely to collide with an internal package. The generator must detect these
    collisions and alias imports where necessary — but only where necessary — to
    produce valid, non-colliding output.
servers:
  - url: https://api.example.com
paths:
  /ns/utils:
    post:
      operationId: utils
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
  /ns/lib:
    post:
      operationId: lib
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/lib_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/lib_Widget"
  /ns/sdk:
    post:
      operationId: sdk
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/sdk_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/sdk_Widget"
  /ns/models:
    post:
      operationId: models
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
  /ns/operations:
    post:
      operationId: operations
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
      operationId: errors
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
        "4XX":
          description: Error
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
                  code:
                    type: integer
  /ns/funcs:
    post:
      operationId: funcs
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/funcs_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/funcs_Widget"
  /ns/types:
    post:
      operationId: types
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
  /ns/timeout:
    post:
      operationId: timeout
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/timeout_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/timeout_Widget"
  /ns/retry:
    post:
      operationId: retry
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/retry_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/retry_Widget"
  /ns/retries:
    post:
      operationId: retries
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/retries_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/retries_Widget"
  /ns/hooks:
    post:
      operationId: hooks
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
  /ns/webhooks:
    post:
      operationId: webhooks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/webhooks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/webhooks_Widget"
  /ns/callbacks:
    post:
      operationId: callbacks
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/callbacks_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/callbacks_Widget"
  /ns/enums:
    post:
      operationId: enums
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/enums_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/enums_Widget"
  /ns/unions:
    post:
      operationId: unions
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/unions_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/unions_Widget"
  /ns/core:
    post:
      operationId: core
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/core_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/core_Widget"
  /ns/test:
    post:
      operationId: test
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Test_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Test_Widget"
  /ns/async:
    post:
      operationId: async
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Async_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Async_Widget"
  /ns/auth:
    post:
      operationId: auth
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Auth_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Auth_Widget"
  /ns/docs:
    post:
      operationId: docs
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Docs_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Docs_Widget"
  /ns/components:
    post:
      operationId: components
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/components_Widget"
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/components_Widget"
components:
  schemas:
    utils_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: utils
      type: object
      properties:
        utils:
          type: string
    lib_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: lib
      type: object
      properties:
        lib:
          type: string
    sdk_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: sdk
      type: object
      properties:
        sdk:
          type: string
    models_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: models
      type: object
      properties:
        models:
          type: string
    operations_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: operations
      type: object
      properties:
        operations:
          type: string
    errors_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: errors
      type: object
      properties:
        errors:
          type: string
    funcs_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: funcs
      type: object
      properties:
        funcs:
          type: string
    types_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: types
      type: object
      properties:
        types:
          type: string
    timeout_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: timeout
      type: object
      properties:
        timeout:
          type: string
    retry_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: retry
      type: object
      properties:
        retry:
          type: string
    retries_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: retries
      type: object
      properties:
        retries:
          type: string
    hooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: hooks
      type: object
      properties:
        hooks:
          type: string
    webhooks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: webhooks
      type: object
      properties:
        webhooks:
          type: string
    callbacks_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: callbacks
      type: object
      properties:
        callbacks:
          type: string
    enums_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: enums
      type: object
      properties:
        enums:
          type: string
    unions_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: unions
      type: object
      properties:
        unions:
          type: string
    core_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: core
      type: object
      properties:
        core:
          type: string
    Test_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Test
      type: object
      properties:
        test:
          type: string
    Async_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Async
      type: object
      properties:
        async:
          type: string
    Auth_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Auth
      type: object
      properties:
        auth:
          type: string
    Docs_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: Docs
      type: object
      properties:
        docs:
          type: string
    components_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: components
      type: object
      properties:
        components:
          type: string`

	genYaml := `python:
  packageName: nstest
`

	expectedSnapshotFiles := []string{
		"src/nstest/sdk.py",
	}

	expectedSnapshot := `--- src/nstest/sdk.py ---
"""Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT."""
# Generated under the AGPL-3.0-only license.
# SPDX-License-Identifier: AGPL-3.0-only

from .basesdk import BaseSDK
from .httpclient import AsyncHttpClient, ClientOwner, HttpClient, close_clients
from .sdkconfiguration import SDKConfiguration
from .utils.logger import Logger, get_default_logger
from .utils.retries import RetryConfig
import httpx
from nstest import errors as errors_, models as models_, utils as utils_
from nstest._hooks import HookContext, SDKHooks
from nstest.types import OptionalNullable, UNSET
from nstest.utils.unmarshal_json_response import unmarshal_json_response
from typing import Any, Dict, Mapping, Optional, cast
import weakref


class SDK(BaseSDK):
    r"""Namespace Collision Stress Test: Stress-tests namespace collision detection by using x-speakeasy-model-namespace values that match common internal SDK package/module names (utils, models, errors, types, hooks, sdk, etc.). Each Widget model is placed into a namespace that is likely to collide with an internal package. The generator must detect these collisions and alias imports where necessary — but only where necessary — to produce valid, non-colliding output."""

    def __init__(
        self,
        server_idx: Optional[int] = None,
        url_params: Optional[Dict[str, str]] = None,
        server_url: Optional[str] = None,
        client: Optional[HttpClient] = None,
        async_client: Optional[AsyncHttpClient] = None,
        retry_config: OptionalNullable[RetryConfig] = UNSET,
        timeout_ms: Optional[int] = None,
        debug_logger: Optional[Logger] = None,
    ) -> None:
        r"""Instantiates the SDK configuring it with the provided parameters.

        :param server_idx: The index of the server to use for all methods
        :param server_url: The server URL to use for all methods
        :param url_params: Parameters to optionally template the server URL with
        :param client: The HTTP client to use for all synchronous methods
        :param async_client: The Async HTTP client to use for all asynchronous methods
        :param retry_config: The retry configuration to use for all supported methods
        :param timeout_ms: Optional request timeout applied to each operation in milliseconds
        """
        client_supplied = True
        if client is None:
            client = httpx.Client(follow_redirects=True)
            client_supplied = False

        assert issubclass(
            type(client), HttpClient
        ), "The provided client must implement the HttpClient protocol."

        async_client_supplied = True
        if async_client is None:
            async_client = httpx.AsyncClient(follow_redirects=True)
            async_client_supplied = False

        if debug_logger is None:
            debug_logger = get_default_logger()

        assert issubclass(
            type(async_client), AsyncHttpClient
        ), "The provided async_client must implement the AsyncHttpClient protocol."

        if server_url is not None:
            if url_params is not None:
                server_url = utils_.template_url(server_url, url_params)

        BaseSDK.__init__(
            self,
            SDKConfiguration(
                client=client,
                client_supplied=client_supplied,
                async_client=async_client,
                async_client_supplied=async_client_supplied,
                server_url=server_url,
                server_idx=server_idx,
                retry_config=retry_config,
                timeout_ms=timeout_ms,
                debug_logger=debug_logger,
            ),
            parent_ref=self,
        )

        hooks = SDKHooks()

        # pylint: disable=protected-access
        self.sdk_configuration.__dict__["_hooks"] = hooks

        self.sdk_configuration = hooks.sdk_init(self.sdk_configuration)

        weakref.finalize(
            self,
            close_clients,
            cast(ClientOwner, self.sdk_configuration),
            self.sdk_configuration.client,
            self.sdk_configuration.client_supplied,
            self.sdk_configuration.async_client,
            self.sdk_configuration.async_client_supplied,
        )

    def __enter__(self):
        return self

    async def __aenter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if (
            self.sdk_configuration.client is not None
            and not self.sdk_configuration.client_supplied
        ):
            self.sdk_configuration.client.close()
        self.sdk_configuration.client = None

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if (
            self.sdk_configuration.async_client is not None
            and not self.sdk_configuration.async_client_supplied
        ):
            await self.sdk_configuration.async_client.aclose()
        self.sdk_configuration.async_client = None

    def utils(
        self,
        *,
        utils: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.utils.Widget:
        r"""
        :param utils:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.utils.Widget(
            utils=utils,
        )

        req = self._build_request(
            method="POST",
            path="/ns/utils",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.utils.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="utils",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.utils.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def utils_async(
        self,
        *,
        utils: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.utils.Widget:
        r"""
        :param utils:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.utils.Widget(
            utils=utils,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/utils",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.utils.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="utils",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.utils.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def lib(
        self,
        *,
        lib: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.lib.Widget:
        r"""
        :param lib:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.lib.Widget(
            lib=lib,
        )

        req = self._build_request(
            method="POST",
            path="/ns/lib",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.lib.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="lib",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.lib.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def lib_async(
        self,
        *,
        lib: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.lib.Widget:
        r"""
        :param lib:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.lib.Widget(
            lib=lib,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/lib",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.lib.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="lib",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.lib.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def sdk(
        self,
        *,
        sdk: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.sdk.Widget:
        r"""
        :param sdk:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.sdk.Widget(
            sdk=sdk,
        )

        req = self._build_request(
            method="POST",
            path="/ns/sdk",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.sdk.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="sdk",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.sdk.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def sdk_async(
        self,
        *,
        sdk: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.sdk.Widget:
        r"""
        :param sdk:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.sdk.Widget(
            sdk=sdk,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/sdk",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.sdk.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="sdk",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.sdk.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def models(
        self,
        *,
        models: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.Widget:
        r"""
        :param models:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.Widget(
            models=models,
        )

        req = self._build_request(
            method="POST",
            path="/ns/models",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="models",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def models_async(
        self,
        *,
        models: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.Widget:
        r"""
        :param models:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.Widget(
            models=models,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/models",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="models",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def operations(
        self,
        *,
        operations: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.operations.Widget:
        r"""
        :param operations:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.operations.Widget(
            operations=operations,
        )

        req = self._build_request(
            method="POST",
            path="/ns/operations",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.operations.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="operations",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.operations.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def operations_async(
        self,
        *,
        operations: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.operations.Widget:
        r"""
        :param operations:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.operations.Widget(
            operations=operations,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/operations",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.operations.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="operations",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.operations.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def errors(
        self,
        *,
        errors: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> errors_.Widget:
        r"""
        :param errors:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = errors_.Widget(
            errors=errors,
        )

        req = self._build_request(
            method="POST",
            path="/ns/errors",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", errors_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="errors",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        response_data: Any = None
        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(errors_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "application/json"):
            response_data = unmarshal_json_response(errors_.ClientErrorData, http_res)
            raise errors_.ClientError(response_data, http_res)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def errors_async(
        self,
        *,
        errors: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> errors_.Widget:
        r"""
        :param errors:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = errors_.Widget(
            errors=errors,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/errors",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", errors_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="errors",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        response_data: Any = None
        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(errors_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "application/json"):
            response_data = unmarshal_json_response(errors_.ClientErrorData, http_res)
            raise errors_.ClientError(response_data, http_res)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def funcs(
        self,
        *,
        funcs: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.funcs.Widget:
        r"""
        :param funcs:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.funcs.Widget(
            funcs=funcs,
        )

        req = self._build_request(
            method="POST",
            path="/ns/funcs",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.funcs.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="funcs",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.funcs.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def funcs_async(
        self,
        *,
        funcs: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.funcs.Widget:
        r"""
        :param funcs:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.funcs.Widget(
            funcs=funcs,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/funcs",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.funcs.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="funcs",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.funcs.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def types(
        self,
        *,
        types: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.types.Widget:
        r"""
        :param types:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.types.Widget(
            types=types,
        )

        req = self._build_request(
            method="POST",
            path="/ns/types",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.types.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="types",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.types.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def types_async(
        self,
        *,
        types: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.types.Widget:
        r"""
        :param types:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.types.Widget(
            types=types,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/types",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.types.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="types",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.types.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def timeout(
        self,
        *,
        timeout: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.timeout.Widget:
        r"""
        :param timeout:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.timeout.Widget(
            timeout=timeout,
        )

        req = self._build_request(
            method="POST",
            path="/ns/timeout",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.timeout.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="timeout",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.timeout.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def timeout_async(
        self,
        *,
        timeout: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.timeout.Widget:
        r"""
        :param timeout:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.timeout.Widget(
            timeout=timeout,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/timeout",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.timeout.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="timeout",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.timeout.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def retry(
        self,
        *,
        retry: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.retry.Widget:
        r"""
        :param retry:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.retry.Widget(
            retry=retry,
        )

        req = self._build_request(
            method="POST",
            path="/ns/retry",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.retry.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="retry",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.retry.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def retry_async(
        self,
        *,
        retry: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.retry.Widget:
        r"""
        :param retry:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.retry.Widget(
            retry=retry,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/retry",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.retry.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="retry",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.retry.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def retries(
        self,
        *,
        retries_: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.retries.Widget:
        r"""
        :param retries:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.retries.Widget(
            retries=retries_,
        )

        req = self._build_request(
            method="POST",
            path="/ns/retries",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.retries.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="retries",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.retries.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def retries_async(
        self,
        *,
        retries_: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.retries.Widget:
        r"""
        :param retries:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.retries.Widget(
            retries=retries_,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/retries",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.retries.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="retries",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.retries.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def hooks(
        self,
        *,
        hooks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.hooks.Widget:
        r"""
        :param hooks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.hooks.Widget(
            hooks=hooks,
        )

        req = self._build_request(
            method="POST",
            path="/ns/hooks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.hooks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="hooks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.hooks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def hooks_async(
        self,
        *,
        hooks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.hooks.Widget:
        r"""
        :param hooks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.hooks.Widget(
            hooks=hooks,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/hooks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.hooks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="hooks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.hooks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def webhooks(
        self,
        *,
        webhooks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.webhooks.Widget:
        r"""
        :param webhooks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.webhooks.Widget(
            webhooks=webhooks,
        )

        req = self._build_request(
            method="POST",
            path="/ns/webhooks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.webhooks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="webhooks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.webhooks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def webhooks_async(
        self,
        *,
        webhooks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.webhooks.Widget:
        r"""
        :param webhooks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.webhooks.Widget(
            webhooks=webhooks,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/webhooks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.webhooks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="webhooks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.webhooks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def callbacks(
        self,
        *,
        callbacks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.callbacks.Widget:
        r"""
        :param callbacks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.callbacks.Widget(
            callbacks=callbacks,
        )

        req = self._build_request(
            method="POST",
            path="/ns/callbacks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.callbacks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="callbacks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.callbacks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def callbacks_async(
        self,
        *,
        callbacks: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.callbacks.Widget:
        r"""
        :param callbacks:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.callbacks.Widget(
            callbacks=callbacks,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/callbacks",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.callbacks.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="callbacks",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.callbacks.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def enums(
        self,
        *,
        enums: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.enums.Widget:
        r"""
        :param enums:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.enums.Widget(
            enums=enums,
        )

        req = self._build_request(
            method="POST",
            path="/ns/enums",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.enums.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="enums",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.enums.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def enums_async(
        self,
        *,
        enums: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.enums.Widget:
        r"""
        :param enums:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.enums.Widget(
            enums=enums,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/enums",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.enums.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="enums",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.enums.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def unions(
        self,
        *,
        unions: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.unions.Widget:
        r"""
        :param unions:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.unions.Widget(
            unions=unions,
        )

        req = self._build_request(
            method="POST",
            path="/ns/unions",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.unions.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="unions",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.unions.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def unions_async(
        self,
        *,
        unions: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.unions.Widget:
        r"""
        :param unions:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.unions.Widget(
            unions=unions,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/unions",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.unions.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="unions",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.unions.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def core(
        self,
        *,
        core: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.core.Widget:
        r"""
        :param core:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.core.Widget(
            core=core,
        )

        req = self._build_request(
            method="POST",
            path="/ns/core",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.core.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="core",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.core.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def core_async(
        self,
        *,
        core: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.core.Widget:
        r"""
        :param core:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.core.Widget(
            core=core,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/core",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.core.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="core",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.core.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def test(
        self,
        *,
        test: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.test.Widget:
        r"""
        :param test:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.test.Widget(
            test=test,
        )

        req = self._build_request(
            method="POST",
            path="/ns/test",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.test.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="test",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.test.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def test_async(
        self,
        *,
        test: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.test.Widget:
        r"""
        :param test:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.test.Widget(
            test=test,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/test",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.test.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="test",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.test.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def async_(
        self,
        *,
        async_: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.async_.Widget:
        r"""
        :param async_:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.async_.Widget(
            async_=async_,
        )

        req = self._build_request(
            method="POST",
            path="/ns/async",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.async_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="async",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.async_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def async__async(
        self,
        *,
        async_: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.async_.Widget:
        r"""
        :param async_:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.async_.Widget(
            async_=async_,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/async",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.async_.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="async",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.async_.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def auth(
        self,
        *,
        auth: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.auth.Widget:
        r"""
        :param auth:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.auth.Widget(
            auth=auth,
        )

        req = self._build_request(
            method="POST",
            path="/ns/auth",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.auth.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="auth",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.auth.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def auth_async(
        self,
        *,
        auth: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.auth.Widget:
        r"""
        :param auth:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.auth.Widget(
            auth=auth,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/auth",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.auth.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="auth",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.auth.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def docs(
        self,
        *,
        docs: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.docs.Widget:
        r"""
        :param docs:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.docs.Widget(
            docs=docs,
        )

        req = self._build_request(
            method="POST",
            path="/ns/docs",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.docs.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="docs",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.docs.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def docs_async(
        self,
        *,
        docs: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.docs.Widget:
        r"""
        :param docs:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.docs.Widget(
            docs=docs,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/docs",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.docs.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="docs",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.docs.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    def components(
        self,
        *,
        components: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.components.Widget:
        r"""
        :param components:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.components.Widget(
            components=components,
        )

        req = self._build_request(
            method="POST",
            path="/ns/components",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.components.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = self.do_request(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="components",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.components.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = utils_.stream_to_text(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)

    async def components_async(
        self,
        *,
        components: Optional[str] = None,
        retries: OptionalNullable[utils_.RetryConfig] = UNSET,
        server_url: Optional[str] = None,
        timeout_ms: Optional[int] = None,
        http_headers: Optional[Mapping[str, str]] = None,
    ) -> models_.components.Widget:
        r"""
        :param components:
        :param retries: Override the default retry configuration for this method
        :param server_url: Override the default server URL for this method
        :param timeout_ms: Override the default request timeout configuration for this method in milliseconds
        :param http_headers: Additional headers to set or replace on requests.
        """
        base_url = None
        url_variables = None
        if timeout_ms is None:
            timeout_ms = self.sdk_configuration.timeout_ms

        if server_url is not None:
            base_url = server_url
        else:
            base_url = self._get_url(base_url, url_variables)

        request = models_.components.Widget(
            components=components,
        )

        req = self._build_request_async(
            method="POST",
            path="/ns/components",
            base_url=base_url,
            url_variables=url_variables,
            request=request,
            request_body_required=True,
            request_has_path_params=False,
            request_has_query_params=False,
            user_agent_header="user-agent",
            accept_header_value="application/json",
            http_headers=http_headers,
            get_serialized_body=lambda: utils_.serialize_request_body(
                request, False, False, "json", models_.components.Widget
            ),
            allow_empty_value=None,
            timeout_ms=timeout_ms,
        )

        if retries == UNSET:
            if self.sdk_configuration.retry_config is not UNSET:
                retries = self.sdk_configuration.retry_config

        retry_config = None
        if isinstance(retries, utils_.RetryConfig):
            retry_config = (retries, ["429", "500", "502", "503", "504"])

        http_res = await self.do_request_async(
            hook_ctx=HookContext(
                config=self.sdk_configuration,
                base_url=base_url or "",
                operation_id="components",
                oauth2_scopes=None,
                security_source=None,
                tags=None,
                extensions=None,
            ),
            request=req,
            is_error_status_code=lambda c: utils_.match_status_codes(["4XX", "5XX"], c),
            retry_config=retry_config,
        )

        if utils_.match_response(http_res, "200", "application/json"):
            return unmarshal_json_response(models_.components.Widget, http_res)
        if utils_.match_response(http_res, "4XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)
        if utils_.match_response(http_res, "5XX", "*"):
            http_res_text = await utils_.stream_to_text_async(http_res)
            raise errors_.SDKDefaultError("API error occurred", http_res, http_res_text)

        raise errors_.SDKDefaultError("Unexpected response received", http_res)


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
