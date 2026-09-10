import { expect, test } from "vitest";
import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Redirects Are Followed", async () => {
  recordTest("redirects-are-followed");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.redirects.redirectsAreFollowed();

  expect(result).toBeDefined();
  expect(result.name).toEqual("John Doe");
});
