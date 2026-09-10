import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { APIError } from "../sdk/models/errors/apierror.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Headers Override Request Headers", async () => {
  recordTest("headers-override-request-headers");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.methods.methodGet({
    fetchOptions: {
      headers: {
        "x-inject-header-1": "foo",
        "x-inject-header-2": "bar",
      },
    },
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.object).toEqual({ status: "OK" });

  expect(res.httpMeta.request.headers.get("x-inject-header-1")).toEqual("foo");
  expect(res.httpMeta.request.headers.get("x-inject-header-2")).toEqual("bar");
});

test("Headers Empty Response Body With Headers", async () => {
  recordTest("headers-empty-response-body-with-headers");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.responseHeaders.responseBodyEmptyWithHeaders(
    1.1,
    "hello",
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.httpMeta.response.headers.get("x-string-header")).toEqual("hello");
  expect(res.httpMeta.response.headers.get("x-number-header")).toEqual("1.1");

  // In Typescript, the `headers` dictionary is not templated when httpMeta exists
  expect("headers" in res).toBe(false);
});

test("Headers Response Without Headers", async () => {
  recordTest("headers-response-headers-none");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  // The 200 response in `errorResponseHeaders` does not include any headers.
  // Because it has a response body and some other responses in the same operation
  // include headers, this test case was previously raising a zod validation error.
  const res200 = await s.responseHeaders.errorResponseHeaders(true, 200);
  expect(res200.httpMeta.response.status).toBe(200);
  expect(res200.authToken?.token).toEqual("test-token");
});

test("Headers Error Responses Without Headers", async () => {
  recordTest("headers-error-response-headers-none");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  // 1. Basic error response (no body nor custom headers) whose
  // operation includes another response that has headers.
  await s.responseHeaders.responseHeaders(true, 400).then(
    () => {
      expect.unreachable("expected APIError");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err.httpMeta.response.status).toBe(400);
      expect(err.httpMeta.body).toBe("");
    },
  );

  // 2. Error response with a body but no custom headers whose
  // operation includes another response that has headers.
  await s.responseHeaders.responseHeaders(true, 500).then(
    () => {
      expect.unreachable("expected APIError");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err.httpMeta.response.status).toBe(500);
      expect(err.httpMeta.body).toEqual('"Internal server error."\n');
    },
  );
});

test("Headers Response Headers Optional", async () => {
  recordTest("headers-response-headers-optional");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  // 1. Success response with required header included
  const res1 = await s.responseHeaders.responseHeaders(true, 200);
  expect(res1.httpMeta.response.status).toBe(200);
  expect(res1.authToken?.token).toEqual("test-token");
  expect(res1.httpMeta.response.headers.get("x-required-header")).toEqual(
    "required",
  );
  expect(res1.httpMeta.response.headers.get("x-optional-header")).toBeNull();

  // 2. Success response with required header omitted - SDK should not throw
  const res2 = await s.responseHeaders.responseHeaders(false, 200);
  expect(res2.httpMeta.response.status).toBe(200);
  expect(res2.authToken?.token).toEqual("test-token");
  expect(res2.httpMeta.response.headers.get("x-required-header")).toBeNull();
  expect(res2.httpMeta.response.headers.get("x-optional-header")).toBeNull();
});

test("Headers Error Response Headers Optional", async () => {
  recordTest("headers-error-response-headers-optional");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  // 1. Error response with Retry-After header included
  await s.responseHeaders.errorResponseHeaders(true, 429).then(
    () => {
      expect.unreachable("expected APIError");
    },
    (err1) => {
      expect(err1).toBeInstanceOf(APIError);
      expect(err1.httpMeta.response.status).toBe(429);
      expect(err1.httpMeta.response.headers.get("retry-after")).toEqual("60");
      expect(err1.httpMeta.body).toEqual(
        '"Too many attempts. Please try again later."\n',
      );
    },
  );

  // 2. Error response with Retry-After header omitted - SDK should not throw
  await s.responseHeaders.errorResponseHeaders(false, 429).then(
    () => {
      expect.unreachable("expected APIError");
    },
    (err2) => {
      expect(err2).toBeInstanceOf(APIError);
      expect(err2.httpMeta.response.status).toBe(429);
      expect(err2.httpMeta.response.headers.get("retry-after")).toBeNull();
      expect(err2.httpMeta.body).toEqual(
        '"Too many attempts. Please try again later."\n',
      );
    },
  );
});
