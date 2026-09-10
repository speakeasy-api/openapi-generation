import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";
import { ResultObject } from "../sdk/models/operations/paginationlimitoffsetunionoutputpageparams.js";

test("Test Pagination LimitOffset Page Params Flat", async () => {
  recordTest("pagination-limit-offset-page-params-flat");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res = await sdk.pagination.paginationLimitOffsetPageParams(1);

  expect(res.result).toBeDefined();
  expect(res.result!.resultArray).toBeDefined();
  expect(res.result!.resultArray!.length).toBe(serverLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  expect(nextRes!.result).toBeDefined();
  expect(nextRes!.result!.resultArray).toBeDefined();
  expect(nextRes!.result!.resultArray.length).toBe(0);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});

test("Test Pagination LimitOffset Union Output Page Params Flat", async () => {
  recordTest("pagination-limit-offset-union-output-page-params-flat");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const serverLimit = 20;

  const res =
    await sdk.pagination.paginationLimitOffsetUnionOutputPageParams(1);

  const result = res.result as ResultObject;
  expect(result).toBeDefined();
  expect(result!.resultArray.length).toBe(serverLimit);

  const nextRes = await res.next();
  expect(nextRes).not.toBeNull();
  const nextResult = nextRes!.result as ResultObject;
  expect(nextResult).toBeDefined();
  expect(nextResult!.resultArray.length).toBe(0);

  const nullRes = await nextRes!.next();
  expect(nullRes).toBeNull();
});
