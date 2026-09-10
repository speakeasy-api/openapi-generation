import { expect, test, vi } from "vitest";

import { SDK } from "../index.js";
import { HTTPClient } from "../lib/http.js";
import {
  ConnectionError,
  RequestTimeoutError,
} from "../sdk/models/errors/httpclienterrors.js";
import { APIError } from "../sdk/models/errors/apierror.js";
import { RetriesAfterResponse } from "../sdk/models/operations/retriesafter.js";
import { RetriesAttemptCountResponse } from "../sdk/models/operations/retriesattemptcount.js";
import { RetriesGetResponse } from "../sdk/models/operations/retriesget.js";
import { RetriesPostResponse } from "../sdk/models/operations/retriespost.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Test Retries Succeeds", async () => {
  recordTest("retries-succeeds");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res: RetriesGetResponse = await sdk.retries.retriesGet(pseudoUUID());

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(3);
});

test("Test Retries Succeeds With Body", async () => {
  recordTest("retries-succeeds-with-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res: RetriesPostResponse = await sdk.retries.retriesGet(pseudoUUID());

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(3);
});

test("Test Retries Request Timeout", async () => {
  recordTest("retries-request-timeout");

  const fetcher = vi.fn(async () => {
    const timeoutError = new Error("Request timed out");
    timeoutError.name = "TimeoutError";
    throw timeoutError;
  });
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: new HTTPClient({ fetcher }),
  });

  try {
    await sdk.retries.retriesGet(pseudoUUID(), 10, { timeoutMs: 1 });
    expect.unreachable("SDK call should not have succeeded");
  } catch (e) {
    expect(e).toBeInstanceOf(RequestTimeoutError);
  }
  expect(fetcher).toHaveBeenCalledTimes(1);
});

test("Test Retries Timeout", async () => {
  recordTest("retries-timeout");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  try {
    await sdk.retries.retriesGet(pseudoUUID(), 1000000000, {
      retries: {
        strategy: "backoff",
        backoff: {
          initialInterval: 1,
          maxInterval: 50,
          exponent: 1.1,
          maxElapsedTime: 100,
        },
        retryConnectionErrors: false,
      },
    });

    expect.unreachable("SDK call should not have succeeded");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);

    const sdkErr = e as APIError;
    expect(sdkErr.httpMeta.response.status).toBe(503);
  }
});

test("Test Respect Retry After", async () => {
  recordTest("retries-header");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res: RetriesAfterResponse = await sdk.retries.retriesAfter(
    pseudoUUID(),
    3,
    1,
    {
      retries: {
        strategy: "backoff",
        backoff: {
          initialInterval: 5000,
          maxInterval: 10000,
          exponent: 1.1,
          maxElapsedTime: 10000,
        },
        retryConnectionErrors: false,
      },
    },
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(3);
});

test("Test Attempt Count Backoff", async () => {
  recordTest("retries-attempt-count-backoff");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res: RetriesAttemptCountResponse =
    await sdk.retries.retriesAttemptCount(pseudoUUID(), 3, 1);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(3);
});

test("Test Attempt Count Backoff Max Retries Zero", async () => {
  recordTest("retries-attempt-count-zero");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  try {
    await sdk.retries.retriesAttemptCountZero(pseudoUUID(), 2, 1);

    expect.unreachable("SDK call should not have succeeded");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);

    const sdkErr = e as APIError;
    expect(sdkErr.httpMeta.response.status).toBe(503);
  }
});

test("TestGlobalRetryConfigDisable", async () => {
  recordTest("retries-global-config-disable");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    retryConfig: { strategy: "none" },
  });

  try {
    await sdk.retries.retriesGet(pseudoUUID(), 2);

    expect.unreachable("SDK call should not have succeeded");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);

    const sdkErr = e as APIError;
    expect(sdkErr.httpMeta.response.status).toBe(503);
  }
});

test("TestGlobalRetryConfigSuccess", async () => {
  recordTest("retries-global-config-success");
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    retryConfig: {
      strategy: "backoff",
      backoff: {
        initialInterval: 1,
        maxInterval: 50,
        exponent: 1.1,
        maxElapsedTime: 200,
      },
      retryConnectionErrors: false,
    },
  });

  const res: RetriesGetResponse = await sdk.retries.retriesGet(pseudoUUID(), 2);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(2);
});

test("TestGlobalRetryConfigTimeout", async () => {
  recordTest("retries-global-config-timeout");
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    retryConfig: {
      strategy: "backoff",
      backoff: {
        initialInterval: 1,
        maxInterval: 50,
        exponent: 1.1,
        maxElapsedTime: 100,
      },
      retryConnectionErrors: false,
    },
  });

  try {
    await sdk.retries.retriesGet(pseudoUUID(), 30);

    expect.unreachable("SDK call should not have succeeded");
  } catch (e) {
    expect(e).toBeInstanceOf(APIError);

    const sdkErr = e as APIError;
    expect(sdkErr.httpMeta.response.status).toBe(503);
  }
});

test("Test Retries With Timeout Uses Fresh Signal Per Attempt", async () => {
  recordTest("retries-timeout-fresh-signal");

  const observed: Array<{ signal: AbortSignal; abortedAtSend: boolean }> = [];
  let attempt = 0;
  const httpClient = new HTTPClient({
    fetcher: async (input) => {
      const req = input as Request;
      observed.push({
        signal: req.signal,
        abortedAtSend: req.signal.aborted,
      });
      attempt++;
      if (attempt < 3) {
        return new Response("retry me", { status: 503 });
      }
      return new Response(JSON.stringify({ retries: 3 }), {
        status: 200,
        headers: { "content-type": "application/json" },
      });
    },
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });

  const res: RetriesGetResponse = await sdk.retries.retriesGet(
    pseudoUUID(),
    3,
    {
      timeoutMs: 50,
      retries: {
        strategy: "backoff",
        backoff: {
          initialInterval: 100,
          maxInterval: 200,
          exponent: 1,
          maxElapsedTime: 5000,
        },
        retryConnectionErrors: false,
      },
    },
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.retries?.retries).toBe(3);
  expect(observed.length).toBe(3);
  for (const o of observed) {
    expect(o.abortedAtSend).toBe(false);
  }
  expect(new Set(observed.map((o) => o.signal)).size).toBe(observed.length);
});

test("Test Retries Connect Error", async () => {
  recordTest("retries-connect-error");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    retryConfig: {
      strategy: "backoff",
      backoff: {
        initialInterval: 1,
        maxInterval: 50,
        exponent: 1.1,
        maxElapsedTime: 1000,
      },
      retryConnectionErrors: false,
    },
  });

  await sdk.retries.retriesConnectErrorGet().then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(ConnectionError);
      expect(`${err}`).toBe(
        "ConnectionError: Unable to make request: TypeError: fetch failed",
      );
    },
  );
});

function pseudoUUID() {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0,
      v = c == "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
