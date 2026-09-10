import { expect, test } from "vitest";

import { SDKCore } from "../core.js";
import { recordTest } from "./common_helpers.js";
import { authBasicAuthOptional } from "../funcs/authBasicAuthOptional.js";
import { authBasicAuthGlobal } from "../funcs/authBasicAuthGlobal.js";
import { authBasicAuthHoisted } from "../funcs/authBasicAuthHoisted.js";

test("Basic Auth Flattened Global", async () => {
  recordTest("auth-basic-auth-flattened-global");

  const sdk = new SDKCore({
    security: {
      username: "testUser",
      password: "testPass",
    },
  });

  const res = await authBasicAuthGlobal(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.basicAuthResponse).toBeDefined();
  expect(res.value.basicAuthResponse!.authenticated).toBe(true);
  expect(res.value.basicAuthResponse!.user).toBe("testUser");
});

test("Basic Auth Flattened Hoisted", async () => {
  recordTest("auth-basic-auth-flattened-hoisted");

  const sdk = new SDKCore({
    security: {
      username: "testUser",
      password: "testPass",
    },
  });

  const res = await authBasicAuthHoisted(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.basicAuthResponse).toBeDefined();
  expect(res.value.basicAuthResponse!.authenticated).toBe(true);
  expect(res.value.basicAuthResponse!.user).toBe("testUser");
});

test("Basic Auth Operation Optional", async () => {
  recordTest("auth-basic-auth-operation-optional");

  const sdk = new SDKCore({
    security: {
      username: "wrongUser",
      password: "wrongPass",
    },
  });

  const res = await authBasicAuthOptional(sdk, {
    username: "testUser",
    password: "testPass",
  });
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.basicAuthResponse).toBeDefined();
  expect(res.value.basicAuthResponse!.authenticated).toBe(true);
  expect(res.value.basicAuthResponse!.user).toBe("testUser");

  const sdk2 = new SDKCore({
    security: {
      username: "testUser",
      password: "testPass",
    },
  });

  const res2 = await authBasicAuthOptional(sdk2, {});
  if (!res2.ok) throw res2.error;
  expect(res2.value.StatusCode).toBe(401);
});
