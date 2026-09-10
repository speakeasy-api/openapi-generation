import { expect, test } from "vitest";
import { SDK } from "../index.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

// In the no-zod variant, jsonl/ndjson streams are not auto-wrapped into a
// JsonLStream iterator — the matcher passes the raw response body through.
// We therefore read the stream as bytes and split it on newlines.
async function collectNdjson(
  stream: ReadableStream<Uint8Array>,
): Promise<Array<Record<string, unknown>>> {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  const events: Array<Record<string, unknown>> = [];
  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
  }
  buffer += decoder.decode();
  for (const line of buffer.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) {
      continue;
    }
    events.push(JSON.parse(trimmed) as Record<string, unknown>);
  }
  return events;
}

test("x-ndjson stream data", async () => {
  recordTest("x-ndjson-stream-data-envelope-http-responses");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.jsonl.xNdjsonStream();

  expect(result).toBeDefined();
  expect(result.httpMeta).toBeDefined();
  expect(result.httpMeta.response).toBeDefined();
  expect(result.httpMeta.response.status).toBe(200);
  expect(result.object).toBeDefined();

  if (!result.object) {
    expect.fail("Expected object to be defined");
  }

  const events = await collectNdjson(
    result.object as unknown as ReadableStream<Uint8Array>,
  );

  expect(events.length).toBe(2);

  // Assert first event
  expect(events[0]?.name).toBe("Peter");
  expect(events[0]?.skills).toEqual(["Go", "Python"]);

  // Assert second event
  expect(events[1]?.name).toBe("John");
  expect(events[1]?.skills).toEqual(["Go", "Rust"]);
});

test("x-ndjson stream data chunks", async () => {
  recordTest("x-ndjson-stream-data-chunks-envelope-http-responses");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.jsonl.xNdjsonStreamChunks();

  expect(result).toBeDefined();
  expect(result.httpMeta).toBeDefined();
  expect(result.httpMeta.response).toBeDefined();
  expect(result.httpMeta.response.status).toBe(200);
  expect(result.object).toBeDefined();

  if (!result.object) {
    expect.fail("Expected object to be defined");
  }

  const events = await collectNdjson(
    result.object as unknown as ReadableStream<Uint8Array>,
  );

  expect(events.length).toBeGreaterThanOrEqual(2);

  // Assert first event
  expect(events[0]?.name).toBe("Peter");
  expect(events[0]?.skills).toEqual(["Go", "Python"]);

  // Assert second event
  expect(events[1]?.name).toBe("John");
  expect(events[1]?.skills).toEqual(["Go", "Rust"]);
});
