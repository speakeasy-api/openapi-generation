import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { SDKError } from "../sdk/models/errors/sdkerror.js";
import { recordTest } from "./common_helpers.js";

test("Global Security Basic Http Success", async () => {
  recordTest("auth-basic-http-global-option");

  // Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
  const sdk = new SDK({
    security: {
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
      accessToken: {
        accessToken: "Bearer ignored",
      },
    },
  });

  const res = await sdk.auth.globalSecurityOptionBasicHttp();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.basicAuth).toBeDefined();
  expect(res.basicAuth!.authenticated).toBe(true);
  expect(res.basicAuth!.user).toBe("testUser");
});

test("Global Security Fields Ordering", async () => {
  recordTest("auth-global-security-option-fields-ordering");

  // Expected to fail since APIKeyAuth takes priority over BasicHTTP in global security definition
  const sdk = new SDK({
    security: {
      apiKeyAuth: {
        apiKeyAuth: "Bearer test_api_key",
      },
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  try {
    await sdk.auth.globalSecurityOptionBasicHttp();
    expect.unreachable("Expected request to fail with 401");
  } catch (e) {
    expect(e).toBeInstanceOf(SDKError);
    expect((e as SDKError).httpMeta.response.status).toBe(401);
  }
});

test("Hoisted Security Access Token First", async () => {
  recordTest("auth-hoisted-security-option-access-token-first");

  const sdk = new SDK({
    security: {
      accessToken: {
        accessToken: "Bearer ghp_xxxx",
      },
    },
  });

  const res = await sdk.auth.hoistedSecurityOptionAccessTokenFirst();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.tokenAuthResponse).toBeDefined();
  expect(res.tokenAuthResponse!.token).toBe("Bearer ghp_xxxx");
});

test("Hoisted Security Api Key First", async () => {
  recordTest("auth-hoisted-security-option-api-key-first");

  const sdk = new SDK({
    security: {
      apiKeyAuth: {
        apiKeyAuth: "testApiKey",
      },
    },
  });

  const res = await sdk.auth.hoistedSecurityOptionApiKeyFirst();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.tokenAuthResponse).toBeDefined();
  expect(res.tokenAuthResponse!.token).toBe("testApiKey");
});

test("Hoisted Security Basic Http Only", async () => {
  recordTest("auth-hoisted-security-option-basic-http-only");

  const sdk = new SDK({
    security: {
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  const res = await sdk.auth.hoistedSecurityOptionBasicHttpOnly();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.basicAuth).toBeDefined();
  expect(res.basicAuth!.authenticated).toBe(true);
  expect(res.basicAuth!.user).toBe("testUser");
});

test("Hoisted Security Invalid Option", async () => {
  recordTest("auth-hoisted-security-invalid-option");

  // Provide only basicHttp — not valid for accessTokenFirst which expects bearer/apiKey
  const sdk = new SDK({
    security: {
      basicHttp: {
        username: "user",
        password: "pass",
      },
    },
  });

  try {
    await sdk.auth.hoistedSecurityOptionAccessTokenFirst();
    expect.unreachable("Expected request to fail with 401");
  } catch (e) {
    expect(e).toBeInstanceOf(SDKError);
    expect((e as SDKError).httpMeta.response.status).toBe(401);
  }
});
