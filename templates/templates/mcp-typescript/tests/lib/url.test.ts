import { describe, expect, test } from "vitest";

import { pathToFunc } from "../../lib/url.js";

test.each<
  [
    title: string,
    pattern: string,
    args: Record<string, string | number>,
    output: string | Error,
  ]
>([
  ["no placeholders", "/sample/path", {}, "/sample/path"],
  ["one placeholder", "/author/{id}/likes", { id: 123 }, "/author/123/likes"],
  [
    "infix placeholder",
    "/entity/user-{id}/",
    { id: "1893" },
    "/entity/user-1893/",
  ],
  [
    "placeholder with hyphen",
    "/foo/bar/{thing-blah}",
    { "thing-blah": "test-value" },
    "/foo/bar/test-value",
  ],
  [
    "insufficient params",
    "/blog/{slug}/comments/{commentId}",
    { slug: "hello-world" },
    new Error("Parameter 'commentId' is required"),
  ],
])("pathToFunc: %s", (_, pattern, args, output) => {
  if (output instanceof Error) {
    expect(() => pathToFunc(pattern)(args)).toThrow(output);
  } else {
    expect(pathToFunc(pattern)(args)).toEqual(output);
  }
});

describe("pathToFunc dot segments", () => {
  const pattern = "/stores/{store_id}/items/{item_id}/details";
  const dotSegmentError =
    /^Parameter 'store_id' must not form a '\.' or '\.\.' path segment$/;
  const checked = (
    path: string,
    options?: { charEncoding?: "percent" | "none" },
  ) => pathToFunc(path, { ...options, rejectDotSegments: true });

  test.each(["..", ".", "%2e%2e", ".%2E", "a/../b", "..\\b", "..?x", ".\t."])(
    "rejects %j",
    (value) => {
      expect(() =>
        checked(pattern)({ store_id: value, item_id: "item-1" }),
      ).toThrow(dotSegmentError);
    },
  );

  test.each([".", ".."])("rejects %j under percent encoding", (value) => {
    expect(() =>
      checked(pattern, { charEncoding: "percent" })({
        store_id: value,
        item_id: "item-1",
      }),
    ).toThrow(dotSegmentError);
  });

  test("rejects dot segments from hyphenated placeholders", () => {
    expect(() => checked("/x/{thing-blah}")({ "thing-blah": ".." })).toThrow(
      /^Parameter 'thing-blah' must not form/,
    );
  });

  test("names every placeholder that contributes to the dot segment", () => {
    expect(() => checked("/x/{a}{b}/y")({ a: ".", b: "." })).toThrow(
      "Parameters 'a', 'b' must not form a '.' or '..' path segment",
    );
    expect(() => checked("/x/{a}{b}/y")({ a: "..", b: "" })).toThrow(
      /^Parameter 'a' must not form/,
    );
  });

  test.each(["...", ".hidden", "v1.2.3", "a..b", "safe?next=/.."])(
    "allows %j",
    (value) => {
      expect(checked(pattern)({ store_id: value, item_id: "item-1" })).toEqual(
        `/stores/${value}/items/item-1/details`,
      );
    },
  );

  test("leaves literal template segments unchanged", () => {
    expect(checked("/v1/../{id}")({ id: "x" })).toEqual("/v1/../x");
    expect(checked("/entity/user-{id}")({ id: ".." })).toEqual(
      "/entity/user-..",
    );
  });

  test("does not check templates unless rejectDotSegments is set", () => {
    expect(pathToFunc("https://{host}/x")({ host: ".." })).toEqual(
      "https://../x",
    );
    expect(pathToFunc(pattern)({ store_id: "..", item_id: "item-1" })).toEqual(
      "/stores/../items/item-1/details",
    );
  });
});
