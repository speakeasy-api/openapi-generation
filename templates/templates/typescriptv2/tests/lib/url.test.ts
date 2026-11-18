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
    /^Parameter 'store_id' must not form a '\.' or '\.\.' path segment$/;
  const checked = (
    path: string,
    options?: { charEncoding?: "percent" | "none" },
  ) => pathToFunc(path, { ...options, rejectDotSegments: true });

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
    ".\t.",
    ".\n",
    "\r..",
    "..\x00",
    ".. ",
    "\x00..",
    " ..",
    "a/.. ",
    ".. \n",
  ])("rejects %j when values are pre-encoded", (value) => {
    expect(() =>
      checked(pattern)({ store_id: value, item_id: "item-1" }),
    ).toThrow(dotSegmentError);
  });

  test.each(["..?x=1", "..#frag", "safe?next=/..", "safe#/../x", "a?b", "#"])(
    "rejects unencoded '?' or '#' in %j",
    (value) => {
      expect(() =>
        checked(pattern)({ store_id: value, item_id: "item-1" }),
      ).toThrow(
        /^Parameter 'store_id' must not contain an unencoded '\?' or '#'$/,
      );
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

  test.each([
    ["simple, percent", encodeSimple("id", "..", { charEncoding: "percent" })],
    [
      "simple, allowReserved",
      encodeSimple("id", "a/../b", { charEncoding: "percentExceptReserved" }),
    ],
    ["label, percent", encodeLabel("id", ".", { charEncoding: "percent" })],
  ])("rejects encoder output: %s", (_, encoded) => {
    expect(() =>
      checked(pattern)({ store_id: encoded!, item_id: "item-1" }),
    ).toThrow(dotSegmentError);
  });

  test("names every placeholder that contributes to the dot segment", () => {
    expect(() => checked("/x/{a}{b}/y")({ a: ".", b: "." })).toThrow(
      "Parameters 'a', 'b' must not form a '.' or '..' path segment",
    );
  });

  test("omits empty placeholders from the error", () => {
    expect(() => checked("/x/{a}{b}/y")({ a: "..", b: "" })).toThrow(
      /^Parameter 'a' must not form/,
    );
  });

  test("allows a literal dot segment when the placeholder is empty", () => {
    expect(checked("/x/.{a}/y")({ a: "" })).toEqual("x/./y");
  });

  test("rejects dot segments in any parameter position", () => {
    expect(() =>
      checked(pattern)({ store_id: "store_1", item_id: ".." }),
    ).toThrow(/^Parameter 'item_id'/);
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
    expect(checked(pattern)({ store_id: value, item_id: "item-1" })).toEqual(
      `stores/${segment}/items/item-1/details`,
    );
  });

  test("percent encoding neutralises encoded dots", () => {
    expect(
      checked(pattern, { charEncoding: "percent" })({
        store_id: "%2e%2e",
        item_id: "item-1",
      }),
    ).toEqual("stores/%252e%252e/items/item-1/details");
  });

  test("allows dots alongside literal text in the same segment", () => {
    expect(checked("/entity/user-{id}")({ id: ".." })).toEqual(
      "entity/user-..",
    );
    expect(checked("/items/{name}.{ext}")({ name: "a", ext: "b" })).toEqual(
      "items/a.b",
    );
  });

  test.each(["/x/{id}?y", "/x/{id}#y", "/x/{id}?y/z"])(
    "rejects a dot segment before a literal '?' or '#' in %j",
    (path) => {
      expect(() => checked(path)({ id: ".." })).toThrow(
        /^Parameter 'id' must not form a '\.' or '\.\.' path segment$/,
      );
      expect(() => checked(path)({ id: "." })).toThrow(
        /^Parameter 'id' must not form/,
      );
    },
  );

  test("leaves literal template segments unchanged", () => {
    expect(checked("/v1/../{id}")({ id: "x" })).toEqual("v1/../x");
  });

  test("does not check templates unless rejectDotSegments is set", () => {
    expect(pathToFunc(pattern)({ store_id: "..", item_id: "item-1" })).toEqual(
      "stores/../items/item-1/details",
    );
    expect(pathToFunc("https://{host}/x")({ host: ".." })).toEqual(
      "https://../x",
    );
    expect(
      pathToFunc("http://{hostname}:{port}/{basePath}", {
        charEncoding: "percent",
      })({ hostname: "localhost", port: 8080, basePath: "." }),
    ).toEqual("http://localhost:8080/.");
  });

  test("keeps encoded '?' and '#' inside the path segment", () => {
    const path = checked(pattern)({
      store_id: encodeSimple("id", "a?b=1#c", {
        charEncoding: "percentExceptReserved",
      })!,
      item_id: "item-1",
    });
    const url = new URL(path, "https://api.example.com/v2/");
    expect(url.pathname).toEqual("/v2/stores/a%3Fb=1%23c/items/item-1/details");
    expect(url.search).toEqual("");
    expect(url.hash).toEqual("");
  });

  test.each([
    "https://evil.example.com",
    "http:evil.example.com",
    "javascript:alert(1)",
    " https://evil.example.com",
    "\x00http:evil.example.com",
    "ht\ttp://evil.example.com",
    "\\\\evil.example.com",
  ])("rejects %j overriding the base URL", (value) => {
    expect(() => checked("/{id}/details")({ id: value })).toThrow(
      /^Parameter 'id' must not start the path with a URL scheme or '\/\/'$/,
    );
  });

  test("rejects a scheme formed after empty leading parameters", () => {
    expect(() =>
      checked("/{a}/{b}/details")({ a: "", b: "http:evil.example.com" }),
    ).toThrow(/^Parameter 'b' must not start the path/);
    expect(() => checked("/user{id}")({ id: ":x" })).toThrow(
      /^Parameter 'id' must not start the path/,
    );
  });

  test("allows ':' outside the first path segment", () => {
    expect(
      checked("/v1/models/{id}")({ id: "model-a:generateContent" }),
    ).toEqual("v1/models/model-a:generateContent");
    expect(checked("/v1/items/{id}")({ id: "foo:bar" })).toEqual(
      "v1/items/foo:bar",
    );
  });

  test("allows ordinary first segment values", () => {
    expect(checked("/{id}/details")({ id: "v1.2.3" })).toEqual(
      "v1.2.3/details",
    );
    expect(checked("/{id}/details")({ id: "/v1" })).toEqual("v1/details");
  });

  test("resolves allowed values to the intended route", () => {
    const path = checked(pattern)({
      store_id: "...",
      item_id: "item.1",
    });
    expect(new URL(path, "https://api.example.com/v2/").pathname).toEqual(
      "/v2/stores/.../items/item.1/details",
    );
  });
});
