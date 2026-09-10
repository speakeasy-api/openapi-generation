namespace EnvTestTemplating {
  type EnvTestCase = {
    name: string;
    typeStr: string;
    input: string;
    expected: string;
    hasDefault?: boolean;
    defaultValue?: any;
  };

  const testCases: EnvTestCase[] = [
    {
      name: "string value",
      typeStr: "string",
      input: '"hello"',
      expected: '"hello"',
    },
    { name: "string empty", typeStr: "string", input: '""', expected: '""' },
    {
      name: "string default",
      typeStr: "string",
      input: "undefined",
      expected: '"default"',
      hasDefault: true,
      defaultValue: "default",
    },

    {
      name: "bigint positive",
      typeStr: "bigint",
      input: '"12345"',
      expected: "BigInt(12345)",
    },
    {
      name: "bigint negative",
      typeStr: "bigint",
      input: '"-67890"',
      expected: "BigInt(-67890)",
    },
    {
      name: "bigint default",
      typeStr: "bigint",
      input: "undefined",
      expected: "BigInt(0)",
      hasDefault: true,
      defaultValue: 0,
    },

    {
      name: "decimal positive",
      typeStr: "decimal",
      input: '"3.141592653589793"',
      expected: 'new Decimal("3.141592653589793")',
    },
    {
      name: "decimal negative",
      typeStr: "decimal",
      input: '"-2.718281828"',
      expected: 'new Decimal("-2.718281828")',
    },
    {
      name: "decimal default",
      typeStr: "decimal",
      input: "undefined",
      expected: 'new Decimal("0")',
      hasDefault: true,
      defaultValue: "0",
    },

    {
      name: "number float",
      typeStr: "number",
      input: '"123.45"',
      expected: "123.45",
    },
    {
      name: "number negative",
      typeStr: "number",
      input: '"-67.89"',
      expected: "-67.89",
    },
    {
      name: "number default",
      typeStr: "number",
      input: "undefined",
      expected: "0",
      hasDefault: true,
      defaultValue: 0,
    },

    {
      name: "integer positive",
      typeStr: "integer",
      input: '"123"',
      expected: "123",
    },
    {
      name: "integer negative",
      typeStr: "integer",
      input: '"-456"',
      expected: "-456",
    },
    {
      name: "integer default",
      typeStr: "integer",
      input: "undefined",
      expected: "0",
      hasDefault: true,
      defaultValue: 0,
    },

    {
      name: "boolean true",
      typeStr: "boolean",
      input: '"true"',
      expected: "true",
    },
    {
      name: "boolean false",
      typeStr: "boolean",
      input: '"false"',
      expected: "false",
    },
    {
      name: "boolean default true",
      typeStr: "boolean",
      input: "undefined",
      expected: "true",
      hasDefault: true,
      defaultValue: "true",
    },
    {
      name: "boolean default false",
      typeStr: "boolean",
      input: "undefined",
      expected: "false",
      hasDefault: true,
      defaultValue: "false",
    },

    {
      name: "date valid",
      typeStr: "date",
      input: '"2025-01-15"',
      expected: '"2025-01-15"',
    },
    {
      name: "date default",
      typeStr: "date",
      input: "undefined",
      expected: '"2025-01-01"',
      hasDefault: true,
      defaultValue: "2025-01-01",
    },

    {
      name: "datetime valid",
      typeStr: "date-time",
      input: '"2025-01-15T12:30:00.000Z"',
      expected: '"2025-01-15T12:30:00.000Z"',
    },
    {
      name: "datetime default",
      typeStr: "date-time",
      input: "undefined",
      expected: '"2025-01-01T00:00:00.000Z"',
      hasDefault: true,
      defaultValue: "2025-01-01T00:00:00.000Z",
    },

    {
      name: "any string",
      typeStr: "any",
      input: '"string"',
      expected: '"string"',
    },
    {
      name: "any default",
      typeStr: "any",
      input: "undefined",
      expected: '"fallback"',
      hasDefault: true,
      defaultValue: "fallback",
    },
  ];

  function createTypeDef(typeStr: string): TypeDef {
    const typeDef = createZeroTypeDef();
    typeDef.Type = typeStr as any;
    return typeDef;
  }

  function createFieldDefForType(
    typeStr: string,
    defaultValue?: any,
  ): FieldDef {
    const typeDef = createTypeDef(typeStr);
    const fieldDef = typeDefToFieldDef(typeDef);
    fieldDef.Name = "testField";

    if (defaultValue !== undefined) {
      fieldDef.Default = { Value: defaultValue };
    }

    return fieldDef;
  }

  export function templateEnvCoercionTests(): string {
    const usageLocation = "models";
    let tests = "";

    for (const testCase of testCases) {
      const fieldDef = createFieldDefForType(
        testCase.typeStr,
        testCase.hasDefault ? testCase.defaultValue : undefined,
      );

      const schema = toEnvZodSchema(fieldDef, usageLocation);

      tests += `
  test("${testCase.name}", () => {
    const schema = ${schema};
    const result = schema.parse(${testCase.input});
    expect(result).toEqual(${testCase.expected});
  });
`;
    }

    return tests;
  }
}

registerTemplateFunc(
  "templateEnvCoercionTests",
  EnvTestTemplating.templateEnvCoercionTests,
);
