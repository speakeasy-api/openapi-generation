import { expect, test } from "vitest";

import { APIPromise, SDK } from "../index.js";
import { HTTPBIN_URL } from "./common_helpers.js";

test("APIPromise is re-exported from the package root", () => {
  expect(typeof APIPromise).toBe("function");
});

test("SDK method returns an APIPromise that awaits to the parsed body", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const p = sdk.auth.noAuth();
  expect(p).toBeInstanceOf(APIPromise);

  await p;
});

test("SDK method .withResponse() returns parsed body and raw Response", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const { response } = await sdk.auth.noAuth().withResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
  expect(typeof response.headers.get("content-type")).toBe("string");
});

test("SDK method .asResponse() returns the raw Response", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const response = await sdk.auth.noAuth().asResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
});

test("params-object endpoint returns APIPromise and .withResponse() works", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const p = sdk.requestBodies.formRequestBodyStringArray({
    stuff: ["a", "b"],
  });
  expect(p).toBeInstanceOf(APIPromise);

  const { data, response } = await p.withResponse();
  expect(response.status).toBe(200);
  expect(data.form?.["stuff"]).toContain("a,b");
});

test("pure-SSE method returns APIPromise wrapping the stream", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const p = sdk.eventstreams.optionalData();
  expect(p).toBeInstanceOf(APIPromise);

  const stream = await p;
  expect(stream).toBeDefined();
  expect(
    typeof (stream as { [Symbol.asyncIterator]?: unknown })[
      Symbol.asyncIterator
    ],
  ).toBe("function");
});

test("multi-overload SSE method returns APIPromise on non-streaming overload", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const p = sdk.eventstreams.sseOverloadChat({ prompt: "x" });
  expect(p).toBeInstanceOf(APIPromise);

  const { response } = await p.withResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
});

test("multi-overload SSE method returns APIPromise on streaming overload", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const p = sdk.eventstreams.sseOverloadChat({ prompt: "x", stream: true });
  expect(p).toBeInstanceOf(APIPromise);

  const response = await p.asResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
  expect(response.headers.get("content-type")).toContain("text/event-stream");
});

test("_thenUnwrap maps the parsed value while preserving raw Response access", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const mapped = sdk.auth.noAuth()._thenUnwrap((v) => ({ wrapped: v }));
  expect(mapped).toBeInstanceOf(APIPromise);

  const value = await mapped;
  expect(value).toHaveProperty("wrapped");

  const response = await sdk.auth
    .noAuth()
    ._thenUnwrap((v) => v)
    .asResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
});

test("_thenUnwrap forwards callSource so .asResponse() resolves when transform throws", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const original = sdk.auth.noAuth();
  const mapped = original._thenUnwrap(() => {
    throw new Error("post-process boom");
  });

  await expect(mapped).rejects.toThrow(/post-process boom/);

  const response = await sdk.auth
    .noAuth()
    ._thenUnwrap(() => {
      throw new Error("post-process boom");
    })
    .asResponse();
  expect(response).toBeInstanceOf(Response);
  expect(response.status).toBe(200);
});

test("plain await, .withResponse() and .asResponse() agree on the response", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const [, wrapped, raw] = await Promise.all([
    sdk.auth.noAuth(),
    sdk.auth.noAuth().withResponse(),
    sdk.auth.noAuth().asResponse(),
  ]);

  expect(wrapped.response.status).toBe(200);
  expect(raw.status).toBe(200);
});
