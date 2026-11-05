import { afterEach, expect, test, vi } from "vitest";

import { SDK } from "../index.js";
import { envSchema, resetEnv } from "../lib/env.js";
import { SDKError } from "../sdk/models/errors/sdkerror.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

afterEach(() => {
  resetEnv();
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

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

test("Test Debug Env Var False Keeps Logging Off", async () => {
  recordTest("debug-env-var-false-keeps-logging-off");

  for (const value of ["false", "0", "off", "no", "", "debug", false, 0]) {
    expect(envSchema.parse({ SEC_DEBUG: value }).SEC_DEBUG).toBe(false);
  }

  for (const value of ["true", "1", "yes", "on", "TRUE", " on ", true, 1]) {
    expect(envSchema.parse({ SEC_DEBUG: value }).SEC_DEBUG).toBe(true);
  }

  expect(envSchema.parse({}).SEC_DEBUG).toBeUndefined();

  const group = vi.spyOn(console, "group").mockImplementation(() => {});
  vi.spyOn(console, "log").mockImplementation(() => {});
  vi.spyOn(console, "groupEnd").mockImplementation(() => {});

  vi.stubEnv("SEC_DEBUG", "false");
  await expect(
    new SDK({ serverURL: HTTPBIN_URL }).auth.apiKeyAuthGlobal(),
  ).rejects.toBeInstanceOf(SDKError);
  expect(group).not.toHaveBeenCalled();

  resetEnv();
  vi.stubEnv("SEC_DEBUG", "true");
  await expect(
    new SDK({ serverURL: HTTPBIN_URL }).auth.apiKeyAuthGlobal(),
  ).rejects.toBeInstanceOf(SDKError);
  expect(group).toHaveBeenCalled();
});
