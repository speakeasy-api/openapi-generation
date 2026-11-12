import { describe, test, expect } from "vitest";

import { encodeLabel, encodeSimple } from "../../lib/encodings.js";
import { pathToFunc } from "../../lib/url.js";

test.each<
  [
    title: string,
    pattern: string,
    args: Record<string, string | number>,
    output: string | Error,
  ]
>([
  ["no placeholders", "/sample/path", {}, "sample/path"],
  ["one placeholder end", "/author/{id}", { id: 123 }, "author/123"],
  [
    "one placeholder mid",
    "/author/{id}/likes",
    { id: "123" },
    "author/123/likes",
  ],
  [
    "multiple placeholders",
    "/{a}/{b}/{c}",
    { a: "foo", b: "bar", c: "baz" },
    "foo/bar/baz",
  ],
  [
    "infix placeholder",
    "/entity/user-{id}/",
    { id: "1893" },
    "entity/user-1893/",
  ],
  [
    "placeholder with hyphen",
    "/foo/bar/{thing-blah}",
    { "thing-blah": "test-value" },
    "foo/bar/test-value",
  ],
  [
    "insufficient params",
    "/blog/{slug}/comments/{commentId}",
    { slug: "hello-world" },
    new Error("Parameter 'commentId' is required"),
  ],
  [
    // This test covers javascript users that may pass opaque objects into the
    // path builder function.
    "invalid param type",
    "/blog/{slug}/comments/{commentId}",
    { slug: [] as unknown as string, commentId: 456 },
    new Error("Parameter 'slug' must be a string or number"),
  ],
])("pathToFunc: %s", (_, pattern, args, output) => {
  if (output instanceof Error) {
    expect(() => pathToFunc(pattern)(args)).toThrow(output);
  } else {
    expect(pathToFunc(pattern)(args)).toEqual(output);
  }
});

test("pathToFunc encodes param values", () => {
  const expected = "channel/product%3Aupdates/entity/user-u%2F123/";
  const actual = pathToFunc("/channel/{cid}/entity/user-{uid}/", {
    charEncoding: "percent",
  })({
    cid: "product:updates",
    uid: "u/123",
  });

  expect(actual).toEqual(expected);
});

describe("pathToFunc dot segments", () => {
  const pattern = "/stores/{store_id}/items/{item_id}/details";
  const dotSegmentError =
    /Parameter 'store_id' must not form a '\.' or '\.\.' path segment/;

  test.each([
    "..",
    ".",
    "%2e%2e",
    "%2E%2E",
    ".%2e",
    "%2E.",
    "%2e",
    "a/../b",
    "a/./b",
    "..\\b",
    "..?x=1",
    "..#frag",
    ".\t.",
    ".\n",
    "\r..",
  ])("rejects %j when values are pre-encoded", (value) => {
    expect(() =>
      pathToFunc(pattern)({ store_id: value, item_id: "item-1" }),
    ).toThrow(dotSegmentError);
  });

  test.each([".", ".."])("rejects %j under percent encoding", (value) => {
    expect(() =>
      pathToFunc(pattern, { charEncoding: "percent" })({
        store_id: value,
        item_id: "item-1",
      }),
    ).toThrow(dotSegmentError);
  });

  test.each([
    ["simple, percent", encodeSimple("id", "..", { charEncoding: "percent" })],
    [
      "simple, allowReserved",
      encodeSimple("id", "a/../b", { charEncoding: "percentExceptReserved" }),
    ],
    ["label, percent", encodeLabel("id", ".", { charEncoding: "percent" })],
  ])("rejects encoder output: %s", (_, encoded) => {
    expect(() =>
      pathToFunc(pattern)({ store_id: encoded!, item_id: "item-1" }),
    ).toThrow(dotSegmentError);
  });

  test("names every placeholder that shares the dot segment", () => {
    expect(() => pathToFunc("/x/{a}{b}/y")({ a: ".", b: "." })).toThrow(
      "Parameter 'a', 'b' must not form a '.' or '..' path segment",
    );
  });

  test("rejects dot segments in any parameter position", () => {
    expect(() =>
      pathToFunc(pattern)({ store_id: "store_1", item_id: ".." }),
    ).toThrow(/Parameter 'item_id'/);
  });

  test.each<[value: string | number, segment: string]>([
    ["store_1", "store_1"],
    ["...", "..."],
    ["....", "...."],
    [".hidden", ".hidden"],
    ["file.txt", "file.txt"],
    ["v1.2.3", "v1.2.3"],
    ["a..b", "a..b"],
    ["..a", "..a"],
    ["%2e%2e%2e", "%2e%2e%2e"],
    ["%2e%2", "%2e%2"],
    [1.5, "1.5"],
  ])("allows %j", (value, segment) => {
    expect(pathToFunc(pattern)({ store_id: value, item_id: "item-1" })).toEqual(
      `stores/${segment}/items/item-1/details`,
    );
  });

  test("percent encoding neutralises encoded dots", () => {
    expect(
      pathToFunc(pattern, { charEncoding: "percent" })({
        store_id: "%2e%2e",
        item_id: "item-1",
      }),
    ).toEqual("stores/%252e%252e/items/item-1/details");
  });

  test("allows dots alongside literal text in the same segment", () => {
    expect(pathToFunc("/entity/user-{id}")({ id: ".." })).toEqual(
      "entity/user-..",
    );
    expect(pathToFunc("/items/{name}.{ext}")({ name: "a", ext: "b" })).toEqual(
      "items/a.b",
    );
  });

  test("leaves literal template segments unchanged", () => {
    expect(pathToFunc("/v1/../{id}")({ id: "x" })).toEqual("v1/../x");
  });

  test("allows server URL templates", () => {
    expect(
      pathToFunc("http://{hostname}:{port}/{basePath}")({
        hostname: "localhost",
        port: 8080,
        basePath: "v1",
      }),
    ).toEqual("http://localhost:8080/v1");
  });

  test("resolves allowed values to the intended route", () => {
    const path = pathToFunc(pattern)({
      store_id: "...",
      item_id: "item.1",
    });
    expect(new URL(path, "https://api.example.com/v2/").pathname).toEqual(
      "/v2/stores/.../items/item.1/details",
    );
  });
});
