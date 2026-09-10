import {
  AfterErrorContext,
  AfterErrorHook,
  AfterSuccessContext,
  AfterSuccessHook,
  BeforeCreateRequestContext,
  BeforeCreateRequestHook,
  BeforeRequestContext,
  BeforeRequestHook,
  SDKInitHook,
  SDKInitOptions,
} from "./types.js";

import { HTTPClient, RequestInput } from "../lib/http.js";

export class TestHook
  implements
    SDKInitHook,
    BeforeRequestHook,
    AfterSuccessHook,
    AfterErrorHook,
    BeforeCreateRequestHook
{
  sdkInit(opts: SDKInitOptions): SDKInitOptions {
    opts.client = new TestClient(opts.client ?? new HTTPClient());

    return opts;
  }

  beforeCreateRequest(
    hookCtx: BeforeCreateRequestContext,
    input: RequestInput,
  ): RequestInput {
    if (hookCtx.operationID === "testHooksBeforeCreateRequestPaths") {
      input.options ??= {};

      const hdrs = new Headers(input.options.headers);
      hdrs.set("old-pathname", input.url.pathname);
      input.options.headers = hdrs;

      input.url.pathname = decodeURIComponent(input.url.pathname);
    }

    return input;
  }

  beforeRequest(hookCtx: BeforeRequestContext, request: Request): Request {
    request.headers.set("Idempotency-Key", "some-key");

    switch (hookCtx.operationID) {
      case "testHooks": {
        // add an additional parameter `someParam` to the request with value `overriddenParam`
        const u = new URL(request.url);
        u.searchParams.set("someParam", "overriddenParam");

        request = new Request(u.toString(), request);
        break;
      }
      case "authorizationHeaderModification": {
        request.headers.set(
          "Authorization",
          request.headers.get("Authorization") + " modified",
        );
        break;
      }
    }

    return request;
  }

  afterSuccess(hookCtx: AfterSuccessContext, response: Response): Response {
    switch (hookCtx.operationID) {
      case "testHooksAfterResponse":
        throw new Error("validation failed");
    }

    return response;
  }

  afterError(
    hookCtx: AfterErrorContext,
    response: Response | null,
    error: unknown,
  ): { response: Response | null; error: unknown } {
    switch (hookCtx.operationID) {
      case "testHooksError":
        if (response && response?.status != 400) {
          return {
            response: null,
            error: new Error("expected status code 400"),
          };
        }
        return { response: null, error: new Error("special test error case") };
    }

    return { response, error };
  }
}

class TestClient extends HTTPClient {
  private client: HTTPClient;

  constructor(client: HTTPClient) {
    super();
    this.client = client;
  }

  override async request(request: Request): Promise<Response> {
    request.headers.set("Client-Level-Header", "added by client");
    return this.client.request(request);
  }
}
