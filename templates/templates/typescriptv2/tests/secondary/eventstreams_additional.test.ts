import { expect, test } from "vitest";
import { SDK } from "../index.js";
import { OptionalDataEvent } from "../sdk/models/shared/optionaldataevent.js";
import { SseEvent } from "../sdk/models/shared/sseevent.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

test("event stream with optional data field", async () => {
  recordTest("event-stream-optional-data-field");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.optionalData();
  if (result == null) {
    expect.fail("Expected result to be defined");
  }

  const events: OptionalDataEvent[] = [];
  for await (const event of result) {
    events.push(event);
  }

  // Server returns 5 events - some with data, some without
  // Since sseFlatResponse is false (default), we get the full event object
  // including event, id, and optional data fields
  expect(events).toEqual([
    {
      event: "message",
      data: { content: "Hello, this event has data" },
      id: "event-1",
    },
    { event: "heartbeat", id: "event-2" },
    {
      event: "message",
      data: { content: "Another message with data" },
      id: "event-3",
    },
    { event: "ping", id: "event-4" },
    { event: "complete", data: { content: "Stream finished" }, id: "event-5" },
  ]);
});

test("event stream wpt compliance", async () => {
  recordTest("event-stream-wpt-compliance");

  const s = new SDK({ serverURL: HTTPBIN_URL });
  const result = await s.eventstreams.wptCompliance();
  if (result == null) {
    expect.fail("Expected result to be defined");
  }

  const actual: SseEvent[] = [];
  let expected: SseEvent[] = [];

  for await (const event of result) {
    if (event.event === "expected") {
      expected = JSON.parse(JSON.parse(event.data));
    } else {
      actual.push(event);
    }
  }

  expect(expected.length).toBeGreaterThan(0);
  expect(actual).toEqual(expected);
});
