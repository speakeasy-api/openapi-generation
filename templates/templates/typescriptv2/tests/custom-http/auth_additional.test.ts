import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { CustomHttpOnlyResponseBody } from "../models/operations/customhttponly.js";
import { recordTest, API_TEST_SERVICE_URL } from "./common_helpers.js";

test("Custom Http Scheme Only", async () => {
  recordTest("auth-custom-security-scheme-only");

  const testScopes = ["read:products", "write:products"];

  const sdk = new SDK({
    serverURL: API_TEST_SERVICE_URL,
    customHttp: {
      userID: 54321,
      role: "manager",
      passphrase: "secure-passphrase-123",
      accessCode: 104,
      scopes: testScopes,
    },
  });

  const res: CustomHttpOnlyResponseBody = await sdk.auth.customHttpOnly();

  expect(res.grant).toBe("access_granted");
  expect(res.scopes).toEqual(testScopes);
});
