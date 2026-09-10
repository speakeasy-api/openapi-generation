import { expect, test } from "vitest";
import { SDK } from "../index.js";
import { JsonlStreamResponseBody } from "../sdk/models/operations/jsonlstream.js";
import { JsonlStreamChunksResponseBody } from "../sdk/models/operations/jsonlstreamchunks.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

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

  const events: JsonlStreamResponseBody[] = [];
  for await (const event of result.object) {
    events.push(event);
  }

  expect(events.length).toBe(2);

  // Assert first event
  expect(events.length).toBeGreaterThan(0);
  const firstEvent = events[0];
  expect(firstEvent?.name).toBe("Peter");
  expect(firstEvent?.skills).toEqual(["Go", "Python"]);

  // Assert second event
  expect(events.length).toBeGreaterThan(1);
  const secondEvent = events[1];
  expect(secondEvent?.name).toBe("John");
  expect(secondEvent?.skills).toEqual(["Go", "Rust"]);
});

// This test checks that the stream of data is processed as soon as it comes
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

  const iterator = result.object[Symbol.asyncIterator]();

  // Get first event
  const firstEventResult = await iterator.next();

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

  const events = result.object[Symbol.asyncIterator]();
  const firstEvent = await events.next();
  expect(firstEvent.value?.isFinished).toBe("yes");
});
