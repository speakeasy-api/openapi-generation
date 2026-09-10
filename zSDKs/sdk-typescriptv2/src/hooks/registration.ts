import { AfterSuccessContext, AfterSuccessHook, Hooks } from "./types.js";
import { IdempotencyHook } from "./idempotency.js";

export function initHooks(hooks: Hooks) {
  hooks.registerBeforeRequestHook(new IdempotencyHook());
  hooks.registerAfterSuccessHook(new FixInputTcpJsonResponseHook());
}

class FixInputTcpJsonResponseHook implements AfterSuccessHook {
  async afterSuccess(_: AfterSuccessContext, response: Response) {
    if (!response.url.includes("/system/input")) {
      return response;
    }

    let originalData: unknown;
    try {
      // Clone before consumption so upstream callers can still read their copy if needed
      const clone = response.clone();
      originalData = await clone.json();
    } catch {
      // If parsing fails, return the original response untouched
      return response;
    }

    const fixed = deepFix(originalData);

    const headers = new Headers(response.headers);
    headers.set("content-type", "application/json");
    headers.delete("content-length");

    const newResponse = new Response(JSON.stringify(fixed), {
      status: response.status,
      statusText: response.statusText,
      headers,
    });

    return newResponse;
  }
}

function deepFix(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(deepFix);
  }
  if (isPlainObject(value)) {
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value)) {
      if (k === "connections" && Array.isArray(v)) {
        out[k] = fixConnections(v);
      } else {
        out[k] = deepFix(v);
      }
    }
    return out;
  }
  return value;
}

function fixConnections(arr: unknown[]): unknown[] {
  return arr.map((item) => {
    if (isEmptyPlainObject(item)) return { output: "" };
    return deepFix(item);
  });
}

function isPlainObject(val: unknown): val is Record<string, unknown> {
  return typeof val === "object" && val !== null;
}

function isEmptyPlainObject(val: unknown): boolean {
  return isPlainObject(val) && Object.keys(val).length === 0;
}
