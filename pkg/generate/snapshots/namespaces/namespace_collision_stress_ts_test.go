package namespaces

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapTsNamespaceCollisionStress stress-tests TypeScript namespace collision detection
// by using x-speakeasy-model-namespace values that match common internal SDK directory
// names (types, funcs, hooks, lib, sdk) and standard scope names (errors, operations).
// Each "Widget" model is placed into a namespace that is likely to collide with an
// internal import alias. When collisions are detected, the generator switches all
// namespace references to go through a parent barrel import with dotted access
// (e.g., "import * as models$ from ../models/index.js" then "models$.types.Widget").
// Non-colliding SDKs are completely unaffected.
//
// Only affects UseIndexModules=true mode (where import * as <alias> barrel imports are used).
func TestSnapTsNamespaceCollisionStress(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Namespace Collision Stress Test
  version: 1.0.0
  description: >
    Stress-tests namespace collision detection by using x-speakeasy-model-namespace
    values that match common internal SDK directory names (types, funcs, hooks, lib,
    sdk, etc.). Each Widget model is placed into a namespace that is likely to collide
    with an internal import alias. The generator must detect these collisions and
    produce unique aliases where necessary.
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
    core_Widget:
      x-speakeasy-name-override: Widget
      x-speakeasy-model-namespace: core
      type: object
      properties:
        core:
          type: string`

	genYaml := `typescript:
  packageName: nstest
`

	expectedSnapshotFiles := []string{
		"src/sdk/sdk.ts",
		"src/funcs/utils.ts",
		"src/models/types/widget.ts",
	}

	expectedSnapshot := `--- src/funcs/utils.ts ---
/*
 * Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
 * Generated under the AGPL-3.0-only license.
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import * as z from "zod/v4-mini";
import { SDKCore } from "../core.js";
import { encodeJSON } from "../lib/encodings.js";
import { matchStatusCode } from "../lib/http.js";
import * as M from "../lib/matchers.js";
import { compactMap } from "../lib/primitives.js";
import { safeParse } from "../lib/schemas.js";
import { RequestOptions } from "../lib/sdks.js";
import { pathToFunc } from "../lib/url.js";
import {
  ConnectionError,
  InvalidRequestError,
  RequestAbortedError,
  RequestTimeoutError,
  UnexpectedClientError,
} from "../models/errors/http-client-errors.js";
import { ResponseValidationError } from "../models/errors/response-validation-error.js";
import { SDKError } from "../models/errors/sdk-error.js";
import { SDKValidationError } from "../models/errors/sdk-validation-error.js";
import * as models$ from "../models/index.js";
import { APICall, APIPromise } from "../types/async.js";
import { Result } from "../types/fp.js";

export function utils(
  client: SDKCore,
  request: models$.utils.Widget,
  options?: RequestOptions,
): APIPromise<
  Result<
    models$.utils.Widget,
    | SDKError
    | ResponseValidationError
    | ConnectionError
    | RequestAbortedError
    | RequestTimeoutError
    | InvalidRequestError
    | UnexpectedClientError
    | SDKValidationError
  >
> {
  return new APIPromise($do(
    client,
    request,
    options,
  ));
}

async function $do(
  client: SDKCore,
  request: models$.utils.Widget,
  options?: RequestOptions,
): Promise<
  [
    Result<
      models$.utils.Widget,
      | SDKError
      | ResponseValidationError
      | ConnectionError
      | RequestAbortedError
      | RequestTimeoutError
      | InvalidRequestError
      | UnexpectedClientError
      | SDKValidationError
    >,
    APICall,
  ]
> {
  const parsed = safeParse(
    request,
    (value) => z.parse(models$.utils.Widget$outboundSchema, value),
    "Input validation failed",
  );
  if (!parsed.ok) {
    return [parsed, { status: "invalid" }];
  }
  const payload = parsed.value;
  const body = encodeJSON("body", payload, { explode: true });

  const path = pathToFunc("/ns/utils")();

  const headers = new Headers(compactMap({
    "Content-Type": "application/json",
    Accept: "application/json",
  }));

  const context = {
    options: client._options,
    baseURL: options?.serverURL ?? client._baseURL ?? "",
    operationID: "utils",
    oAuth2Scopes: null,

    resolvedSecurity: null,

    securitySource: null,
    retryConfig: options?.retries
      || client._options.retryConfig
      || { strategy: "none" },
    retryCodes: options?.retryCodes || ["429", "500", "502", "503", "504"],
  };

  const requestRes = client._createRequest(context, {
    method: "POST",
    baseURL: options?.serverURL,
    path: path,
    headers: headers,
    body: body,
    userAgent: client._options.userAgent,
    timeoutMs: options?.timeoutMs || client._options.timeoutMs || -1,
  }, options);
  if (!requestRes.ok) {
    return [requestRes, { status: "invalid" }];
  }
  const req = requestRes.value;

  const doResult = await client._do(req, {
    context,
    isErrorStatusCode: (statusCode: number) =>
      matchStatusCode({ status: statusCode } as Response, ["4XX", "5XX"]),
    retryConfig: context.retryConfig,
    retryCodes: context.retryCodes,
  });
  if (!doResult.ok) {
    return [doResult, { status: "request-error", request: req }];
  }
  const response = doResult.value;

  const [result] = await M.match<
    models$.utils.Widget,
    | SDKError
    | ResponseValidationError
    | ConnectionError
    | RequestAbortedError
    | RequestTimeoutError
    | InvalidRequestError
    | UnexpectedClientError
    | SDKValidationError
  >(
    M.json(200, models$.utils.Widget$inboundSchema),
    M.fail("4XX"),
    M.fail("5XX"),
  )(response, req);
  if (!result.ok) {
    return [result, { status: "complete", request: req, response }];
  }

  return [result, { status: "complete", request: req, response }];
}


--- src/models/types/widget.ts ---
/*
 * Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
 * Generated under the AGPL-3.0-only license.
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import * as z from "zod/v4-mini";
import { safeParse } from "../../lib/schemas.js";
import { Result as SafeParseResult } from "../../types/fp.js";
import * as types from "../../types/primitives.js";
import { SDKValidationError } from "../errors/sdk-validation-error.js";

export type Widget = {
  types?: string | undefined;
};

/** @internal */
export const Widget$inboundSchema: z.ZodMiniType<Widget, unknown> = z.object({
  types: types.optional(types.string()),
});
/** @internal */
export type Widget$Outbound = {
  types?: string | undefined;
};

/** @internal */
export const Widget$outboundSchema: z.ZodMiniType<Widget$Outbound, Widget> = z
  .object({
    types: z.optional(z.string()),
  });

export function widgetToJSON(widget: Widget): string {
  return JSON.stringify(Widget$outboundSchema.parse(widget));
}
export function widgetFromJSON(
  jsonString: string,
): SafeParseResult<Widget, SDKValidationError> {
  return safeParse(
    jsonString,
    (x) => Widget$inboundSchema.parse(JSON.parse(x)),
    ` + "`" + `Failed to parse 'Widget' from JSON` + "`" + `,
  );
}


--- src/sdk/sdk.ts ---
/*
 * Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
 * Generated under the AGPL-3.0-only license.
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { callbacks } from "../funcs/callbacks.js";
import { core } from "../funcs/core.js";
import { errors } from "../funcs/errors.js";
import { funcs } from "../funcs/funcs.js";
import { hooks } from "../funcs/hooks.js";
import { lib } from "../funcs/lib.js";
import { models } from "../funcs/models.js";
import { operations } from "../funcs/operations.js";
import { sdk } from "../funcs/sdk.js";
import { types } from "../funcs/types.js";
import { utils } from "../funcs/utils.js";
import { webhooks } from "../funcs/webhooks.js";
import { ClientSDK, RequestOptions } from "../lib/sdks.js";
import * as models$ from "../models/index.js";
import { unwrapAsync } from "../types/fp.js";

export class SDK extends ClientSDK {
  async utils(
    request: models$.utils.Widget,
    options?: RequestOptions,
  ): Promise<models$.utils.Widget> {
    return unwrapAsync(utils(
      this,
      request,
      options,
    ));
  }

  async lib(
    request: models$.lib.Widget,
    options?: RequestOptions,
  ): Promise<models$.lib.Widget> {
    return unwrapAsync(lib(
      this,
      request,
      options,
    ));
  }

  async sdk(
    request: models$.sdk.Widget,
    options?: RequestOptions,
  ): Promise<models$.sdk.Widget> {
    return unwrapAsync(sdk(
      this,
      request,
      options,
    ));
  }

  async models(
    request: models$.models.Widget,
    options?: RequestOptions,
  ): Promise<models$.models.Widget> {
    return unwrapAsync(models(
      this,
      request,
      options,
    ));
  }

  async operations(
    request: models$.operations.Widget,
    options?: RequestOptions,
  ): Promise<models$.operations.Widget> {
    return unwrapAsync(operations(
      this,
      request,
      options,
    ));
  }

  async errors(
    request: models$.errors.Widget,
    options?: RequestOptions,
  ): Promise<models$.errors.Widget> {
    return unwrapAsync(errors(
      this,
      request,
      options,
    ));
  }

  async funcs(
    request: models$.funcs.Widget,
    options?: RequestOptions,
  ): Promise<models$.funcs.Widget> {
    return unwrapAsync(funcs(
      this,
      request,
      options,
    ));
  }

  async types(
    request: models$.types.Widget,
    options?: RequestOptions,
  ): Promise<models$.types.Widget> {
    return unwrapAsync(types(
      this,
      request,
      options,
    ));
  }

  async hooks(
    request: models$.hooks.Widget,
    options?: RequestOptions,
  ): Promise<models$.hooks.Widget> {
    return unwrapAsync(hooks(
      this,
      request,
      options,
    ));
  }

  async webhooks(
    request: models$.webhooks.Widget,
    options?: RequestOptions,
  ): Promise<models$.webhooks.Widget> {
    return unwrapAsync(webhooks(
      this,
      request,
      options,
    ));
  }

  async callbacks(
    request: models$.callbacks.Widget,
    options?: RequestOptions,
  ): Promise<models$.callbacks.Widget> {
    return unwrapAsync(callbacks(
      this,
      request,
      options,
    ));
  }

  async core(
    request: models$.core.Widget,
    options?: RequestOptions,
  ): Promise<models$.core.Widget> {
    return unwrapAsync(core(
      this,
      request,
      options,
    ));
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
