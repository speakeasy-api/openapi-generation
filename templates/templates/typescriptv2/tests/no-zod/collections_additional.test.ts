import { expect, test } from "vitest";

import { SDK } from "../index.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Collections Containing Null", async () => {
  recordTest("collections-containing-null");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const value = {
    requiredArray: ["foo", null],
    requiredMap: { foo: null, bar: 123 },
    optionalArray: ["foo", null],
    optionalMap: { foo: null, bar: 123 },
    arrayOfNullUnion: ["foo", null],
    mapOfNullUnion: { foo: null, bar: 123 },
  };

  const res = await s.collections.collectionsContainingNull(value);

  expect(res.object?.json).toEqual(value);
});
