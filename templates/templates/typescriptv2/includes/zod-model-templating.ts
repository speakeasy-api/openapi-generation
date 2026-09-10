// Zod model-specific schema generation functions

// @ts-ignore
function templateZodModelInboundObjectFields(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
): string {
  const fieldLines: string[] = [];

  for (const field of fields) {
    if (field.IsAdditionalProperties) {
      continue;
    }

    const keyName = isParameterField(field)
      ? field.Name
      : outboundSecurityKeyName(field, fields);
    const isOptional = isInputOptional(field);

    const fieldType = resolveInbound({
      usageLocation,
      typeDef: field.Type,
      rootTypeDef,
      optional: isOptional,
      nullable: field.Nullable,
      constValue: field.Const,
      defaultValue: field.Default,
      encoding: getFieldContentType(field),
    });

    if (field.IsResponseHeaders) {
      // Prevent `safeParseResponse` from raising a `ResponseValidationError` when receiving
      // `undefined` for the required `Headers` field. This can happen if a response expects
      // content but no headers, while other responses in the same operation do define headers.
      fieldType.zod = z.default(z.record(z.array(z.string())), "{}");
      // Don't call importZodTypes since we're using a hardcoded schema
      addZodPackageImport();
    } else {
      fieldType.importZodTypes();
    }
    fieldLines.push(`  ${sanitizeKey(keyName)}: ${fieldType.zod}`);
  }

  return fieldLines.join(",\n");
}

// @ts-ignore
function templateZodModelInboundObject(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  inboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsInbound: ResolvedTypes | false,
): string {
  const fieldsCode = templateZodModelInboundObjectFields(
    fields,
    rootTypeDef,
    usageLocation,
  );

  let schema = z.object(`{\n${fieldsCode}}`);

  // Handle additional properties
  if (inboundRemaps.additionalPropsField && additionalPropsInbound) {
    additionalPropsInbound.importZodTypes();
    schema = z.catchall(schema, additionalPropsInbound.zod);

    // For non-flat mode, wrap with collectExtraKeys to gather into a nested field
    if (
      !shouldFlattenAdditionalProperties(inboundRemaps.additionalPropsField)
    ) {
      addInternalImport(
        "schemas",
        "collectExtraKeys as collectExtraKeys$",
        usageLocation,
      );
      const fieldName = sanitizeFieldName(
        inboundRemaps.additionalPropsField.Name,
      );
      const isOptional = inboundRemaps.additionalPropsField.Optional;
      schema = `collectExtraKeys$(${schema}, "${fieldName}", ${isOptional})`;
    }
  }

  // Handle remapping
  if (inboundRemaps.mappingObject) {
    addInternalImport("primitives", "remap as remap$", usageLocation);
    const transformFn = `(v) => {\n    return remap$(v, ${inboundRemaps.mappingObject});\n  }`;
    schema = z.transform(schema, transformFn);
  }

  return schema;
}

// @ts-ignore
function templateZodModelOutboundObjectFields(
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
        isInputOptional(field),
      ),
    );
  }

  return fieldLines.join(",\n");
}

// @ts-ignore
function templateZodModelOutboundObject(
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
): string {
  const fieldsCode = templateZodModelOutboundObjectFields(
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
function templateInboundSchemaDeclaration(
  zodName: string,
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  inboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsInbound: ResolvedTypes | false,
): string {
  const schemaImpl = templateZodModelInboundObject(
    fields,
    rootTypeDef,
    usageLocation,
    inboundRemaps,
    additionalPropsInbound,
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
function templateOutboundSchemaDeclaration(
  zodName: string,
  typeName: string,
  fields: FieldDef[],
  rootTypeDef: TypeDef,
  usageLocation: string,
  outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
  additionalPropsOutbound: ResolvedTypes | false,
): string {
  const schemaImpl = templateZodModelOutboundObject(
    fields,
    rootTypeDef,
    usageLocation,
    outboundRemaps,
    additionalPropsOutbound,
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
  "templateInboundSchemaDeclaration",
  templateInboundSchemaDeclaration,
);
registerTemplateFunc(
  "templateOutboundSchemaDeclaration",
  templateOutboundSchemaDeclaration,
);
