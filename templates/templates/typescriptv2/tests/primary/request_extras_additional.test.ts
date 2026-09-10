import { HTTPBIN_URL, recordTest } from "./common_helpers.js";
import { expect, test } from "vitest";

import { ClientSDK } from "../lib/sdks.js";
import { HTTPClient } from "../lib/http.js";
import { SDK } from "../index.js";

function createCapturingClient(): {
  httpClient: HTTPClient;
  captured: { body?: string; contentType?: string };
} {
  const captured: { body?: string; contentType?: string } = {};
  const httpClient = new HTTPClient({
    fetcher: async (request) => {
      captured.body = await request.text();
      captured.contentType = request.headers.get("content-type") ?? "";
      return new Response(JSON.stringify({ status: "OK" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    },
  });
  return { httpClient, captured };
}

test("Request Extras Extra Query Appended", async () => {
  recordTest("request-extras-extra-query-appended");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.methods.methodGet({
    extraQuery: {
      strParam: "extra",
      intParam: 1,
      boolParam: true,
    },
  });

  expect(res.httpMeta.response.status).toBe(200);

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.get("strParam")).toBe("extra");
  expect(url.searchParams.get("intParam")).toBe("1");
  expect(url.searchParams.get("boolParam")).toBe("true");
});

test("Request Extras Extra Query Merged With Operation Params", async () => {
  recordTest("request-extras-extra-query-merged");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.parameters.formQueryParamsPrimitive(
    true,
    1,
    1.1,
    "test",
    {
      extraQuery: {
        extraParam: "extra-value",
      },
    },
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res?.args).toMatchObject({
    strParam: "test",
    boolParam: "true",
    intParam: "1",
    numParam: "1.1",
  });

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.get("strParam")).toBe("test");
  expect(url.searchParams.get("extraParam")).toBe("extra-value");
});

test("Request Extras Extra Query Skips Nullish And Explodes Arrays", async () => {
  recordTest("request-extras-extra-query-nullish-arrays");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.methods.methodGet({
    extraQuery: {
      skipped: null,
      alsoSkipped: undefined,
      arrParam: ["one", "two"],
    },
  });

  expect(res.httpMeta.response.status).toBe(200);

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.has("skipped")).toBe(false);
  expect(url.searchParams.has("alsoSkipped")).toBe(false);
  expect(url.searchParams.getAll("arrParam")).toEqual(["one", "two"]);
});

test("Request Extras Extra Query Overrides Operation Params", async () => {
  recordTest("request-extras-extra-query-overrides");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.parameters.formQueryParamsPrimitive(
    true,
    1,
    1.1,
    "test",
    {
      extraQuery: {
        strParam: "overridden",
      },
    },
  );

  expect(res.httpMeta.response.status).toBe(200);

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.getAll("strParam")).toEqual(["overridden"]);
  expect(url.searchParams.get("intParam")).toBe("1");
});

test("Request Extras Extra Query Preserves Server URL Query", async () => {
  recordTest("request-extras-extra-query-preserves-server-url-query");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.parameters.formQueryParamsPrimitive(
    true,
    1,
    1.1,
    "test",
    {
      serverURL: `${HTTPBIN_URL}?api-version=1`,
      extraQuery: {
        foo: "bar",
      },
    },
  );

  expect(res.httpMeta.response.status).toBe(200);

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.get("api-version")).toBe("1");
  expect(url.searchParams.get("strParam")).toBe("test");
  expect(url.searchParams.get("foo")).toBe("bar");
});

test("Request Extras Extra Query Does Not Override Security Params", () => {
  recordTest("request-extras-extra-query-security-overrides");

  const s = new ClientSDK({ serverURL: HTTPBIN_URL });
  const request = s._createRequest(
    {
      baseURL: HTTPBIN_URL,
      operationID: "requestExtrasSecurityPrecedence",
      options: {},
      oAuth2Scopes: [],
      resolvedSecurity: null,
      retryConfig: { strategy: "none" },
    },
    {
      method: "GET",
      path: "/method/get",
      security: {
        basic: {},
        headers: {},
        queryParams: {
          api_key: "secure-value",
        },
        cookies: {},
        oauth2: { type: "none" },
      },
    },
    {
      extraQuery: {
        api_key: "extra-value",
      },
    },
  );

  expect(request.ok).toBe(true);
  if (!request.ok) {
    throw request.error;
  }

  const url = new URL(request.value.url);
  expect(url.searchParams.getAll("api_key")).toEqual(["secure-value"]);
});

test("Request Extras Extra Body Merged Into JSON Body", async () => {
  recordTest("request-extras-extra-body-merged");

  const { httpClient, captured } = createCapturingClient();
  const s = new SDK({ httpClient });

  await s.methods.methodPost(
    { id: "test123" },
    {
      extraBody: {
        extraStr: "extra-value",
        extraObj: { nested: 1 },
        extraNull: null,
        skipped: undefined,
      },
    },
  );

  expect(captured.contentType).toContain("application/json");
  expect(JSON.parse(captured.body ?? "{}")).toEqual({
    id: "test123",
    extraStr: "extra-value",
    extraObj: { nested: 1 },
    extraNull: null,
  });
});

test("Request Extras Extra Body Overrides Operation Body Fields", async () => {
  recordTest("request-extras-extra-body-overrides");

  const { httpClient, captured } = createCapturingClient();
  const s = new SDK({ httpClient });

  await s.methods.methodPost(
    { id: "test123" },
    {
      extraBody: {
        id: "overridden",
      },
    },
  );

  expect(JSON.parse(captured.body ?? "{}")).toEqual({ id: "overridden" });

  // An undefined entry is skipped entirely: it must not delete the
  // operation-produced field of the same name.
  await s.methods.methodPost(
    { id: "test123" },
    {
      extraBody: {
        id: undefined,
        extraStr: "extra-value",
      },
    },
  );

  expect(JSON.parse(captured.body ?? "{}")).toEqual({
    id: "test123",
    extraStr: "extra-value",
  });
});

test("Request Extras Extra Body Rejects Non JSON Object Bodies", async () => {
  recordTest("request-extras-extra-body-rejects-non-json");

  const { httpClient } = createCapturingClient();
  const s = new SDK({ httpClient, serverURL: HTTPBIN_URL });

  await expect(
    s.methods.methodGet({
      // @ts-expect-error operations without a JSON request body do not accept extraBody
      extraBody: { extraStr: "extra-value" },
    }),
  ).rejects.toThrow(/JSON object request bodies/);
});

test("Request Extras Extra Query JSON Encodes Objects", async () => {
  recordTest("request-extras-extra-query-object-json");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.methods.methodGet({
    extraQuery: {
      objParam: { str: "val", int: 1 },
    },
  });

  expect(res.httpMeta.response.status).toBe(200);

  const url = new URL(res.httpMeta.request.url);
  expect(url.searchParams.get("objParam")).toBe('{"str":"val","int":1}');
});
