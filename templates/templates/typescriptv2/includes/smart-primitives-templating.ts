function templateSmartOptional(): string {
  // Wrap with z.optional so the field is treated as optional at the object
  // level. As of Zod 4.4 a bare `z.union([z.undefined(), ...])` is no longer
  // sufficient — the object schema requires `z.optional` to allow the key to
  // be missing entirely, otherwise it raises `expected: 'nonoptional'`.
  return `
    export function optional<T extends ${z.ZodType()}>(t: T) {
      return ${z.optional(`z.union([
        // Null -> undefined
        ${z.transform(z.null(), `() => unrecognized(undefined)`)},
        t
      ])`)};
    }`;
}

function templateSmartNullable(): string {
  return `
    export function nullable<T extends ${z.ZodType()}>(t: T) {
      return z.union([
        ${z.null()},

        // Undefined -> null
        ${z.transform(z.undefined(), `() => defaultToZeroValue(null)`)},
        t
      ]);
    }`;
}

function templateSmartString(): string {
  return `
    export function string(): ${z.ZodType("string")} {
      return z.union([
        ${z.string()},

        // Null or undefined -> ""
        zodDefaultToZeroValue(""),

        // Any other value -> String(x)
        ${z.transform(z.any(), `(x) => unrecognized(JSON.stringify(x))`)}
      ]);
    }`;
}

function templateSmartBoolean(): string {
  return `
    export function boolean(): ${z.ZodType("boolean")} {
      return z.union([
        ${z.boolean()},

        // String "true" (case insensitive) -> true, "false" -> false
        ${z.transform(
          z.string(),
          `(x, ctx) => {
            const lower = x.toLowerCase();
            if (lower === "true") return unrecognized(true);
            if (lower === "false") return unrecognized(false);
            ${z.addIssue({
              code: "invalid_type",
              input: "x",
              expected: "boolean",
              received: "string",
            })};
            return ${z.NEVER()};
          }`,
        )},

        zodDefaultToZeroValue(false)
      ]);
    }`;
}

function templateSmartLiteral(): string {
  return `
    export function literal<T extends string | number | boolean>(value: T): ${z.ZodType(
      "T",
    )} {
      return z.union([${z.literal("value")}, zodDefaultToZeroValue(value)]);
    }`;
}

function templateSmartLiteralBigInt(): string {
  return `
    export function literalBigInt<T extends bigint>(value: T): ${z.ZodType(
      "T",
    )} {
      return ${z.transform(
        `z.literal(String(value))`,
        "(x) => BigInt(x)",
      )} as any;
    }`;
}

function templateSmartNumber(): string {
  return `
    export function number(): ${z.ZodType("number")} {
      return z.union([
        ${z.number()},

        // String -> Number
        ${z.transform(
          z.string(),
          `(x, ctx) => {
            const num = Number(x);
            if (isNaN(num)) {
              ${z.addIssue({
                code: "invalid_type",
                input: "x",
                expected: "number",
                received: "string",
              })};
              return ${z.NEVER()};
            }
            return unrecognized(num);
          }`,
        )},

        // Null or undefined -> 0
        zodDefaultToZeroValue(0)
      ]);
    }`;
}

function templateSmartBigint(): string {
  return `
    export function bigint(): ${z.ZodType("bigint")} {
      return z.union([
        ${z.transform(
          z.string(),
          `(x, ctx) => {
          try {
            return BigInt(x);
          } catch (error) {
            ${z.addIssue({
              code: "invalid_type",
              input: "x",
              expected: "bigint",
              received: "string",
            })};
            return ${z.NEVER()};
          }
        }`,
        )},
        zodDefaultToZeroValue(BigInt(0))
      ]);
    }`;
}

function templateSmartDate(): string {
  return `
    export function date(): ${z.ZodType("Date")} {
      return z.union([
        ${z.pipe(
          z.transform(
            z.union([z.string(), `zodDefaultToZeroValue(0)`]),
            `(x) => new Date(x)`,
          ),
          z.date(),
        )},
        ${z.transform(
          z.number(),
          `(x, ctx) => {
            const date = new Date(x);
            if (isNaN(date.getTime())) {
              ${z.addIssue({
                code: "invalid_type",
                input: "x",
                expected: "date",
                received: "number",
              })};
              return ${z.NEVER()};
            }
            return unrecognized(date);
          }`,
        )}
      ]);
    }`;
}

function templateZodDefaultToZeroValue(): string {
  return `
    function zodDefaultToZeroValue<T>(value: T): ${z.ZodType("T")} {
      return ${z.transform(
        z.any(),
        `(input, ctx) => {
        if (input === undefined) return defaultToZeroValue(value);
        if (input === null) return defaultToZeroValue(value);
        ${z.addIssue({
          code: "invalid_type",
          input: "input",
          expected: "undefined",
          received: "unknown",
        })};
        return z.NEVER;
      }`,
      )};
    }`;
}

registerTemplateFunc("templateSmartOptional", templateSmartOptional);
registerTemplateFunc("templateSmartNullable", templateSmartNullable);
registerTemplateFunc("templateSmartString", templateSmartString);
registerTemplateFunc("templateSmartBoolean", templateSmartBoolean);
registerTemplateFunc("templateSmartLiteral", templateSmartLiteral);
registerTemplateFunc("templateSmartLiteralBigInt", templateSmartLiteralBigInt);
registerTemplateFunc("templateSmartNumber", templateSmartNumber);
registerTemplateFunc("templateSmartBigint", templateSmartBigint);
registerTemplateFunc("templateSmartDate", templateSmartDate);
registerTemplateFunc(
  "templateZodDefaultToZeroValue",
  templateZodDefaultToZeroValue,
);
