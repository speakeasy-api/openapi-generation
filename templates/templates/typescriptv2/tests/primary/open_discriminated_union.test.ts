import { expect, test } from "vitest";
import { z } from "zod";

import {
  discriminatedUnion,
  isUnknown,
} from "../sdk/types/discriminatedUnion.js";
import { smartUnion } from "../sdk/types/smartUnion.js";
import { recordTest } from "./common_helpers.js";

// Test: Known discriminator values parse, serialize, and validate correctly
test("discriminatedUnion - known variant handling", () => {
  recordTest("open-union-known-variant");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const union = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const catPayload = { type: "cat", name: "whiskers" };
  const catResult = union.parse(catPayload);

  expect(isUnknown(catResult)).toBe(false);
  expect(catResult).toEqual(catPayload);

  const dogPayload = { type: "dog", age: 5 };
  const dogResult = union.parse(dogPayload);

  expect(isUnknown(dogResult)).toBe(false);
  expect(dogResult).toEqual(dogPayload);
});

// Test: Unknown discriminator values produce Unknown fallback with raw payload
test("discriminatedUnion - unknown variant handling", () => {
  recordTest("open-union-unknown-discriminator");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const union = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const payload = { type: "bird", wingspan: 5 };
  const result = union.parse(payload);

  expect(isUnknown(result)).toBe(true);
  if (isUnknown(result)) {
    expect(result.type).toBe("UNKNOWN");
    expect(result.raw).toEqual(payload);
  }
});

// Test: Capture payload with missing discriminator property
test("discriminatedUnion - captures payload with missing discriminator", () => {
  recordTest("open-union-missing-discriminator");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const union = discriminatedUnion("type", {
    cat: catSchema,
  });

  // Object without the 'type' discriminator
  const payload = { name: "whiskers" };
  const result = union.parse(payload);

  expect(isUnknown(result)).toBe(true);
  if (isUnknown(result)) {
    expect(result.type).toBe("UNKNOWN");
    expect(result.raw).toEqual(payload);
  }
});

// Test: Invalid payloads (non-object, non-string discriminator, null, array) -> Unknown
test("discriminatedUnion - captures invalid payloads", () => {
  recordTest("open-union-invalid-payload");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const union = discriminatedUnion("type", {
    cat: catSchema,
  });

  // String payload instead of object
  const strResult = union.parse("not an object");
  expect(isUnknown(strResult)).toBe(true);
  if (isUnknown(strResult)) {
    expect(strResult.type).toBe("UNKNOWN");
    expect(strResult.raw).toEqual("not an object");
  }

  // Non-string discriminator
  const numDiscResult = union.parse({ type: 123, name: "whiskers" });
  expect(isUnknown(numDiscResult)).toBe(true);
  if (isUnknown(numDiscResult)) {
    expect(numDiscResult.type).toBe("UNKNOWN");
  }

  // Null payload
  const nullResult = union.parse(null);
  expect(isUnknown(nullResult)).toBe(true);
  if (isUnknown(nullResult)) {
    expect(nullResult.type).toBe("UNKNOWN");
    expect(nullResult.raw).toBeNull();
  }

  // Array payload
  const arrayResult = union.parse([1, 2, 3]);
  expect(isUnknown(arrayResult)).toBe(true);
  if (isUnknown(arrayResult)) {
    expect(arrayResult.type).toBe("UNKNOWN");
    expect(arrayResult.raw).toEqual([1, 2, 3]);
  }
});

// Test: Known discriminator but schema validation fails -> falls back to Unknown
test("discriminatedUnion - known discriminator with invalid payload falls back to Unknown", () => {
  recordTest("open-union-known-disc-invalid-schema");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
    requiredField: z.number(), // This field is required
  });

  const union = discriminatedUnion("type", {
    cat: catSchema,
  });

  // Has correct discriminator but missing requiredField
  const payload = { type: "cat", name: "whiskers" };
  const result = union.parse(payload);

  expect(isUnknown(result)).toBe(true);
  if (isUnknown(result)) {
    expect(result.type).toBe("UNKNOWN");
    expect(result.raw).toEqual(payload);
  }
});

// Test: smartUnion with discriminatedUnion (A) | string (B)
// Payload is a string, should match B
test("smartUnion - discriminatedUnion | string - string payload matches string option", () => {
  recordTest("open-union-smart-union-interop");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const discUnion = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const union = smartUnion([discUnion, z.string()] as const);

  const stringPayload = "hello world";
  const result = union.parse(stringPayload);

  expect(result).toBe(stringPayload);
  expect(typeof result).toBe("string");
});

// Test: smartUnion with discriminatedUnion (A) | string (B)
// Payload is an object with known discriminator, should match A
test("smartUnion - discriminatedUnion | string - object payload matches discriminatedUnion", () => {
  recordTest("open-union-smart-union-interop");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const discUnion = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const union = smartUnion([discUnion, z.string()] as const);

  const objectPayload = { type: "cat", name: "whiskers" };
  const result = union.parse(objectPayload);

  expect(isUnknown(result)).toBe(false);
  expect(result).toEqual(objectPayload);
});

// Test: smartUnion with discriminatedUnion (A) | { foo: string } (B)
// Payload is { foo: string }, should match B
test("smartUnion - discriminatedUnion | object - object without discriminator matches simpler object", () => {
  recordTest("open-union-smart-union-interop");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const discUnion = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const simpleObjectSchema = z.object({
    foo: z.string(),
  });

  const union = smartUnion([discUnion, simpleObjectSchema] as const);

  const simplePayload = { foo: "bar" };
  const result = union.parse(simplePayload);

  expect(result).toEqual(simplePayload);
  expect(isUnknown(result)).toBe(false);
});

// Test: smartUnion with discriminatedUnion (A) | { foo: string } (B)
// Payload has known discriminator, should match A
test("smartUnion - discriminatedUnion | object - object with discriminator matches discriminatedUnion", () => {
  recordTest("open-union-smart-union-interop");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const discUnion = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const simpleObjectSchema = z.object({
    foo: z.string(),
  });

  const union = smartUnion([discUnion, simpleObjectSchema] as const);

  const discriminatedPayload = { type: "dog", age: 3 };
  const result = union.parse(discriminatedPayload);

  expect(result).toEqual(discriminatedPayload);
  expect(isUnknown(result)).toBe(false);
});

// Test: smartUnion prefers exact match over Unknown from discriminatedUnion
test("smartUnion - prefers exact match over Unknown wrapper", () => {
  recordTest("open-union-smart-union-interop");

  const discUnion = discriminatedUnion("type", {
    cat: z.object({ type: z.literal("cat"), meow: z.boolean() }),
  });

  const valueSchema = z.object({
    value: z.number(),
  });

  const union = smartUnion([discUnion, valueSchema] as const);

  const payload = { value: 42 };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
  expect(isUnknown(result)).toBe(false);
});

// Test: Open union works when composed in parent models and lists
test("discriminatedUnion - embedded in parent object", () => {
  recordTest("open-union-embedded");

  const catSchema = z.object({
    type: z.literal("cat"),
    name: z.string(),
  });

  const dogSchema = z.object({
    type: z.literal("dog"),
    age: z.number(),
  });

  const vehicleUnion = discriminatedUnion("type", {
    cat: catSchema,
    dog: dogSchema,
  });

  const containerSchema = z.object({
    label: z.string(),
    animal: vehicleUnion,
  });

  // Known variant embedded
  const knownResult = containerSchema.parse({
    label: "my pet",
    animal: { type: "cat", name: "whiskers" },
  });
  expect(isUnknown(knownResult.animal)).toBe(false);
  expect(knownResult.animal).toEqual({ type: "cat", name: "whiskers" });

  // Unknown variant embedded
  const unknownResult = containerSchema.parse({
    label: "mystery",
    animal: { type: "bird", wingspan: 5 },
  });
  expect(isUnknown(unknownResult.animal)).toBe(true);

  // List of mixed known and unknown
  const listSchema = z.array(vehicleUnion);
  const listResult = listSchema.parse([
    { type: "cat", name: "whiskers" },
    { type: "bird", wingspan: 5 },
    { type: "dog", age: 3 },
  ]);
  expect(listResult).toHaveLength(3);
  expect(isUnknown(listResult[0]!)).toBe(false);
  expect(isUnknown(listResult[1]!)).toBe(true);
  expect(isUnknown(listResult[2]!)).toBe(false);
});
