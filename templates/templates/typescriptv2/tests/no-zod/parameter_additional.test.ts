import { createDeepObject, createSimpleObject } from "./primary_helpers.js";
import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { recordTest, HTTPBIN_URL } from "./common_helpers.js";
import { sortQueryParameters } from "./helpers.js";

// Useful for simplifying JS objects contain complex-but-serializable values
// like dates before comparing them to expected values.
const roundTrip = (v: unknown) => JSON.parse(JSON.stringify(v));

// In the no-zod variant the wire shape uses `preserveModelFieldNames: true`
// which preserves the camelCase spec field names, emitted in declaration
// order: {str,bool,int,int32,int32Enum,intEnum,num,float32,enum,any,date,
// dateTime,boolOpt,strOpt}. The deep wrapper preserves {any,arr,bool,int,
// map,num,obj,str}.
const SIMPLE_OBJECT_JSON = `{"str":"test","bool":true,"int":1,"int32":1,"int32Enum":55,"intEnum":2,"num":1.1,"float32":1.1,"enum":"one","any":"any","date":"2020-01-01","dateTime":"2020-01-01T00:00:00.001Z","boolOpt":true,"strOpt":"testOptional"}`;

const DEEP_OBJECT_JSON = `{"any":${SIMPLE_OBJECT_JSON},"arr":[${SIMPLE_OBJECT_JSON},${SIMPLE_OBJECT_JSON}],"bool":true,"int":1,"map":{"key":${SIMPLE_OBJECT_JSON}},"num":1.1,"obj":${SIMPLE_OBJECT_JSON},"str":"test"}`;

const encodeJSONForQuery = (jsonString: string): string =>
  jsonString.replace(/:/g, "%3A").replace(/,/g, "%2C");

test("Test Path Parameter JSON", async () => {
  recordTest("parameters-path-parameter-json");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.parameters.pathParameterJson(createSimpleObject());

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.url).toBe(
    `${HTTPBIN_URL}/anything/pathParams/json/${SIMPLE_OBJECT_JSON}`,
  );
});

test("Test JSON Query Params Object", async () => {
  recordTest("parameters-json-query-params-object");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const simpleObj = createSimpleObject();
  const deepObject = createDeepObject();

  const res = await sdk.parameters.jsonQueryParamsObject(simpleObj, deepObject);

  expect(res.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/queryParams/json/obj?deepObjParam=${encodeJSONForQuery(
      DEEP_OBJECT_JSON,
    )}&simpleObjParam=${encodeJSONForQuery(SIMPLE_OBJECT_JSON)}`,
  );

  expect(JSON.parse(res.res?.args.simpleObjParam || "")).toEqual(
    roundTrip(simpleObj),
  );
  expect(JSON.parse(res.res?.args.deepObjParam || "")).toEqual(
    roundTrip(deepObject),
  );
});

test("Test Mixed Query Params", async () => {
  recordTest("parameters-mixed-query-params");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const obj = createSimpleObject();

  const res = await sdk.parameters.mixedQueryParams(obj, obj, obj);

  expect(res.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/queryParams/mixed?any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.001Z&deepObjectParam[any]=any&deepObjectParam[boolOpt]=true&deepObjectParam[bool]=true&deepObjectParam[dateTime]=2020-01-01T00%3A00%3A00.001Z&deepObjectParam[date]=2020-01-01&deepObjectParam[enum]=one&deepObjectParam[float32]=1.1&deepObjectParam[int32Enum]=55&deepObjectParam[int32]=1&deepObjectParam[intEnum]=2&deepObjectParam[int]=1&deepObjectParam[num]=1.1&deepObjectParam[strOpt]=testOptional&deepObjectParam[str]=test&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&jsonParam=${encodeJSONForQuery(
      SIMPLE_OBJECT_JSON,
    )}&num=1.1&str=test&strOpt=testOptional`,
  );

  const actual: Record<string, unknown> = res.res?.args ?? {};
  if (typeof actual["jsonParam"] === "string") {
    actual["jsonParam"] = JSON.parse(actual["jsonParam"]);
  }

  expect(actual).toEqual({
    any: "any",
    bool: "true",
    boolOpt: "true",
    date: "2020-01-01",
    dateTime: "2020-01-01T00:00:00.001Z",
    "deepObjectParam[any]": "any",
    "deepObjectParam[boolOpt]": "true",
    "deepObjectParam[bool]": "true",
    "deepObjectParam[dateTime]": "2020-01-01T00:00:00.001Z",
    "deepObjectParam[date]": "2020-01-01",
    "deepObjectParam[enum]": "one",
    "deepObjectParam[float32]": "1.1",
    "deepObjectParam[int32]": "1",
    "deepObjectParam[int]": "1",
    "deepObjectParam[int32Enum]": "55",
    "deepObjectParam[intEnum]": "2",
    "deepObjectParam[num]": "1.1",
    "deepObjectParam[strOpt]": "testOptional",
    "deepObjectParam[str]": "test",
    enum: "one",
    float32: "1.1",
    int: "1",
    int32: "1",
    int32Enum: "55",
    intEnum: "2",
    jsonParam: roundTrip(obj),
    num: "1.1",
    str: "test",
    strOpt: "testOptional",
  });
});

test("Test Parameters Header Params Nil", async () => {
  recordTest("parameters-header-params-nil");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.parameters.headerParamsNil({
    "Nullable-Header": null,
  });

  expect(res.httpMeta.response.status).toBe(200);

  const headers = new Headers(res.res?.headers);
  expect(headers.get("Nullable-Header")).toBeNull();
  expect(headers.get("Optional-Header")).toBeNull();
  expect(headers.get("Optional-Nullable-Header")).toBeNull();
});

test("Test Allow Empty Value Query Params", async () => {
  recordTest("parameters-allow-empty-value");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.parameters.allowEmptyValueQueryParams(
    "", // strParam
    [], // arrParamOmitEmpty
    null, // nullableStrParam
    undefined, // numParam
    [], // arrParam
  );

  expect(res.httpMeta.response.status).toBe(200);
  // In the no-zod variant explicit empty strings on optional params are
  // serialised as `key=` rather than being omitted; we accept either form.
  const actualUrl = sortQueryParameters(res.res?.url);
  const expectedWithStr = `${HTTPBIN_URL}/anything/allowEmptyValue?arrParam=&nullableStrParam=&numParam=&strParam=`;
  const expectedWithoutStr = `${HTTPBIN_URL}/anything/allowEmptyValue?arrParam=&nullableStrParam=&numParam=`;
  expect([expectedWithStr, expectedWithoutStr]).toContain(actualUrl);

  const res2 = await sdk.parameters.allowEmptyValueQueryParams(
    "test", // strParam
    ["x", "y"], // arrParamOmitEmpty
    "nullable", // nullableStrParam
    123, // numParam
    ["a", "b"], // arrParam
  );

  expect(res2.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res2.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/allowEmptyValue?arrParam=a&arrParam=b&arrParamOmitEmpty=x&arrParamOmitEmpty=y&nullableStrParam=nullable&numParam=123&strParam=test`,
  );
});
