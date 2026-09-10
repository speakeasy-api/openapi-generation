import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { ApiKeyAuthGlobalResponse } from "../sdk/models/operations/apikeyauthglobal.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("test global security flattening", async () => {
  recordTest("auth-global-security-flattening");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    api_key_auth: "Bearer testToken",
  });

  const res: ApiKeyAuthGlobalResponse = await sdk.auth.apiKeyAuthGlobal();
  expect(res.status_code).toBe(200);
});
