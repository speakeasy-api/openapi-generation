import { RequestInput } from "../lib/http.js";
import { PermanentError, TemporaryError } from "../lib/retries.js";
import {
  AfterErrorContext,
  AfterErrorHook,
  AfterSuccessContext,
  AfterSuccessHook,
  BeforeCreateRequestContext,
  BeforeCreateRequestHook,
} from "./types.js";

export class ForceRetryHook
  implements BeforeCreateRequestHook, AfterSuccessHook, AfterErrorHook
{
  nextState: "passthrough" | "force-retry" | "skip-retry" = "passthrough";

  beforeCreateRequest(
    hookCtx: BeforeCreateRequestContext,
    input: RequestInput,
  ) {
    const hdrInit = input.options?.headers || {};
    switch (new Headers(hdrInit).get("x-force-retry-hook")) {
      case "force-retry":
        this.nextState = "force-retry";
        break;
      case "skip-retry":
        this.nextState = "skip-retry";
        break;
      default:
        this.nextState = "passthrough";
        break;
    }

    if (hookCtx.retryConfig.strategy === "none") {
      this.nextState = "passthrough";
    }

    return input;
  }

  afterSuccess(_: AfterSuccessContext, response: Response): Response {
    const forceRetry = this.nextState === "force-retry";
    this.nextState = "passthrough";

    if (forceRetry) {
      throw new TemporaryError("Forcing retry", response);
    }

    return response;
  }

  afterError(_: AfterErrorContext, response: Response | null, error: unknown) {
    const skipRetry = this.nextState === "skip-retry";
    this.nextState = "passthrough";

    if (skipRetry) {
      throw new PermanentError("Terminating retry loop", {
        cause: new Error("Permanent error encountered"),
      });
    }

    return { response, error };
  }
}
