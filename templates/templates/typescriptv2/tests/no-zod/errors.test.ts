import { test, expect } from "vitest";

import { SDK } from "../index.js";
import { APIError } from "../sdk/models/errors/apierror.js";
import { ErrorUnionPostRequestBody } from "../sdk/models/operations/errorunionpost.js";
import { ErrorUnionDiscriminatedPostRequestBody } from "../sdk/models/operations/erroruniondiscriminatedpost.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Test Status Get Error Default Error Codes", async () => {
  recordTest("errors-status-get-error-default-error-codes");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // When `clientServerStatusCodesAsErrors: true` is set (default), 4XX and 5XX ranges
  // are automatically treated as errors.

  // 400 and 500 responses are explicitly defined
  await sdk.errors.statusGetError(400).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 400 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );

  await sdk.errors.statusGetError(500).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 500 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );

  // 404 and 503 responses are undefined but still treated as errors by default
  await sdk.errors.statusGetError(404).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 404 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );

  await sdk.errors.statusGetError(503).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 503 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );
});

test("Test Status Get Error 300 Non Error", async () => {
  recordTest("errors-status-get-error300-non-error");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.errors.statusGetError(300);
  expect(res.httpMeta.response.status).toBe(300);
});

test("Test Status Get Error X-Speakeasy-Errors", async () => {
  recordTest("errors-status-get-error-x-speakeasy-errors");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // 400 response is explicitly defined and is marked as an error in `x-speakeasy-errors`
  await sdk.errors.statusGetXSpeakeasyErrors(400).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 400);
      expect(`${err}`).toEqual(
        `APIError: API error occurred: Status 400. Body: {"message":"an error occurred","code":"400","type":"internal"}`,
      );
    },
  );

  // 401 response is undefined but it is marked as an error in `x-speakeasy-errors`
  await sdk.errors.statusGetXSpeakeasyErrors(401).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 401);
      expect(`${err}`).toEqual(
        `APIError: API error occurred: Status 401. Body: {"message":"an error occurred","code":"401","type":"internal"}`,
      );
    },
  );

  // 402 response is undefined and is not treated as an API error since it's not listed in `x-speakeasy-errors`.
  // Instead we raise a "Unexpected Status" exception.

  await sdk.errors.statusGetXSpeakeasyErrors(402).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 402);
      expect(`${err}`).toEqual(
        `APIError: Unexpected Status or Content-Type: Status 402. Body: {"message":"an error occurred","code":"402","type":"internal"}`,
      );
    },
  );

  // Both 500 and 501 responses are marked as errors since `5XX` is listed `x-speakeasy-errors`.
  // In no-zod the matcher schemas are passthroughs, so the rejected value is a
  // plain object carrying the wire-shape fields rather than an `ErrorT`/typed
  // class instance. We assert on the wire-shape instead.
  await sdk.errors.statusGetXSpeakeasyErrors(500).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toHaveProperty("code", "500");
      expect(err).toHaveProperty("message", "an error occurred");
    },
  );

  await sdk.errors.statusGetXSpeakeasyErrors(501).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toHaveProperty("code", "501");
      expect(err).toHaveProperty("message", "an error occurred");
    },
  );
});

test("Test Status Get Success X-Speakeasy-Errors Empty Status Code List", async () => {
  recordTest("errors-status-get-error-x-speakeasy-errors-none");

  const sdk = new SDK();

  // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
  // with a dummy list, meaning all responses are treated as non-errors.

  const res200 = await sdk.errors.statusGetNonError(200);
  expect(res200.httpMeta.response.status).toBe(200);

  const res400 = await sdk.errors.statusGetNonError(400);
  expect(res400.httpMeta.response.status).toBe(400);

  const res500 = await sdk.errors.statusGetNonError(500);
  expect(res500.httpMeta.response.status).toBe(500);
});

test("Test Status Get Success X-Speakeasy-Errors Unspecified Responses", async () => {
  recordTest("errors-status-get-error-x-speakeasy-errors-default");

  const sdk = new SDK();

  // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
  // by marking all *unspecified* responses as errors.

  // 200 and 400 responses are explicitly defined, so they are treated as non-errors.
  const res200 = await sdk.errors.statusGetDefaultError(200);
  expect(res200.httpMeta.response.status).toBe(200);

  const res400 = await sdk.errors.statusGetDefaultError(400);
  expect(res400.httpMeta.response.status).toBe(400);

  // 404 and 500 responses are undefined, so they are treated as errors.
  await sdk.errors.statusGetDefaultError(404).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 404);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 404 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );

  await sdk.errors.statusGetDefaultError(500).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(APIError);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 500);
      expect(`${err}`).toBe(
        `APIError: API error occurred: Status 500 Content-Type "text/html; charset=utf-8". Body: ""`,
      );
    },
  );

  // The TestHook (ported from primary's tests/.hooks override) installs an
  // AfterError handler that rewrites a 418 from this operation into a 200.
  // Confirm the recovery actually fires under no-zod too.
  const res418 = await sdk.errors.statusGetDefaultError(418);
  expect(res418.httpMeta.response.status).toBe(200);
});

test("Test Errors Connection Error", async () => {
  recordTest("errors-connection-error");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  await expect(
    sdk.errors.connectionErrorGet({ timeoutMs: 1000 }),
  ).rejects.toThrow(/(fetch failed|operation was aborted due to timeout)/);
});

test("Test Union Of Errors", async () => {
  recordTest("errors-union-of-errors");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod, error union types are not wrapped in dedicated `Error*` classes;
  // the rejected value is a plain object containing httpMeta and the wire fields.
  const req1: ErrorUnionPostRequestBody = { error: "Error1" };
  await sdk.errors.errorUnionPost(req1).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 500);
      expect(err).toHaveProperty(["error"], "Error1");
    },
  );

  const req2: ErrorUnionPostRequestBody = {
    errorType2Message: { message: "Error2" },
  };
  await sdk.errors.errorUnionPost(req2).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 500);
      expect(err).toHaveProperty(["errorType2Message", "message"], "Error2");
    },
  );
});

test("Test Discriminated Union Of Errors", async () => {
  recordTest("errors-union-of-errors-discriminated");

  // For discriminated unions, zod typings are altered to explicitly
  // list the discriminator key as having a literal string type.
  // The errors thrown are thus expected to be of type `Object`, e.g.:
  //   TaggedError1$inboundSchema.and(
  //     z.object({ tag: z.literal("tag1") }).transform((v) => ({ tag: v.tag })),
  //   ),

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // In no-zod, discriminated union errors are plain wire-shape objects with
  // the discriminator field still set, but without a class-name wrapper.
  const req1: ErrorUnionDiscriminatedPostRequestBody = {
    tag: "tag1",
    error: "Error1",
  };
  await sdk.errors.errorUnionDiscriminatedPost(req1).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(Object);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 400);
      expect(err).toHaveProperty(["tag"], "tag1");
      expect(err).toHaveProperty(["error"], "Error1");
    },
  );

  const req2: ErrorUnionDiscriminatedPostRequestBody = {
    tag: "tag2",
    taggedError2Message: { message: "Error2" },
  };
  await sdk.errors.errorUnionDiscriminatedPost(req2).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(Object);
      expect(err).toHaveProperty(["httpMeta", "response", "status"], 400);
      expect(err).toHaveProperty(["tag"], "tag2");
      expect(err).toHaveProperty(["taggedError2Message", "message"], "Error2");
    },
  );
});
