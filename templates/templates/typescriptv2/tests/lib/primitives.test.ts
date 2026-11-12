import { test, expect } from "vitest";
import { abortSignalAny, matchesPattern } from "../../lib/primitives.js";

test("matchesPattern tests values against the pattern", () => {
  const matches = matchesPattern("^[a-z0-9-]+$");

  expect(matches("abc-1")).toBe(true);
  expect(matches("..")).toBe(false);
  expect(matches("ABC")).toBe(false);
});

test("matchesPattern uses Unicode mode when the pattern allows it", () => {
  expect(matchesPattern("^.$")("😀")).toBe(true);
});

test("matchesPattern falls back to legacy mode", () => {
  // A class range from an escape is only a syntax error in Unicode mode.
  const matches = matchesPattern("^[\\w-]+$");

  expect(matches("a_b-c")).toBe(true);
  expect(matches("a b")).toBe(false);
});

test("matchesPattern accepts every value for an uncompilable pattern", () => {
  expect(matchesPattern("(?i)abc")("anything")).toBe(true);
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
