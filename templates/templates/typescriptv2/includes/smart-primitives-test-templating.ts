// Smart Primitives test case generation functions

namespace SmartPrimitivesTestTemplating {
  type CoercionCase<T = unknown> = [
    {
      type:
        | "string"
        | "number"
        | "boolean"
        | "bigint"
        | "date"
        | { literal: T };
      optional?: 1;
      nullable?: 1;
    },
    (
      | string
      | number
      | boolean
      | bigint
      | Date
      | undefined
      | null
      | {
          input: any;
          expected?:
            | string
            | number
            | boolean
            | bigint
            | Date
            | undefined
            | null;
        }
      | { error: any }
    ),
  ];

  const coercionCases: CoercionCase[] = [
    // String: Optional and nullable
    [{ type: "string", optional: 1, nullable: 1 }, "foo"],
    [{ type: "string", optional: 1, nullable: 1 }, ""],
    [
      { type: "string", optional: 1, nullable: 1 },
      { input: {}, expected: "{}" },
    ],
    [{ type: "string", optional: 1, nullable: 1 }, undefined],
    [{ type: "string", optional: 1, nullable: 1 }, null],

    // String: Optional
    [{ type: "string", optional: 1 }, "foo"],
    [{ type: "string", optional: 1 }, ""],
    [
      { type: "string", optional: 1 },
      { input: {}, expected: "{}" },
    ],
    [{ type: "string", optional: 1 }, undefined],
    [
      { type: "string", optional: 1 },
      { input: null, expected: undefined },
    ],

    // String: Nullable
    [{ type: "string", nullable: 1 }, "foo"],
    [{ type: "string", nullable: 1 }, ""],
    [
      { type: "string", nullable: 1 },
      { input: {}, expected: "{}" },
    ],
    [
      { type: "string", nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: "string", nullable: 1 }, null],

    // String: Required
    [{ type: "string" }, "foo"],
    [{ type: "string" }, ""],
    [{ type: "string" }, { input: {}, expected: "{}" }],
    [{ type: "string" }, { input: undefined, expected: "" }],
    [{ type: "string" }, { input: null, expected: "" }],
    // Boolean: Optional and nullable
    [{ type: "boolean", optional: 1, nullable: 1 }, true],
    [{ type: "boolean", optional: 1, nullable: 1 }, false],
    [{ type: "boolean", optional: 1, nullable: 1 }, { error: {} }],
    [{ type: "boolean", optional: 1, nullable: 1 }, undefined],
    [{ type: "boolean", optional: 1, nullable: 1 }, null],

    // Boolean: Optional
    [{ type: "boolean", optional: 1 }, true],
    [{ type: "boolean", optional: 1 }, false],
    [{ type: "boolean", optional: 1 }, undefined],
    [{ type: "boolean", optional: 1 }, { error: {} }],
    [
      { type: "boolean", optional: 1 },
      { input: null, expected: undefined },
    ],
    // Boolean: Nullable
    [{ type: "boolean", nullable: 1 }, true],
    [{ type: "boolean", nullable: 1 }, false],
    [{ type: "boolean", nullable: 1 }, { error: {} }],
    [
      { type: "boolean", nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: "boolean", nullable: 1 }, null],
    // Boolean: Required
    [{ type: "boolean" }, true],
    [{ type: "boolean" }, false],
    [{ type: "boolean" }, { error: {} }],
    [{ type: "boolean" }, { input: undefined, expected: false }],
    [{ type: "boolean" }, { input: null, expected: false }],

    // Number: Optional and nullable
    [{ type: "number", optional: 1, nullable: 1 }, 1],
    [{ type: "number", optional: 1, nullable: 1 }, 0],
    [{ type: "number", optional: 1, nullable: 1 }, { error: {} }],
    [{ type: "number", optional: 1, nullable: 1 }, undefined],
    [{ type: "number", optional: 1, nullable: 1 }, null],
    // Number: Optional
    [{ type: "number", optional: 1 }, 1],
    [{ type: "number", optional: 1 }, 0],
    [{ type: "number", optional: 1 }, { error: {} }],
    [{ type: "number", optional: 1 }, undefined],
    [
      { type: "number", optional: 1 },
      { input: null, expected: undefined },
    ],
    // Number: Nullable
    [{ type: "number", nullable: 1 }, 1],
    [{ type: "number", nullable: 1 }, 0],
    [{ type: "number", nullable: 1 }, { error: {} }],
    [
      { type: "number", nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: "number", nullable: 1 }, null],
    // Number: Required
    [{ type: "number" }, 1],
    [{ type: "number" }, 0],
    [{ type: "number" }, { error: {} }],
    [{ type: "number" }, { input: undefined, expected: 0 }],
    [{ type: "number" }, { input: null, expected: 0 }],

    // BigInt: Optional and nullable
    [
      { type: "bigint", optional: 1, nullable: 1 },
      { input: "1", expected: BigInt(1) },
    ],
    [
      { type: "bigint", optional: 1, nullable: 1 },
      { input: "0", expected: BigInt(0) },
    ],
    [{ type: "bigint", optional: 1, nullable: 1 }, { error: {} }],
    [{ type: "bigint", optional: 1, nullable: 1 }, undefined],
    [{ type: "bigint", optional: 1, nullable: 1 }, null],

    // BigInt: Optional
    [
      { type: "bigint", optional: 1 },
      { input: "1", expected: BigInt(1) },
    ],
    [
      { type: "bigint", optional: 1 },
      { input: "0", expected: BigInt(0) },
    ],
    [{ type: "bigint", optional: 1 }, { error: {} }],
    [{ type: "bigint", optional: 1 }, undefined],
    [
      { type: "bigint", optional: 1 },
      { input: null, expected: undefined },
    ],
    // BigInt: Nullable
    [
      { type: "bigint", nullable: 1 },
      { input: "1", expected: BigInt(1) },
    ],
    [
      { type: "bigint", nullable: 1 },
      { input: "0", expected: BigInt(0) },
    ],
    [{ type: "bigint", nullable: 1 }, { error: {} }],
    [
      { type: "bigint", nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: "bigint", nullable: 1 }, null],
    // BigInt: Required
    [{ type: "bigint" }, { input: "1", expected: BigInt(1) }],
    [{ type: "bigint" }, { input: "0", expected: BigInt(0) }],
    [{ type: "bigint" }, { error: {} }],
    [{ type: "bigint" }, { input: undefined, expected: BigInt(0) }],
    [{ type: "bigint" }, { input: null, expected: BigInt(0) }],

    // Date: Optional and nullable
    [
      { type: "date", optional: 1, nullable: 1 },
      { input: "2021-01-01", expected: new Date("2021-01-01") },
    ],
    [
      { type: "date", optional: 1, nullable: 1 },
      { input: "0", expected: new Date("0") },
    ],
    [
      { type: "date", optional: 1, nullable: 1 },
      { input: 1763130865965, expected: new Date(1763130865965) },
    ],
    [{ type: "date", optional: 1, nullable: 1 }, { error: {} }],
    [{ type: "date", optional: 1, nullable: 1 }, undefined],
    [{ type: "date", optional: 1, nullable: 1 }, null],

    // Date: Optional
    [
      { type: "date", optional: 1 },
      { input: "2021-01-01", expected: new Date("2021-01-01") },
    ],
    [
      { type: "date", optional: 1 },
      { input: "0", expected: new Date("0") },
    ],
    [
      { type: "date", optional: 1 },
      { input: 1763130865965, expected: new Date(1763130865965) },
    ],
    [{ type: "date", optional: 1 }, { error: {} }],
    [
      { type: "date", optional: 1 },
      { input: undefined, expected: undefined },
    ],
    [
      { type: "date", optional: 1 },
      { input: null, expected: undefined },
    ],

    // Date: Nullable
    [
      { type: "date", nullable: 1 },
      { input: "2021-01-01", expected: new Date("2021-01-01") },
    ],
    [
      { type: "date", nullable: 1 },
      { input: "0", expected: new Date("0") },
    ],
    [
      { type: "date", nullable: 1 },
      { input: 1763130865965, expected: new Date(1763130865965) },
    ],
    [{ type: "date", nullable: 1 }, { error: {} }],
    [
      { type: "date", nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: "date", nullable: 1 }, null],

    // Date: Required
    [
      { type: "date" },
      { input: "2021-01-01", expected: new Date("2021-01-01") },
    ],
    [{ type: "date" }, { input: "0", expected: new Date("0") }],
    [
      { type: "date" },
      { input: 1763130865965, expected: new Date(1763130865965) },
    ],
    [{ type: "date" }, { error: {} }],
    [{ type: "date" }, { input: undefined, expected: new Date(0) }],
    [{ type: "date" }, { input: null, expected: new Date(0) }],

    // Literal: Optional and nullable
    [{ type: { literal: "foo" }, optional: 1, nullable: 1 }, "foo"],
    [{ type: { literal: 22 }, optional: 1, nullable: 1 }, 22],
    [{ type: { literal: true }, optional: 1, nullable: 1 }, true],
    [{ type: { literal: "foo" }, optional: 1, nullable: 1 }, { error: {} }],
    [{ type: { literal: "foo" }, optional: 1, nullable: 1 }, undefined],
    [{ type: { literal: "foo" }, optional: 1, nullable: 1 }, null],

    // Literal: Optional
    [{ type: { literal: "foo" }, optional: 1 }, "foo"],
    [{ type: { literal: 22 }, optional: 1 }, 22],
    [{ type: { literal: true }, optional: 1 }, true],
    [{ type: { literal: "foo" }, optional: 1 }, { error: {} }],
    [{ type: { literal: "foo" }, optional: 1 }, undefined],
    [
      { type: { literal: "foo" }, optional: 1 },
      { input: null, expected: undefined },
    ],

    // Literal: Nullable
    [{ type: { literal: "foo" }, nullable: 1 }, "foo"],
    [{ type: { literal: 22 }, nullable: 1 }, 22],
    [{ type: { literal: true }, nullable: 1 }, true],
    [{ type: { literal: "foo" }, nullable: 1 }, { error: {} }],
    [
      { type: { literal: "foo" }, nullable: 1 },
      { input: undefined, expected: null },
    ],
    [{ type: { literal: "foo" }, nullable: 1 }, null],

    // Literal: Required
    [{ type: { literal: "foo" } }, "foo"],
    [{ type: { literal: 22 } }, 22],
    [{ type: { literal: true } }, true],
    [{ type: { literal: "foo" } }, { error: {} }],
    [{ type: { literal: "foo" } }, { input: undefined, expected: "foo" }],
    [{ type: { literal: "foo" } }, { input: null, expected: "foo" }],
  ];

  function templateCase({
    typeDescription,
    schema,
    inputString,
    expectedString,
    shouldError,
  }: {
    typeDescription: string;
    schema: string;
    inputString: string;
    expectedString: string;
    shouldError: boolean;
  }) {
    const testName = shouldError
      ? `Test coercion for ${typeDescription} - ${inputString} should error`
      : `Test coercion for ${typeDescription} - ${inputString} coerces to ${expectedString}`;

    return `
      test(\`${testName}\`, () => {
        const result = ${schema}.safeParse(${inputString});
        ${
          shouldError
            ? `expect(result.success).toBe(false);`
            : `expect(result.success).toBe(true);`
        }
        ${shouldError ? "" : `expect(result.data).toEqual(${expectedString});`}
      });
    `;
  }

  // Helper to serialize values (including BigInt and Date)
  function serializeValue(val: any): string {
    if (typeof val === "bigint") {
      return `BigInt(${val})`;
    }
    if (val instanceof Date) {
      return `new Date("${val.toISOString()}")`;
    }
    return JSON.stringify(val);
  }

  // Helper to create a TypeDef from a type configuration
  function createTypeDefFromConfig(typeConfig: CoercionCase[0]): {
    typeDef: TypeDef;
    constValue: any;
  } {
    const { type } = typeConfig;
    let typeDefType: string;
    let constValueObj: any = undefined;

    if (typeof type === "string") {
      // Map the test type names to TypeDef type names
      if (type === "date") {
        typeDefType = "date-time";
      } else {
        typeDefType = type;
      }
    } else {
      // It's a literal type
      typeDefType = typeof type.literal;
      constValueObj = { Value: type.literal };
    }

    // In lax mode, bigint types need Format: "string" to generate bigint instead of number
    const format = typeDefType === "bigint" ? "string" : "";

    const typeDef = createZeroTypeDef();
    typeDef.Name = typeDefType;
    typeDef.OriginalName = typeDefType;
    typeDef.Type = dataType(typeDefType as DataType);
    typeDef.Scope = "shared" as Scope;
    typeDef.Format = format;

    return { typeDef, constValue: constValueObj };
  }

  export function templateSmartPrimitiveCoercionTests(): string {
    return coercionCases
      .map((testCase) => {
        const [typeConfig, inputOrExpected] = testCase;
        const { type, optional, nullable } = typeConfig;
        let typeDescription =
          typeof type === "string"
            ? type
            : `literal(${JSON.stringify(type.literal)})`;
        if (optional && nullable) {
          typeDescription += " (optional and nullable)";
        } else if (optional) {
          typeDescription += " (optional)";
        } else if (nullable) {
          typeDescription += " (nullable)";
        } else {
          typeDescription += " (required)";
        }
        let inputString = "";
        let expectedString = "";
        let shouldError = false;

        if (typeof inputOrExpected === "object" && inputOrExpected !== null) {
          if ("error" in inputOrExpected) {
            shouldError = true;
            // For error cases, use the error object as input
            inputString = serializeValue(inputOrExpected.error);
          } else if (
            "input" in inputOrExpected &&
            "expected" in inputOrExpected
          ) {
            inputString = serializeValue(inputOrExpected.input);
            expectedString = serializeValue(inputOrExpected.expected);
          }
        } else {
          inputString = serializeValue(inputOrExpected);
          expectedString = serializeValue(inputOrExpected);
        }

        // Create a proper TypeDef object using the utility function
        const { typeDef, constValue } = createTypeDefFromConfig(typeConfig);

        let schema = resolveInbound({
          nullable: Boolean(nullable),
          optional: Boolean(optional),
          usageLocation: "models",
          rootTypeDef: typeDef,
          typeDef: typeDef,
          constValue: constValue,
        }).zod;

        return templateCase({
          typeDescription,
          schema,
          inputString,
          expectedString,
          shouldError,
        });
      })
      .join("\n");
  }
}

// Register template functions
registerTemplateFunc(
  "templateSmartPrimitiveCoercionTests",
  SmartPrimitivesTestTemplating.templateSmartPrimitiveCoercionTests,
);
