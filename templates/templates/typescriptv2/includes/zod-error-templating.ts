// Zod error-specific schema generation functions
// @ts-ignore
function templateZodErrorInboundObjectFields(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
): string {
  const fieldLines: string[] = [];

  for (const field of fields) {
    if (field.IsAdditionalProperties) {
      continue;
    }

    const isOptional = isErrorFieldOptional(field);

    const fieldType = resolveInbound({
      usageLocation,
      typeDef: field.Type,
      rootTypeDef,
      optional: isOptional,
      nullable: field.Nullable,
      constValue: field.Const,
      defaultValue: field.Default,
    });

    fieldType.importZodTypes();
    fieldLines.push(
      `  ${sanitizeKey(originalFieldName(field))}: ${fieldType.zod}`,
    );
  }

  return fieldLines.join(",\n");
}

// @ts-ignore
function templateZodErrorInboundSchema(
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  inboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsInbound: ResolvedTypes | boolean,
  hasFields: boolean,
  emptyType: boolean,
): string {
  const fieldsCode = templateZodErrorInboundObjectFields(
    fields,
    rootTypeDef,
    usageLocation,
  );

  // Build the base object schema with error-specific fields
  let schema = z.object(
    `{\n${fieldsCode}${fieldsCode ? ",\n" : ""}  request$: ${z.instanceof(
      "Request",
    )},\n  response$: ${z.instanceof("Response")},\n  body$: ${z.string()}}`,
  );

  // Handle additional properties with collectExtraKeys
  if (
    inboundRemaps.additionalPropsField &&
    typeof additionalPropsInbound == "object"
  ) {
    addInternalImport(
      "schemas",
      "collectExtraKeys as collectExtraKeys$",
      usageLocation,
    );
    additionalPropsInbound.importZodTypes();
    const fieldName = sanitizeFieldName(
      inboundRemaps.additionalPropsField.Name,
    );
    const catchallSchema = z.catchall(schema, additionalPropsInbound.zod, true);
    schema = `collectExtraKeys$(${catchallSchema}, "${fieldName}", false)`;
  }

  // Build the transform function that creates the error instance
  // Always need 'v' to access request$, response$, body$ properties
  let transformParam = "v";
  let transformBody = "";

  if (inboundRemaps.mappingObject && !emptyType) {
    addInternalImport("primitives", "remap as remap$", usageLocation);
    transformBody = `const remapped = remap$(v, ${inboundRemaps.mappingObject});\n    `;
  }

  let constructorArg = "";
  if (hasFields) {
    if (inboundRemaps.mappingObject && !emptyType) {
      constructorArg = "remapped, ";
    } else if (!emptyType) {
      constructorArg = "v, ";
    }
  } else {
    // For empty types, we still need to pass an empty object as the first arg
    // since the error constructor expects (data, httpMeta) signature
    constructorArg = "{}, ";
  }

  transformBody += `\nreturn new ${typeName}(${constructorArg}{ request: v.request$, response: v.response$, body: v.body$ });`;

  const transformFn = `(${transformParam}) => {\n    ${transformBody}\n  }`;
  return z.transform(schema, transformFn, true);
}

// @ts-ignore
function templateZodErrorOutboundObjectFields(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
): string {
  const fieldLines: string[] = [];

  for (const field of fields) {
    if (field.IsAdditionalProperties) {
      if (shouldFlattenAdditionalProperties(field)) {
        continue;
      }
      const additionalPropsCode =
        zodTemplating.templateOutboundAdditionalPropertiesField(
          outboundRemaps,
          additionalPropsOutbound,
        );
      if (additionalPropsCode) {
        fieldLines.push(additionalPropsCode);
      }
      continue;
    }

    fieldLines.push(
      zodTemplating.templateOutboundField(
        field,
        rootTypeDef,
        usageLocation,
        isErrorFieldOptional(field),
      ),
    );
  }

  return fieldLines.join(",\n");
}

// @ts-ignore
function templateZodErrorOutboundObject(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
): string {
  const fieldsCode = templateZodErrorOutboundObjectFields(
    fields,
    rootTypeDef,
    usageLocation,
    outboundRemaps,
    additionalPropsOutbound,
  );

  let baseSchema = z.object(`{\n${fieldsCode}}`);

  // Apply catchall for flat additional properties
  if (
    outboundRemaps.additionalPropsField &&
    additionalPropsOutbound &&
    shouldFlattenAdditionalProperties(outboundRemaps.additionalPropsField)
  ) {
    const itemType = resolveOutbound({
      usageLocation,
      typeDef: outboundRemaps.additionalPropsField.Type.ItemType,
      rootTypeDef,
      optional: false,
      nullable: outboundRemaps.additionalPropsField.Type.ContainsNull,
    });
    itemType.importZodTypes();
    baseSchema = z.catchall(baseSchema, itemType.zod);
  }

  return zodTemplating.applyRemapTransform(
    baseSchema,
    outboundRemaps,
    additionalPropsOutbound,
    usageLocation,
  );
}

// @ts-ignore
function templateZodErrorOutboundSchema(
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
  hasFields: boolean,
): string {
  if (!hasFields) {
    return z.transform(z.instanceof(typeName), `v => v.data$`, true);
  }

  const transformedSchema = z.transform(
    z.instanceof(typeName),
    `v => v.data$`,
    true,
  );
  const fieldsSchema = templateZodErrorOutboundObject(
    fields,
    rootTypeDef,
    usageLocation,
    outboundRemaps,
    additionalPropsOutbound,
  );

  return z.pipe(transformedSchema, fieldsSchema, true);
}

// @ts-ignore
function templateErrorInboundSchemaDeclaration(
  zodName: string,
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  inboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsInbound: ResolvedTypes | false,
  hasFields: boolean,
  emptyType: boolean,
): string {
  const schemaImpl = templateZodErrorInboundSchema(
    typeName,
    fields,
    rootTypeDef,
    usageLocation,
    inboundRemaps,
    additionalPropsInbound,
    hasFields,
    emptyType,
  );

  const lines = [
    "/** @internal */",
    `export const ${zodName}inboundSchema: ${z.ZodType(
      typeName,
      "unknown",
    )} = ${schemaImpl}`,
  ];

  return lines.join("\n");
}

// @ts-ignore
function templateErrorOutboundSchemaDeclaration(
  zodName: string,
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
  hasFields: boolean,
): string {
  const schemaImpl = templateZodErrorOutboundSchema(
    typeName,
    fields,
    rootTypeDef,
    usageLocation,
    outboundRemaps,
    additionalPropsOutbound,
    hasFields,
  );

  const outboundTypeName = `${zodName}Outbound`;
  const lines = [
    "/** @internal */",
    `export const ${zodName}outboundSchema: ${z.ZodType(
      outboundTypeName,
      typeName,
    )} = ${schemaImpl}`,
  ];

  return lines.join("\n");
}

registerTemplateFunc(
  "templateZodErrorInboundSchema",
  templateZodErrorInboundSchema,
);
registerTemplateFunc(
  "templateZodErrorOutboundSchema",
  templateZodErrorOutboundSchema,
);
registerTemplateFunc(
  "templateErrorInboundSchemaDeclaration",
  templateErrorInboundSchemaDeclaration,
);
registerTemplateFunc(
  "templateErrorOutboundSchemaDeclaration",
  templateErrorOutboundSchemaDeclaration,
);
