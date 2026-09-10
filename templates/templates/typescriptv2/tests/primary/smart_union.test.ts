import { expect, test, describe } from "vitest";
import { z } from "zod";
import * as types from "../sdk/types/primitives.js";

import { SDK } from "../index.js";
import { smartUnion } from "../sdk/types/smartUnion.js";
import * as enums from "../sdk/types/enums.js";

import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

// Helper to create an open enum schema that matches specific values
function openEnumSchema<T extends Record<string, string>>(values: T) {
  return enums.inboundSchema(values);
}

test("smartUnion - basic discriminator with OpenEnum", () => {
  recordTest("smart-union-open-enums");
  // types: { kind: OpenEnum["cat"] } | { kind: OpenEnum["dog"] }
  // payload: { kind: "dog" }
  // expect: should match the dog variant

  const catSchema = z.object({
    kind: enums.inboundSchema({ cat: "cat" } as const),
  });

  const dogSchema = z.object({
    kind: enums.inboundSchema({ dog: "dog" } as const),
  });

  const union = smartUnion([catSchema, dogSchema] as const);

  const payload = { kind: "dog" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - discriminator with additional field and unrecognized enum", async () => {
  recordTest("smart-union-open-enums-and-size");
  // types: { kind: OpenEnum["cat"], name: string } | { kind: OpenEnum["dog"] }
  // payload: { kind: "bat", name: "asdf" }
  // expect: should match cat variant (has name field)

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const req = { kind: "bat", name: "asdf" } as any;

  const result = await sdk.unions.smartUnionOpenEnumsAndSize(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - array with more fields wins", () => {
  recordTest("smart-union-deeply-nested-array");
  // types: Array<{ x: { a: string } }> | Array<{ x: { a: string; b: boolean } }>
  // payload: [{ x: { a: "", b: false } }]
  // expect: should match the second variant (has more fields)

  const arraySchemaA = z.array(
    z.object({
      x: z.object({
        a: z.string(),
      }),
    }),
  );

  const arraySchemaB = z.array(
    z.object({
      x: z.object({
        a: z.string(),
        b: z.boolean(),
      }),
    }),
  );

  const union = smartUnion([arraySchemaA, arraySchemaB] as const);

  const payload = [{ x: { a: "", b: false } }];
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - simple object field discriminator", () => {
  recordTest("smart-union-empty-string");
  // types: { a: string } | { b: string }
  // payload: { b: "" }
  // expect: should match the second variant (has b field)

  const schemaA = z.object({
    a: z.string(),
  });

  const schemaB = z.object({
    b: z.string(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: "" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - lax mode", () => {
  // types: { a: optional(string) } | { b: nullable(string) }
  // payload: { }
  // expect: { a: undefined }

  const schemaA = z.object({
    a: types.optional(types.string()),
  });

  const schemaB = z.object({
    b: types.nullable(types.string()),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: null };
  const result = union.parse(payload);

  expect(result).toEqual({ b: null });
});

test("smartUnion - lax mode - getting zero default undefined should be counted lower than exact", () => {
  const schemaA = z.object({
    a: types.string(),
  });

  const schemaB = z.object({
    b: types.optional(types.string()),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: undefined };
  const result = union.parse(payload);

  expect(result).toEqual({ b: undefined });
});

test("smartUnion - lax mode - getting real string should be preferred to zero default string", () => {
  const schemaA = z.object({
    a: types.string(),
  });

  const schemaB = z.object({
    b: types.string(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: "" };
  const result = union.parse(payload);

  expect(result).toEqual({ b: "" });
});

test("smartUnion - lax mode - getting string should be preferred to coerced number", () => {
  const schemaA = z.object({
    a: types.string(),
  });

  const schemaB = z.object({
    a: types.number(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { a: "1" };
  const result = union.parse(payload);

  expect(result).toEqual({ a: "1" });
});

test("smartUnion - lax mode - exact match preferred over inexact with more fields", () => {
  // Schema A should win - exact match beats schema B which has zero-default for missing 'b'
  const schemaA = z.object({ a: types.string() });
  const schemaB = z.object({ a: types.string(), b: types.string() });

  const union = smartUnion([schemaA, schemaB] as const);
  const result = union.parse({ a: "hello" });

  expect(result).toEqual({ a: "hello" });
});

test("smartUnion - lax mode - object coerced to string should lose to exact string match", () => {
  // Schema B should win - exact object match beats string coercion via JSON.stringify
  const schemaA = z.object({ a: types.string() });
  const schemaB = z.object({ a: z.object({}).passthrough() });

  const union = smartUnion([schemaA, schemaB] as const);
  const result = union.parse({ a: {} });

  expect(result).toEqual({ a: {} });
});

test("smartUnion - lax mode - actually null should be preferred to defaulting to null", () => {
  const schemaA = z.object({
    a: types.optional(types.string()),
  });

  const schemaB = z.object({
    b: types.nullable(types.string()),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { a: null };
  const result = union.parse(payload);

  expect(result).toEqual({ a: undefined });
});

test("smartUnion - nested union should only count winning inner option's unrecognized", () => {
  recordTest("smart-union-nested-union");
  // Bug: Without proper count propagation, nested unions accumulate unrecognized
  // counts from ALL inner options tried, not just the winning one.
  // This causes the parent union to incorrectly prefer simpler options.
  //
  // Test structure:
  // OuterUnion:
  //   - Option A: { data: InnerUnion } where InnerUnion has 3 options with open enums
  //   - Option B: { data: { kind: OpenEnum, name: OpenEnum } } has 2 open enum fields
  //
  // InnerUnion (inside Option A):
  //   - { kind: OpenEnum["cat"], name: string } - 2 fields, wins for payload with name
  //   - { kind: OpenEnum["dog"] } - 1 field
  //   - { kind: OpenEnum["bird"] } - 1 field
  //
  // Payload: { data: { kind: "unknown", name: "also_unknown" } }
  //
  // Expected: Option A should win because:
  //   - Inner union picks first variant (has name field), counts 1 unrecognized (kind="unknown")
  //   - Option B would count 2 unrecognized (both kind and name are unknown enum values)
  //   - With correct counting: A=1 < B=2, so A wins
  //   - With buggy counting: A=3 (accumulated from all inner options) > B=2, so B would wrongly win
  //
  // To verify which option won, we add a marker field via transform

  // Inner union: 3 options with open enums
  const innerUnion = smartUnion([
    z.object({
      kind: enums.inboundSchema({ cat: "cat" } as const),
      name: z.string(),
    }),
    z.object({ kind: enums.inboundSchema({ dog: "dog" } as const) }),
    z.object({ kind: enums.inboundSchema({ bird: "bird" } as const) }),
  ] as const);

  // Outer union with markers to identify which option was selected
  const outerUnion = smartUnion([
    z.pipe(
      z.object({ data: innerUnion }),
      z.transform((v) => ({ ...v, marker: "a" as const })),
    ),
    z.pipe(
      z.object({
        data: z.object({
          kind: enums.inboundSchema({ x: "x" } as const),
          name: enums.inboundSchema({ y: "y" } as const),
        }),
      }),
      z.transform((v) => ({ ...v, marker: "b" as const })),
    ),
  ] as const);

  // Payload with unknown enum values
  const payload = { data: { kind: "unknown", name: "also_unknown" } };
  const result = outerUnion.parse(payload) as { data: unknown; marker: string };

  // With correct counting: A has 1 unrecognized, B has 2 -> A should win (marker = "a")
  // With buggy counting: A has 3 unrecognized (accumulated), B has 2 -> B would wrongly win (marker = "b")
  expect(result.marker).toEqual("a");
});

// ============================================================================
// Tests ported from Go union_test.go
// ============================================================================

describe("smartUnion - ported from Go tests", () => {
  test("selects type with more matched fields", () => {
    recordTest("smart-union-selects-more-matched-fields");
    // Go: TestPickBestUnionCandidate_SelectsTypeWithMoreMatchedFields
    const schemaA = z.object({
      foo: z.string(),
    });
    const schemaB = z.object({
      foo: z.string(),
      bar: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "", bar: "" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has more matched fields
  });

  test("prefers fewer unmatched fields", () => {
    recordTest("smart-union-prefers-fewer-unmatched-fields");
    // Go: TestPickBestUnionCandidate_PrefersFewerUnmatchedFields
    const schemaA = z.object({
      foo: z.string(),
      bar: z.string().optional(),
    });
    const schemaB = z.object({
      foo: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "test" };
    const result = union.parse(payload);

    expect(result).toEqual({ foo: "test" });
    // B should win because it has fewer unmatched fields
  });

  test("nested structs - deeper match wins", () => {
    recordTest("smart-union-nested-structs");
    // Go: TestPickBestUnionCandidate_NestedStructs
    const schemaA = z.object({
      nested: z.object({
        value: z.string(),
      }),
    });
    const schemaB = z.object({
      nested: z.object({
        value: z.string(),
        extra: z.string(),
      }),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { nested: { value: "test", extra: "data" } };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because nested has more fields
  });

  test("array fields - more fields in items wins", () => {
    recordTest("smart-union-array-fields");
    // Go: TestPickBestUnionCandidate_ArrayFields
    const schemaA = z.array(
      z.object({
        name: z.string(),
      }),
    );
    const schemaB = z.array(
      z.object({
        name: z.string(),
        value: z.string(),
      }),
    );

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = [
      { name: "a", value: "1" },
      { name: "b", value: "2" },
    ];
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because items have more fields
  });

  test("preserves order on tie - first wins", () => {
    recordTest("smart-union-preserves-order-on-tie");
    // Go: TestPickBestUnionCandidate_PreservesOrderOnTie
    const schemaA = z.object({
      foo: z.string(),
    });
    const schemaB = z.object({
      foo: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "test" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // A should win on tie (first wins)
  });

  test("optional pointer fields", () => {
    recordTest("smart-union-optional-pointer-fields");
    // Go: TestPickBestUnionCandidate_OptionalPointerFields
    const schemaA = z.object({
      foo: z.string().optional(),
    });
    const schemaB = z.object({
      foo: z.string().optional(),
      bar: z.string().optional(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "test", bar: "value" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has more matched fields
  });

  test("optional pointer structs", () => {
    recordTest("smart-union-optional-pointer-structs");
    // Go: TestPickBestUnionCandidate_OptionalPointerStructs
    const schemaA = z.object({
      nested: z
        .object({
          name: z.string(),
        })
        .optional(),
    });
    const schemaB = z.object({
      nested: z
        .object({
          name: z.string(),
          value: z.string(),
        })
        .optional(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { nested: { name: "test", value: "data" } };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because nested struct has more fields
  });

  test("null pointer field - matches nullable schema", () => {
    // Go: TestPickBestUnionCandidate_NullPointerField
    const schemaA = z.object({
      bar: z.string().nullable().optional(),
    });
    const schemaB = z.object({
      foo: z.string().nullable().optional(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: null };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has the foo field
  });

  test("null nested pointer field - first wins on tie", () => {
    // Go: TestPickBestUnionCandidate_NullNestedPointerField
    const schemaA = z.object({
      foo: z.object({
        bar: z.string().nullable().optional(),
      }),
    });
    const schemaB = z.object({
      foo: z.object({
        bar: z.boolean().nullable().optional(),
      }),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: { bar: null } };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // A should win on tie (first wins)
  });

  test("null non-pointer field", () => {
    // Go: TestPickBestUnionCandidate_NullNonPointerField
    const schemaA = z.object({
      foo: z.string(),
    });
    const schemaB = z.object({
      foo: z.string(),
      bar: z.string().nullable(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "", bar: null };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has more matched fields
  });

  test("null nested field different structs", () => {
    // Go: TestPickBestUnionCandidate_NullNestedFieldDifferentStructs
    const schemaA = z.object({
      foo: z.object({
        bar: z.string(),
      }),
    });
    const schemaB = z.object({
      foo: z.object({
        baz: z.string().nullable().optional(),
      }),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: { baz: null } };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it matches the baz field
  });

  test("null nested field both present - first wins", () => {
    // Go: TestPickBestUnionCandidate_NullNestedFieldBothPresent
    const schemaA = z.object({
      foo: z.object({
        bar: z.string().nullable().optional(),
      }),
    });
    const schemaB = z.object({
      foo: z.object({
        baz: z.string().nullable().optional(),
      }),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: { bar: null, baz: null } };
    // Both match one field, A should win on tie
    const result = union.parse(payload);

    expect(result).toEqual({ foo: { bar: null } });
  });

  test("enum discrimination", () => {
    // Go: TestPickBestUnionCandidate_EnumDiscrimination
    // EnumA matches 1 or 2, EnumB matches 3 or 4
    const schemaA = z.object({
      a: openEnumSchema({ "1": "1", "2": "2" } as const),
    });
    const schemaB = z.object({
      a: openEnumSchema({ "3": "3", "4": "4" } as const),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { a: "4" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because 4 is an exact match for EnumB
  });

  test("three-way field discrimination", () => {
    recordTest("smart-union-three-way-field-discrimination");
    // Go: TestPickBestUnionCandidate_ThreeWayFieldDiscrimination
    const schemaA = z.object({
      a: z.string(),
      b: z.string(),
    });
    const schemaB = z.object({
      a: z.string(),
      c: z.string(),
    });
    const schemaC = z.object({
      b: z.string(),
      c: z.string(),
    });

    const union = smartUnion([schemaA, schemaB, schemaC] as const);
    const payload = { b: "", c: "" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // C should win because it matches both b and c
  });

  test("const field discrimination", () => {
    recordTest("smart-union-const-field-discrimination");
    // Go: TestPickBestUnionCandidate_ConstFieldDiscrimination
    // Using literal types to simulate const fields
    const schemaA = z.object({
      a: z.literal("x"),
      b: z.literal("1"),
    });
    const schemaB = z.object({
      a: z.literal("x"),
      c: z.literal("1"),
    });
    const schemaC = z.object({
      b: z.literal("1"),
      c: z.literal("1"),
    });

    const union = smartUnion([schemaA, schemaB, schemaC] as const);
    const payload = { b: "1", c: "1" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // C should win because it matches both b and c with correct const values
  });

  test("const field discrimination with inexact values", () => {
    // Go: TestPickBestUnionCandidate_ConstFieldDiscriminationInexactValues
    // Using open enums to simulate const fields that can accept inexact values
    // payload has fields b and c, but with wrong values
    // A: matched=1 (b exists), inexact=1 (b!="1"), unmatched=1 (a missing)
    // B: matched=1 (c exists), inexact=1 (c!="1"), unmatched=1 (a missing)
    // C: matched=2 (b,c exist), inexact=2 (b!="1", c!="1"), unmatched=0
    // C wins because it has more matched fields
    const schemaA = z.object({
      a: openEnumSchema({ x: "x" } as const),
      b: openEnumSchema({ "1": "1" } as const),
    });
    const schemaB = z.object({
      a: openEnumSchema({ x: "x" } as const),
      c: openEnumSchema({ "1": "1" } as const),
    });
    const schemaC = z.object({
      b: openEnumSchema({ "1": "1" } as const),
      c: openEnumSchema({ "1": "1" } as const),
    });

    const union = smartUnion([schemaA, schemaB, schemaC] as const);
    const payload = { b: "hey", c: "ho" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // C should win because it has more matched fields (2 vs 1)
  });

  test("null payload - nullable type wins", () => {
    // Go: TestPickBestUnionCandidate_NullPayload
    const schemaA = z.object({
      foo: z.string(),
    });
    const schemaB = z.null();

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = null;
    const result = union.parse(payload);

    expect(result).toBeNull();
    // B (nullable type) should win for null payload
  });

  test("primitive string types - first wins on tie", () => {
    // Go: TestPickBestUnionCandidate_PrimitiveStringTypes
    const schemaA = z.string();
    const schemaB = z.string();

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = "asdf";
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // A should win on tie (first wins)
  });

  test("null pointer primitives - first wins on tie", () => {
    // Go: TestPickBestUnionCandidate_NullPointerPrimitives
    const schemaA = z.string().nullable();
    const schemaB = z.number().nullable();

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = null;
    const result = union.parse(payload);

    expect(result).toBeNull();
    // A should win on tie (first wins)
  });

  test("null pointers match over missing fields", () => {
    // Go: TestPickBestUnionCandidate_NullPointersMatchOverMissingFields
    const schemaA = z.object({
      a: z.string(),
    });
    const schemaB = z.object({
      b: z.string().nullable().optional(),
      c: z.string().nullable().optional(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { b: null, c: null };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has 2 matched fields (even though null), A has 0
  });

  test("map of structs - matching inner field wins", () => {
    // Go: TestPickBestUnionCandidate_MapOfStructs
    const schemaA = z.record(
      z.string(),
      z
        .object({
          a: z.string(),
        })
        .nullable(),
    );
    const schemaB = z.record(
      z.string(),
      z
        .object({
          b: z.string(),
        })
        .nullable(),
    );

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { x: null, y: null, z: { b: "b" } };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has a matched field in the nested struct
  });

  test("null pointer struct vs string - first wins on null", () => {
    // Go: TestPickBestUnionCandidate_NullPointerStructVsString
    const schemaA = z
      .object({
        foo: z.string(),
      })
      .nullable();
    const schemaB = z.string().nullable();

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = null;
    const result = union.parse(payload);

    expect(result).toBeNull();
    // A should win on tie (first wins)
  });

  test("fewer unmatched fields wins", () => {
    // Go: TestPickBestUnionCandidate_FewerUnmatchedFieldsWins
    const schemaA = z.object({
      a: z.string(),
      b: z.string().optional(),
    });
    const schemaB = z.object({
      a: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { a: "" };
    const result = union.parse(payload);

    expect(result).toEqual({ a: "" });
    // B should win because it has fewer unmatched fields
  });

  test("map with nested structs vs simple field", () => {
    // Go: TestPickBestUnionCandidate_MapWithNestedStructsVsSimpleField
    // A: { a: string }
    // B: { b: map[string]{ id: string, name: string } }
    // payload: { "a": "", "b": { "foo": { "id": "", "name": "" } } }
    // B should win because it has more matched fields (id, name in nested struct)
    const schemaA = z.object({
      a: z.string(),
    });
    const schemaB = z.object({
      b: z.record(
        z.string(),
        z.object({
          id: z.string(),
          name: z.string(),
        }),
      ),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { a: "", b: { foo: { id: "", name: "" } } };
    const result = union.parse(payload);

    // B should win because it has more matched fields (id, name in nested struct)
    // zod strips fields not in the winning schema, so result won't have 'a'
    expect(result).toEqual({ b: { foo: { id: "", name: "" } } });
  });

  test("additional properties wins over simple struct", () => {
    // Go: TestPickBestUnionCandidate_AdditionalPropertiesWins
    // A: { id: string }
    // B: { id: string, additionalProperties: true }
    // payload: { "id": "", "foo": "" }
    // B should win because it captures the additional "foo" field
    const schemaA = z.object({
      id: z.string(),
    });
    const schemaB = z
      .object({
        id: z.string(),
      })
      .passthrough();

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { id: "", foo: "" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it captures the additional "foo" field
  });

  test("struct vs map of strings", () => {
    // Go: TestPickBestUnionCandidate_StructVsMapOfStrings
    // A: map[string]string
    // B: { id: string }
    // payload: { "id": "", "foo": "" }
    // B should win because struct fields are more specific than map
    const schemaA = z.record(z.string(), z.string());
    const schemaB = z.object({
      id: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { id: "", foo: "" };
    const result = union.parse(payload);

    // B should win because struct fields are more specific than map
    expect(result).toEqual(payload);
  });

  test("additional properties vs exact field", () => {
    // Go: TestPickBestUnionCandidate_AdditionalPropertiesVsExactField
    // A: { id: *string, additionalProperties: true }
    // B: { foo: string }
    // payload: { "foo": "" }
    // B should win because it has an exact field match, while A only matches via additionalProperties
    const schemaA = z
      .object({
        id: z.string().optional(),
      })
      .passthrough();
    const schemaB = z.object({
      foo: z.string(),
    });

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = { foo: "" };
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has an exact field match
  });

  test("array of nullable structs", () => {
    // Go: TestPickBestUnionCandidate_ArrayOfNullableStructs
    // A: Array<{ foo: *string } | null>
    // B: Array<{ bar: *string } | null>
    // payload: [null, null, { "bar": "" }]
    // B should win because it has a matching field in the non-null element
    const schemaA = z.array(
      z
        .object({
          foo: z.string().optional(),
        })
        .nullable(),
    );
    const schemaB = z.array(
      z
        .object({
          bar: z.string().optional(),
        })
        .nullable(),
    );

    const union = smartUnion([schemaA, schemaB] as const);
    const payload = [null, null, { bar: "" }];
    const result = union.parse(payload);

    expect(result).toEqual(payload);
    // B should win because it has a matching field in the non-null element
  });
});

// ============================================================================
// SDK-call tests for smart union operations
// ============================================================================

test("smartUnion - all consts", async () => {
  recordTest("smart-union-all-consts");
  // Types: { a: const "A", b: const "B" } | { b: const "B", c: const "C" }
  // Payload: { b: "B", c: "C" }
  // Expect: second variant (matches both b and c)

  const sdk = new SDK();
  const req = { b: "B", c: "C" } as any;

  const result = await sdk.unions.smartUnionAllConsts(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - any field type", async () => {
  recordTest("smart-union-any-field-type");
  // Types: { a: optional string } | { b: any }
  // Payload: { b: "asdf" }
  // Expect: second variant (field b matches)

  const sdk = new SDK();
  const req = { b: "asdf" } as any;

  const result = await sdk.unions.smartUnionAnyFieldType(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - nested union vs flat struct", async () => {
  recordTest("smart-union-nested-union-vs-flat-struct");
  // Outer: { data: Union[x-only | y-only | z-only] } | { data: { x: string, y: string } }
  // Payload: { data: { x: "", y: "" } }
  // Expect: flat struct (matches 2 fields vs nested union's 1)

  const sdk = new SDK();
  const req = { data: { x: "", y: "" } } as any;

  const result = await sdk.unions.smartUnionNestedUnionVsFlatStruct(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("smartUnion - union vs union", async () => {
  recordTest("smart-union-union-vs-union");
  // Outer: Union[{ a } | { b }] | Union[{ a, b } | { a, b, c }]
  // Payload: { a: "", b: "", c: "" }
  // Expect: second union's second variant (matches all 3 fields)

  const sdk = new SDK();
  const req = { a: "", b: "", c: "" } as any;

  const result = await sdk.unions.smartUnionUnionVsUnion(req);

  expect(result.httpMeta.response.status).toEqual(200);

  if (result.res == null) {
    expect.unreachable("Server should have returned a valid response");
  }

  expect(result.res.json).toEqual(req);
});

test("unions - discriminated open enum", async () => {
  recordTest("unions-discriminated-open-enum");
  // Discriminator on "status" field with open enum values
  // active/pending -> Status1, inactive/archived -> Status2

  const sdk = new SDK();

  // active status -> Status1
  const reqActive = {
    status: "active",
    userId: "user-123",
    activeAt: new Date("2024-01-15T10:30:00Z"),
  } as any;

  const resActive = await sdk.unions.discriminatedOpenEnum(reqActive);
  expect(resActive.httpMeta.response.status).toEqual(200);
  expect(resActive.res?.json).toBeTruthy();
  // Verify active variant fields
  const activeJson = resActive.res!.json as any;
  expect(activeJson.status).toEqual("active");
  expect(activeJson.userId).toEqual("user-123");

  // inactive status -> Status2
  const reqInactive = {
    status: "inactive",
    reason: "User requested deactivation",
    inactiveSince: new Date("2024-01-10T15:45:00Z"),
  } as any;

  const resInactive = await sdk.unions.discriminatedOpenEnum(reqInactive);
  expect(resInactive.httpMeta.response.status).toEqual(200);
  expect(resInactive.res?.json).toBeTruthy();
  // Verify inactive variant fields
  const inactiveJson = resInactive.res!.json as any;
  expect(inactiveJson.status).toEqual("inactive");
  expect(inactiveJson.reason).toEqual("User requested deactivation");
});
