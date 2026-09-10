import { expect, test } from "vitest";
import { SDK } from "../index.js";
import {
  JsonlStreamResponseBody,
  JsonlStreamChunksResponseBody,
} from "../sdk/models/operations/index.js";
import { JsonLStream } from "../lib/jsonl.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("jsonl stream data", async () => {
  recordTest("jsonl-stream-data-flat-responses");
  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.jsonl.jsonlStream();

  expect(res).toBeDefined();
  expect(res).toBeInstanceOf(JsonLStream);

  const result: JsonlStreamResponseBody[] = [];
  for await (const item of res) {
    result.push(item);
  }
  expect(result.length).toBe(2);

  // Assert first event
  expect(result[0]?.name).toBe("Peter");
  expect(result[0]?.skills).toEqual(["Go", "Python"]);

  // Assert second event
  expect(result[1]?.name).toBe("John");
  expect(result[1]?.skills).toEqual(["Go", "Rust"]);
});

test("jsonl stream data with x-ndjson", async () => {
  recordTest("jsonl-stream-data-async-flat-response");
  const s = new SDK({ serverURL: HTTPBIN_URL });

  const res = await s.jsonl.xNdjsonStream();

  expect(res).toBeDefined();
  expect(res).toBeInstanceOf(JsonLStream);

  const result: JsonlStreamResponseBody[] = [];
  for await (const item of res) {
    result.push(item);
  }
  expect(result.length).toBe(2);

  // Assert first event
  expect(result[0]?.name).toBe("Peter");
  expect(result[0]?.skills).toEqual(["Go", "Python"]);

  // Assert second event
  expect(result[1]?.name).toBe("John");
  expect(result[1]?.skills).toEqual(["Go", "Rust"]);
});

// This test checks that the stream of data is processed as soon as it comes
test("jsonl stream data chunks", async () => {
  recordTest("jsonl-stream-data-chunks-flat-response");
  const s = new SDK({ serverURL: HTTPBIN_URL });

  const startTime = Date.now();
  const res = await s.jsonl.jsonlStreamChunks();

  expect(res).toBeInstanceOf(JsonLStream);

  const iterator = res[Symbol.asyncIterator]();
  const firstEventResult = await iterator.next();
  const firstEventTime = Date.now() - startTime;

  // First event should take longer than 100ms because the server sends a chunk and then waits
  expect(firstEventTime).toBeGreaterThanOrEqual(100);
  expect(firstEventTime).toBeLessThan(190);

  expect(firstEventResult.done).toBe(false);
  const firstEvent = firstEventResult.value as JsonlStreamChunksResponseBody;

  // Assert first event
  expect(firstEvent.name).toBe("Peter");
  expect(firstEvent.skills).toEqual(["Go", "Python"]);

  // Get second event
  const secondEventResult = await iterator.next();
  expect(secondEventResult.done).toBe(false);
  const secondEvent = secondEventResult.value as JsonlStreamChunksResponseBody;

  // Assert second event
  expect(secondEvent.name).toBe("John");
  expect(secondEvent.skills).toEqual(["Go", "Rust"]);
});

// This test checks that the stream of data is processed as soon as it comes
test("jsonl stream data chunks with x-ndjson", async () => {
  recordTest("jsonl-stream-data-async-chunks-flat-response");
  const s = new SDK({ serverURL: HTTPBIN_URL });

  const startTime = Date.now();
  const res = await s.jsonl.xNdjsonStreamChunks();

  expect(res).toBeInstanceOf(JsonLStream);

  const iterator = res[Symbol.asyncIterator]();
  const firstEventResult = await iterator.next();
  const firstEventTime = Date.now() - startTime;

  // First event should take longer than 100ms because the server sends a chunk and then waits
  expect(firstEventTime).toBeGreaterThanOrEqual(100);
  expect(firstEventTime).toBeLessThan(190);

  expect(firstEventResult.done).toBe(false);
  const firstEvent = firstEventResult.value as JsonlStreamChunksResponseBody;

  // Assert first event
  expect(firstEvent.name).toBe("Peter");
  expect(firstEvent.skills).toEqual(["Go", "Python"]);

  // Get second event
  const secondEventResult = await iterator.next();
  expect(secondEventResult.done).toBe(false);
  const secondEvent = secondEventResult.value as JsonlStreamChunksResponseBody;

  // Assert second event
  expect(secondEvent.name).toBe("John");
  expect(secondEvent.skills).toEqual(["Go", "Rust"]);
});
