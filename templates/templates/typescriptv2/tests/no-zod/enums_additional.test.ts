import { expect, test } from "vitest";

import { SDK } from "../index.js";
import type { Theme } from "../sdk/models/shared/theme.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Open Enums Round Trip", async () => {
  recordTest("open-enums-round-trip-string-union");

  // No-zod: open enums inline their primitive fallback type. This assignment
  // compile-checks that unknown values remain assignable while the wire
  // round-trip checks they pass through unchanged.
  const unknownTheme: Theme = {
    color: "purple",
    icon: "tick",
    heroWidth: 2160,
  };

  const s = new SDK({ serverURL: HTTPBIN_URL });
  let result = await s.enums.enumsPostOpenEnumUnrecognized(unknownTheme);

  const theme = result.themeResponse?.json;
  if (!theme) {
    expect.fail("Expected result.themeResponse?.json to be set");
  }
  expect(theme).toEqual({
    color: "purple",
    icon: "tick",
    heroWidth: 2160,
  });

  result = await s.enums.enumsPostOpenEnumUnrecognized(theme);
  expect(theme, "expected unrecognized enum values to roundtrip").toEqual({
    color: "purple",
    icon: "tick",
    heroWidth: 2160,
  });
});
