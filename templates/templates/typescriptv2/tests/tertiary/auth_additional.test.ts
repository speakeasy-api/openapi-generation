import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { SDKError } from "../sdk/models/errors/sdkerror.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Test API Key Auth Global", async () => {
  recordTest("auth-api-key-auth-global");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  try {
    await sdk.auth.apiKeyAuthGlobal();

    expect.unreachable(
      "Unexpected API response status or content-type: Status 401 Content-Type application/json",
    );
  } catch (e) {
    expect(e).toBeInstanceOf(SDKError);

    const sdkErr = e as SDKError;
    expect(sdkErr.statusCode).toBe(401);
  }
});
