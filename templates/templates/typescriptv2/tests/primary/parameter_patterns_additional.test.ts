import { describe, expect, test } from "vitest";

import { SDK } from "../index.js";
import { SDKValidationError } from "../sdk/models/errors/sdkvalidationerror.js";
import { CliPatternMessagesRequest } from "../sdk/models/operations/clipatternmessages.js";
import {
  HTTPBIN_URL,
  RequestLogEntry,
  newRequestRecorderClient,
} from "./common_helpers.js";

function newSDK(log: RequestLogEntry[]): SDK {
  return new SDK({
    serverURL: HTTPBIN_URL,
    httpClient: newRequestRecorderClient(log),
  });
}

async function expectRejected(
  request: CliPatternMessagesRequest,
  message: string,
): Promise<void> {
  const log: RequestLogEntry[] = [];
  const err = await newSDK(log)
    .cliHygiene.cliPatternMessages(request)
    .then(
      () => undefined,
      (e: unknown) => e,
    );

  expect(err).toBeInstanceOf(SDKValidationError);
  expect((err as SDKValidationError).pretty()).toContain(message);
  expect(log).toHaveLength(0);
}

describe("request parameter patterns", () => {
  test("sends values that match every pattern", async () => {
    const log: RequestLogEntry[] = [];

    const res = await newSDK(log).cliHygiene.cliPatternMessages({
      code: "abc-1",
      tag: "abc",
      xLabel: "abc",
      changed: "123",
      composed: "abc",
      same: "abc",
      replaced: "123",
      wrapper: "abc",
      unionCode: "123",
      unionSame: "abc",
      slug: "abc",
    });

    expect(res.httpMeta.response.status).toBe(200);
    expect(log).toHaveLength(1);
  });

  test("rejects a path parameter before sending the request", async () => {
    await expectRejected(
      { code: ".." },
      "Use lowercase letters, digits or hyphens.",
    );
  });

  test("rejects a query parameter with its custom message", async () => {
    await expectRejected(
      { code: "abc", tag: "ABC" },
      'Use "a" through "z" with 100% accuracy; café.',
    );
  });

  test("rejects a header parameter with a multi-line message", async () => {
    await expectRejected(
      { code: "abc", xLabel: "ABC" },
      'Use lowercase letters.\nTry "sample".',
    );
  });

  test("uses a default message when none is declared", async () => {
    await expectRejected(
      { code: "abc", changed: "abc" },
      "Value must match pattern ^[0-9]+$",
    );
  });

  test("rejects a string oneOf parameter against its merged pattern", async () => {
    await expectRejected(
      { code: "abc", unionCode: "ABC" },
      "Value must match pattern (^[a-z]+$|^[0-9]+$)",
    );
  });

  test("leaves non-portable patterns such as lookaheads to the server", async () => {
    const log: RequestLogEntry[] = [];

    const res = await newSDK(log).cliHygiene.cliPatternMessages({
      code: "abc",
      slug: "xyz",
    });

    expect(res.httpMeta.response.status).toBe(200);
    expect(log).toHaveLength(1);
  });

  test("leaves request body properties unconstrained", async () => {
    const log: RequestLogEntry[] = [];

    const res = await newSDK(log).cliHygiene.cliPatternMessages({
      code: "abc",
      requestBody: { name: "ABC", fallback: "ABC" },
    });

    expect(res.httpMeta.response.status).toBe(200);
    expect(log).toHaveLength(1);
  });
});
