import { expect, test, vi } from "vitest";

import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";
import { HTTPClient } from "../lib/http.js";
import { ResultObject } from "../sdk/models/operations/paginationlimitoffsetunionoutputpageparams.js";

test("Test Pagination LimitOffset Page Params", async () => {
  recordTest("pagination-limit-offset-page-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res = await sdk.pagination.paginationLimitOffsetPageParams(1);
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(serverLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(0);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Union Output Page Params", async () => {
  recordTest("pagination-limit-offset-union-output-page-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res =
    await sdk.pagination.paginationLimitOffsetUnionOutputPageParams(1);
  expect(res.httpMeta.response.status).toBe(200);
  const result = res.res as ResultObject;
  expect(result).toBeDefined();
  expect(result!.resultArray.length).toBe(serverLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  const nextResult = nextRes!.res as ResultObject;
  expect(nextResult).toBeDefined();
  expect(nextResult!.resultArray.length).toBe(0);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Nil Page Params", async () => {
  recordTest("pagination-limit-offset-nil-page-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res = await sdk.pagination.paginationLimitOffsetOptionalPageParams();
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(serverLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(0);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Zero Page Params", async () => {
  recordTest("pagination-limit-offset-zero-page-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res = await sdk.pagination.paginationLimitOffsetOptionalPageParams(0);
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(serverLimit);

  let reqURL = new URL(res.httpMeta.request.url);
  expect(reqURL.searchParams.get("page")).toEqual("0");

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20);

  reqURL = new URL(nextRes?.httpMeta.request.url || "");
  expect(reqURL.searchParams.get("page")).toEqual("1");
});

test("Test Pagination LimitOffset Page Body", async () => {
  recordTest("pagination-limit-offset-page-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationLimitOffsetPageBody({
    limit,
    page: 1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Page Body Nullable", async () => {
  recordTest("pagination-limit-offset-page-body-nullable");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  // first request sends a null body; later pages materialize one with the
  // advanced page (wire sequence enforced by the test service)
  const res = await sdk.pagination.paginationLimitOffsetPageBodyNullable(null);
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray).toEqual([0, 1, 2, 3, 4, 5, 6]);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray).toEqual([7, 8, 9, 10, 11, 12, 13]);

  const lastRes = await nextRes!.next();
  expect(lastRes).not.toBeNull();
  expect(lastRes!.httpMeta.response.status).toBe(200);
  expect(lastRes!.res).toBeDefined();
  expect(lastRes!.res!.resultArray).toEqual([14, 15, 16, 17, 18, 19]);

  const nullRes = await lastRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Deep Outputs Page Body", async () => {
  recordTest("pagination-limit-offset-deep-outputs-page-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationLimitOffsetDeepOutputsPageBody({
    limit,
    page: 1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Offset Params", async () => {
  recordTest("pagination-limit-offset-offset-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationLimitOffsetOffsetParams(limit, 0);
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Nil Offset Params", async () => {
  recordTest("pagination-limit-offset-nil-offset-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const defaultLimit = 20;

  const res = await sdk.pagination.paginationLimitOffsetOffsetParams();
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(defaultLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - defaultLimit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Offset Body", async () => {
  recordTest("pagination-limit-offset-offset-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationLimitOffsetOffsetBody({
    limit,
    offset: 0,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Default Offset Body", async () => {
  recordTest("pagination-limit-offset-default-offset-body");

  let capturedRequest: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", async (req) => {
    capturedRequest = req.clone();

    return req;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });
  const res = await sdk.pagination.paginationLimitOffsetDefaultOffsetBody({});

  if (capturedRequest == null) {
    expect.fail("Expected request to be captured");
  }

  const body = await capturedRequest.json();

  expect(res.httpMeta.response.status).toBe(200);
  expect(body).toHaveProperty("limit", 15);
  expect(body).toHaveProperty("offset", 10);
});

test("Test Pagination LimitOffset Default Offset Params", async () => {
  recordTest("pagination-limit-offset-default-offset-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const res = await sdk.pagination.paginationLimitOffsetDefaultOffsetParams();

  const reqURL = new URL(res.httpMeta.request.url);
  const limit = reqURL.searchParams.get("limit");
  const offset = reqURL.searchParams.get("offset");

  expect(res.httpMeta.response.status).toBe(200);
  expect(limit).toEqual("15");
  expect(offset).toEqual("10");
});

test("Test Pagination Cursor Params", async () => {
  recordTest("pagination-cursor-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationCursorParams(-1);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBeLessThan(limit);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination URL", async () => {
  recordTest("pagination-url");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.pagination.paginationURLParams(3);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res!.res!.resultArray.length).toBe(9);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(6);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(3);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();

  const res2 = await sdk.pagination.paginationURLParams(3, "true");
  expect(res2.httpMeta.response.status).toBe(200);
  expect(res2.res).toBeDefined();
  expect(res2!.res!.resultArray.length).toBe(9);

  const nextRes2 = await res2.next();
  expect(nextRes2).not.toBeNull();
  expect(nextRes2!.httpMeta.response.status).toBe(200);
  expect(nextRes2!.res).toBeDefined();
  expect(nextRes2!.res!.resultArray.length).toBe(6);

  const penultimateRes2 = await nextRes2!.next();
  expect(penultimateRes2).not.toBeNull();
  expect(penultimateRes2!.httpMeta.response.status).toBe(200);
  expect(penultimateRes2!.res).toBeDefined();
  expect(penultimateRes2!.res!.resultArray.length).toBe(3);

  const nullRe2 = await penultimateRes2!.next();
  expect(nullRe2).toBeNull();
});

test("Test Pagination Cursor Body", async () => {
  recordTest("pagination-cursor-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationCursorBody({
    cursor: -1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBeLessThan(limit);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Cursor Non-Numeric", async () => {
  recordTest("pagination-cursor-non-numeric");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationCursorNonNumeric("-1");

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(5);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Cursor Non-Numeric Nullable", async () => {
  recordTest("pagination-cursor-non-numeric-nullable");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationCursorNonNumericNullable("2");

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);
  expect(res!.res!.cursor).toBe("17");

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(2);
  expect(nextRes!.res!.cursor).toBeNull();

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Cursor Non-Numeric Empty String", async () => {
  recordTest("pagination-cursor-non-numeric-empty-string");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.pagination.paginationCursorNonNumericEmptyString(
    "",
    "2",
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(15);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(2);
  expect(nextRes!.res!.cursor).toBe("");

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination With Retries", async () => {
  recordTest("pagination-with-retries");

  const reqID = `${Math.random()}`;

  const spy = vi.fn<(status: number, method: string, path: string) => void>();

  const httpClient = new HTTPClient();
  httpClient.addHook("beforeRequest", (req) => {
    req.headers.set("request-id", reqID);
  });
  httpClient.addHook("response", (res, req) => {
    spy(res.status, req.method, new URL(req.url).pathname);
  });

  const s = new SDK({ serverURL: HTTPBIN_URL, httpClient });

  const result = await s.pagination.paginationWithRetries();
  let count = 0;
  for await (const page of result) {
    count += page.res?.resultArray.length ?? 0;
  }

  expect(count).toBe(20);
  expect(spy).toHaveBeenCalledTimes(6);
  expect(spy.mock.calls).toEqual([
    ...Array(3).fill([503, "GET", "/pagination/cursor_non_numeric"]),
    ...Array(3).fill([200, "GET", "/pagination/cursor_non_numeric"]),
  ]);
});

test("Test Pagination Body Wrapped Request", async () => {
  recordTest("pagination-body-wrapped-request");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationBodyWrappedRequest({
    limitOffsetConfig: {
      limit,
      page: 1,
    },
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Params Wrapped Request", async () => {
  recordTest("pagination-params-wrapped-request");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const available = 20;
  const limit = 15;
  const offset = 1;

  const res = await sdk.pagination.paginationParamsWrappedRequest({
    limit,
    offset,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(available - offset - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Body Flattened With Security", async () => {
  recordTest("pagination-body-flattened-with-security");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationBodyFlattenedWithSecurity(
    { paginationAuth: "test" },
    limit,
    0,
  );
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Body Flattened Optional Security", async () => {
  recordTest("pagination-body-flattened-optional-security");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationBodyFlattenedOptionalSecurity(
    limit,
    0,
    { paginationAuth: "test" },
  );
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBe(20 - limit);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Ambiguous Input", async () => {
  recordTest("pagination-ambiguous-input");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationAmbiguousInput({
    cursor: -1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBeLessThan(limit);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Wrapped Optional Body", async () => {
  recordTest("pagination-wrapped-optional-body");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.pagination.paginationWrappedOptionalBody();
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(20);

  const penultimateRes = await res!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("supports async iterators", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const iterator = await sdk.pagination.paginationLimitOffsetOffsetParams(3);
  expect(iterator.httpMeta.response.status).toBe(200);

  const collect: number[] = [];

  for await (const page of iterator) {
    if (!page.res) {
      expect.unreachable("expected to receive a page containing results");
    }
    collect.push(...page.res.resultArray);
  }

  expect(collect.length).toBe(serverLimit);
  expect(collect).toEqual(Array.from({ length: serverLimit }, (_, i) => i));
});

test("Test Pagination Cursor Nullable", async () => {
  recordTest("pagination-cursor-nullable-limit");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const defaultLimit = 10;

  const res = await sdk.pagination.paginationCursorNullableLimit();
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.paginationCursorNullableLimitNextCursor).toBeDefined();
  expect(res.paginationCursorNullableLimitNextCursor!.results.length).toBe(
    defaultLimit,
  );

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.paginationCursorNullableLimitNextCursor).toBeDefined();
  expect(nextRes!.paginationCursorNullableLimitNextCursor!.results.length).toBe(
    defaultLimit,
  );

  // Paginate until exhausted
  let lastRes = nextRes!;
  while (true) {
    const n = await lastRes.next();
    if (n === null) break;
    lastRes = n;
  }

  expect(lastRes.paginationCursorNullableLimitNextCursor).toBeDefined();
});

test("Test Pagination Encapsulated Parameter", async () => {
  recordTest("pagination-encapsulated-parameter");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const limit = 15;

  const res = await sdk.pagination.paginationEncapsulatedParameter({
    cursor: -1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.resultArray.length).toBe(limit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.httpMeta.response.status).toBe(200);
  expect(nextRes!.res).toBeDefined();
  expect(nextRes!.res!.resultArray.length).toBeLessThan(limit);

  const penultimateRes = await nextRes!.next();
  expect(penultimateRes).not.toBeNull();
  expect(penultimateRes!.httpMeta.response.status).toBe(200);
  expect(penultimateRes!.res).toBeDefined();
  expect(penultimateRes!.res!.resultArray.length).toBe(0);

  const nullRes = await penultimateRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination Cursor Deep Nested Outputs", async () => {
  recordTest("pagination-cursor-deep-nested-outputs");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const defaultLimit = 5;
  const totalItems = 20;

  // Walk every page for each final-page shape: the optional meta object
  // absent entirely, the nullable paging object explicitly null, and the
  // nullable next_cursor leaf explicitly null. In every case the inlined
  // next-cursor access must resolve to a nullish value and end pagination.
  for (const style of ["omit-meta", "null-paging", "null-cursor"] as const) {
    const res = await sdk.pagination.paginationCursorDeepNestedOutputs(
      undefined,
      style,
    );
    expect(res.httpMeta.response.status).toBe(200);
    expect(res.paginationCursorDeepNestedOutputsResult).toBeDefined();
    expect(res.paginationCursorDeepNestedOutputsResult!.data.items.length).toBe(
      defaultLimit,
    );

    const collected = [
      ...res.paginationCursorDeepNestedOutputsResult!.data.items,
    ];
    let page = res;
    while (true) {
      const next = await page.next();
      if (next === null) break;
      expect(next.httpMeta.response.status).toBe(200);
      expect(next.paginationCursorDeepNestedOutputsResult).toBeDefined();
      collected.push(
        ...next.paginationCursorDeepNestedOutputsResult!.data.items,
      );
      page = next;
    }

    expect(collected, `final_page_style=${style}`).toEqual(
      Array.from({ length: totalItems }, (_, i) => `item_${i}`),
    );
  }
});

test("Test Pagination Cursor Deep Nested Outputs Iterator", async () => {
  recordTest("pagination-cursor-deep-nested-outputs-iterator");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const totalItems = 20;

  const collected: string[] = [];
  const iterator = await sdk.pagination.paginationCursorDeepNestedOutputs(
    undefined,
    "null-paging",
  );
  for await (const page of iterator) {
    if (!page.paginationCursorDeepNestedOutputsResult) {
      expect.unreachable("expected to receive a page containing results");
    }
    collected.push(...page.paginationCursorDeepNestedOutputsResult.data.items);
  }

  expect(collected).toEqual(
    Array.from({ length: totalItems }, (_, i) => `item_${i}`),
  );
});
