import { expect, test } from "vitest";
import { ProxyAgent } from "undici";

import { SDK } from "../index.js";
import { HTTPClient } from "../lib/http.js";
import { CustomClientPostResponse } from "../sdk/models/operations/customclientpost.js";

import {
  recordTest,
  HTTPBIN_URL,
  API_TEST_SERVICE_URL,
} from "./common_helpers.js";
import { createSimpleObject } from "./primary_helpers.js";

test("CustomClient", async () => {
  recordTest("customclient-request-parameters-retained");

  const httpClient = new HTTPClient({
    fetcher: (request) => {
      return fetch(request);
    },
  });

  httpClient.addHook("beforeRequest", (request) => {
    request.headers.set("X-Custom-Header", "someValue");

    return request;
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient: httpClient });

  const res: CustomClientPostResponse = await sdk.customClient.customClientPost(
    "headerValue",
    "pathValue",
    "queryValue",
    createSimpleObject(),
  );
  expect(res.httpMeta.response.status).toBe(200);

  expect(res.res?.url).toBe(
    `${HTTPBIN_URL}/anything/customClient/pathValue?queryStringParam=queryValue`,
  );
  expect(res.res?.args?.queryStringParam).toBe("queryValue");
  expect(res.res?.headers["Headerparam"]).toBe("headerValue");
  expect(res.res?.headers["X-Custom-Header"]).toBe("someValue");
});

test("CustomClient ProxyAgent", async () => {
  recordTest("custom-client-http-proxy");

  const dispatcher = new ProxyAgent(`${API_TEST_SERVICE_URL}`);

  const httpClient = new HTTPClient({
    fetcher: (input, init) =>
      fetch(input, { ...init, dispatcher } as RequestInit),
  });

  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });

  const res = await sdk.customClient.proxy("proxied");

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.string).toBeDefined();
  expect(res.string).toContain("proxied");
});
