import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { HTTPBIN_URL } from "./common_helpers.js";

test("Test Form Encoded String Array", async () => {
  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.requestBodies.formRequestBodyStringArray({
    stuff: ["item1", "item2", "item3"],
  });

  expect(res.form).toBeDefined();
  expect(res.form["stuff"]).toContain("item1,item2,item3");
});
