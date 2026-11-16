import { test, expect } from "vitest";
import {
  abortSignalAny,
  matchesPattern,
  withinMaxLength,
} from "../../lib/primitives.js";

test("matchesPattern tests values against the pattern", () => {
  const matches = matchesPattern("^[a-z0-9-]+$");

  expect(matches("abc-1")).toBe(true);
  expect(matches("..")).toBe(false);
  expect(matches("ABC")).toBe(false);
});

test("matchesPattern uses Unicode mode", () => {
  expect(matchesPattern("^.$")("😀")).toBe(true);
});

test("withinMaxLength counts code points", () => {
  const within = withinMaxLength(2);

  expect(within("ab")).toBe(true);
  expect(within("😀😀")).toBe(true);
  expect(within("abc")).toBe(false);
  expect(withinMaxLength(0)("")).toBe(true);
});

test("matchesPattern skips the pattern for values over maxLength", () => {
  // Exponential backtracking on a non-matching suffix without the bound.
  const matches = matchesPattern("^(a+)+$", 8);

  expect(matches("aaaa")).toBe(true);
  expect(matches("aaab")).toBe(false);

  const start = Date.now();
  expect(matches(`${"a".repeat(64)}b`)).toBe(true);
  expect(Date.now() - start).toBeLessThan(100);
});

test("abortSignalAny with one signal", () => {
  const sig = new AbortController().signal;

  const result = abortSignalAny([sig]);

  expect(result).toBe(sig);
});

test("abortSignalAny with one aborted signal", () => {
  const controller = new AbortController();
  const sig = controller.signal;
  controller.abort("simulated signal abort");

  const result = abortSignalAny([sig]);

  expect(result).toBe(sig);
  expect(result?.reason).toBe("simulated signal abort");
});

test("abortSignalAny short circuits on first aborted signal", () => {
  const sig1 = new AbortController().signal;
  const sig2 = new AbortController().signal;

  const controller = new AbortController();
  const sig3 = controller.signal;
  controller.abort("simulated signal abort");

  const result = abortSignalAny([sig1, sig2, sig3]);

  expect(result).toBe(sig3);
  expect(result?.reason).toBe("simulated signal abort");
});

test("abortSignalAny composes signals", () => {
  const sig1 = new AbortController().signal;

  const controller = new AbortController();
  const sig2 = controller.signal;

  const sig3 = new AbortController().signal;

  const result = abortSignalAny([sig1, sig2, sig3]);
  controller.abort("simulated signal abort");

  expect(result).not.toBe(sig1);
  expect(result).not.toBe(sig2);
  expect(result).not.toBe(sig3);
  expect(result?.reason).toBe("simulated signal abort");
});
