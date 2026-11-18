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

  test.each([
    "..",
    ".",
    "%2e%2e",
    ".%2E",
    "a/../b",
    "..\\b",
    ".\t.",
    "..\x00",
    ".. ",
    "\x00..",
    " ..",
    "a/.. ",
    ".. \n",
  ])("rejects %j", (value) => {
    expect(() =>
      checked(pattern)({ store_id: value, item_id: "item-1" }),
    ).toThrow(dotSegmentError);
  });

  test.each(["..?x", "..#frag", "safe?next=/..", "safe#/../x", "a?b", "#"])(
    "rejects unencoded '?' or '#' in %j",
    (value) => {
      expect(() =>
        checked(pattern)({ store_id: value, item_id: "item-1" }),
      ).toThrow(
        /^Parameter 'store_id' must not contain an unencoded '\?' or '#'$/,
      );
    },
  );

  test.each([
    "https://evil.example.com",
    "http:evil.example.com",
    "//evil.example.com",
    "\\\\evil.example.com",
    " //evil.example.com",
    "\x00http:evil.example.com",
  ])("rejects %j overriding the base URL", (value) => {
    expect(() => checked("{id}/details")({ id: value })).toThrow(
      /^Parameter 'id' must not start the path with a URL scheme or '\/\/'$/,
    );
  });

  test.each(["/evil.example.com", "\\evil.example.com"])(
    "rejects %j forming '//' after a leading slash",
    (value) => {
      expect(() => checked("/{id}/details")({ id: value })).toThrow(
        /^Parameter 'id' must not start the path/,
      );
    },
  );

  test("rejects a protocol-relative path formed after empty parameters", () => {
    expect(() =>
      checked("/{a}/{b}/details")({ a: "", b: "//evil.example.com" }),
    ).toThrow(/^Parameter 'b' must not start the path/);
  });

  test("allows ':' outside the first path segment", () => {
    expect(checked("/v1/items/{id}")({ id: "foo:bar" })).toEqual(
      "/v1/items/foo:bar",
    );
  });

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

  test.each(["...", ".hidden", "v1.2.3", "a..b"])("allows %j", (value) => {
    expect(checked(pattern)({ store_id: value, item_id: "item-1" })).toEqual(
      `/stores/${value}/items/item-1/details`,
    );
  });

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
