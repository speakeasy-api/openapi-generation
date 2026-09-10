/** Defines a Zod schema and its imports. */
type ZodSchemaDef = {
  /** Templated Zod schema, e.g. z.number(). */
  zod: string;

  /** Imports for Zod schema. Includes zod itself and reference imports. */
  importZodTypes: () => string;
};

/**
 * Returns the Zod schema of zodSchema.and(intersectionZodSchema).
 */
function zodAnd(zodSchema: string, intersectionZodSchema: string): string {
  return `${zodSchema}.and(${intersectionZodSchema})`;
}

/**
 * Returns the Zod schema of z.any().
 */
function zodAny(): string {
  return `z.any()`;
}

/**
 * Returns the Zod schema of z.array(zodSchema).
 */
function zodArray(zodSchema: string): string {
  return `z.array(${zodSchema})`;
}

/**
 * Returns the Zod schema of z.string().base64().
 */
function zodBase64(): string {
  return `z.string().base64()`;
}

/**
 * Returns the Zod schema of z.bigint().
 */
function zodBigint(): string {
  return `z.bigint()`;
}

/**
 * Returns the Zod schema of z.boolean().
 */
function zodBoolean(): string {
  return `z.boolean()`;
}

/**
 * Returns the Zod schema as zodSchema.catchall(catchallZodSchema).
 */
function zodCatchall(zodSchema: string, catchallZodSchema: string): string {
  return `${zodSchema}.catchall(${catchallZodSchema})`;
}

/**
 * Returns the Zod schema of z.coerce.bigint().
 */
function zodCoerceBigint(): string {
  return `z.coerce.bigint()`;
}

/**
 * Returns the Zod schema of z.coerce.number().int().
 */
function zodCoerceInteger(): string {
  return `z.coerce.number().int()`;
}

/**
 * Returns the Zod schema of z.coerce.number().
 */
function zodCoerceNumber(): string {
  return `z.coerce.number()`;
}

/**
 * Returns the Zod schema of z.string().date().
 */
function zodDate(): string {
  return `z.string().date()`;
}

/**
 * Returns the Zod schema of z.iso.datetime({ offset: true }).
 */
function zodDatetime(): string {
  return `z.iso.datetime({ offset: true })`;
}

/**
 * Returns the Zod schema as zodSchema.default(value).
 */
function zodDefault(zodSchema: string, defaultValue: unknown): string {
  if (typeof defaultValue === "undefined") {
    return zodSchema;
  }

  return `${zodSchema}.default(${defaultValue})`;
}

/**
 * Returns the Zod schema as zodSchema.describe(description).
 */
function zodDescribe(
  zodSchema: string,
  description: string | undefined,
): string {
  if (!description) {
    return zodSchema;
  }

  return `${zodSchema}.describe(${templateStringValue(description, null)})`;
}

/**
 * Returns the description for the Zod schema for the given TypeDef.
 */
function zodDescribeFromTypeDef(zodSchema: string, typeDef: TypeDef): string {
  const description = typeDef.Comments?.Description;

  return zodDescribe(zodSchema, sanitizeComments(description));
}

/**
 * Returns the Zod schema of z.enum([...values]). Only valid for string enums.
 */
function zodEnum(values: string[]): string {
  return `z.enum([\n` + values.map((v) => `"${v}",\n`).join("") + `])`;
}

/**
 * Returns the Zod schema for an enum using z.enum(). If non-string enum, which
 * is unsupported by Zod, returns z.union() of z.literal(value).
 */
function zodEnumFromTypeDef(typeDef: TypeDef): string {
  addImport("zod", "z", "aliasImport");

  if (!typeDef.Enum || typeDef.Enum == null || typeDef.Enum.Values == null) {
    return "z.enum([])";
  }

  const enumType = typeDef.Enum.Type.Type.toString();

  let zod =
    enumType === "string"
      ? zodEnum(typeDef.Enum.Values)
      : typeDef.Enum.Values.length === 1
      ? zodLiteral(
          sanitizeEnumValue(typeDef.Enum.Values[0], typeDef.Enum.Type.Type),
        )
      : zodUnion(typeDef.Enum.Values.map((value) => zodLiteral(value)));

  if (typeDef.Enum.Open) {
    const enumTypeZod = enumType === "string" ? zodString() : zodNumber();
    const usageLocation = getModelsLocation(typeDef.OutputLocation);

    addTypeImport("enums", "catchUnrecognizedEnum", usageLocation);

    zod = zodUnion([zod, zodTransform(enumTypeZod, "catchUnrecognizedEnum")]);
  }

  zod = zodDescribe(zod, sanitizeComments(typeDef.Comments?.Description));

  return zod;
}

registerTemplateFunc("zodEnumFromTypeDef", zodEnumFromTypeDef);

/**
 * Returns the Zod schema using z.custom for instanceof check.
 */
function zodInstanceof(value: any): string {
  // Strip generics for the instanceof check by taking everything before first <
  const str = String(value);
  const stripped = str.includes("<") ? str.substring(0, str.indexOf("<")) : str;
  return `z.custom<${value}>(x => x instanceof ${stripped})`;
}

/**
 * Returns the Zod schema of z.int().
 */
function zodInteger(): string {
  return `z.int()`;
}

/**
 * Returns the Zod schema as z.lazy(() => zodSchemaRef).
 */
function zodLazy(zodSchemaRef: string): string {
  return `z.lazy(() => ${zodSchemaRef})`;
}

/**
 * Returns the Zod schema of z.literal(value) where value is unsanitized.
 */
function zodLiteral(value: any): string {
  return `z.literal(${value})`;
}

/**
 * Returns the Zod schema of z.object({...zodSchemas}).
 */
function zodObject(zodSchemas: Record<string, string>): string {
  return (
    `z.object({\n` +
    Object.keys(zodSchemas)
      .sort()
      .map((key) => `${key}: ${zodSchemas[key]},\n`)
      .join("") +
    `})`
  );
}

/**
 * Returns the Zod schema of z.object() from the TypeDef. If the TypeDef has
 * additional properties, that field is omitted from the Zod object and it is
 * handled via catchall and collectExtraKeys.
 */
function zodObjectFromTypeDef(typeDef: TypeDef): string {
  let additionalPropertiesField: FieldDef = undefined;
  const usageLocation = getModelsLocation(typeDef.OutputLocation);
  const zodSchemas: Record<string, string> = {};

  for (const field of typeDef.Fields) {
    if (field.IsAdditionalProperties) {
      additionalPropertiesField = field;

      continue;
    }

    const ann = field.Annotations?.Get("security");
    const isSecurityOption = ann && isSecurityAnnotation(ann) && ann.Option;
    const name = isSecurityOption
      ? field.Name
      : sanitizeFieldName(originalFieldName(field));
    let zodSchema: string;

    if (field.IsResponseHeaders) {
      // Prevent validation errors when Headers field is undefined
      // This can happen if a response expects content but no headers,
      // while other responses in the same operation do define headers.
      zodSchema = "z.record(z.string(), z.array(z.string())).default({})";
    } else {
      const zod = resolveInbound({
        constValue: field.Const,
        defaultValue: field.Default,
        encoding: getFieldContentType(field),
        nullable: field.Nullable,
        optional: isInputOptional(field),
        rootTypeDef: typeDef,
        typeDef: field.Type,
        usageLocation,
      });

      callImportTypesFunc(zod.importZodTypes);
      zodSchema = zod.zod;

      if (!zodSchema.includes(".describe(")) {
        const fieldDescription = sanitizeComments(
          [field.Comments?.Summary, field.Comments?.Description]
            .filter(Boolean)
            .join("\n\n"),
        );
        if (fieldDescription) {
          zodSchema = zodDescribe(zodSchema, fieldDescription);
        }
      }
    }

    zodSchemas[name] = zodSchema;
  }

  addImport("zod", "z", "aliasImport");
  let zod = zodObject(zodSchemas);

  zod = zodDescribe(zod, sanitizeComments(typeDef.Comments?.Description));

  if (additionalPropertiesField) {
    const additionalPropertiesZod = resolveInbound({
      nullable: false,
      optional: false,
      rootTypeDef: typeDef,
      typeDef: additionalPropertiesField.Type.ItemType,
      usageLocation,
    });

    addInternalImport(
      "schemas",
      "collectExtraKeys as collectExtraKeys$",
      usageLocation,
    );
    callImportTypesFunc(additionalPropertiesZod.importZodTypes);

    zod = zodCatchall(zod, additionalPropertiesZod.zod);
    zod = `collectExtraKeys$(\n${zod}, "${sanitizeFieldName(
      originalFieldName(additionalPropertiesField),
    )}", ${additionalPropertiesField.Optional})`;
  }

  return zod;
}

registerTemplateFunc("zodObjectFromTypeDef", zodObjectFromTypeDef);

/**
 * Returns the Zod schema of z.object() for outbound serialization.
 * Uses transform to flatten additional properties back out.
 */
function zodObjectFromTypeDefOutbound(typeDef: TypeDef): string {
  let additionalPropertiesField: FieldDef = undefined;
  const usageLocation = getModelsLocation(typeDef.OutputLocation);
  const zodSchemas: Record<string, string> = {};

  for (const field of typeDef.Fields) {
    if (field.IsAdditionalProperties) {
      additionalPropertiesField = field;
      continue;
    }

    const ann = field.Annotations?.Get("security");
    const isSecurityOption = ann && isSecurityAnnotation(ann) && ann.Option;
    const name = isSecurityOption
      ? field.Name
      : sanitizeFieldName(originalFieldName(field));
    let zodSchema: string;

    if (field.IsResponseHeaders) {
      zodSchema = "z.record(z.string(), z.array(z.string())).default({})";
    } else {
      const zod = resolveOutbound({
        constValue: field.Const,
        defaultValue: field.Default,
        nullable: field.Nullable,
        optional: isInputOptional(field),
        rootTypeDef: typeDef,
        typeDef: field.Type,
        usageLocation,
      });

      callImportTypesFunc(zod.importZodTypes);
      zodSchema = zod.zod;
    }

    zodSchemas[name] = zodSchema;
  }

  addImport("zod", "z", "aliasImport");
  let zod = "";

  if (additionalPropertiesField) {
    const additionalPropertiesZod = resolveOutbound({
      nullable: false,
      optional: false,
      rootTypeDef: typeDef,
      typeDef: additionalPropertiesField.Type.ItemType,
      usageLocation,
    });

    callImportTypesFunc(additionalPropertiesZod.importZodTypes);

    const fieldName = sanitizeFieldName(
      originalFieldName(additionalPropertiesField),
    );
    zodSchemas[fieldName] = zodOptional(zodRecord(additionalPropertiesZod.zod));

    zod = zodObject(zodSchemas);
    zod = zodDescribe(zod, sanitizeComments(typeDef.Comments?.Description));

    const transformFn = `(v) => {\n    const { ${fieldName}, ...rest } = v;\n    return {\n      ...rest,\n      ...${fieldName},\n    };\n  }`;
    zod = zodTransform(zod, transformFn);
  } else {
    zod = zodObject(zodSchemas);
    zod = zodDescribe(zod, sanitizeComments(typeDef.Comments?.Description));
  }

  return zod;
}

registerTemplateFunc(
  "zodObjectFromTypeDefOutbound",
  zodObjectFromTypeDefOutbound,
);

/**
 * Returns the Zod schema as zodSchema.optional().
 */
function zodOptional(zodSchema: string): string {
  return `${zodSchema}.optional()`;
}

/**
 * Returns the Zod schema of z.null().
 */
function zodNull(): string {
  return `z.null()`;
}

/**
 * Returns the Zod schema as zodSchema.nullable().
 */
function zodNullable(zodSchema: string): string {
  return `${zodSchema}.nullable()`;
}

/**
 * Returns the Zod schema of z.number().
 */
function zodNumber(): string {
  return `z.number()`;
}

/**
 * Returns the Zod schema as zodSchema.pipe(targetSchema).
 */
function zodPipe(zodSchema: string, targetSchema: string): string {
  return `${zodSchema}.pipe(${targetSchema})`;
}

/**
 * Returns the Zod schema of z.record(z.string(), zodSchema).
 */
function zodRecord(zodSchema: string): string {
  return `z.record(z.string(), ${zodSchema})`;
}

/**
 * Returns the Zod schema of z.string().
 */
function zodString(): string {
  return `z.string()`;
}

/**
 * Returns the Zod schema as zodSchema.transform(transformFunc) where
 * transformFunc is unsanitized.
 */
function zodTransform(zodSchema: string, transformFunc: string): string {
  return `${zodSchema}.transform(${transformFunc})`;
}

/**
 * Returns the Zod schema of z.union([...zodSchemas]).
 */
function zodUnion(zodSchemas: string[]): string {
  return (
    `z.union([\n` +
    zodSchemas.map((zodSchema) => `${zodSchema},\n`).join("") +
    `])`
  );
}

/**
 * Returns the Zod schema of z.union() from the TypeDef.
 */
function zodUnionFromTypeDef(typeDef: TypeDef): string {
  const zod = resolveInbound({
    nullable: false,
    optional: false,
    rootTypeDef: typeDef,
    typeDef,
    usageLocation: getModelsLocation(typeDef.OutputLocation),
  });

  addImport("zod", "z", "aliasImport");
  callImportTypesFunc(zod.importZodTypes);

  return zodDescribe(zod.zod, sanitizeComments(typeDef.Comments?.Description));
}

registerTemplateFunc("zodUnionFromTypeDef", zodUnionFromTypeDef);

/**
 * Returns the Zod schema of z.void().
 */
function zodVoid(): string {
  return `z.void()`;
}
