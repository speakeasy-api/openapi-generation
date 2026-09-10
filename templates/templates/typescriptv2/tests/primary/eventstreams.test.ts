import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { ChatCompletionStream } from "../sdk/models/shared/chatcompletionstream.js";
import { ChatCompletionResult } from "../sdk/models/shared/chatcompletionresult.js";
import { ChatCompletionEventData } from "../sdk/models/shared/chatcompletionevent.js";
import { SseOverloadChatResponse } from "../sdk/models/operations/sseoverloadchat.js";
import { EventStream } from "../lib/event-streams.js";
import { DifferentDataSchemasData } from "../sdk/models/shared/differentdataschemas.js";
import { JsonEventData } from "../sdk/models/shared/jsonevent.js";
import { PartialWithCommentsEventData } from "../sdk/models/shared/partialwithcommentsevent.js";
import { UnionWithCommentsStream } from "../sdk/models/shared/unionwithcommentsstream.js";
import { RichStream } from "../sdk/models/shared/richstream.js";
import { TeapotJSONError } from "../sdk/models/errors/teapotjsonerror.js";

import {
  HTTPBIN_URL,
  recordTest,
  API_TEST_SERVICE_URL,
} from "./common_helpers.js";

test("event stream json data", async () => {
  recordTest("event-stream-json-data");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.json();
  if (result.jsonEvent == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: JsonEventData[] = [];
  for await (const event of result.jsonEvent) {
    events.push(event);
  }

  expect(events.length).toBe(4);

  const message = events.map((e) => e.content).join("");
  expect(message).toBe("Hello world!");
});

test("event stream text data", async () => {
  recordTest("event-stream-text-data");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.text();
  if (result.textEvent == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: string[] = [];
  for await (const event of result.textEvent) {
    events.push(event);
  }

  expect(events.length).toBe(4);

  const message = events.join("");
  expect(message).toBe("Hello world!");
});

test("event stream multiline data", async () => {
  recordTest("event-stream-multiline-data");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.multiline();
  if (result.textEvent == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: string[] = [];
  for await (const event of result.textEvent) {
    events.push(event);
  }

  expect(events.length).toBe(1);

  const message = events[0];
  expect(message).toBe("YHOO\n+2\n10");
});

test("event stream rich events", async () => {
  recordTest("event-stream-rich-events");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.rich();
  if (result.richStream == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: RichStream[] = [];
  for await (const event of result.richStream) {
    events.push(event);
  }

  expect(events).toEqual([
    {
      data: { completion: "Hello", model: "jeeves-1", stopReason: null },
      event: "completion",
      id: "job-1",
    },
    { data: "ping", event: "heartbeat", retry: 3000 },
    {
      data: {
        completion: "world!",
        model: "jeeves-1",
        stopReason: "stop_sequence",
      },
      event: "completion",
      id: "job-1",
    },
  ]);
});

test("event stream chat with sentinel events", async () => {
  recordTest("event-stream-chat-sentinel-event");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.chat({ prompt: "Print test content" });
  if (result.chatCompletionStream == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: ChatCompletionStream[] = [];
  for await (const event of result.chatCompletionStream) {
    events.push(event);
  }

  expect(events).toEqual([
    { data: { content: "Hello" } },
    { data: { content: " " } },
    { data: { content: "world" } },
    { data: { content: "!" } },
    { data: "[DONE]" },
  ]);
});

test("event stream chat json", async () => {
  recordTest("event-stream-chat-json");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.chat({
    prompt: "Print test content",
    stream: false,
  });

  const expected: ChatCompletionResult[] = [
    { data: { content: "Hello" } },
    { data: { content: " " } },
    { data: { content: "world" } },
    { data: { content: "!" } },
  ];
  expect(result.chatCompletionResult).toEqual(expected);
});

test("event stream chat skip sentinel events", async () => {
  recordTest("event-stream-chat-skip-sentinel");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.chatSkipSentinel({
    prompt: "Print test content",
  });
  if (result.chatCompletionEvent == null) {
    expect.fail("Expected jsonEvent to be defined");
  }

  const events: ChatCompletionEventData[] = [];
  for await (const event of result.chatCompletionEvent) {
    events.push(event);
  }

  expect(events).toEqual([
    { content: "Hello" },
    { content: " " },
    { content: "world" },
    { content: "!" },
  ]);
});

test("event stream chat heartbeat skips dataless events", async () => {
  recordTest("event-stream-chat-heartbeat-skips-dataless");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.chatHeartbeat();
  if (result.chatCompletionEvent == null) {
    expect.fail("Expected chatCompletionEvent to be defined");
  }

  const events: ChatCompletionEventData[] = [];
  for await (const event of result.chatCompletionEvent) {
    events.push(event);
  }

  // Server sends: data("Hello1"), heartbeat (no data — skipped), data("Hello 2"), data("!"), [DONE]
  expect(events.length).toBe(3);
  expect(events).toEqual([
    { content: "Hello1" },
    { content: "Hello 2" },
    { content: "!" },
  ]);
});

test("event stream with different data schemas", async () => {
  recordTest("event-stream-different-data-schemas");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.differentDataSchemas();
  if (result.differentDataSchemas == null) {
    expect.fail("Expected differentDataSchemas to be defined");
  }

  const events: DifferentDataSchemasData[] = [];
  for await (const event of result.differentDataSchemas) {
    events.push(event);
  }

  expect(events).toEqual([
    { id: 123, content: "Here is your url" },
    { url: "https://example.com" },
    { content: "Have a great day!" },
    [1, 2, 3, 4],
    true,
    3.14159,
  ]);
});

test("event stream with mixed data (JSON and plain text)", async () => {
  recordTest("event-stream-mixed-data");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.mixedData();
  if (result.mixedDataEvent == null) {
    expect.fail("Expected mixedDataEvent to be defined");
  }

  const events: any[] = [];
  for await (const event of result.mixedDataEvent) {
    events.push(event);
  }

  expect(events.length).toBe(4);

  // First event: JSON object (completion)
  expect(events[0].content).toBe("Hello world");

  // Second event: plain text
  expect(events[1]).toBe("Processing your request...");

  // Third event: plain text
  expect(events[2]).toBe("Almost done");

  // Fourth event: JSON object (completion)
  expect(events[3].content).toBe("Done!");
});

test("event stream error response", async () => {
  recordTest("event-stream-error-response");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const promise = s.eventstreams.text({
    fetchOptions: {
      headers: {
        "x-teapot": "json",
      },
    },
  });

  await expect(promise).rejects.toThrowError(TeapotJSONError);
});

test("event stream with abort signal", async () => {
  recordTest("event-stream-with-abort-signal");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const controller = new AbortController();

  const result = await s.eventstreams.text({
    fetchOptions: { signal: controller.signal },
  });
  if (result.textEvent == null) {
    expect.fail("Expected textEvent to be defined");
  }

  let chunks = 0;
  let message = "";
  let caughtAbortError = false;
  try {
    for await (const event of result.textEvent) {
      if (chunks === 1) {
        controller.abort();
      }
      message += event;
      chunks += 1;
    }
  } catch (e: any) {
    caughtAbortError = e.name === "AbortError";
  }
  expect(caughtAbortError).toBe(true);
  expect(message).toEqual("Hello ");
});

// This is to prove that the SDK works the same as a plain fetch
test("event stream with abort signal - plain fetch", async () => {
  const controller = new AbortController();
  const response = await fetch(`${API_TEST_SERVICE_URL}/eventstreams/text`, {
    signal: controller.signal,
    headers: { Accept: "text/event-stream" },
    method: "POST",
  });

  if (!response.body) {
    expect.fail("Expected body to be defined");
  }

  let chunks = 0;
  let message = "";
  let caughtAbortError = false;
  try {
    assertStreamHasAyncIterator(response.body);
    for await (const event of response.body) {
      if (chunks === 1) {
        controller.abort();
      }
      message += new TextDecoder().decode(event);
      chunks += 1;
    }
  } catch (e: any) {
    caughtAbortError = e.name === "AbortError";
  }

  expect(caughtAbortError).toBe(true);
  expect(message).toEqual("data: Hello\n\ndata:  \n\n");
});

function assertStreamHasAyncIterator<
  T extends ReadableStream<U> | undefined | null,
  U = Uint8Array<ArrayBufferLike>,
>(
  obj: T,
): asserts obj is T & {
  [Symbol.asyncIterator]: () => AsyncIterableIterator<U>;
} {
  if (!obj) throw new Error("ReadableStream is undefined");
  if (!(Symbol.asyncIterator in obj))
    throw new Error("ReadableStream has no async iterator");
}

test("event stream stayOpen - breaking early exits stream", async () => {
  recordTest("event-stream-stay-open-break-early");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.stayOpen();
  if (!result.textEvent) throw new Error("Expected textEvent to be defined");
  const events: string[] = [];
  for await (const e of result.textEvent) {
    events.push(e);
    break;
  }
  expect(events.length).toBe(1);

  // In Typescript, `break` automatically calls the iterator's `return()` method, which
  // cancels the underlying stream. Attempting to iterate again should yield no more events.
  for await (const e of result.textEvent) {
    events.push(e);
  }
  expect(events.length).toBe(1);
});

// This is to prove that the SDK works the same as a plain fetch
test("event stream stayOpen - breaking early exits stream - plain fetch", async () => {
  const response = await fetch(
    `${API_TEST_SERVICE_URL}/eventstreams/stayopen`,
    {
      method: "POST",
      headers: { Accept: "text/event-stream" },
    },
  );

  if (!response.body) throw new Error("Expected body to be defined");

  const events: string[] = [];
  assertStreamHasAyncIterator(response.body);
  for await (const e of response.body) {
    events.push(new TextDecoder().decode(e));
    break;
  }
  for await (const e of response.body) {
    events.push(new TextDecoder().decode(e));
    break;
  }

  expect(events.length).toBe(1);
});

test("event stream stayOpen - sentinel detection closes stream", async () => {
  recordTest("event-stream-stay-open-sentinel");

  const s = new SDK({ serverURL: HTTPBIN_URL });

  const result = await s.eventstreams.stayOpen();

  const events: string[] = [];
  if (!result.textEvent) throw new Error("Expected textEvent to be defined");
  for await (const e of result.textEvent) {
    events.push(e);
  }

  expect(events).toEqual(["event 1", "event 2", "event 3", "event 4"]);
});

test("event stream partial with comments", async () => {
  recordTest("event-stream-partial-with-comments");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.partialWithComments();

  if (result.partialWithCommentsEvent == null) {
    expect.fail("Expected partialWithCommentsEvent to be defined");
  }

  const actual: PartialWithCommentsEventData[] = [];
  for await (const event of result.partialWithCommentsEvent) {
    actual.push(event);
  }

  const expected: PartialWithCommentsEventData[] = [
    { message: "Hello from SSE" },
    { status: "processing", progress: 50 },
    { status: "complete", progress: 100, result: "Success" },
    { test: "mixed boundaries" },
    { another: "test" },
  ];

  expect(actual).toEqual(expected);
});

test("event stream union with standalone comments", async () => {
  recordTest("event-stream-union-with-standalone-comments");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.unionWithComments();

  if (result.unionWithCommentsStream == null) {
    expect.fail("Expected unionWithCommentsStream to be defined");
  }

  const events: UnionWithCommentsStream[] = [];
  for await (const event of result.unionWithCommentsStream) {
    events.push(event);
  }

  expect(events.length).toBe(4);

  expect(events).toEqual([
    {
      event: "status",
      data: { status: "started", message: "Initializing" },
    },
    {
      event: "progress",
      data: { percent: 50, detail: "Half done" },
    },
    {
      event: "status",
      data: { status: "running", message: "Processing items" },
    },
    {
      event: "progress",
      data: { percent: 100, detail: "Complete" },
    },
  ]);
});

const sseOverloadChatEvents = [
  { data: { content: "Hello" } },
  { data: { content: " " } },
  { data: { content: "world" } },
  { data: { content: "!" } },
] as const;

test("event stream sse overload with streaming response", async () => {
  recordTest("event-stream-sse-overload-streaming-response");

  const s = new SDK();
  const result: SseOverloadChatResponse & {
    chatCompletionStream: EventStream<ChatCompletionStream>;
  } = await s.eventstreams.sseOverloadChat({
    prompt: "Print test content",
    stream: true,
  });

  const events: ChatCompletionStream[] = [];
  for await (const event of result.chatCompletionStream) {
    events.push(event);
  }

  expect(events).toEqual([...sseOverloadChatEvents, { data: "[DONE]" }]);
});

test("event stream sse overload with json response", async () => {
  recordTest("event-stream-sse-overload-json-response");

  const s = new SDK();
  const result: SseOverloadChatResponse & {
    chatCompletionResult: Array<ChatCompletionResult>;
  } = await s.eventstreams.sseOverloadChat({
    prompt: "Print test content",
    stream: false,
  });

  expect(result.chatCompletionResult).toEqual(sseOverloadChatEvents);
});

test("event stream large event small chunks", async () => {
  recordTest("event-stream-large-event-small-chunks");

  const size = 2 * 1024 * 1024;
  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.largeEventSmallChunks(size);
  if (result.largeChunkedEvent == null) {
    expect.fail("Expected largeChunkedEvent to be defined");
  }

  const contents: string[] = [];
  for await (const event of result.largeChunkedEvent) {
    contents.push(event.content);
  }

  expect(contents.length).toBe(3);
  expect(contents[0]).toBe("start");
  expect(contents[2]).toBe("end");
  const big = contents[1] ?? "";
  expect(big.length).toBe(size);
  expect(big.startsWith("Sa")).toBe(true);
  expect(big.endsWith("aE")).toBe(true);
  expect(/^Sa+E$/.test(big)).toBe(true);
}, 60_000);

test("event stream split boundaries", async () => {
  recordTest("event-stream-split-boundaries");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.splitBoundaries();
  if (result.splitBoundaryEvent == null) {
    expect.fail("Expected splitBoundaryEvent to be defined");
  }

  const actualTags: string[] = [];
  let expectedTags: string[] = [];

  for await (const event of result.splitBoundaryEvent) {
    if (event.kind === "expected") {
      expectedTags = event.tags;
    } else {
      actualTags.push(...event.tags);
    }
  }

  expect(expectedTags.length).toBeGreaterThan(0);
  expect(actualTags).toEqual(expectedTags);
});
