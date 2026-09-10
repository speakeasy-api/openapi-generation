import { expect, test } from "vitest";

import { appendForm, normalizeBlob } from "../lib/encodings.js";
import { isBlobLike } from "../sdk/types/blobs.js";

type CrossRealmBlob = Pick<
  Blob,
  "type" | "size" | "stream" | "arrayBuffer" | "slice" | "text"
> & { [Symbol.toStringTag]: string };

type CrossRealmFile = CrossRealmBlob & { name: string };

/**
 * Creates a duck-typed Blob-like object that passes isBlobLike() but fails
 * `instanceof Blob`. This simulates what happens in webpack/Next.js bundled
 * environments where the SDK runs in a separate realm from the user's File
 * objects.
 */
function createCrossRealmBlob(content: string, type: string): CrossRealmBlob {
  const inner = new Blob([content], { type });
  return {
    [Symbol.toStringTag]: "Blob",
    type,
    size: inner.size,
    stream: () => inner.stream(),
    arrayBuffer: () => inner.arrayBuffer(),
    slice: (...args: Parameters<Blob["slice"]>) => inner.slice(...args),
    text: () => inner.text(),
  };
}

/**
 * Same as above but simulates a cross-realm File (has a .name property and
 * toStringTag of "File").
 */
function createCrossRealmFile(
  content: string,
  name: string,
  type: string,
): CrossRealmFile {
  return {
    ...createCrossRealmBlob(content, type),
    [Symbol.toStringTag]: "File",
    name,
  };
}

test("normalizeBlob returns native Blob unchanged", async () => {
  const native = new Blob(["hello"], { type: "text/plain" });
  const result = await normalizeBlob(native);
  expect(result).toBe(native);
});

test("normalizeBlob wraps cross-realm Blob-like in a native Blob", async () => {
  const crossBlob = createCrossRealmBlob("hello world", "text/plain");

  // Sanity: this is NOT a native Blob
  expect(crossBlob instanceof Blob).toBe(false);
  expect(isBlobLike(crossBlob)).toBe(true);

  const result = await normalizeBlob(
    crossBlob as Pick<Blob, "arrayBuffer" | "type">,
  );

  expect(result).toBeInstanceOf(Blob);
  expect(result.type).toBe("text/plain");
  expect(await result.text()).toBe("hello world");
});

test("normalizeBlob preserves MIME type from cross-realm Blob", async () => {
  const crossBlob = createCrossRealmBlob("data", "application/pdf");
  const result = await normalizeBlob(
    crossBlob as Pick<Blob, "arrayBuffer" | "type">,
  );
  expect(result.type).toBe("application/pdf");
});

test("normalizeBlob returns native File unchanged", async () => {
  const native = new File(["data"], "test.txt", { type: "text/plain" });
  const result = await normalizeBlob(native);
  expect(result).toBe(native);
});

test("cross-realm File name is preserved when appendForm is called with name", async () => {
  const fd = new FormData();
  const crossFile = createCrossRealmFile(
    "file content",
    "report.pdf",
    "application/pdf",
  );

  const blob = await normalizeBlob(crossFile);
  appendForm(fd, "file", blob, crossFile.name);

  const entry = fd.get("file") as File;
  expect(entry).toBeInstanceOf(Blob);
  expect(entry.name).toBe("report.pdf");
  expect(await entry.text()).toBe("file content");
});

test("native Blob passes through normalizeBlob and appends correctly", async () => {
  const fd = new FormData();
  const nativeBlob = new Blob(["native content"], { type: "text/plain" });

  const blob = await normalizeBlob(nativeBlob);
  appendForm(fd, "file", blob);

  const entry = fd.get("file") as Blob;
  expect(entry).toBeInstanceOf(Blob);
  expect(await entry.text()).toBe("native content");
});

test("native File passes through with filename", async () => {
  const fd = new FormData();
  const nativeFile = new File(["file data"], "test.txt", {
    type: "text/plain",
  });

  const blob = await normalizeBlob(nativeFile);
  const name =
    "name" in nativeFile ? (nativeFile as { name: string }).name : undefined;
  appendForm(fd, "file", blob, name);

  const entry = fd.get("file") as File;
  expect(entry).toBeInstanceOf(Blob);
  expect(entry.name).toBe("test.txt");
  expect(await entry.text()).toBe("file data");
});

test("appendForm converts non-blob values to strings", () => {
  const fd = new FormData();

  appendForm(fd, "count", 42);

  expect(fd.get("count")).toBe("42");
});

test("appendForm skips null and undefined values", () => {
  const fd = new FormData();

  appendForm(fd, "a", null);
  appendForm(fd, "b", undefined);

  expect(fd.get("a")).toBeNull();
  expect(fd.get("b")).toBeNull();
});
