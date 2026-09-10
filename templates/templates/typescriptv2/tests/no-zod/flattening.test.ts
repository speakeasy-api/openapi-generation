import { expect, test } from "vitest";

import { SDK } from "../index.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";
import { createSimpleObject } from "./primary_helpers.js";

test("Test Component Body and Param No Conflict", async () => {
  recordTest("flattening-component-body-and-param-no-conflict");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const obj = createSimpleObject();
  const res = await s.flattening.componentBodyAndParamNoConflict(
    "param test",
    obj,
  );
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.args["paramStr"]).toEqual("param test");
  expect(res.res!.json).toEqual(obj);
});

test("Test Component Body and Param Conflict", async () => {
  recordTest("flattening-component-body-and-param-conflict");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const obj = createSimpleObject();
  const res = await s.flattening.componentBodyAndParamConflict(
    "param test",
    obj,
  );
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.args["str"]).toEqual("param test");
  expect(res.res!.json).toEqual(obj);
});

test("Test Inline Body and Param Conflict", async () => {
  recordTest("flattening-inline-body-and-param-conflict");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const res = await s.flattening.inlineBodyAndParamConflict("param test", {
    str: "body test",
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.args["str"]).toEqual("param test");
  expect(res.res!.json.str).toEqual("body test");
});

test("Test Inline Body and Param No Conflict", async () => {
  recordTest("flattening-inline-body-and-param-no-conflict");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const res = await s.flattening.inlineBodyAndParamNoConflict("param test", {
    bodyStr: "body test",
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.args["paramStr"]).toEqual("param test");
  expect(res.res!.json.bodyStr).toEqual("body test");
});

test("Test Conflicting Params", async () => {
  recordTest("flattening-conflicting-params");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const res = await s.flattening.conflictingParams("pathParam", "queryParam");
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.url).toContain("/pathParam?");
  expect(res.res!.args["str"]).toEqual("queryParam");
});

test("Test Required Body All Optional", async () => {
  recordTest("flattening-required-body-all-optional");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const res = await s.flattening.requiredBodyAllOptional({
    optStr: "body test",
    optInt: 1,
  });
  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res!.json.optStr).toEqual("body test");
  expect(res.res!.json.optInt).toEqual(1);
});
