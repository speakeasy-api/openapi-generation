package consts

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestSnapTsConstRespectRequired(t *testing.T) {
	t.Parallel()
	spec := `openapi: 3.1.0
info:
  title: Const Values API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: testConst
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ConstModel'
      responses:
        '200':
          description: OK
components:
  schemas:
    ConstModel:
      type: object
      required:
        - requiredConstString
        - requiredConstNumber
        - requiredConstBoolean
        - requiredConstEnum
        - requiredConstDate
        - requiredConstDateTime
        - requiredConstBigint
        - requiredConstDecimal
        - requiredDefaultString
        - requiredDefaultNumber
        - requiredDefaultBoolean
        - requiredDefaultEnum
        - requiredDefaultDate
        - requiredDefaultDateTime
        - requiredDefaultBigint
        - requiredDefaultDecimal
        - requiredConstDefaultString
        - requiredConstDefaultNumber
        - requiredConstDefaultBoolean
        - requiredConstDefaultEnum
        - requiredConstDefaultDate
        - requiredConstDefaultDateTime
        - requiredConstDefaultBigint
        - requiredConstDefaultDecimal
      properties:
        requiredConstString:
          type: string
          const: fixed-string
        requiredConstNumber:
          type: number
          const: 42.5
        requiredConstBoolean:
          type: boolean
          const: true
        requiredConstEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
        requiredConstDate:
          type: string
          format: date
          const: "2023-12-25"
        requiredConstDateTime:
          type: string
          format: date-time
          const: "2023-12-25T10:30:00Z"
        requiredConstBigint:
          type: integer
          format: bigint
          const: 9007199254740991
        requiredConstDecimal:
          type: number
          format: decimal
          const: 123.456789
        optionalConstString:
          type: string
          const: "optional-const-string"
        optionalConstNumber:
          type: number
          const: 123.45
        optionalConstBoolean:
          type: boolean
          const: false
        optionalConstEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
        optionalConstDate:
          type: string
          format: date
          const: "2025-06-15"
        optionalConstDateTime:
          type: string
          format: date-time
          const: "2025-06-15T15:30:00Z"
        optionalConstBigint:
          type: integer
          format: bigint
          const: 987654321098765432
        optionalConstDecimal:
          type: number
          format: decimal
          const: 456.123789
        requiredDefaultString:
          type: string
          default: "default-string"
        requiredDefaultNumber:
          type: number
          default: 99.9
        requiredDefaultBoolean:
          type: boolean
          default: false
        requiredDefaultEnum:
          type: string
          enum: ["A", "B"]
          default: "A"
        requiredDefaultDate:
          type: string
          format: date
          default: "2024-01-01"
        requiredDefaultDateTime:
          type: string
          format: date-time
          default: "2024-01-01T00:00:00Z"
        requiredDefaultBigint:
          type: integer
          format: bigint
          default: 1234567890123456789
        requiredDefaultDecimal:
          type: number
          format: decimal
          default: 999.999999
        optionalDefaultString:
          type: string
          default: "optional-default"
        optionalDefaultNumber:
          type: number
          default: 77.7
        optionalDefaultBoolean:
          type: boolean
          default: true
        optionalDefaultEnum:
          type: string
          enum: ["A", "B"]
          default: "A"
        optionalDefaultDate:
          type: string
          format: date
          default: "2024-12-31"
        optionalDefaultDateTime:
          type: string
          format: date-time
          default: "2024-12-31T23:59:59Z"
        optionalDefaultBigint:
          type: integer
          format: bigint
          default: 555666777888999
        optionalDefaultDecimal:
          type: number
          format: decimal
          default: 111.222333
        optionalString:
          type: string
        optionalNumber:
          type: number
        optionalBoolean:
          type: boolean
        optionalEnum:
          type: string
          enum: ["A", "B"]
        optionalDate:
          type: string
          format: date
        optionalDateTime:
          type: string
          format: date-time
        optionalBigint:
          type: integer
          format: bigint
        optionalDecimal:
          type: number
          format: decimal
        requiredConstDefaultString:
          type: string
          const: "fixed-string"
          default: "fixed-string"
        requiredConstDefaultNumber:
          type: number
          const: 42.5
          default: 42.5
        requiredConstDefaultBoolean:
          type: boolean
          const: true
          default: true
        requiredConstDefaultEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
          default: "A"
        requiredConstDefaultDate:
          type: string
          format: date
          const: "2023-12-25"
          default: "2023-12-25"
        requiredConstDefaultDateTime:
          type: string
          format: date-time
          const: "2023-12-25T10:30:00Z"
          default: "2023-12-25T10:30:00Z"
        requiredConstDefaultBigint:
          type: integer
          format: bigint
          const: 9007199254740991
          default: 9007199254740991
        requiredConstDefaultDecimal:
          type: number
          format: decimal
          const: 123.456789
          default: 123.456789
        optionalConstDefaultString:
          type: string
          const: "optional-const-string"
          default: "optional-const-string"
        optionalConstDefaultNumber:
          type: number
          const: 123.45
          default: 123.45
        optionalConstDefaultBoolean:
          type: boolean
          const: false
          default: false
        optionalConstDefaultEnum:
          type: string
          enum: ["A", "B"]
          const: "A"
          default: "A"
        optionalConstDefaultDate:
          type: string
          format: date
          const: "2025-06-15"
          default: "2025-06-15"
        optionalConstDefaultDateTime:
          type: string
          format: date-time
          const: "2025-06-15T15:30:00Z"
          default: "2025-06-15T15:30:00Z"
        optionalConstDefaultBigint:
          type: integer
          format: bigint
          const: 987654321098765432
          default: 987654321098765432
        optionalConstDefaultDecimal:
          type: number
          format: decimal
          const: 456.123789
          default: 456.123789`

	genYaml := `typescript:
  packageName: consttest
  constFieldsAlwaysOptional: false
`

	// expectedSnapshotFiles defines which generated files to include in the snapshot
	expectedSnapshotFiles := []string{
		"src/models/const-model.ts",
	}

	expectedSnapshot := `--- src/models/const-model.ts ---
/*
 * Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.
 * Generated under the AGPL-3.0-only license.
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import * as z from "zod/v4-mini";
import { Decimal as Decimal$ } from "../types/decimal.js";
import { ClosedEnum } from "../types/enums.js";

export const RequiredConstEnum = {
  A: "A",
  B: "B",
} as const;
export type RequiredConstEnum = ClosedEnum<typeof RequiredConstEnum>;

export const OptionalConstEnum = {
  A: "A",
  B: "B",
} as const;
export type OptionalConstEnum = ClosedEnum<typeof OptionalConstEnum>;

export const RequiredDefaultEnum = {
  A: "A",
  B: "B",
} as const;
export type RequiredDefaultEnum = ClosedEnum<typeof RequiredDefaultEnum>;

export const OptionalDefaultEnum = {
  A: "A",
  B: "B",
} as const;
export type OptionalDefaultEnum = ClosedEnum<typeof OptionalDefaultEnum>;

export const OptionalEnum = {
  A: "A",
  B: "B",
} as const;
export type OptionalEnum = ClosedEnum<typeof OptionalEnum>;

export const RequiredConstDefaultEnum = {
  A: "A",
  B: "B",
} as const;
export type RequiredConstDefaultEnum = ClosedEnum<
  typeof RequiredConstDefaultEnum
>;

export const OptionalConstDefaultEnum = {
  A: "A",
  B: "B",
} as const;
export type OptionalConstDefaultEnum = ClosedEnum<
  typeof OptionalConstDefaultEnum
>;

export type ConstModel = {
  requiredConstString: "fixed-string";
  requiredConstNumber: 42.5;
  requiredConstBoolean: true;
  requiredConstEnum: "A";
  requiredConstDate: Date;
  requiredConstDateTime: Date;
  requiredConstBigint: 9007199254740991;
  requiredConstDecimal: Decimal$ | number;
  optionalConstString?: "optional-const-string" | undefined;
  optionalConstNumber?: 123.45 | undefined;
  optionalConstBoolean?: false | undefined;
  optionalConstEnum?: "A" | undefined;
  optionalConstDate?: Date | undefined;
  optionalConstDateTime?: Date | undefined;
  optionalConstBigint?: 987654321098765400 | undefined;
  optionalConstDecimal?: Decimal$ | number | undefined;
  requiredDefaultString?: string | undefined;
  requiredDefaultNumber?: number | undefined;
  requiredDefaultBoolean?: boolean | undefined;
  requiredDefaultEnum?: RequiredDefaultEnum | undefined;
  requiredDefaultDate?: Date | undefined;
  requiredDefaultDateTime?: Date | undefined;
  requiredDefaultBigint?: number | undefined;
  requiredDefaultDecimal?: Decimal$ | number | undefined;
  optionalDefaultString?: string | undefined;
  optionalDefaultNumber?: number | undefined;
  optionalDefaultBoolean?: boolean | undefined;
  optionalDefaultEnum?: OptionalDefaultEnum | undefined;
  optionalDefaultDate?: Date | undefined;
  optionalDefaultDateTime?: Date | undefined;
  optionalDefaultBigint?: number | undefined;
  optionalDefaultDecimal?: Decimal$ | number | undefined;
  optionalString?: string | undefined;
  optionalNumber?: number | undefined;
  optionalBoolean?: boolean | undefined;
  optionalEnum?: OptionalEnum | undefined;
  optionalDate?: Date | undefined;
  optionalDateTime?: Date | undefined;
  optionalBigint?: number | undefined;
  optionalDecimal?: Decimal$ | number | undefined;
  requiredConstDefaultString?: "fixed-string" | undefined;
  requiredConstDefaultNumber?: 42.5 | undefined;
  requiredConstDefaultBoolean?: true | undefined;
  requiredConstDefaultEnum?: "A" | undefined;
  requiredConstDefaultDate?: Date | undefined;
  requiredConstDefaultDateTime?: Date | undefined;
  requiredConstDefaultBigint?: 9007199254740991 | undefined;
  requiredConstDefaultDecimal?: Decimal$ | number | undefined;
  optionalConstDefaultString?: "optional-const-string" | undefined;
  optionalConstDefaultNumber?: 123.45 | undefined;
  optionalConstDefaultBoolean?: false | undefined;
  optionalConstDefaultEnum?: "A" | undefined;
  optionalConstDefaultDate?: Date | undefined;
  optionalConstDefaultDateTime?: Date | undefined;
  optionalConstDefaultBigint?: 987654321098765400 | undefined;
  optionalConstDefaultDecimal?: Decimal$ | number | undefined;
};

/** @internal */
export const RequiredConstEnum$outboundSchema: z.ZodMiniEnum<
  typeof RequiredConstEnum
> = z.enum(RequiredConstEnum);

/** @internal */
export const OptionalConstEnum$outboundSchema: z.ZodMiniEnum<
  typeof OptionalConstEnum
> = z.enum(OptionalConstEnum);

/** @internal */
export const RequiredDefaultEnum$outboundSchema: z.ZodMiniEnum<
  typeof RequiredDefaultEnum
> = z.enum(RequiredDefaultEnum);

/** @internal */
export const OptionalDefaultEnum$outboundSchema: z.ZodMiniEnum<
  typeof OptionalDefaultEnum
> = z.enum(OptionalDefaultEnum);

/** @internal */
export const OptionalEnum$outboundSchema: z.ZodMiniEnum<typeof OptionalEnum> = z
  .enum(OptionalEnum);

/** @internal */
export const RequiredConstDefaultEnum$outboundSchema: z.ZodMiniEnum<
  typeof RequiredConstDefaultEnum
> = z.enum(RequiredConstDefaultEnum);

/** @internal */
export const OptionalConstDefaultEnum$outboundSchema: z.ZodMiniEnum<
  typeof OptionalConstDefaultEnum
> = z.enum(OptionalConstDefaultEnum);

/** @internal */
export type ConstModel$Outbound = {
  requiredConstString: "fixed-string";
  requiredConstNumber: 42.5;
  requiredConstBoolean: true;
  requiredConstEnum: "A";
  requiredConstDate: string;
  requiredConstDateTime: string;
  requiredConstBigint: 9007199254740991;
  requiredConstDecimal: number;
  optionalConstString?: "optional-const-string" | undefined;
  optionalConstNumber?: 123.45 | undefined;
  optionalConstBoolean?: false | undefined;
  optionalConstEnum?: "A" | undefined;
  optionalConstDate?: string | undefined;
  optionalConstDateTime?: string | undefined;
  optionalConstBigint?: 987654321098765400 | undefined;
  optionalConstDecimal?: number | undefined;
  requiredDefaultString: string;
  requiredDefaultNumber: number;
  requiredDefaultBoolean: boolean;
  requiredDefaultEnum: string;
  requiredDefaultDate: string;
  requiredDefaultDateTime: string;
  requiredDefaultBigint: number;
  requiredDefaultDecimal: number;
  optionalDefaultString: string;
  optionalDefaultNumber: number;
  optionalDefaultBoolean: boolean;
  optionalDefaultEnum: string;
  optionalDefaultDate: string;
  optionalDefaultDateTime: string;
  optionalDefaultBigint: number;
  optionalDefaultDecimal: number;
  optionalString?: string | undefined;
  optionalNumber?: number | undefined;
  optionalBoolean?: boolean | undefined;
  optionalEnum?: string | undefined;
  optionalDate?: string | undefined;
  optionalDateTime?: string | undefined;
  optionalBigint?: number | undefined;
  optionalDecimal?: number | undefined;
  requiredConstDefaultString: "fixed-string";
  requiredConstDefaultNumber: 42.5;
  requiredConstDefaultBoolean: true;
  requiredConstDefaultEnum: "A";
  requiredConstDefaultDate: string;
  requiredConstDefaultDateTime: string;
  requiredConstDefaultBigint: 9007199254740991;
  requiredConstDefaultDecimal: number;
  optionalConstDefaultString: "optional-const-string";
  optionalConstDefaultNumber: 123.45;
  optionalConstDefaultBoolean: false;
  optionalConstDefaultEnum: "A";
  optionalConstDefaultDate: string;
  optionalConstDefaultDateTime: string;
  optionalConstDefaultBigint: 987654321098765400;
  optionalConstDefaultDecimal: number;
};

/** @internal */
export const ConstModel$outboundSchema: z.ZodMiniType<
  ConstModel$Outbound,
  ConstModel
> = z.object({
  requiredConstString: z.literal("fixed-string"),
  requiredConstNumber: z.literal(42.5),
  requiredConstBoolean: z.literal(true),
  requiredConstEnum: z.literal("A"),
  requiredConstDate: z.pipe(
    z.date(),
    z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
  ),
  requiredConstDateTime: z.pipe(
    z.date().check(z.refine((v) =>
      v.getTime() === new Date("2023-12-25T10:30:00Z").getTime(), {
      message: "Value must be equivalent to 2023-12-25T10:30:00Z",
    })),
    z.transform(v =>
      v.toISOString()
    ),
  ),
  requiredConstBigint: z.literal(9007199254740991),
  requiredConstDecimal: z.pipe(
    z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]).check(
      z.refine((v) => v.toString() === "123.456789", {
        message: "Value must be 123.456789",
      }),
    ),
    z.transform(v => typeof v === "number" ? v : v.toNumber()),
  ),
  optionalConstString: z.optional(z.literal("optional-const-string")),
  optionalConstNumber: z.optional(z.literal(123.45)),
  optionalConstBoolean: z.optional(z.literal(false)),
  optionalConstEnum: z.optional(z.literal("A")),
  optionalConstDate: z.optional(
    z.pipe(
      z.date(),
      z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
    ),
  ),
  optionalConstDateTime: z.optional(
    z.pipe(
      z.date().check(
        z.refine((v) =>
          v.getTime() === new Date("2025-06-15T15:30:00Z").getTime(), {
          message: "Value must be equivalent to 2025-06-15T15:30:00Z",
        }),
      ),
      z.transform(v => v.toISOString()),
    ),
  ),
  optionalConstBigint: z.optional(z.literal(987654321098765400)),
  optionalConstDecimal: z.optional(
    z.pipe(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()])
        .check(z.refine((v) =>
          v.toString() === "456.123789", {
          message: "Value must be 456.123789",
        })),
      z.transform(v =>
        typeof v === "number" ? v : v.toNumber()
      ),
    ),
  ),
  requiredDefaultString: z._default(z.string(), "default-string"),
  requiredDefaultNumber: z._default(z.number(), 99.9),
  requiredDefaultBoolean: z._default(z.boolean(), false),
  requiredDefaultEnum: z._default(RequiredDefaultEnum$outboundSchema, "A"),
  requiredDefaultDate: z.pipe(
    z._default(z.date(), new Date("2024-01-01")),
    z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
  ),
  requiredDefaultDateTime: z.pipe(
    z._default(z.date(), () => new Date("2024-01-01T00:00:00Z")),
    z.transform(v => v.toISOString()),
  ),
  requiredDefaultBigint: z._default(z.number(), 1234567890123456800),
  requiredDefaultDecimal: z.pipe(
    z._default(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]),
      () =>
        new Decimal$("999.999999"),
    ),
    z.transform(v => typeof v === "number" ? v : v.toNumber()),
  ),
  optionalDefaultString: z._default(z.string(), "optional-default"),
  optionalDefaultNumber: z._default(z.number(), 77.7),
  optionalDefaultBoolean: z._default(z.boolean(), true),
  optionalDefaultEnum: z._default(OptionalDefaultEnum$outboundSchema, "A"),
  optionalDefaultDate: z.pipe(
    z._default(z.date(), new Date("2024-12-31")),
    z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
  ),
  optionalDefaultDateTime: z.pipe(
    z._default(z.date(), () => new Date("2024-12-31T23:59:59Z")),
    z.transform(v => v.toISOString()),
  ),
  optionalDefaultBigint: z._default(z.number(), 555666777888999),
  optionalDefaultDecimal: z.pipe(
    z._default(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]),
      () =>
        new Decimal$("111.222333"),
    ),
    z.transform(v => typeof v === "number" ? v : v.toNumber()),
  ),
  optionalString: z.optional(z.string()),
  optionalNumber: z.optional(z.number()),
  optionalBoolean: z.optional(z.boolean()),
  optionalEnum: z.optional(OptionalEnum$outboundSchema),
  optionalDate: z.optional(
    z.pipe(
      z.date(),
      z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
    ),
  ),
  optionalDateTime: z.optional(
    z.pipe(z.date(), z.transform(v => v.toISOString())),
  ),
  optionalBigint: z.optional(z.number()),
  optionalDecimal: z.optional(
    z.pipe(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]),
      z.transform(v =>
        typeof v === "number" ? v : v.toNumber()
      ),
    ),
  ),
  requiredConstDefaultString: z._default(
    z.literal("fixed-string"),
    "fixed-string" as const,
  ),
  requiredConstDefaultNumber: z._default(z.literal(42.5), 42.5 as const),
  requiredConstDefaultBoolean: z._default(z.literal(true), true as const),
  requiredConstDefaultEnum: z._default(z.literal("A"), "A"),
  requiredConstDefaultDate: z.pipe(
    z._default(z.date(), new Date("2023-12-25")),
    z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
  ),
  requiredConstDefaultDateTime: z.pipe(
    z._default(z.date(), new Date("2023-12-25T10:30:00Z")).check(z.refine((v) =>
      v.getTime() === new Date("2023-12-25T10:30:00Z").getTime(), {
      message: "Value must be equivalent to 2023-12-25T10:30:00Z",
    })),
    z.transform(v =>
      v.toISOString()
    ),
  ),
  requiredConstDefaultBigint: z._default(
    z.literal(9007199254740991),
    9007199254740991 as const,
  ),
  requiredConstDefaultDecimal: z.pipe(
    z._default(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]),
      new Decimal$("123.456789"),
    ).check(z.refine((v) =>
      v.toString() === "123.456789", { message: "Value must be 123.456789" })),
    z.transform(v =>
      typeof v === "number" ? v : v.toNumber()
    ),
  ),
  optionalConstDefaultString: z._default(
    z.literal("optional-const-string"),
    "optional-const-string" as const,
  ),
  optionalConstDefaultNumber: z._default(z.literal(123.45), 123.45 as const),
  optionalConstDefaultBoolean: z._default(z.literal(false), false as const),
  optionalConstDefaultEnum: z._default(z.literal("A"), "A"),
  optionalConstDefaultDate: z.pipe(
    z._default(z.date(), new Date("2025-06-15")),
    z.transform(v => v.toISOString().slice(0, "YYYY-MM-DD".length)),
  ),
  optionalConstDefaultDateTime: z.pipe(
    z._default(z.date(), new Date("2025-06-15T15:30:00Z")).check(z.refine((v) =>
      v.getTime() === new Date("2025-06-15T15:30:00Z").getTime(), {
      message: "Value must be equivalent to 2025-06-15T15:30:00Z",
    })),
    z.transform(v =>
      v.toISOString()
    ),
  ),
  optionalConstDefaultBigint: z._default(
    z.literal(987654321098765400),
    987654321098765400 as const,
  ),
  optionalConstDefaultDecimal: z.pipe(
    z._default(
      z.union([z.custom<Decimal$>(x => x instanceof Decimal$), z.number()]),
      new Decimal$("456.123789"),
    ).check(z.refine((v) =>
      v.toString() === "456.123789", { message: "Value must be 456.123789" })),
    z.transform(v =>
      typeof v === "number" ? v : v.toNumber()
    ),
  ),
});

export function constModelToJSON(constModel: ConstModel): string {
  return JSON.stringify(ConstModel$outboundSchema.parse(constModel));
}


` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
