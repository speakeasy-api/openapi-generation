import * as fs from "fs";
import * as path from "path";

import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { bytesToBlob } from "../lib/files.js";
import { filesToByteArray } from "./files.js";
import {
  HTTPBIN_URL,
  recordTest,
  API_TEST_SERVICE_URL,
} from "./common_helpers.js";

test("bytesToBlob copies Uint8Array with non-zero byteOffset", async () => {
  // Simulate a Node.js Buffer backed by a pooled ArrayBuffer: create a large
  // ArrayBuffer then take a view into the middle of it.
  const pool = new ArrayBuffer(8192);
  const view = new Uint8Array(pool, 100, 5);
  view.set([104, 101, 108, 108, 111]); // "hello"

  // Sanity: the view shares the larger pool
  expect(view.byteOffset).toBe(100);
  expect(view.buffer.byteLength).toBe(8192);

  const blob = bytesToBlob(view, "application/octet-stream");

  // The Blob must contain exactly the 5 bytes, not the 8192-byte pool
  expect(blob.size).toBe(5);
  const bytes = new Uint8Array(await blob.arrayBuffer());
  expect(bytes).toEqual(new Uint8Array([104, 101, 108, 108, 111]));
});

test("Multipart file upload with offset Uint8Array sends correct bytes", async () => {
  const s = new SDK({ serverURL: HTTPBIN_URL });

  const filePath = path.resolve(__dirname, "./testdata/testUpload.json");
  const fullBytes = new Uint8Array(fs.readFileSync(filePath));

  // Embed the file bytes at a non-zero offset within a larger buffer, as
  // Node.js Buffers commonly do with their internal memory pool.
  const padding = 256;
  const padded = new Uint8Array(padding + fullBytes.length + padding);
  padded.set(fullBytes, padding);
  const offsetView = new Uint8Array(padded.buffer, padding, fullBytes.length);

  // Sanity: the view shares the larger buffer
  expect(offsetView.byteOffset).toBe(padding);
  expect(offsetView.buffer.byteLength).toBe(padded.length);

  const res = await s.requestBodies.requestBodyPutMultipartFile({
    file: {
      content: offsetView,
      fileName: "testUpload.json",
    },
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  // httpbin echoes the file content back — it must match the original file,
  // not include any padding bytes from the surrounding buffer.
  expect(res.res?.files).toEqual({
    file: fs.readFileSync(filePath).toString(),
  });
});

test("Request Body Put Multipart File", async () => {
  recordTest("request-bodies-put-multipart-file");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const filePath = path.resolve(__dirname, "./testdata/testUpload.json");
  const data = fs.readFileSync(filePath);

  const res = await s.requestBodies.requestBodyPutMultipartFile({
    file: {
      content: await filesToByteArray(filePath),
      fileName: "testUpload.json",
    },
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res?.files).toEqual({
    file: data.toString(),
  });
});

test("Request Body Put Multipart File Ref", async () => {
  recordTest("request-bodies-put-multipart-file-ref");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const filePath = path.resolve(__dirname, "./testdata/testUpload.json");
  const data = fs.readFileSync(filePath);

  const res = await s.requestBodies.requestBodyPutMultipartFileRef({
    file: {
      content: await filesToByteArray(filePath),
      fileName: "testUpload.json",
    },
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res?.files).toEqual({
    file: data.toString(),
  });
});

test("Request Body Put Bytes", async () => {
  recordTest("request-bodies-put-bytes");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const filePath = path.resolve(__dirname, "./testdata/testUpload.json");
  const data = fs.readFileSync(filePath);

  const res = await s.requestBodies.requestBodyPutBytes(
    await filesToByteArray(filePath),
  );

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res?.data).toEqual(data.toString());
});

test("Request Body No Body No Content Type", async () => {
  recordTest("request-bodies-no-body-no-content-type");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.methods.methodGet({
    serverURL: API_TEST_SERVICE_URL,
  });

  expect(res.httpMeta.request.body).toBe(null);
  expect(res.httpMeta.request.headers.get("content-type")).toBe(null);

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.object).toEqual({ status: "OK" });
});

test("Request Body Wildcard No Content Type", async () => {
  recordTest("request-bodies-wildcard-no-content-type");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.requestBodies.requestBodyPostWildcard();

  expect(res.httpMeta.request.body).toBe(null);
  expect(res.httpMeta.request.headers.get("content-type")).toBe(null);

  expect(res.httpMeta.response.status).toBe(200);
});

test("Request Bodies File Upload Extract Content Type", async () => {
  recordTest("request-bodies-file-upload-extract-content-type");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.requestBodies.requestBodyPostWildcard(
    new Blob(["a,b,c"], { type: "text/csv" }),
  );

  expect(res.httpMeta.request.body).not.toBeNull();
  expect(res.httpMeta.request.headers.get("content-type")).toEqual("text/csv");

  expect(res.httpMeta.response.status).toEqual(200);
  expect(new Headers(res.res?.headers).get("content-type")).toEqual("text/csv");
});

test("Test Form Encoded String Array", async () => {
  recordTest("request-bodies-form-encoded-string-array");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const res = await sdk.requestBodies.formRequestBodyStringArray({
    stuff: ["item1", "item2", "item3"],
  });

  expect(res.httpMeta.response.status).toBe(200);

  // Validate the request was properly form-encoded
  expect(res.res?.form).toBeDefined();
  expect(res.res?.form["stuff"]?.length).toBe(3);
  expect(res.res?.form["stuff"]).toContain("item1");
  expect(res.res?.form["stuff"]).toContain("item2");
  expect(res.res?.form["stuff"]).toContain("item3");
});

const BASE64_NON_UTF8_BYTES = new Uint8Array([
  0xff, 0xfe, 0x00, 0x01, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0xc3, 0x28,
]);

function bytesToBase64(bytes: Uint8Array): string {
  if (
    typeof (
      globalThis as {
        Buffer?: { from(b: Uint8Array): { toString(enc: string): string } };
      }
    ).Buffer !== "undefined"
  ) {
    return (
      globalThis as unknown as {
        Buffer: { from(b: Uint8Array): { toString(enc: string): string } };
      }
    ).Buffer.from(bytes).toString("base64");
  }
  let bin = "";
  for (const b of bytes) {
    bin += String.fromCharCode(b);
  }
  return btoa(bin);
}

test("Request Bodies Base64 Input Mode File Split Components", async () => {
  recordTest("request-bodies-base64-input-mode-file-split-components");

  const s = new SDK();
  const preEncoded = bytesToBase64(BASE64_NON_UTF8_BYTES);

  const res = await s.requestBodies.postBase64InputMode({
    dataByte: preEncoded,
    dataContentEncoding: preEncoded,
    dataPlain: "plain-value",
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res?.json.dataByte).toBe(preEncoded);
  expect(res.res?.json.dataContentEncoding).toBe(preEncoded);
  expect(res.res?.json.dataPlain).toBe("plain-value");
});

test("Request Bodies Base64 Input Mode File Shared Component", async () => {
  recordTest("request-bodies-base64-input-mode-file-shared-component");

  const s = new SDK();
  const preEncoded = bytesToBase64(BASE64_NON_UTF8_BYTES);

  const res = await s.requestBodies.postBase64InputModeShared({
    dataPlain: "plain-shared",
    dataByte: preEncoded,
    dataContentEncoding: preEncoded,
  });

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.res).toBeDefined();
  expect(res.res?.json.dataPlain).toBe("plain-shared");
  expect(res.res?.json.dataByte).toBe(preEncoded);
  expect(res.res?.json.dataContentEncoding).toBe(preEncoded);
});

test("Request Bodies Base64 File Input Idempotent", async () => {
  recordTest("request-bodies-base64-file-input-idempotent");

  const s = new SDK();
  const preEncoded = bytesToBase64(BASE64_NON_UTF8_BYTES);

  const payload = {
    dataByte: preEncoded,
    dataContentEncoding: preEncoded,
    dataPlain: "plain-idempotent",
  };

  const first = await s.requestBodies.postBase64InputMode(payload);
  const second = await s.requestBodies.postBase64InputMode(payload);

  expect(first.httpMeta.response.status).toBe(200);
  expect(first.res).toBeDefined();
  expect(first.res?.json.dataByte).toBe(preEncoded);
  expect(first.res?.json.dataContentEncoding).toBe(preEncoded);
  expect(first.res?.json.dataPlain).toBe("plain-idempotent");

  expect(second.httpMeta.response.status).toBe(200);
  expect(second.res).toBeDefined();
  expect(second.res?.json.dataByte).toBe(first.res?.json.dataByte);
  expect(second.res?.json.dataContentEncoding).toBe(
    first.res?.json.dataContentEncoding,
  );
  expect(second.res?.json.dataPlain).toBe(first.res?.json.dataPlain);
  expect(second.res?.json.dataByte).toBe(preEncoded);
  expect(second.res?.json.dataContentEncoding).toBe(preEncoded);
  expect(second.res?.json.dataPlain).toBe("plain-idempotent");
});
