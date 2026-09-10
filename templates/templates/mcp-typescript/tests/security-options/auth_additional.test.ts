import { expect, test } from "vitest";

import { SDKCore } from "../core.js";
import { recordTest } from "./common_helpers.js";
import { authGlobalSecurityOptionBasicHttp } from "../funcs/authGlobalSecurityOptionBasicHttp.js";
import { authHoistedSecurityOptionAccessTokenFirst } from "../funcs/authHoistedSecurityOptionAccessTokenFirst.js";
import { authHoistedSecurityOptionApiKeyFirst } from "../funcs/authHoistedSecurityOptionApiKeyFirst.js";
import { authHoistedSecurityOptionBasicHttpOnly } from "../funcs/authHoistedSecurityOptionBasicHttpOnly.js";

test("Global Security Basic Http Success", async () => {
  recordTest("auth-basic-http-global-option");

  // Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
  const sdk = new SDKCore({
    security: {
      BasicHttp: {
        username: "testUser",
        password: "testPass",
      },
      AccessToken: {
        AccessToken: "Bearer ignored",
      },
    },
  });

  const res = await authGlobalSecurityOptionBasicHttp(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.basicAuth).toBeDefined();
  expect(res.value.basicAuth!.authenticated).toBe(true);
  expect(res.value.basicAuth!.user).toBe("testUser");
});

test("Global Security Fields Ordering", async () => {
  recordTest("auth-global-security-option-fields-ordering");

  // Expected to fail since APIKeyAuth takes priority over BasicHTTP in global security definition
  const sdk = new SDKCore({
    security: {
      ApiKeyAuth: {
        ApiKeyAuth: "Bearer test_api_key",
      },
      BasicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  const res = await authGlobalSecurityOptionBasicHttp(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(401);
});

test("Hoisted Security Access Token First", async () => {
  recordTest("auth-hoisted-security-option-access-token-first");

  const sdk = new SDKCore({
    security: {
      AccessToken: {
        AccessToken: "Bearer ghp_xxxx",
      },
    },
  });

  const res = await authHoistedSecurityOptionAccessTokenFirst(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.tokenAuthResponse).toBeDefined();
  expect(res.value.tokenAuthResponse!.token).toBe("Bearer ghp_xxxx");
});

test("Hoisted Security Api Key First", async () => {
  recordTest("auth-hoisted-security-option-api-key-first");

  const sdk = new SDKCore({
    security: {
      ApiKeyAuth: {
        ApiKeyAuth: "testApiKey",
      },
    },
  });

  const res = await authHoistedSecurityOptionApiKeyFirst(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.tokenAuthResponse).toBeDefined();
  expect(res.value.tokenAuthResponse!.token).toBe("testApiKey");
});

test("Hoisted Security Basic Http Only", async () => {
  recordTest("auth-hoisted-security-option-basic-http-only");

  const sdk = new SDKCore({
    security: {
      BasicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  const res = await authHoistedSecurityOptionBasicHttpOnly(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(200);
  expect(res.value.basicAuth).toBeDefined();
  expect(res.value.basicAuth!.authenticated).toBe(true);
  expect(res.value.basicAuth!.user).toBe("testUser");
});

test("Hoisted Security Invalid Option", async () => {
  recordTest("auth-hoisted-security-invalid-option");

  // Provide only BasicHttp — not valid for accessTokenFirst which expects bearer/apiKey
  const sdk = new SDKCore({
    security: {
      BasicHttp: {
        username: "user",
        password: "pass",
      },
    },
  });

  const res = await authHoistedSecurityOptionAccessTokenFirst(sdk);
  if (!res.ok) throw res.error;

  expect(res.value.StatusCode).toBe(401);
});
