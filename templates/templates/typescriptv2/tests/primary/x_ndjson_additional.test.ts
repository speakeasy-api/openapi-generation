import { expect, test } from "vitest";
import { SDK } from "../index.js";
import { XNdjsonStreamResponseBody } from "../sdk/models/operations/xndjsonstream.js";
import { XNdjsonStreamChunksResponseBody } from "../sdk/models/operations/xndjsonstreamchunks.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

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

  const events: XNdjsonStreamResponseBody[] = [];
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

  const iterator = result.object[Symbol.asyncIterator]();

  // Get first event
  const firstEventResult = await iterator.next();

  expect(firstEventResult.done).toBe(false);
  const firstEvent = firstEventResult.value as XNdjsonStreamChunksResponseBody;

  // Assert first event
  expect(firstEvent.name).toBe("Peter");
  expect(firstEvent.skills).toEqual(["Go", "Python"]);

  // Get second event
  const secondEventResult = await iterator.next();
  expect(secondEventResult.done).toBe(false);
  const secondEvent =
    secondEventResult.value as XNdjsonStreamChunksResponseBody;

  // Assert second event
  expect(secondEvent.name).toBe("John");
  expect(secondEvent.skills).toEqual(["Go", "Rust"]);
});
