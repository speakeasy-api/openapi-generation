// In the no-zod variant `smartUnion` doesn't exist — unions are resolved
// left-to-right by declaration order, so the populated-fields heuristic
// tests from the primary suite can't be ported verbatim. We keep the SDK
// integration tests and remove the pure smartUnion unit tests that rely on
// the zod-based helper.

import { expect, test } from "vitest";

import { SDK } from "../index.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("smartUnion - discriminator with additional field and unrecognized enum", async () => {
  recordTest("smart-union-open-enums-and-size");
  // types: { kind: OpenEnum["cat"], name: string } | { kind: OpenEnum["dog"] }
  // payload: { kind: "bat", name: "asdf" }

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req = { kind: "bat", name: "asdf" } as any;

  const result = await sdk.unions.smartUnionOpenEnumsAndSize(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - all consts", async () => {
  recordTest("smart-union-all-consts");

  const sdk = new SDK();
  const req = { b: "B", c: "C" } as any;

  const result = await sdk.unions.smartUnionAllConsts(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - any field type", async () => {
  recordTest("smart-union-any-field-type");

  const sdk = new SDK();
  const req = { b: "asdf" } as any;

  const result = await sdk.unions.smartUnionAnyFieldType(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - nested union vs flat struct", async () => {
  recordTest("smart-union-nested-union-vs-flat-struct");

  const sdk = new SDK();
  const req = { data: { x: "", y: "" } } as any;

  const result = await sdk.unions.smartUnionNestedUnionVsFlatStruct(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - union vs union", async () => {
  recordTest("smart-union-union-vs-union");

  const sdk = new SDK();
  const req = { a: "", b: "", c: "" } as any;

  const result = await sdk.unions.smartUnionUnionVsUnion(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("unions - discriminated open enum", async () => {
  recordTest("unions-discriminated-open-enum");

  const sdk = new SDK();

  // active status -> Status1. In no-zod, datetimes are ISO strings on the wire.
  const reqActive = {
    status: "active",
    userId: "user-123",
    activeAt: "2024-01-15T10:30:00.000Z",
  } as any;

  const resActive = await sdk.unions.discriminatedOpenEnum(reqActive);
  expect(resActive.httpMeta.response.status).toEqual(200);
  expect(resActive.res?.json).toBeTruthy();
  const activeJson = resActive.res!.json as any;
  expect(activeJson.status).toEqual("active");
  expect(activeJson.userId).toEqual("user-123");

  // inactive status -> Status2
  const reqInactive = {
    status: "inactive",
    reason: "User requested deactivation",
    inactiveSince: "2024-01-10T15:45:00.000Z",
  } as any;

  const resInactive = await sdk.unions.discriminatedOpenEnum(reqInactive);
  expect(resInactive.httpMeta.response.status).toEqual(200);
  expect(resInactive.res?.json).toBeTruthy();
  const inactiveJson = resInactive.res!.json as any;
  expect(inactiveJson.status).toEqual("inactive");
  expect(inactiveJson.reason).toEqual("User requested deactivation");
});
