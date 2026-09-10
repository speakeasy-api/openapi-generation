import { createDeepObject, createSimpleObject } from "./primary_helpers.js";
import { expect, test } from "vitest";

import { SDK } from "../index.js";
import {
  recordTest,
  HTTPBIN_URL,
  RequestLogEntry,
  newRequestRecorderClient,
} from "./common_helpers.js";
import { sortQueryParameters } from "./helpers.js";

// Useful for simplifying JS objects contain complex-but-serializable values
// like dates before comparing them to expected values.
const roundTrip = (v: unknown) => JSON.parse(JSON.stringify(v));

test("Test Path Parameter JSON", async () => {
  recordTest("parameters-path-parameter-json");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.parameters.pathParameterJson(createSimpleObject());

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.url).toBe(
    `${HTTPBIN_URL}/anything/pathParams/json/{"any":"any","bool":true,"boolOpt":true,"date":"2020-01-01","dateTime":"2020-01-01T00:00:00.001Z","enum":"one","float32":1.1,"int":1,"int32":1,"int32Enum":55,"intEnum":2,"num":1.1,"str":"test","strOpt":"testOptional"}`,
  );
});

test("Test JSON Query Params Object", async () => {
  recordTest("parameters-json-query-params-object");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const simpleObj = createSimpleObject();
  const deepObject = createDeepObject();

  const res = await sdk.parameters.jsonQueryParamsObject(deepObject, simpleObj);

  expect(res.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/queryParams/json/obj?deepObjParam={"any"%3A{"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}%2C"arr"%3A[{"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}%2C{"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}]%2C"bool"%3Atrue%2C"int"%3A1%2C"map"%3A{"key"%3A{"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}}%2C"num"%3A1.1%2C"obj"%3A{"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}%2C"str"%3A"test"}&simpleObjParam={"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}`,
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
    `${HTTPBIN_URL}/anything/queryParams/mixed?any=any&bool=true&boolOpt=true&date=2020-01-01&dateTime=2020-01-01T00%3A00%3A00.001Z&deepObjectParam[any]=any&deepObjectParam[boolOpt]=true&deepObjectParam[bool]=true&deepObjectParam[dateTime]=2020-01-01T00%3A00%3A00.001Z&deepObjectParam[date]=2020-01-01&deepObjectParam[enum]=one&deepObjectParam[float32]=1.1&deepObjectParam[int32Enum]=55&deepObjectParam[int32]=1&deepObjectParam[intEnum]=2&deepObjectParam[int]=1&deepObjectParam[num]=1.1&deepObjectParam[strOpt]=testOptional&deepObjectParam[str]=test&enum=one&float32=1.1&int=1&int32=1&int32Enum=55&intEnum=2&jsonParam={"any"%3A"any"%2C"bool"%3Atrue%2C"boolOpt"%3Atrue%2C"date"%3A"2020-01-01"%2C"dateTime"%3A"2020-01-01T00%3A00%3A00.001Z"%2C"enum"%3A"one"%2C"float32"%3A1.1%2C"int"%3A1%2C"int32"%3A1%2C"int32Enum"%3A55%2C"intEnum"%3A2%2C"num"%3A1.1%2C"str"%3A"test"%2C"strOpt"%3A"testOptional"}&num=1.1&str=test&strOpt=testOptional`,
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
    nullableHeader: null,
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
    [], // arrParam
    [], // arrParamOmitEmpty
    null, // nullableStrParam
    undefined, // numParam
    "", // strParam
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/allowEmptyValue?arrParam=&nullableStrParam=&numParam=&strParam=`,
  );

  const res2 = await sdk.parameters.allowEmptyValueQueryParams(
    ["a", "b"], // arrParam
    ["x", "y"], // arrParamOmitEmpty
    "nullable", // nullableStrParam
    123, // numParam
    "test", // strParam
  );

  expect(res2.httpMeta.response.status).toBe(200);
  expect(sortQueryParameters(res2.res?.url)).toBe(
    `${HTTPBIN_URL}/anything/allowEmptyValue?arrParam=a&arrParam=b&arrParamOmitEmpty=x&arrParamOmitEmpty=y&nullableStrParam=nullable&numParam=123&strParam=test`,
  );
});

const reservedChars = "abc 123:/?#[]@!$&'()*+,;=";

// Reserved characters with no structural meaning inside a path segment: they
// survive URL parsing, so the request reaches the server intact and the raw
// form can be compared against the percent-encoded one.
const nonStructuralReserved = ":@!$&'()*+,;=";

test("Test Path Encoding", async () => {
  recordTest("parameters-path-encoding");

  const log: RequestLogEntry[] = [];
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: newRequestRecorderClient(log),
  });

  const res = await sdk.parameters.pathEncoding(
    nonStructuralReserved,
    nonStructuralReserved,
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(log).toHaveLength(1);

  const segments = log[0]!.url.split("/anything/pathencoding/")[1]!.split("/");
  const [encoded, raw] = segments;
  // param1 percent-encodes every reserved character; param2 carries them raw.
  expect(decodeURIComponent(encoded!)).toBe(nonStructuralReserved);
  expect(encoded).not.toBe(raw);
  expect(raw).toBe(nonStructuralReserved);
});

test("Test Path Encoding Structural Characters", async () => {
  const log: RequestLogEntry[] = [];
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: newRequestRecorderClient(log),
  });

  // "?" and "#" keep their structural meaning when passed through raw, so the
  // value is truncated on the wire. Callers using allowReserved own that.
  await sdk.parameters
    .pathEncoding(reservedChars, reservedChars)
    .catch(() => {});

  expect(log).toHaveLength(1);
  expect(log[0]!.url).toBe(
    HTTPBIN_URL +
      "/anything/pathencoding" +
      "/abc%20123%3A%2F%3F%23%5B%5D%40!%24%26'()*%2B%2C%3B%3D" +
      "/abc%20123:/?",
  );
});

test("Test Query Encoding", async () => {
  recordTest("parameters-query-encoding");

  const log: RequestLogEntry[] = [];
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: newRequestRecorderClient(log),
  });

  const res = await sdk.parameters.queryEncoding(nonStructuralReserved);

  expect(res.httpMeta.response.status).toBe(200);
  expect(log).toHaveLength(1);
  // log[0].url is the normalized Fetch Request.url captured by the recorder,
  // which has apostrophes percent-encoded as %27 for query parameters.
  expect(log[0]!.url).toBe(
    HTTPBIN_URL +
      "/anything/queryencoding?param1=" +
      nonStructuralReserved.replace(/'/g, "%27"),
  );
});
