import { vi, expect, test, afterEach } from "vitest";

import { SDK } from "../index.js";
import { envSchema, resetEnv } from "../lib/env.js";
import { HTTPClient } from "../lib/http.js";
import { GlobalPathParameterGetResponse } from "../sdk/models/operations/globalpathparameterget.js";
import { GlobalsHeaderGetResponse } from "../sdk/models/operations/globalsheaderget.js";
import { GlobalsHiddenPostResponse } from "../sdk/models/operations/globalshiddenpost.js";
import { GlobalsQueryParameterGetResponse } from "../sdk/models/operations/globalsqueryparameterget.js";

import { recordTest, HTTPBIN_URL } from "./common_helpers.js";

afterEach(() => {
  resetEnv();
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
});

test("Test Globals Query Parameter Get Uses Global", async () => {
  recordTest("globals-query-parameter-get-uses-global");

  vi.stubEnv("SPEAKEASY_GLOBAL_QUERY_PARAM", "test");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res: GlobalsQueryParameterGetResponse =
    await sdk.globals.globalsQueryParameterGet();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.args.globalQueryParam).toBe("test");
});

test("Test Globals Query Parameter Get Uses Local", async () => {
  recordTest("globals-query-parameter-get-uses-local");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalQueryParam: "test",
  });

  const res: GlobalsQueryParameterGetResponse =
    await sdk.globals.globalsQueryParameterGet("local");

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.args.globalQueryParam).toBe("local");
});

test("Test Globals Path Parameter Get Uses Global", async () => {
  recordTest("globals-path-parameter-get-uses-global");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalPathParam: 1,
  });

  const res: GlobalPathParameterGetResponse =
    await sdk.globals.globalPathParameterGet();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.url).toBe(`${HTTPBIN_URL}/anything/globals/pathParameter/1`);
});

test("Test Globals Path Parameter Get Uses Local", async () => {
  recordTest("globals-path-parameter-get-uses-local");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalPathParam: 1,
  });

  const res: GlobalPathParameterGetResponse =
    await sdk.globals.globalPathParameterGet(2);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.url).toBe(`${HTTPBIN_URL}/anything/globals/pathParameter/2`);
});

test("Test Globals Header Get Uses Global", async () => {
  recordTest("globals-header-get-uses-global");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalHeaderParam: true,
  });

  const res: GlobalsHeaderGetResponse = await sdk.globals.globalsHeaderGet();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.headers?.["Globalheaderparam"]).toBe("true");
});

test("Test Globals Header Get Uses Local", async () => {
  recordTest("globals-header-get-uses-local");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalHeaderParam: true,
  });

  const res: GlobalsHeaderGetResponse =
    await sdk.globals.globalsHeaderGet(false);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.headers?.["Globalheaderparam"]).toBe("false");
});

test("Test Global Header Keeps Custom Client Headers", async () => {
  recordTest("globals-header-keeps-custom-client-headers");

  const httpClient = new HTTPClient({
    fetcher: (request) => {
      return fetch(request);
    },
  });

  httpClient.addHook("beforeRequest", (request) => {
    request.headers.set("x-custom-header", "customValue");

    return request;
  });

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalHeaderParam: true,
    httpClient: httpClient,
  });

  const res: GlobalsHeaderGetResponse = await sdk.globals.globalsHeaderGet();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.headers?.["Globalheaderparam"]).toBe("true");
  expect(res.res?.headers?.["X-Custom-Header"]).toBe("customValue");
});

test("Test Globals Hidden Post", async () => {
  recordTest("globals-hidden-post");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalHiddenQueryParam: "hello",
    globalHiddenHeaderParam: "world",
    globalHiddenPathParam: "test",
  });

  const res: GlobalsHiddenPostResponse = await sdk.globals.globalsHiddenPost({
    test: "friend",
    other: 37,
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.args.globalHiddenQueryParam).toBe("hello");
  expect(res.res?.json?.test).toBe("friend");
  expect(res.res?.json?.other).toBe(37);
  expect(res.res?.headers?.["Globalhiddenheaderparam"]).toBe("world");
  expect(res.res?.url).toBe(
    `${HTTPBIN_URL}/anything/globals/hidden/test?globalHiddenQueryParam=hello`,
  );
});

test("Test Globals Operation Params Only", async () => {
  recordTest("globals-operation-params-only");

  // Initialize SDK with ALL global parameters
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    globalQueryParam: "globalQueryValue",
    globalPathParam: 999,
    globalHeaderParam: true,
    globalHiddenQueryParam: "hiddenQueryValue",
    globalHiddenHeaderParam: "hiddenHeaderValue",
    globalHiddenPathParam: "hiddenPathValue",
  });

  // Call operation with operation-specific parameters
  const res = await sdk.globals.globalsOperationScopedExclusive(
    "operationHeader",
    "operationPath",
    "operationQuery",
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();

  // Verify that ONLY operation-specific parameters are present in the request
  // Query params should NOT contain global params
  expect(res.res?.args["globalQueryParam"]).toBeUndefined();
  expect(res.res?.args["globalHiddenQueryParam"]).toBeUndefined();

  // URL should contain operation path param, NOT global path param
  expect(res.res?.url).toBe(
    `${HTTPBIN_URL}/anything/globals/operationScopedExclusive/operationPath?operationQueryParam=operationQuery`,
  );

  // Headers should NOT contain global headers
  expect(res.res?.headers?.["Globalheaderparam"]).toBeUndefined();
  expect(res.res?.headers?.["Globalhiddenheaderparam"]).toBeUndefined();
});

test("Test Globals Kebab Case Param Get", async () => {
  recordTest("globals-kebab-case-param-get");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    kebabCaseParam: "kebab-case-value",
  });

  const res = await sdk.globals.globalsKebabCaseParamGet();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.args["kebab-case-param"]).toBe("kebab-case-value");
});

test("Test Debug Env Var False Keeps Logging Off", async () => {
  recordTest("debug-env-var-false-keeps-logging-off");

  const group = vi.spyOn(console, "group").mockImplementation(() => {});
  vi.spyOn(console, "log").mockImplementation(() => {});
  vi.spyOn(console, "groupEnd").mockImplementation(() => {});

  for (const value of ["false", "0", "off", "no", "", "debug", "FALSE"]) {
    resetEnv();
    vi.stubEnv("SPEAKEASY_DEBUG", value);

    const sdk = new SDK({ serverURL: HTTPBIN_URL });
    const res = await sdk.globals.globalsQueryParameterGet("local");

    expect(res.httpMeta.response.status).toBe(200);
    expect(group).not.toHaveBeenCalled();
  }

  for (const value of ["true", "1", "yes", "on", "TRUE", "Yes", " on "]) {
    group.mockClear();
    resetEnv();
    vi.stubEnv("SPEAKEASY_DEBUG", value);

    const sdk = new SDK({ serverURL: HTTPBIN_URL });
    const res = await sdk.globals.globalsQueryParameterGet("local");

    expect(res.httpMeta.response.status).toBe(200);
    expect(group).toHaveBeenCalled();
  }

  for (const value of [true, 1]) {
    expect(envSchema.parse({ SPEAKEASY_DEBUG: value }).SPEAKEASY_DEBUG).toBe(
      true,
    );
  }

  for (const value of [false, 0]) {
    expect(envSchema.parse({ SPEAKEASY_DEBUG: value }).SPEAKEASY_DEBUG).toBe(
      false,
    );
  }
});
