import { afterEach, expect, test, vi } from "vitest";

import { SDK } from "../index.js";
import { envSchema, resetEnv } from "../lib/env.js";
import { ApiKeyAuthGlobalResponse } from "../sdk/models/operations/apikeyauthglobal.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

afterEach(() => {
  resetEnv();
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

test("test global security flattening", async () => {
  recordTest("auth-global-security-flattening");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    api_key_auth: "Bearer testToken",
  });

  const res: ApiKeyAuthGlobalResponse = await sdk.auth.apiKeyAuthGlobal();
  expect(res.status_code).toBe(200);
});

test("Test Debug Env Var False Keeps Logging Off", async () => {
  recordTest("debug-env-var-false-keeps-logging-off");

  for (const value of ["false", "0", "off", "no", "", "debug", false, 0]) {
    expect(envSchema.parse({ QUATERNARY_DEBUG: value }).QUATERNARY_DEBUG).toBe(
      false,
    );
  }

  for (const value of ["true", "1", "yes", "on", "TRUE", " on ", true, 1]) {
    expect(envSchema.parse({ QUATERNARY_DEBUG: value }).QUATERNARY_DEBUG).toBe(
      true,
    );
  }

  expect(envSchema.parse({}).QUATERNARY_DEBUG).toBeUndefined();

  const group = vi.spyOn(console, "group").mockImplementation(() => {});
  vi.spyOn(console, "log").mockImplementation(() => {});
  vi.spyOn(console, "groupEnd").mockImplementation(() => {});

  vi.stubEnv("QUATERNARY_DEBUG", "false");
  await new SDK({
    serverURL: HTTPBIN_URL,
    api_key_auth: "Bearer testToken",
  }).auth.apiKeyAuthGlobal();
  expect(group).not.toHaveBeenCalled();

  resetEnv();
  vi.stubEnv("QUATERNARY_DEBUG", "true");
  await new SDK({
    serverURL: HTTPBIN_URL,
    api_key_auth: "Bearer testToken",
  }).auth.apiKeyAuthGlobal();
  expect(group).toHaveBeenCalled();
});
