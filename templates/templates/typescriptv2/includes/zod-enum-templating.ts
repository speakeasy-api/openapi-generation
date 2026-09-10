function templateOpenEnumInboundHelperFunc(): string {
  // No-zod skips this entirely: `enums.ts.stmpl` is gated off (templating.ts),
  // so this function is only called from the zod branch.
  return `
    export function inboundSchema<T extends Record<string, string>>(
      enumObj: T,
    ): ${z.ZodType("OpenEnum<T>", "unknown")} {
      const options = Object.values(enumObj);
      return z.union([
        ...options.map(x => ${z.literal("x")}),
        ${z.transform(z.string(), "x => unrecognized(x)")},
      ] as any);
    }

    export function inboundSchemaInt<T extends Record<string, number | string>>(
      enumObj: T,
    ): ${z.ZodType("OpenEnum<T>", "unknown")} {
      // For numeric enums, Object.values returns both numbers and string keys
      const options = Object.values(enumObj).filter(v => typeof v === 'number');
      return z.union([
        ...options.map(x => ${z.literal("x")}),
        ${z.transform(z.int(), "x => unrecognized(x)")},
      ] as any);
    }

    export function outboundSchema<T extends Record<string, string>>(
      _: T,
    ): ${z.ZodType("string", "OpenEnum<T>")} {
      return ${z.string()} as any;
    }

    export function outboundSchemaInt<T extends Record<string, number | string>>(
      _: T,
    ): ${z.ZodType("number", "OpenEnum<T>")} {
      return ${z.int()} as any;
    }
`;
}

registerTemplateFunc(
  "templateOpenEnumInboundHelperFunc",
  templateOpenEnumInboundHelperFunc,
);

// @ts-ignore
function templateOpenEnumInboundSchema(
  typeDef: TypeDef,
  typeName: string,
  usageLocation: string,
): string {
  addTypeImport("enums", "openEnums", usageLocation, "aliasImport");
  const dataType = getEnumDataType(typeDef);
  const helperFunc =
    dataType === "number" ? "inboundSchemaInt" : "inboundSchema";
  return `openEnums.${helperFunc}(${typeName})`;
}

registerTemplateFunc(
  "templateOpenEnumInboundSchema",
  templateOpenEnumInboundSchema,
);

// @ts-ignore
function templateOpenEnumOutboundSchema(
  typeDef: TypeDef,
  typeName: string,
  usageLocation: string,
): string {
  const dataType = getEnumDataType(typeDef);

  addTypeImport("enums", "openEnums", usageLocation, "aliasImport");

  if (dataType === "number") {
    return "openEnums.outboundSchemaInt(" + typeName + ")";
  }

  return "openEnums.outboundSchema(" + typeName + ")";
}

registerTemplateFunc(
  "templateOpenEnumOutboundSchema",
  templateOpenEnumOutboundSchema,
);
