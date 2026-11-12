import { describe, expect, test } from "vitest";

import { SDK } from "../index.js";
import { SDKValidationError } from "../sdk/models/errors/sdkvalidationerror.js";
import {
  HTTPBIN_URL,
  RequestLogEntry,
  newRequestRecorderClient,
} from "./common_helpers.js";

function newSDK(log: RequestLogEntry[]): SDK {
  return new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: newRequestRecorderClient(log),
  });
}

describe("request parameter patterns", () => {
  test("rejects a path parameter before sending the request", async () => {
    const log: RequestLogEntry[] = [];

    await expect(
      newSDK(log).parameterPatterns.parameterPatternsSharedComponent(".."),
    ).rejects.toThrow(SDKValidationError);
    expect(log).toHaveLength(0);
  });

  test("sends the default of an omitted patterned parameter", async () => {
    const log: RequestLogEntry[] = [];

    await newSDK(log).parameterPatterns.parameterPatternsSharedComponent(
      "store_1",
    );

    expect(log).toHaveLength(1);
    expect(new URL(log[0]!.url).searchParams.get("sort")).toBe("name");
  });

  test("leaves properties of an object parameter shared with the body unconstrained", async () => {
    const log: RequestLogEntry[] = [];

    await newSDK(log).parameterPatterns.parameterPatternsSharedComponent(
      "store_1",
      { name: "ABC" },
      undefined,
      { name: "ABC" },
    );

    expect(log).toHaveLength(1);
  });
});
