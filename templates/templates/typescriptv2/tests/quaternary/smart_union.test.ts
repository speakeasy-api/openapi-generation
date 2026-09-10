import { expect, test } from "vitest";
import { z } from "zod";

import { smartUnion } from "../sdk/types/smartUnion.js";
import * as enums from "../sdk/types/enums.js";
import { SDK } from "../index.js";
import { HTTPBIN_URL, recordTest } from "./common_helpers.js";

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
  expect(result?.res?.json).toEqual(req);
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

test("smartUnion - selects type with more matched fields", () => {
  recordTest("smart-union-selects-more-matched-fields");
  // types: { foo: string } | { foo: string, bar: string }
  // payload: { foo: "", bar: "" }
  // expect: should match B (has more fields)

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
});

test("smartUnion - prefers fewer unmatched fields", () => {
  recordTest("smart-union-prefers-fewer-unmatched-fields");
  // types: { foo: string, bar: string } | { foo: string }
  // payload: { foo: "test" }
  // expect: should match B (fewer unmatched fields)

  const schemaA = z.object({
    foo: z.string(),
    bar: z.string(),
  });

  const schemaB = z.object({
    foo: z.string(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: "test" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - nested structs", () => {
  recordTest("smart-union-nested-structs");
  // types: { nested: { value: string } } | { nested: { value: string, extra: string } }
  // payload: { nested: { value: "test", extra: "data" } }
  // expect: should match B (more fields in nested)

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
});

test("smartUnion - preserves order on tie", () => {
  recordTest("smart-union-preserves-order-on-tie");
  // types: { foo: string } | { foo: string }
  // payload: { foo: "test" }
  // expect: first wins on tie

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
});

test("smartUnion - optional fields", () => {
  recordTest("smart-union-optional-pointer-fields");
  // types: { foo?: string } | { foo?: string, bar?: string }
  // payload: { foo: "test", bar: "value" }
  // expect: should match B (more matched fields)

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
});

test("smartUnion - optional nested structs", () => {
  recordTest("smart-union-optional-pointer-structs");
  // types: { nested?: { name: string } } | { nested?: { name: string, value: string } }
  // payload: { nested: { name: "test", value: "data" } }
  // expect: should match B (more fields in nested)

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
});

test("smartUnion - null pointer field", () => {
  // types: { bar?: string } | { foo?: string }
  // payload: { foo: null }
  // expect: should match B (has foo field)

  const schemaA = z.object({
    bar: z.string().optional().nullable(),
  });

  const schemaB = z.object({
    foo: z.string().optional().nullable(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: null };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - null nested pointer field", () => {
  // types: { foo: { bar?: string } } | { foo: { bar?: boolean } }
  // payload: { foo: { bar: null } }
  // expect: first wins on tie (both can have null bar)

  const schemaA = z.object({
    foo: z.object({
      bar: z.string().optional().nullable(),
    }),
  });

  const schemaB = z.object({
    foo: z.object({
      bar: z.boolean().optional().nullable(),
    }),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: { bar: null } };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - null non-pointer field", () => {
  // types: { foo: string } | { foo: string, bar: string }
  // payload: { foo: "", bar: null }
  // expect: should match B (has bar field even if null)

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
});

test("smartUnion - null nested field different structs", () => {
  // types: { foo: { bar: string } } | { foo: { baz?: string } }
  // payload: { foo: { baz: null } }
  // expect: should match B (has baz field)

  const schemaA = z.object({
    foo: z.object({
      bar: z.string(),
    }),
  });

  const schemaB = z.object({
    foo: z.object({
      baz: z.string().optional().nullable(),
    }),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: { baz: null } };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - null nested field both present", () => {
  // types: { foo: { bar?: string } } | { foo: { baz?: string } }
  // payload: { foo: { bar: null, baz: null } }
  // expect: first wins on tie, and Zod strips fields not in schema

  const schemaA = z.object({
    foo: z.object({
      bar: z.string().optional().nullable(),
    }),
  });

  const schemaB = z.object({
    foo: z.object({
      baz: z.string().optional().nullable(),
    }),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: { bar: null, baz: null } };
  const result = union.parse(payload);

  // schemaA wins on tie, so baz is stripped from the result
  expect(result).toEqual({ foo: { bar: null } });
});

test("smartUnion - three way field discrimination", () => {
  recordTest("smart-union-three-way-field-discrimination");
  // types: { a: string, b: string } | { a: string, c: string } | { b: string, c: string }
  // payload: { b: "", c: "" }
  // expect: should match C (has both b and c)

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
});

test("smartUnion - null payload", () => {
  // types: { foo: string } | null
  // payload: null
  // expect: should match nullable type

  const schemaA = z.object({
    foo: z.string(),
  });

  const schemaB = z.null();

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = null;
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - primitive string types", () => {
  // types: string | string
  // payload: "asdf"
  // expect: first wins on tie

  const schemaA = z.string();
  const schemaB = z.string();

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = "asdf";
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - null pointer primitives", () => {
  // types: string | null | number | null
  // payload: null
  // expect: first wins on tie

  const schemaA = z.string().nullable();
  const schemaB = z.number().nullable();

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = null;
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - null pointers match over missing fields", () => {
  // types: { a: string } | { b?: string, c?: string }
  // payload: { b: null, c: null }
  // expect: B wins (has 2 matched fields, A has 0)

  const schemaA = z.object({
    a: z.string(),
  });

  const schemaB = z.object({
    b: z.string().optional().nullable(),
    c: z.string().optional().nullable(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: null, c: null };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - map of structs", () => {
  // types: Record<string, { a: string } | null> | Record<string, { b: string } | null>
  // payload: { x: null, y: null, z: { b: "b" } }
  // expect: B wins (has matching field in nested struct)

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
});

test("smartUnion - null pointer struct vs string", () => {
  // types: { foo: string } | null | string | null
  // payload: null
  // expect: first wins on tie

  const schemaA = z
    .object({
      foo: z.string(),
    })
    .nullable();

  const schemaB = z.string().nullable();

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = null;
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - map with nested structs vs simple field", () => {
  // types: { a: string } | { b: Record<string, { id: string, name: string }> }
  // payload: { a: "", b: { foo: { id: "", name: "" } } }
  // expect: B wins (more matched fields in nested struct), Zod strips field not in schema

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

  // B wins, so 'a' is stripped from the result
  expect(result).toEqual({ b: { foo: { id: "", name: "" } } });
});

test("smartUnion - additional properties wins", () => {
  // types: { a: string } | { b: string, [key: string]: string }
  // payload: { a: "", b: "", c: "" }
  // expect: B wins (has additionalProperties to match c)

  const schemaA = z.object({
    a: z.string(),
  });

  const schemaB = z
    .object({
      b: z.string(),
    })
    .catchall(z.string());

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { a: "", b: "", c: "" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - struct vs map of strings", () => {
  // types: Record<string, string> | { id: string }
  // payload: { id: "", foo: "" }
  // Note: This test documents current behavior - map wins because it matches all fields

  const schemaA = z.record(z.string(), z.string());

  const schemaB = z.object({
    id: z.string(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { id: "", foo: "" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - additional properties vs exact field", () => {
  // types: { id?: string, [key: string]: string } | { foo: string }
  // payload: { foo: "" }
  // expect: B wins (has exact field match, A only matches via additionalProperties)

  const schemaA = z
    .object({
      id: z.string().optional(),
    })
    .catchall(z.string());

  const schemaB = z.object({
    foo: z.string(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { foo: "" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});

test("smartUnion - array of nullable structs", () => {
  // types: Array<{ foo?: string } | null> | Array<{ bar?: string } | null>
  // payload: [null, null, { bar: "" }]
  // expect: B wins (has matching field in non-null element)

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
});

test("smartUnion - any field type", () => {
  recordTest("smart-union-any-field-type");
  // types: { a: string } | { b: any }
  // payload: { b: "asdf" }
  // expect: B wins (has matching field)

  const schemaA = z.object({
    a: z.string(),
  });

  const schemaB = z.object({
    b: z.any(),
  });

  const union = smartUnion([schemaA, schemaB] as const);

  const payload = { b: "asdf" };
  const result = union.parse(payload);

  expect(result).toEqual(payload);
});
