import * as fs from "fs";
import { HTTPClient } from "../lib/http.js";

// Test service URLs - read from environment with defaults for local development
export const HTTPBIN_PORT = process.env["HTTPBIN_PORT"] || "35123";
export const API_TEST_SERVICE_PORT =
  process.env["API_TEST_SERVICE_PORT"] || "35456";
export const HTTPBIN_URL = `http://localhost:${HTTPBIN_PORT}`;
export const API_TEST_SERVICE_URL = `http://localhost:${API_TEST_SERVICE_PORT}`;

export function recordTest(id: string) {
  fs.appendFileSync("test-typescript-record.txt", id + "\n");
}

export async function consumeStream(
  readableStream: ReadableStream<Uint8Array>,
) {
  const reader = readableStream.getReader();

  const chunks = [];
  let totalLength = 0;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    chunks.push(value);
    totalLength += value.length;
  }

  const fullArray = new Uint8Array(totalLength);
  let position = 0;

  for (const chunk of chunks) {
    fullArray.set(chunk, position);
    position += chunk.length;
  }

  return fullArray;
}

export function randSeq(length: number) {
  const chars = "abcdefghijklmnopqrstuvwxyz";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }

  return result;
}

export interface RequestLogEntry {
  method: string;
  url: string;
  body: string;
}

export function newRequestRecorderClient(log: RequestLogEntry[]): HTTPClient {
  return new HTTPClient().addHook("beforeRequest", async (req) => {
    const body = req.body ? await req.clone().text() : "";

    log.push({
      method: req.method,
      url: req.url,
      body: body,
    });

    return req;
  });
}
