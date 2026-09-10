import { expect, test } from "vitest";

import { SDK } from "../index.js";

import { recordTest, HTTPBIN_PORT } from "./common_helpers.js";
import {
  ConnectionError,
  RequestTimeoutError,
} from "../sdk/models/errors/httpclienterrors.js";

test("Test Select Global Server By Name With Templates Defaults", async () => {
  recordTest("servers-select-global-server-by-name-with-templates-defaults");

  const sdk = new SDK({
    server: "TEMPLATED",
    hostname: "localhost",
    port: HTTPBIN_PORT,
  });
  const res = await sdk.servers.selectGlobalServer();
  expect(res.status_code).toBe(200);
});

test("Test Select Global Server By Name With Templates", async () => {
  recordTest("servers-select-global-server-by-name-with-templates-valid");

  const sdk = new SDK({
    server: "TEMPLATED",
    hostname: "localhost",
    port: HTTPBIN_PORT,
  });
  const res = await sdk.servers.selectGlobalServer();
  expect(res.status_code).toBe(200);
});

test("Test Select Server By Name With Templates Broken", async () => {
  recordTest("servers-select-global-server-by-name-with-templates-broken");

  const sdk = new SDK({
    server: "TEMPLATED",
    hostname: "broken",
    port: "12345",
  });

  try {
    await sdk.servers.selectGlobalServer({ timeout_ms: 1000 });
    expect.unreachable("expected error to be thrown");
  } catch (err: unknown) {
    switch (true) {
      case err instanceof ConnectionError:
      case err instanceof RequestTimeoutError:
        break;
      default:
        expect.fail(
          "expected error to be ConnectionError or RequestTimeoutError",
        );
    }
  }
});
