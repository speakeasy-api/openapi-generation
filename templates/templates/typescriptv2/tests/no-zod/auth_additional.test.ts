import { afterEach, expect, test, vi } from "vitest";

import { SDK } from "../index.js";
import { Security } from "../sdk/models/shared/security.js";
import { APIError } from "../sdk/models/errors/apierror.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";
import { resetEnv } from "../lib/env.js";

afterEach(() => {
  resetEnv();
  vi.unstubAllEnvs();
});

test("function callbacks for oAuth support are invoked for global security", async () => {
  recordTest("auth-function-callbacks-oauth-global-security");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    security: async (): Promise<Security> => ({
      oauth2: "Bearer global",
    }),
  });

  const res = await sdk.auth.globalBearerAuth();

  expect(res.token?.token).toBe("global");
});

test("Hoisted Operation Security Basic Auth", async () => {
  recordTest("auth-hoisted-security-basic-http-only");

  const sdk = new SDK({
    security: {
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  const res = await sdk.auth.hoistedSecurityBasicHttpOnly();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.basicAuth).toBeDefined();
  expect(res.basicAuth!.authenticated).toBe(true);
  expect(res.basicAuth!.user).toBe("testUser");
});

test("Global Security Fields Ordering", async () => {
  recordTest("auth-global-security-fields-ordering");

  // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
  // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
  // which is not valid authentication for this endpoint.
  const sdk = new SDK({
    security: {
      apiKeyAuth: "testApiKey",
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
    },
  });

  try {
    await sdk.auth.globalSecurityBasicHttp();
    expect.unreachable("Expected request to fail with 401");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);
    expect((e as APIError).httpMeta.response.status).toBe(401);
  }
});

test("Hoisted Security Access Token First", async () => {
  recordTest("auth-hoisted-security-access-token-first");

  const sdk = new SDK({
    security: {
      apiKeyAuth: "testApiKey",
      accessToken: "Bearer ghp_xxxx",
    },
  });

  const res = await sdk.auth.hoistedSecurityAccessTokenFirst();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.token).toBeDefined();
  expect(res.token!.token).toBe("Bearer ghp_xxxx");
});

test("Hoisted Security Access Token Only", async () => {
  recordTest("auth-hoisted-security-access-token-only");

  const sdk = new SDK({
    security: async (): Promise<Security> => ({
      apiKeyAuth: "testApiKey",
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
      accessToken: "Bearer ghp_xxxx",
    }),
  });

  const res = await sdk.auth.hoistedSecurityAccessTokenOnly();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.token).toBeDefined();
  expect(res.token!.token).toBe("Bearer ghp_xxxx");
});

test("Hoisted Security Api Key First", async () => {
  recordTest("auth-hoisted-security-api-key-first");

  const sdk = new SDK({
    security: {
      apiKeyAuth: "testApiKey",
      accessToken: "Bearer ghp_xxxx",
    },
  });

  const res = await sdk.auth.hoistedSecurityApiKeyFirst();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.token).toBeDefined();
  expect(res.token!.token).toBe("testApiKey");
});

test("Hoisted Security Basic Http Only", async () => {
  recordTest("auth-hoisted-security-basic-http-only");

  const sdk = new SDK({
    security: {
      basicHttp: {
        username: "testUser",
        password: "testPass",
      },
      apiKeyAuth: "testApiKey",
      accessToken: "Bearer ghp_xxxx",
    },
  });

  const res = await sdk.auth.hoistedSecurityBasicHttpOnly();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.basicAuth).toBeDefined();
  expect(res.basicAuth!.authenticated).toBe(true);
  expect(res.basicAuth!.user).toBe("testUser");
});

test("Hoisted Security Invalid Field", async () => {
  recordTest("auth-hoisted-security-invalid-field");

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
    await sdk.auth.hoistedSecurityAccessTokenFirst();
    expect.unreachable("Expected request to fail with 401");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);
    expect((e as APIError).httpMeta.response.status).toBe(401);
  }
});
