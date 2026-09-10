type ResolvedTypes = {
  inputType: string;
  importInputTypes: () => string;
  outputType: string;
  importOutputTypes: () => string;
  zod: string;
  importZodTypes: () => string;
};

function callImportTypesFunc(func: () => string): string {
  assert(func, "func missing for callImportTypesFunc");
  assert(
    typeof func === "function",
    "func is not a function in callImportTypesFunc",
  );
  return func();
}

registerTemplateFunc("callImportTypesFunc", callImportTypesFunc);

function resolveBigintZod(
  baseZod: string,
  optional: boolean,
  nullable: boolean,
  usageLocation: string,
  origImportZodTypes: () => string,
): { zod: string; importZodTypes: () => string } {
  if (optional && nullable) {
    return {
      zod: baseZod.replace("bigint()", "z.optional(z.nullable(bigint()))"),
      importZodTypes: () => {
        addImport("zod", "z", "aliasImport");
        addTypeImport("bigint", "bigint", usageLocation);
        return "";
      },
    };
  } else if (optional) {
    return {
      zod: baseZod.replace("bigint()", "bigintOptional()"),
      importZodTypes: () => {
        addTypeImport("bigint", "bigintOptional", usageLocation);
        return "";
      },
    };
  } else if (nullable) {
    return {
      zod: baseZod.replace("bigint()", "bigintNullable()"),
      importZodTypes: () => {
        addTypeImport("bigint", "bigintNullable", usageLocation);
        return "";
      },
    };
  }
  return { zod: baseZod, importZodTypes: origImportZodTypes };
}

function resolveDecimalZod(
  baseZod: string,
  optional: boolean,
  nullable: boolean,
  usageLocation: string,
  origImportZodTypes: () => string,
  isFormatString: boolean,
): { zod: string; importZodTypes: () => string } {
  const baseFn = isFormatString ? "decimalStr" : "decimal";
  const optionalFn = isFormatString ? "decimalStrOptional" : "decimalOptional";
  const nullableFn = isFormatString ? "decimalStrNullable" : "decimalNullable";

  if (optional && nullable) {
    return {
      zod: baseZod.replace(
        /decimal(?:Str)?\(\)/,
        `z.optional(z.nullable(${baseFn}()))`,
      ),
      importZodTypes: () => {
        addImport("zod", "z", "aliasImport");
        addTypeImport("decimal", baseFn, usageLocation);
        if (baseZod.includes("new Decimal")) {
          addTypeImport("decimal", "Decimal", usageLocation);
        }
        return "";
      },
    };
  } else if (optional) {
    return {
      zod: baseZod.replace(/decimal(?:Str)?\(\)/, `${optionalFn}()`),
      importZodTypes: () => {
        addTypeImport("decimal", optionalFn, usageLocation);
        if (baseZod.includes("new Decimal")) {
          addTypeImport("decimal", "Decimal", usageLocation);
        }
        return "";
      },
    };
  } else if (nullable) {
    return {
      zod: baseZod.replace(/decimal(?:Str)?\(\)/, `${nullableFn}()`),
      importZodTypes: () => {
        addTypeImport("decimal", nullableFn, usageLocation);
        if (baseZod.includes("new Decimal")) {
          addTypeImport("decimal", "Decimal", usageLocation);
        }
        return "";
      },
    };
  }
  return { zod: baseZod, importZodTypes: origImportZodTypes };
}

function enumLiteralFromValue(
  enumName: string,
  typeDef: TypeDef,
  value: unknown,
) {
  return typeof value === "string" ? quote(value) : `${value}`;
}

function getEnumNamesFromValues(values: string[]): string[] {
  let enumNames = [];

  let names = {};
  for (const value of values) {
    let name = sanitizeEnumName(value);
    if (!names[name]) {
      names[name] = 0;
    }

    names[name] += 1;
  }

  for (const value of values) {
    let name = sanitizeEnumName(value);
    if (names[name] > 1) {
      name = `${name}${caser().ToPascal(getCasing(value))}`;
    }

    enumNames.push(name);
  }

  return enumNames;
}

function getEnumDataType(typeDef: TypeDef): "string" | "number" {
  const underlying = typeDef.Enum?.Type.Type.toString();
  switch (underlying) {
    case "string":
      return "string";
    case "int32":
    case "integer":
      return "number";
    default:
      throw new Error(`Unsupported enum type: ${underlying}`);
  }
}

registerTemplateFunc("getEnumDataType", getEnumDataType);

function isInputOptional(fieldDef: FieldDef) {
  if (
    fieldDef.Type?.Type.toString() === "any" ||
    // Any unions formed with an `any` member are effectively reduced to `any`
    // which means we treat them as optional.
    fieldDef.Type?.AssociatedTypes?.some((at) => at.Type.toString() === "any")
  ) {
    return true;
  }

  const hasConst = fieldDef.Const?.Value !== undefined;

  const hasDefault = fieldDef.Default?.Value !== undefined;

  return fieldDef.Optional || hasConst || hasDefault;
}

registerTemplateFunc("isInputOptional", isInputOptional);

function isOutputOptional(fieldDef: FieldDef) {
  if (
    fieldDef.Type?.Type.toString() === "any" ||
    // Any unions formed with an `any` member are effectively reduced to `any`
    // which means we treat them as optional.
    fieldDef.Type?.AssociatedTypes?.some((at) => at.Type.toString() === "any")
  ) {
    return true;
  }

  const isConstant = fieldDef.Const?.Value !== undefined;
  if (isConstant) {
    return false;
  }

  const hasDefault = fieldDef.Default?.Value !== undefined;

  return fieldDef.Optional && !hasDefault;
}

registerTemplateFunc("isOutputOptional", isOutputOptional);

function needsLazyRef(usageLocation: string, a: TypeDef, b: TypeDef) {
  if (!a.IsCustomType() || !b.IsCustomType()) {
    return false;
  }

  if (!a.ResolvedModel || !b.ResolvedModel) {
    return false;
  }

  const aLoc = `${usageLocation}${a.ResolvedModel}`;
  const bLoc = `${getModelsLocation(b.OutputLocation)}${b.ResolvedModel}`;

  if (aLoc === bLoc) {
    return true;
  }

  return areCircular(a, b);
}

function resolveDiscriminator(
  usageLocation: string,
  typeDef: TypeDef,
  rootTypeDef: TypeDef,
  fieldName: string,
  fieldValue: string,
): { literal: string; type: TypeDef; importFunc: () => string } {
  const type = typeDef.Type.toString();
  if (type !== "class" && type !== "union" && !isUnionOfErrors(rootTypeDef)) {
    throw new Error(
      "Expected discriminator to map over a union of objects but got: " + type,
    );
  }

  let resolvedTypeDef = typeDef;
  if (type === "union") {
    // resolvedTypeDef is used to lookup the discriminator field.
    // the discriminator should be homogenous(present, and all of the same type) across union members.
    // we pick the first union member in such cases.
    resolvedTypeDef = typeDef.AssociatedTypes[0];
  }
  const field = resolvedTypeDef.Fields.find((f) => f.Name === fieldName);

  switch (field.Type?.Type.toString()) {
    case "enum": {
      const enumName = sanitizeClass(field.Type, usageLocation, true);
      return {
        literal: enumLiteralFromValue(enumName, field.Type, fieldValue),
        type: field.Type,
        importFunc: () => "",
      };
    }
    default: {
      return {
        literal: `"${fieldValue}"`,
        type: field.Type,
        importFunc: () => "",
      };
    }
  }
}

function resolveEnvSchema(
  fieldDef: FieldDef,
  usageLocation: string,
  options?: {
    /**
     * Useful when we want the schema without worrying about optionality.
     * Specific example of this is the MCP CLI the `parse` callback for a flag
     * is only called when a value is passed in and expects a non-optional
     * return value.
     */
    ignoreOptional?: boolean;
  },
): string {
  let schema = toEnvZodSchema(fieldDef, usageLocation);
  if (fieldDef.Nullable) {
    schema = zodNullable(schema);
  }
  if (!options?.ignoreOptional && fieldDef.Default == null) {
    schema = zodOptional(schema);
  }
  return schema;
}

registerTemplateFunc("resolveEnvSchema", resolveEnvSchema);

function resolveInbound(options: {
  usageLocation: string;
  typeDef?: TypeDef;
  rootTypeDef: TypeDef;
  optional: boolean;
  nullable: boolean;
  visited?: Set<string>;
  encoding?: string;
  constValue?: AnyValue;
  defaultValue?: AnyValue;
}): ResolvedTypes & { inputType: "unknown" } {
  const {
    usageLocation,
    rootTypeDef,
    typeDef,
    optional,
    nullable,
    constValue,
    defaultValue,
    visited,
    encoding,
  } = options;

  if (!typeDef) {
    return {
      inputType: "unknown",
      outputType: "void",
      zod: zodVoid(),
      importInputTypes: () => "",
      importOutputTypes: () => "",
      importZodTypes: () => addImport("zod", "z", "aliasImport"),
    };
  }

  const isOptional = optional && defaultValue?.Value === undefined;

  // We can short-circuit type resolution if a field is pinned to null using
  // const in the schema.
  if (constValue?.Value === null) {
    return {
      inputType: "unknown",
      outputType: "null",
      zod: zodDefault(zodNull(), "null"),
      importInputTypes: () => "",
      importOutputTypes: () => "",
      importZodTypes: () => addImport("zod", "z", "aliasImport"),
    };
  }

  if (usageLocation == null) {
    throw new Error("usage location cannot be null for: " + typeDef.Name);
  }
  let result = {
    ...toInbound({
      usageLocation,
      rootTypeDef,
      typeDef,
      constValue: constValue?.Value,
      defaultValue: defaultValue?.Value,
      visited: visited || new Set(),
    }),
    inputType: "unknown" as const,
    importInputTypes: () => "",
  };

  const isNullable =
    nullable || defaultValue?.Value === null || constValue?.Value === null;

  const hasConst = constValue?.Value !== undefined;

  if (typeDef.Type?.toString() === "bigint" && !hasConst) {
    const resolved = resolveBigintZod(
      result.zod,
      isOptional,
      isNullable,
      usageLocation,
      result.importZodTypes,
    );
    result.zod = resolved.zod;
    result.importZodTypes = resolved.importZodTypes;

    if (defaultValue?.Value === null) {
      result.zod = zodDefault(result.zod, "null");
    }
  } else if (typeDef.Type?.toString() === "decimal" && !hasConst) {
    const resolved = resolveDecimalZod(
      result.zod,
      isOptional,
      isNullable,
      usageLocation,
      result.importZodTypes,
      typeDef.Format === "string",
    );
    result.zod = resolved.zod;
    result.importZodTypes = resolved.importZodTypes;

    if (defaultValue?.Value === null) {
      result.zod = zodDefault(result.zod, "null");
    }
  } else {
    if (isNullable) {
      // If the default value is null then each type resolver needs to incorporate
      // that into its zod pipeline.
      result.zod = zodNullable(result.zod);
    }

    if (defaultValue?.Value === null) {
      result.zod = zodDefault(result.zod, "null");
    }

    if (isOptional) {
      result.zod = zodOptional(result.zod);
    }
  }

  if (encoding === "application/json") {
    const issueStatement =
      "ctx.addIssue({code: z.ZodIssueCode.custom, message: `malformed json: ${err}`})";

    result.zod = zodPipe(
      zodTransform(
        zodString(),
        `(v, ctx) => {try { return JSON.parse(v) } catch (err) { ${issueStatement}; return z.NEVER; }}`,
      ),
      result.zod,
    );
  }

  return result;
}

registerTemplateFunc("resolveInbound", resolveInbound);

function resolveOutbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  optional: boolean;
  nullable: boolean;
  streamable?: boolean;
  visited?: Set<string>;
  constValue?: AnyValue;
  defaultValue?: AnyValue;
}): Omit<ResolvedTypes, "outputType" | "importOutputTypes"> {
  const {
    usageLocation,
    rootTypeDef,
    typeDef,
    optional,
    nullable,
    streamable,
    visited,
    constValue,
    defaultValue,
  } = options;

  if (usageLocation == null) {
    throw new Error("usage location cannot be null for: " + typeDef.Name);
  }

  // We can short-circuit type resolution if a field is pinned to null using
  // const in the schema.
  if (constValue?.Value === null) {
    return {
      inputType: "null | undefined",
      importInputTypes: () => "",
      zod: "z.literal(null).default(null)",
      importZodTypes: () => "",
    };
  }

  let result = toOutbound({
    usageLocation,
    rootTypeDef,
    typeDef,
    constValue: constValue?.Value,
    defaultValue: defaultValue?.Value,
    visited: visited || new Set(),
  });

  if (streamable) {
    result.inputType += " | Blob";
    result.zod += ".or(blobLikeSchema)";
    const origImportZodTypes = result.importZodTypes;
    result.importZodTypes = () => {
      origImportZodTypes();
      addTypeImport("blobs", "blobLikeSchema", usageLocation);
      return "";
    };
  }

  const isNullable = nullable || defaultValue?.Value === null;
  if (isNullable) {
    result.inputType += " | null";
    result.zod = `z.nullable(${result.zod})`;
    let importZodTypes = result.importZodTypes;
    result.importZodTypes = () => {
      addImport("zod", "z", "aliasImport");
      return importZodTypes();
    };
  }
  if (defaultValue?.Value === null) {
    result.zod += ".default(null)";
  }

  switch (true) {
    // For outbound data, i.e. from the client, we consider const fields optional
    // and fill them in so they become required/set on the way out.
    case constValue?.Value !== undefined:
      result.inputType += " | undefined";
      break;
    case optional && defaultValue?.Value === undefined:
      result.inputType += " | undefined";
      result.zod += ".optional()";
      break;
    case defaultValue?.Value !== undefined:
      result.inputType += " | undefined";
      break;
  }

  return result;
}

registerTemplateFunc("resolveOutbound", resolveOutbound);

function toInbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  visited: Set<string>;
  constValue?: unknown;
  defaultValue?: unknown;
}): Omit<ResolvedTypes, "importInputTypes" | "inputType"> {
  const {
    usageLocation,
    typeDef,
    rootTypeDef,
    constValue,
    defaultValue,
    visited,
  } = options;
  const rootModel = rootTypeDef.ResolvedModel;
  const typeName = typeDef.Type.toString();
  const zodRef = sanitizeZodRef(typeDef, usageLocation);
  const isFormatString = typeDef.Format === "string";
  let zod = "";

  switch (typeName) {
    case "enum": {
      const enumName = sanitizeClass(typeDef, usageLocation, true);
      const literal = enumLiteralFromValue(enumName, typeDef, constValue);

      if (constValue !== undefined) {
        zod = zodDefault(zodLiteral(literal), literal);
      } else {
        zod = zodRef;
        if (defaultValue) {
          zod = zodDefault(zod, JSON.stringify(defaultValue));
        }
      }

      return {
        outputType: constValue !== undefined ? literal : enumName,
        importOutputTypes: () =>
          addTypeDefImport(typeDef, usageLocation, rootModel),
        zod,
        importZodTypes: () =>
          constValue !== undefined
            ? ""
            : addZodImport(typeDef, usageLocation, rootModel),
      };
    }
    case "union": {
      const regid = typeDef.GetRegistrationID();
      const shouldTruncate = visited.has(regid);

      if (shouldTruncate) {
        return {
          outputType: sanitizeClassName(typeDef.Name),
          importOutputTypes: () => "",
          zod: zodLazy(zodRef),
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      } else {
        visited.add(regid);
      }

      // If we're in a different location to the TypeDef we're resolving then it
      // implies we want to import the union type. Therefore, we need to a
      // reference to the union instead of building it out. For example, this
      // code path may be hit when an operation takes a request that is a union.
      if (!inSameModel(rootModel, usageLocation, typeDef)) {
        const lazy = needsLazyRef(usageLocation, typeDef, rootTypeDef);

        return {
          outputType: sanitizeClass(typeDef, usageLocation, true),
          importOutputTypes: () =>
            addTypeDefImport(typeDef, usageLocation, rootModel),
          zod: lazy ? zodLazy(zodRef) : zodRef,
          importZodTypes: () => {
            addZodImport(typeDef, usageLocation, rootModel);
            if (lazy) {
              addImport("zod", "z", "aliasImport");
            }
            return "";
          },
        };
      }

      // Unions with a single member should be reduced to resolving the
      // underlying type.
      if (typeDef.AssociatedTypes.length === 1) {
        return resolveInbound({
          usageLocation,
          typeDef: typeDef.AssociatedTypes[0],
          rootTypeDef,
          optional: false,
          nullable: false,
          visited,
        });
      }

      // Iterate over each union member and resolve the runtime type info then
      // it's a matter of concatenating the results to build typescript/zod
      // unions.
      const discr = typeDef.Discriminator?.TypePropertyName;
      const unionOutput: string[] = [];
      const unionImportOutputTypes: (() => string)[] = [];
      const unionZod: string[] = [];
      const unionImportZodTypes: (() => string)[] = [];

      sortUnionMembers(typeDef).forEach(([at, mapping]) => {
        let { outputType, importOutputTypes, zod, importZodTypes } =
          resolveInbound({
            usageLocation,
            typeDef: at,
            rootTypeDef,
            optional: false,
            nullable: false,
            visited,
          });

        const atName = mapping?.Name;

        if (discr && !atName) {
          throw new Error(
            `Discriminated union member does not have a discriminator mapping: ${typeDef.Name} > ${at.Name}`,
          );
        }

        // If we're dealing with a discriminated union, we expect union members
        // to be objects. Each object is expected to have a discriminator field
        // the value of that value is explicitly defined in the spec and our ast
        // in the disciminator mappings list. So if these expectations are met
        // then we alter the typings to more explicitly list the discriminator
        // key as having a literal string type.
        if (discr) {
          const dkey = sanitizeKey(discr);
          const dfield = sanitizeFieldName(discr);
          const dacc = sanitizeAccessor("v", discr);
          const transform = `(v) => ({ ${dfield}: ${dacc} })`;

          let { literal } = resolveDiscriminator(
            usageLocation,
            mapping.Type,
            rootTypeDef,
            discr,
            atName,
          );

          zod = zodAnd(
            zod,
            zodTransform(zodObject({ [dkey]: zodLiteral(literal) }), transform),
          );

          outputType = `(${outputType} & {${dfield}: ${literal}})`;
        }

        unionOutput.push(outputType);
        unionImportOutputTypes.push(importOutputTypes);
        unionZod.push(zod);
        unionImportZodTypes.push(importZodTypes);
      });

      return {
        outputType: unionOutput.join(" | "),
        importOutputTypes: () => {
          unionImportOutputTypes.forEach((f) => f());
          return "";
        },
        zod: zodUnion(unionZod),
        importZodTypes: () => {
          addImport("zod", "z", "aliasImport");
          unionImportZodTypes.forEach((f) => f());
          return "";
        },
      };
    }
    case "error":
    case "class": {
      zod = needsLazyRef(usageLocation, typeDef, rootTypeDef)
        ? zodLazy(zodRef)
        : zodRef;

      return {
        outputType: sanitizeClass(typeDef, usageLocation, true),
        importOutputTypes: () =>
          addTypeDefImport(typeDef, usageLocation, rootModel),
        zod,
        importZodTypes: () => {
          zod.startsWith("z.") && addImport("zod", "z", "aliasImport");
          addZodImport(typeDef, usageLocation, rootModel);
          return "";
        },
      };
    }
    case "string": {
      const literal = templateStringValue(String(constValue), null);

      if (constValue !== undefined) {
        zod = zodDefault(zodLiteral(literal), literal);
      } else {
        zod = zodString();
        if (defaultValue != null) {
          zod = zodDefault(
            zod,
            templateStringValue(String(defaultValue), null),
          );
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? literal : "string",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "date": {
      if (constValue !== undefined) {
        zod = zodDefault(zodLiteral(`"${constValue}"`), `"${constValue}"`);
      } else {
        zod = zodDate();
        if (defaultValue != null) {
          zod = zodDefault(zod, `"${defaultValue}"`);
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? `"${constValue}"` : "string",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "date-time": {
      if (constValue !== undefined) {
        zod = zodDefault(zodLiteral(`"${constValue}"`), `"${constValue}"`);
      } else {
        zod = zodDatetime();
        if (defaultValue != null) {
          zod = zodDefault(zod, `"${defaultValue}"`);
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? `"${constValue}"` : "string",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "bigint": {
      if (constValue !== undefined) {
        zod = `bigintConst(BigInt("${constValue}") as ${constValue}n)`;
        zod = zodDefault(zod, `BigInt("${constValue}") as ${constValue}n`);
        return {
          outputType: `${constValue}n`,
          importOutputTypes: () => "",
          zod,
          importZodTypes: () => {
            addTypeImport("bigint", "bigintConst", usageLocation);
            return "";
          },
        };
      }

      zod = "bigint()";
      if (defaultValue != null) {
        zod = zodDefault(zod, `BigInt("${defaultValue}")`);
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: "bigint | string",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => {
          addTypeImport("bigint", "bigint", usageLocation);
          return "";
        },
      };
    }
    case "decimal": {
      const outputType = isFormatString
        ? "Decimal | string"
        : "Decimal | number";
      const fnName = isFormatString ? "decimalStr" : "decimal";
      const constFnName = isFormatString ? "decimalStrConst" : "decimalConst";

      if (constValue !== undefined) {
        zod = `${constFnName}(new Decimal("${constValue}"))`;
        zod = zodDefault(zod, `new Decimal("${constValue}")`);
        return {
          outputType: `Decimal`,
          importOutputTypes: () =>
            addTypeImport("decimal", "Decimal", usageLocation),
          zod,
          importZodTypes: () => {
            addTypeImport("decimal", constFnName, usageLocation);
            addTypeImport("decimal", "Decimal", usageLocation);
            return "";
          },
        };
      }

      zod = `${fnName}()`;
      if (defaultValue != null) {
        zod = zodDefault(zod, `new Decimal("${defaultValue}")`);
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType,
        importOutputTypes: () =>
          addTypeImport("decimal", "Decimal", usageLocation),
        zod,
        importZodTypes: () => {
          addTypeImport("decimal", fnName, usageLocation);
          if (defaultValue != null) {
            addTypeImport("decimal", "Decimal", usageLocation);
          }
          return "";
        },
      };
    }
    case "integer":
    case "int32": {
      if (constValue !== undefined) {
        const literalValue = isFormatString ? `"${constValue}"` : constValue;
        zod = zodDefault(zodLiteral(literalValue), literalValue);
      } else if (isFormatString) {
        zod = `z.union([${zodString()}, ${zodInteger()}]).transform(v => typeof v === 'string' ? parseInt(v, 10) : v)`;
        if (defaultValue != null) {
          zod = zodDefault(zod, `"${defaultValue}"`);
        }
      } else {
        zod = zodInteger();
        if (defaultValue != null) {
          zod = zodDefault(zod, defaultValue);
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? `${constValue}` : "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "number":
    case "float32": {
      if (constValue !== undefined) {
        const literalValue = isFormatString ? `"${constValue}"` : constValue;
        zod = zodDefault(zodLiteral(literalValue), literalValue);
      } else if (isFormatString) {
        zod = `z.union([${zodString()}, ${zodNumber()}]).transform(v => typeof v === 'string' ? parseFloat(v) : v)`;
        if (defaultValue != null) {
          zod = zodDefault(zod, `"${defaultValue}"`);
        }
      } else {
        zod = zodNumber();
        if (defaultValue != null) {
          zod = zodDefault(zod, defaultValue);
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? `${constValue}` : "number",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "boolean": {
      if (constValue !== undefined) {
        zod = zodDefault(zodLiteral(constValue), constValue);
      } else {
        zod = zodBoolean();
        if (defaultValue != null) {
          zod = zodDefault(zod, defaultValue);
        }
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: constValue !== undefined ? `${constValue}` : "boolean",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "bytes":
      zod =
        'z.string().describe("Base64-encoded binary content").transform(b64$.bytesFromBase64)';

      return {
        outputType: "Uint8Array",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => {
          addImport("zod", "z", "aliasImport");
          addInternalImport("base64", "b64$", usageLocation, aliasImport);
          return "";
        },
      };
    case "map":
      const mapType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        visited,
      });

      zod = zodRecord(mapType.zod);

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: `{ [k: string]: ${mapType.outputType} }`,
        importOutputTypes: () => mapType.importOutputTypes(),
        zod,
        importZodTypes: () => {
          mapType.importZodTypes();
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    case "array":
      const arrType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        visited,
      });

      zod = zodArray(arrType.zod);

      if (defaultValue != null && typeDef.ItemType.IsPrimitive) {
        zod = zodDefault(zod, JSON.stringify(defaultValue));
      }

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: `Array<${arrType.outputType}>`,
        importOutputTypes: () => arrType.importOutputTypes(),
        zod,
        importZodTypes: () => {
          arrType.importZodTypes();
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    case "any":
      zod = zodAny();

      zod = zodDescribeFromTypeDef(zod, typeDef);

      return {
        outputType: "any",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "response":
      zod = zodInstanceof(`Response`);

      return {
        outputType: "Response",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "request":
      zod = zodInstanceof(`Request`);

      return {
        outputType: "Request",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "request-stream":
      zod = zodUnion([
        zodInstanceof(`ReadableStream<Uint8Array>`),
        zodInstanceof(`Blob`),
        zodInstanceof(`ArrayBuffer`),
        zodInstanceof(`Uint8Array`),
      ]);

      return {
        outputType:
          "ReadableStream<Uint8Array> | Blob | ArrayBuffer | Uint8Array",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "response-stream":
      zod = zodInstanceof("ReadableStream<Uint8Array>");

      return {
        outputType: "ReadableStream<Uint8Array>",
        importOutputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "event-stream":
      const eventType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        visited,
      });

      const decoder = `decoder(rawEvent) { const schema =  ${eventType.zod}; return schema.parse(rawEvent); }`;

      zod = zodTransform(
        zodInstanceof(`ReadableStream<Uint8Array>`),
        `stream => { return new EventStream({stream, ${decoder}}) }`,
      );

      return {
        outputType: `EventStream<${eventType.outputType}>`,
        importOutputTypes: () => {
          addInternalImport("event-streams", "EventStream", usageLocation);
          return eventType.importOutputTypes();
        },
        zod,
        importZodTypes: () => {
          eventType.importZodTypes();
          addInternalImport("event-streams", "EventStream", usageLocation);
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    case "jsonl":
      const jsonlEventType = resolveInbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        visited,
      });
      const jsonLDecoder = `decoder(rawEvent) { const schema =  ${jsonlEventType.zod}; return schema.parse(rawEvent); }`;

      zod = zodTransform(
        zodInstanceof(`ReadableStream<Uint8Array>`),
        `stream => { return new JsonLStream({stream, ${jsonLDecoder}}) }`,
      );

      return {
        outputType: `JsonLStream<${jsonlEventType.outputType}>`,
        importOutputTypes: () => {
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          return jsonlEventType.importOutputTypes();
        },
        zod,
        importZodTypes: () => {
          jsonlEventType.importZodTypes();
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    default:
      throw new Error(`Unknown type: ${typeName}`);
  }
}

//@ts-ignore
function sortUnionMembers(
  union: TypeDef,
): Array<[TypeDef, DiscriminatorMapping?]> {
  if (union.Discriminator == null) {
    // Sort the TypeDefs first, then map to the desired format
    const sortedTypes = sortTypeDefRequiredFieldsDescending(
      union.AssociatedTypes,
    );
    return sortedTypes.map((t): [TypeDef, DiscriminatorMapping?] => [t]);
  } else {
    const sortedDiscriminators =
      sortDiscriminatorMappingRequiredFieldsDescending(
        union.Discriminator.Mapping,
      );
    return sortedDiscriminators.map((m): [TypeDef, DiscriminatorMapping?] => [
      m.Type,
      m,
    ]);
  }
}

/**
 * Returns the environment Zod schema for the given FieldDef.
 */
function toEnvZodSchema(fieldDef: FieldDef, usageLocation: string): string {
  const type = fieldDef.Type.Type.toString();
  const dv = fieldDef.Default?.Value;
  const hasDefault = dv !== undefined;
  let zod: string | undefined = undefined;

  switch (type) {
    case "enum": {
      const types = toInbound({
        typeDef: fieldDef.Type,
        rootTypeDef: context.Global.AST.MainSDK.Type,
        defaultValue: dv,
        visited: new Set(),
        usageLocation,
      });
      types.importZodTypes();
      return types.zod;
    }
    case "any":
      return hasDefault
        ? zodDefault(zodAny(), typeof dv === "string" ? `"${dv}"` : dv)
        : zodAny();
    case "string":
      zod = zodString();

      return hasDefault ? zodDefault(zod, `"${dv}"`) : zod;
    case "date":
      zod = zodDate();

      return hasDefault ? zodDefault(zod, `"${dv}"`) : zod;
    case "date-time":
      zod = zodDatetime();

      return hasDefault ? zodDefault(zod, `"${dv}"`) : zod;
    case "bigint":
      zod = zodCoerceBigint();

      return hasDefault ? zodDefault(zod, `BigInt("${dv}")`) : zod;
    case "decimal":
      addTypeImport("decimal", "Decimal", usageLocation);
      zod = `z.string().transform(v => new Decimal(v))`;

      return hasDefault ? zodDefault(zod, `new Decimal("${dv}")`) : zod;
    case "number":
    case "float32":
      return zodDefault(zodCoerceNumber(), dv);
    case "integer":
    case "int32":
      return zodDefault(zodCoerceInteger(), dv);
    case "boolean": {
      let zod = zodEnum(["true", "false"]);

      if (hasDefault) {
        zod = zodDefault(zod, `"${String(dv)}"`);
      }

      return zodTransform(zod, `v => v === "true"`);
    }
    case "array": {
      const zodArrayItem = toEnvZodSchema(
        typeDefToFieldDef(fieldDef.Type.ItemType, fieldDef),
        usageLocation,
      );

      return zodPipe(
        zodTransform(zodString(), `v => v ? v.split(",") : undefined`),
        zodDefault(zodArray(zodArrayItem), dv),
      );
    }
    default:
      throw new Error(
        `Global paramater has unsupported type: ${fieldDef.Name}: ${type}`,
      );
  }
}

function toOutbound(options: {
  usageLocation: string;
  typeDef: TypeDef;
  rootTypeDef: TypeDef;
  constValue?: unknown;
  defaultValue?: unknown;
  visited: Set<string>;
}): Omit<ResolvedTypes, "outputType" | "importOutputTypes"> {
  const {
    usageLocation,
    typeDef,
    rootTypeDef,
    constValue,
    defaultValue,
    visited,
  } = options;

  const rootModel = rootTypeDef.ResolvedModel;
  const typeName = typeDef.Type.toString();
  const zodRef = sanitizeZodRef(typeDef, usageLocation);
  const isFormatString = typeDef.Format === "string";

  switch (typeName) {
    case "enum": {
      const enumName = sanitizeClass(typeDef, usageLocation, true);

      if (constValue !== undefined) {
        const literal = enumLiteralFromValue(enumName, typeDef, constValue);
        return {
          inputType: literal,
          importInputTypes: () => "",
          zod: `z.literal(${literal}).default(${literal})`,
          importZodTypes: () => {
            addImport("zod", "z", "aliasImport");
            return "";
          },
        };
      }

      let zod = zodRef;
      if (defaultValue != null) {
        const literal = enumLiteralFromValue(enumName, typeDef, defaultValue);
        zod += `.default(${literal})`;
      }

      return {
        inputType: enumName,
        importInputTypes: () =>
          addEnumTypeImport(typeDef, usageLocation, rootModel),
        zod,
        importZodTypes: () => {
          addZodImport(typeDef, usageLocation, rootModel);
          return "";
        },
      };
    }
    case "union": {
      const regid = typeDef.GetRegistrationID();
      const shouldTruncate = visited.has(regid);
      if (shouldTruncate) {
        return {
          inputType: sanitizeClassName(typeDef.Name),
          importInputTypes: () => "",
          zod: `z.lazy(() => ${zodRef})`,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      } else {
        visited.add(regid);
      }

      // If we're in a different location to the TypeDef we're resolving then it
      // implies we want to import the union type. Therefore, we need to import
      // a reference to the union instead of building it out. For example, this
      // code path may be hit when an operation takes a request that is a union.
      if (!inSameModel(rootModel, usageLocation, typeDef)) {
        const lazy = needsLazyRef(usageLocation, typeDef, rootTypeDef);
        return {
          inputType: sanitizeClass(typeDef, usageLocation, true),
          importInputTypes: () =>
            addTypeDefImport(typeDef, usageLocation, rootModel),
          zod: lazy ? `z.lazy(() => ${zodRef})` : zodRef,
          importZodTypes: () => {
            addZodImport(typeDef, usageLocation, rootModel);
            if (lazy) {
              addImport("zod", "z", "aliasImport");
            }
            return "";
          },
        };
      }

      // Unions with a single member should be reduced to resolving the
      // underlying type.
      if (typeDef.AssociatedTypes.length === 1) {
        return resolveOutbound({
          usageLocation,
          typeDef: typeDef.AssociatedTypes[0],
          rootTypeDef,
          optional: false,
          nullable: false,
          streamable: false,
          visited,
        });
      }

      // Iterate over each union member and resolve the runtime type info then
      // it's a matter of concatenating the results to build typescript/zod
      // unions.
      const discr = typeDef.Discriminator?.TypePropertyName;
      const unionInput: string[] = [];
      const unionImportInputTypes: (() => string)[] = [];
      const unionZod: string[] = [];
      const unionImportZodTypes: (() => string)[] = [];
      sortUnionMembers(typeDef).forEach(([at, mapping]) => {
        let { inputType, importInputTypes, zod, importZodTypes } =
          resolveOutbound({
            usageLocation,
            typeDef: at,
            rootTypeDef,
            optional: false,
            nullable: false,
            streamable: false,
            visited,
          });

        const atName = mapping?.Name;
        if (discr && !atName) {
          throw new Error(
            `union member does not have a discriminator mapping: ${typeDef.Name} > ${at.Name}`,
          );
        }

        // If we're dealing with a discriminated union, we expect union members
        // to be objects. Each object is expected to have a discriminator field
        // the value of that value is explicitly defined in the spec and our ast
        // in the disciminator mappings list. So if these expectations are met
        // then we alter the typings to more explicitly list the discriminator
        // key as having a literal string type.
        if (discr) {
          const dkey = sanitizeKey(discr);
          const dfield = sanitizeFieldName(discr);
          const dacc = sanitizeAccessor("v", dfield);
          const transform = `(v) => ({ ${dkey}: ${dacc} })`;
          let { literal, importFunc } = resolveDiscriminator(
            usageLocation,
            mapping.Type,
            rootTypeDef,
            discr,
            atName,
          );

          inputType = `(${inputType} & {${dfield}: ${literal}})`;
          zod += `.and(z.object({ ${dfield}: z.literal(${literal}) }).transform(${transform}))`;

          unionImportInputTypes.push(importFunc);
          unionImportZodTypes.push(importFunc);
        }

        unionInput.push(inputType);
        unionImportInputTypes.push(importInputTypes);
        unionZod.push(zod);
        unionImportZodTypes.push(importZodTypes);
      });

      return {
        inputType: unionInput.join(" | "),
        importInputTypes: () => {
          unionImportInputTypes.forEach((f) => f());
          return "";
        },
        zod: `z.union([${unionZod.join(", ")}])`,
        importZodTypes: () => {
          addImport("zod", "z", "aliasImport");
          unionImportZodTypes.forEach((f) => f());
          return "";
        },
      };
    }
    case "error":
    case "class": {
      const hasAdditionalProps = typeDef.Fields?.some(
        (f) => f.IsAdditionalProperties,
      );
      let zod = zodRef;
      if (hasAdditionalProps) {
        zod += "Outbound";
      }
      if (needsLazyRef(usageLocation, typeDef, rootTypeDef)) {
        zod = `z.lazy(() => ${zod})`;
      }

      return {
        inputType: sanitizeClass(typeDef, usageLocation, true),
        importInputTypes: () =>
          addTypeDefImport(typeDef, usageLocation, rootModel),
        zod: zod,
        importZodTypes: () => {
          zod.startsWith("z.") && addImport("zod", "z", "aliasImport");
          if (hasAdditionalProps) {
            const zodName = sanitizeZodName(typeDef.Name);
            const outboundName = zodName + "Outbound";

            if (
              rootModel !== "" &&
              inSameModel(rootModel, usageLocation, typeDef)
            ) {
              return "";
            }

            const { outputLocation, mfilename, extpath } = resolveModelImport({
              def: typeDef,
              usageLocation,
              name: outboundName,
            });

            const nsPrefix = getNamespacePrefix(
              typeDef.OutputLocation,
              usageLocation,
            );
            const outboundAlias = nsPrefix
              ? `${nsPrefix}_${outboundName}`
              : undefined;
            const outboundImportType = outboundAlias
              ? namedAliasImport
              : typeImport;

            if (outputLocation == usageLocation) {
              addImport(
                `./${mfilename}.js`,
                outboundName,
                outboundImportType,
                outboundAlias,
              );
            }
            if (outputLocation != usageLocation) {
              addImport(
                extpath,
                outboundName,
                outboundImportType,
                outboundAlias,
              );
            }
          }
          if (!hasAdditionalProps) {
            addZodImport(typeDef, usageLocation, rootModel);
          }
          return "";
        },
      };
    }
    case "string": {
      if (constValue !== undefined) {
        const cv = quote(String(constValue)).replaceAll("{{", `{{"{{"}}`);
        return {
          inputType: cv,
          importInputTypes: () => "",
          zod: `z.literal(${cv}).default(${cv} as const)`,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      let zod = "z.string()";
      if (defaultValue != null) {
        const dv = quote(String(defaultValue)).replaceAll("{{", `{{"{{"}}`);
        zod += `.default(${dv})`;
      }

      return {
        inputType: "string",
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "date": {
      const inputType = "string";
      let zod = "z.string().date()";

      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        zod += `.default(${cv})`;
        zod += `.refine((v) => v === "${constValue}", {message: "Value must be ${constValue}"})`;

        return {
          inputType,
          importInputTypes: () => "",
          zod,
          importZodTypes: () => {
            addImport("zod", "z", "aliasImport");
            return "";
          },
        };
      }

      if (defaultValue != null) {
        zod += `.default("${defaultValue}")`;
      }

      return {
        inputType,
        importInputTypes: () => "",
        zod,
        importZodTypes: () => {
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    }
    case "date-time": {
      const inputType = "string";
      let zod = "z.union([z.date(), z.string().transform(v => new Date(v))])";

      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        zod += `.default(new Date(${cv}))`;
        zod += `.refine((v) => v.getTime() === new Date(${cv}).getTime(), {message: "Value must be equivelant to ${constValue}"})`;
        zod += ".transform(v => v.toISOString())";

        return {
          inputType,
          importInputTypes: () => "",
          zod,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      if (defaultValue != null) {
        zod += `.default(() => new Date("${defaultValue}"))`;
      }
      zod += ".transform(v => v.toISOString())";

      return {
        inputType,
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "bigint": {
      if (constValue !== undefined) {
        let zod = `z.literal(BigInt("${constValue}") as ${constValue}n)`;
        zod += `.default(BigInt("${constValue}") as ${constValue}n)`;
        zod += isFormatString
          ? ".transform(v => `${v}`)"
          : ".transform(v => Number(v))";
        return {
          inputType: `${constValue}n`,
          importInputTypes: () => "",
          zod,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      let zod = `z.bigint()`;
      if (defaultValue != null) {
        zod += `.default(BigInt("${defaultValue}"))`;
      }
      zod += isFormatString
        ? ".transform(v => `${v}`)"
        : ".transform(v => Number(v))";

      return {
        inputType: "bigint | string",
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "decimal": {
      const baseType = isFormatString ? "string" : "number";
      const unionType = isFormatString ? "z.string()" : "z.number()";
      let zod = `z.union([z.custom<Decimal$>(x => x instanceof Decimal$), ${unionType}])`;

      if (constValue !== undefined) {
        const cv = quote(String(constValue));
        zod += `.default(new Decimal$(${cv}))`;
        zod += `.refine((v) => v.toString() === ${cv}, {message: "Value must be ${constValue}"})`;
        zod += isFormatString
          ? ".transform(v => `${v}`)"
          : ".transform(v => typeof v === 'number' ? v : v.toNumber())";

        return {
          inputType: `Decimal$ | ${baseType}`,
          importInputTypes: () =>
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
          zod,
          importZodTypes: () => {
            addImport("zod", "z", "aliasImport");
            addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
            return "";
          },
        };
      }

      if (defaultValue != null) {
        zod += `.default(() => new Decimal$("${defaultValue}"))`;
      }
      zod += isFormatString
        ? ".transform(v => `${v}`)"
        : ".transform(v => typeof v === 'number' ? v : v.toNumber())";

      return {
        inputType: `Decimal$ | ${baseType}`,
        importInputTypes: () =>
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation),
        zod,
        importZodTypes: () => {
          addImport("zod", "z", "aliasImport");
          addTypeImport("decimal", "Decimal as Decimal$", usageLocation);
          return "";
        },
      };
    }
    case "integer":
    case "int32": {
      if (constValue !== undefined) {
        let zod = `z.literal(${constValue}).default(${constValue} as const)`;
        if (isFormatString) {
          zod += ".transform(v => `${v}`)";
        }

        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          zod: zod,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      let zod = "z.int()";
      if (defaultValue != null) {
        zod += `.default(${defaultValue})`;
      }
      if (isFormatString) {
        zod += ".transform(v => `${v}`)";
      }

      return {
        inputType: "number",
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "number":
    case "float32": {
      if (constValue !== undefined) {
        let zod = `z.literal(${constValue}).default(${constValue} as const)`;
        if (isFormatString) {
          zod += ".transform(v => `${v}`)";
        }

        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          zod: zod,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      let zod = "z.number()";
      if (defaultValue != null) {
        zod += `.default(${defaultValue})`;
      }
      if (isFormatString) {
        zod += ".transform(v => `${v}`)";
      }

      return {
        inputType: "number",
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "boolean": {
      if (constValue !== undefined) {
        return {
          inputType: `${constValue}`,
          importInputTypes: () => "",
          zod: `z.literal(${constValue}).default(${constValue} as const)`,
          importZodTypes: () => addImport("zod", "z", "aliasImport"),
        };
      }

      let zod = "z.boolean()";
      if (defaultValue != null) {
        zod += `.default(${defaultValue})`;
      }

      return {
        inputType: "boolean",
        importInputTypes: () => "",
        zod,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    }
    case "bytes":
      return {
        inputType: "Uint8Array | string",
        importInputTypes: () => "",
        zod: "b64$.zodOutbound",
        importZodTypes: () =>
          addInternalImport("base64", "b64$", usageLocation, aliasImport),
      };
    case "map":
      const mapType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        nullable: typeDef.ContainsNull,
        optional: false,
        streamable: false,
        visited,
      });

      return {
        inputType: `{ [k: string]: ${mapType.inputType} }`,
        importInputTypes: () => mapType.importInputTypes(),
        zod: `z.record(z.string(), ${mapType.zod})`,
        importZodTypes: () => {
          mapType.importZodTypes();
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    case "array":
      const arrType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: typeDef.ContainsNull,
        streamable: false,
        visited,
      });

      let zod = `z.array(${arrType.zod})`;
      if (defaultValue != null && typeDef.ItemType.IsPrimitive) {
        zod += `.default(${JSON.stringify(defaultValue)})`;
      }

      return {
        inputType: `Array<${arrType.inputType}>`,
        importInputTypes: () => arrType.importInputTypes(),
        zod,
        importZodTypes: () => {
          arrType.importZodTypes();
          addImport("zod", "z", "aliasImport");
          return "";
        },
      };
    case "any":
      return {
        inputType: "any",
        importInputTypes: () => "",
        zod: "z.any()",
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "response":
      return {
        inputType: "Response",
        importInputTypes: () => "",
        zod: `z.custom<Response>(x => x instanceof Response).transform(() => { throw new Error("Response cannot be serialized") })`,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "request":
      return {
        inputType: "Request",
        importInputTypes: () => "",
        zod: `z.custom<Request>(x => x instanceof Request).transform(() => { throw new Error("Response cannot be serialized") })`,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "request-stream":
      return {
        inputType:
          "ReadableStream<Uint8Array> | Blob | ArrayBuffer | Uint8Array",
        importInputTypes: () => "",
        zod: `z.union([z.custom<ReadableStream<Uint8Array>>(x => x instanceof ReadableStream), z.custom<Blob>(x => x instanceof Blob), z.custom<ArrayBuffer>(x => x instanceof ArrayBuffer), z.custom<Uint8Array>(x => x instanceof Uint8Array)])`,
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "response-stream":
      return {
        inputType: "ReadableStream<Uint8Array>",
        importInputTypes: () => "",
        zod: "z.custom<ReadableStream<Uint8Array>>(x => x instanceof ReadableStream)",
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "event-stream":
      const eventType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        streamable: false,
        visited,
      });
      return {
        inputType: `EventStream<${eventType.inputType}>`,
        importInputTypes: () => {
          addInternalImport("event-streams", "EventStream", usageLocation);
          return eventType.importInputTypes();
        },
        zod: "z.never()",
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    case "jsonl":
      const jsonlEventType = resolveOutbound({
        usageLocation,
        typeDef: typeDef.ItemType,
        rootTypeDef,
        optional: false,
        nullable: false,
        streamable: false,
        visited,
      });
      return {
        inputType: `JsonLStream<${jsonlEventType.inputType}>`,
        importInputTypes: () => {
          addInternalImport("jsonl", "JsonLStream", usageLocation);
          return jsonlEventType.importInputTypes();
        },
        zod: "z.never()",
        importZodTypes: () => addImport("zod", "z", "aliasImport"),
      };
    default:
      throw new Error(`Unknown type: ${typeName}`);
  }
}

// MCPB Manifest Schema Types
interface MCPBAuthor {
  name: string;
  email?: string;
  url?: string;
}

interface MCPBRepository {
  type: string;
  url: string;
}

interface MCPBIcon {
  src: string;
  size: string;
  theme?: string;
}

interface MCPBLocalization {
  resources: string;
  default_locale: string;
}

interface MCPBMcpConfig {
  command: string;
  args?: string[];
  env?: Record<string, string>;
  platform_overrides?: Record<
    string,
    {
      command?: string;
      args?: string[];
      env?: Record<string, string>;
    }
  >;
}

interface MCPBServer {
  type: "python" | "node" | "binary";
  entry_point: string;
  mcp_config: MCPBMcpConfig;
}

interface MCPBTool {
  name: string;
  description?: string;
}

interface MCPBPrompt {
  name: string;
  description?: string;
  arguments?: string[];
  text: string;
}

interface MCPBUserConfigField {
  type: "string" | "number" | "boolean" | "directory" | "file";
  title: string;
  description: string;
  required?: boolean;
  default?: string | number | boolean | string[];
  multiple?: boolean;
  sensitive?: boolean;
  min?: number;
  max?: number;
}

interface MCPBUserConfig {
  [key: string]: MCPBUserConfigField;
}

interface MCPBCompatibility {
  claude_desktop?: string;
  platforms?: ("darwin" | "win32" | "linux")[];
  runtimes?: {
    python?: string;
    node?: string;
  };
}

interface MCPBManifestSchema {
  $schema?: string;
  /** @deprecated Use manifest_version instead */
  dxt_version?: "0.3";
  manifest_version: "0.3";
  name: string;
  display_name?: string;
  version: string;
  description: string;
  long_description?: string;
  author: MCPBAuthor;
  repository?: MCPBRepository;
  homepage?: string;
  documentation?: string;
  support?: string;
  icon?: string;
  icons?: MCPBIcon[];
  screenshots?: string[];
  localization?: MCPBLocalization;
  server: MCPBServer;
  tools?: MCPBTool[];
  tools_generated?: boolean;
  prompts?: MCPBPrompt[];
  prompts_generated?: boolean;
  keywords?: string[];
  license?: string;
  privacy_policies?: string[];
  compatibility?: MCPBCompatibility;
  user_config?: MCPBUserConfig;
  _meta?: Record<string, Record<string, any>>;
}
