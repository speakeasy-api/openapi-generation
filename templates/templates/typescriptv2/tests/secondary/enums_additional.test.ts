import {
  ThemeColor,
  ThemeHeroWidth,
  ThemeIcon,
} from "../sdk/models/shared/theme.js";
import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("Open Enums Round Trip", async () => {
  recordTest("open-enums-round-trip");

  expect(Object.values(ThemeIcon)).toContain("tick");
  expect(Object.values(ThemeColor)).not.toContain("purple");
  expect(Object.values(ThemeHeroWidth)).not.toContain(2160);

  const s = new SDK({ serverURL: HTTPBIN_URL });
  let result = await s.enums.enumsPostOpenEnumUnrecognized({
    color: "purple",
    icon: ThemeIcon.Tick,
    heroWidth: 2160,
  });

  const theme = result?.json;
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
