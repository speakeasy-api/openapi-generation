import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { recordTest } from "./common_helpers.js";

test("Basic Auth Operation Optional", async () => {
  recordTest("auth-basic-auth-operation-optional");

  const sdk = new SDK({
    security: {
      username: "wrongUser",
      password: "wrongPass",
    },
  });

  const res = await sdk.auth.basicAuthOptional({
    username: "testUser",
    password: "testPass",
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.basicAuthResponse).toBeDefined();
  expect(res.basicAuthResponse!.authenticated).toBe(true);
  expect(res.basicAuthResponse!.user).toBe("testUser");

  const sdk2 = new SDK({
    security: {
      username: "testUser",
      password: "testPass",
    },
  });

  await expect(sdk2.auth.basicAuthOptional({})).rejects.toThrow();
});
