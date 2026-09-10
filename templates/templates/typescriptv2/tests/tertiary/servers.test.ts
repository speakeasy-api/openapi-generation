import { expect, test, vi } from "vitest";

import { SDK, ServerList } from "../index.js";

import { recordTest, HTTPBIN_URL } from "./common_helpers.js";
import {
  ConnectionError,
  RequestTimeoutError,
} from "../sdk/models/errors/httpclienterrors.js";
import { HTTPClient } from "../lib/http.js";

test("Test Select Global Server Valid", async () => {
  recordTest("servers-select-global-server-valid");

  const spy = vi.fn();
  const httpClient = new HTTPClient().addHook("response", (res) => {
    spy(res.status);
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });
  await sdk.servers.selectGlobalServer();

  expect(spy).toHaveBeenCalledOnce();
  expect(spy).toHaveBeenCalledWith(200);
});

test("Test Select Global Server Broken", async () => {
  recordTest("servers-select-global-server-broken");

  const url = ServerList[1];
  expect(url).toBe("http://broken");
  const sdk = new SDK({ serverURL: url });

  try {
    await sdk.servers.selectGlobalServer({ timeoutMs: 1000 });
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

test("Test Select Global Server By ID Default", async () => {
  recordTest("servers-select-global-server-by-id-default");

  const spy = vi.fn();
  const httpClient = new HTTPClient().addHook("response", (res) => {
    spy(res.status);
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });
  await sdk.servers.selectGlobalServer();

  expect(spy).toHaveBeenCalledOnce();
  expect(spy).toHaveBeenCalledWith(200);
});

test("Test Select Global Server By ID Valid", async () => {
  recordTest("servers-select-global-server-by-id-valid");

  const spy = vi.fn();
  const httpClient = new HTTPClient().addHook("response", (res) => {
    spy(res.status);
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });
  await sdk.servers.selectGlobalServer();

  expect(spy).toHaveBeenCalledOnce();
  expect(spy).toHaveBeenCalledWith(200);
});

test("Test Select Server By ID Invalid", async () => {
  recordTest("servers-select-global-server-by-id-invalid");

  expect(() => new SDK({ serverIdx: 2 })).toThrowError(
    new Error("Invalid server index 2"),
  );
});

test("Test Select Server By ID Broken", async () => {
  recordTest("servers-select-global-server-by-id-broken");

  const sdk = new SDK({ serverIdx: 1 });

  try {
    await sdk.servers.selectGlobalServer({ timeoutMs: 1000 });
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
