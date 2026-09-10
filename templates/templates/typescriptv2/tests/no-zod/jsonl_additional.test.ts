import { expect, test } from "vitest";
import { SDK } from "../index.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

// In the no-zod variant, jsonl streams aren't auto-wrapped in a JsonLStream
// iterator. We read the raw byte stream and split on newlines instead.
async function collectJsonl(
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

test("jsonl stream data", async () => {
  recordTest("jsonl-stream-data-envelope-http-responses");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.jsonl.jsonlStream();

  expect(result).toBeDefined();
  expect(result.httpMeta).toBeDefined();
  expect(result.httpMeta.response).toBeDefined();
  expect(result.httpMeta.response.status).toBe(200);
  expect(result.object).toBeDefined();

  if (!result.object) {
    expect.fail("Expected object to be defined");
  }

  const events = await collectJsonl(
    result.object as unknown as ReadableStream<Uint8Array>,
  );

  expect(events.length).toBe(2);

  expect(events[0]?.name).toBe("Peter");
  expect(events[0]?.skills).toEqual(["Go", "Python"]);

  expect(events[1]?.name).toBe("John");
  expect(events[1]?.skills).toEqual(["Go", "Rust"]);
});

test("jsonl stream data chunks", async () => {
  recordTest("jsonl-stream-data-chunks-envelope-http-responses");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.jsonl.jsonlStreamChunks();

  expect(result).toBeDefined();
  expect(result.httpMeta).toBeDefined();
  expect(result.httpMeta.response).toBeDefined();
  expect(result.httpMeta.response.status).toBe(200);
  expect(result.object).toBeDefined();

  if (!result.object) {
    expect.fail("Expected object to be defined");
  }

  const events = await collectJsonl(
    result.object as unknown as ReadableStream<Uint8Array>,
  );

  expect(events.length).toBeGreaterThanOrEqual(2);

  expect(events[0]?.name).toBe("Peter");
  expect(events[0]?.skills).toEqual(["Go", "Python"]);

  expect(events[1]?.name).toBe("John");
  expect(events[1]?.skills).toEqual(["Go", "Rust"]);
});

test("jsonl deserialization with camel case properties", async () => {
  recordTest("jsonl-deserialization-camel-case-properties");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.jsonl.jsonlDeserializationVerification();

  expect(result).toBeDefined();
  expect(result.httpMeta).toBeDefined();
  expect(result.httpMeta.response).toBeDefined();
  expect(result.httpMeta.response.status).toBe(200);
  expect(result.object).toBeDefined();

  if (!result.object) {
    expect.fail("Expected object to be defined");
  }

  const events = await collectJsonl(
    result.object as unknown as ReadableStream<Uint8Array>,
  );
  expect(events[0]?.isFinished).toBe("yes");
});
