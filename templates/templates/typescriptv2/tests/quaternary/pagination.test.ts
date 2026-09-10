import { expect, test } from "vitest";

import { HTTPClient } from "../lib/http.js";
import { SDK } from "../index.js";
import { createSimpleObject } from "./quaternary_helpers.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Test Pagination Cursor Snake", async () => {
  recordTest("pagination-cursor-snake");
  let req: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", (request) => {
    req = request.clone();
    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, http_client: httpClient });

  const result = await sdk.pagination.paginationCursorBody({
    cursor: 1,
  });

  expect(result.status_code).toBe(200);
  expect(result.res).toBeDefined();
  expect(result.res!.result_array).toBeDefined();

  if (!req) throw new Error("No request has been made");
  const body = JSON.parse(new TextDecoder().decode(await req.arrayBuffer()));
  expect(body.cursor).toBe(1);
});

test("Test Pagination Cursor Params Snake", async () => {
  recordTest("pagination-cursor-params-snake");
  let req: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", (request) => {
    req = request.clone();
    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, http_client: httpClient });

  const result = await sdk.pagination.paginationCursorParams(1);

  expect(result.status_code).toBe(200);
  expect(result.res).toBeDefined();
  expect(result.res!.result_array).toBeDefined();

  if (!req) throw new Error("No request has been made");
  expect(new URL(req.url).searchParams.get("cursor")).toBe("1");
});

test("Test Pagination Limit Offset Snake", async () => {
  recordTest("pagination-limit-offset-snake");
  let req: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", (request) => {
    req = request.clone();
    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, http_client: httpClient });

  const result = await sdk.pagination.paginationLimitOffsetOffsetBody({
    limit: 5,
    offset: 0,
  });

  expect(result.status_code).toBe(200);
  expect(result.res).toBeDefined();
  expect(result.res!.result_array).toBeDefined();

  if (!req) throw new Error("No request has been made");
  const body = JSON.parse(new TextDecoder().decode(await req.arrayBuffer()));
  expect(body.limit).toBe(5);
  expect(body.offset).toBe(0);
});

test("Test Pagination Limit Offset Params Snake", async () => {
  recordTest("pagination-limit-offset-params-snake");
  let req: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", (request) => {
    req = request.clone();
    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, http_client: httpClient });

  const result = await sdk.pagination.paginationLimitOffsetOffsetParams(5, 0);

  expect(result.status_code).toBe(200);
  expect(result.res).toBeDefined();
  expect(result.res!.result_array).toBeDefined();

  if (!req) throw new Error("No request has been made");
  const url = new URL(req.url);
  expect(url.searchParams.get("limit")).toBe("5");
  expect(url.searchParams.get("offset")).toBe("0");
});

test("Test Pagination with Simple Object Snake", async () => {
  recordTest("pagination-simple-object-snake");
  let req: Request | undefined;
  const httpClient = new HTTPClient().addHook("beforeRequest", (request) => {
    req = request.clone();
    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, http_client: httpClient });

  const simpleObj = createSimpleObject();

  const result = await sdk.pagination.paginationCursorBody({
    cursor: 1,
  });

  expect(result.status_code).toBe(200);
  expect(result.res).toBeDefined();
  expect(result.res!.result_array).toBeDefined();

  if (!req) throw new Error("No request has been made");
  const body = JSON.parse(new TextDecoder().decode(await req.arrayBuffer()));
  expect(body.cursor).toBe(1);
});
