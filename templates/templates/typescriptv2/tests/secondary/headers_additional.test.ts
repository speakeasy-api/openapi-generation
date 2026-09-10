import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("test headers response body with headers flat", async () => {
  recordTest("headers-response-body-with-headers-flat");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  expect(sdk).toBeDefined();

  const res = await sdk.responseHeaders.responseHeaders(true, 200);
  expect(res).toBeDefined();
  expect(res.result).toBeDefined();
  expect(res.result!.token).toBe("test-token");
  expect(res.Headers).toBeDefined();
  expect(res.Headers!["x-required-header"]).toEqual(["required"]);
});

test("test headers empty response body with headers flat", async () => {
  recordTest("headers-empty-response-body-with-headers-flat");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  expect(sdk).toBeDefined();

  const res = await sdk.responseHeaders.responseBodyEmptyWithHeaders(
    1.1,
    "hello",
  );

  expect(res).toBeDefined();
  expect(res?.Headers).toBeDefined();
  expect(res?.Headers!["x-string-header"]).toEqual(["hello"]);
  expect(res?.Headers!["x-number-header"]).toEqual(["1.1"]);
});
