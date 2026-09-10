// Shared helper functions for Zod schema generation

// @ts-ignore
const zodTemplating = {
  templateOutboundAdditionalPropertiesField: function (
    outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
    additionalPropsOutbound: ResolvedTypes | false,
  ): string | null {
    if (additionalPropsOutbound && outboundRemaps.additionalPropsField) {
      const fieldName = declareModelField(outboundRemaps.additionalPropsField);
      additionalPropsOutbound.importZodTypes();
      return `  ${fieldName}: ${additionalPropsOutbound.zod}`;
    }
    return null;
  },

  templateOutboundField: function (
    field: FieldDef,
    rootTypeDef: TypeDef,
    usageLocation: string,
    isOptional: boolean,
  ): string {
    const fieldName = declareModelField(field);

    const fieldType = resolveOutbound({
      usageLocation,
      typeDef: field.Type,
      rootTypeDef,
      optional: isOptional,
      nullable: field.Nullable,
      streamable: isMultipartFileField(field),
      constValue: field.Const,
      defaultValue: field.Default,
    });

    fieldType.importZodTypes();
    return `  ${fieldName}: ${fieldType.zod}`;
  },

  applyRemapTransform: function (
    baseSchema: string,
    outboundRemaps: { mappingObject: string; additionalPropsField: FieldDef },
    additionalPropsOutbound: ResolvedTypes | false,
    usageLocation: string,
  ): string {
    const { mappingObject, additionalPropsField } = outboundRemaps;

    // No transform needed when using flat additional properties or no remapping
    if (!mappingObject && !additionalPropsField) {
      return baseSchema;
    }

    let transformBody = "";
    let appliedTransform = false;

    if (additionalPropsOutbound) {
      transformBody = "return {";
      if (!shouldFlattenAdditionalProperties(additionalPropsField)) {
        // Case 1: Has additional properties (with or without mapping)
        // When using flat additional properties, we can skip remapping
        transformBody += `\n      ...v.${sanitizeModelField(
          additionalPropsField,
        )},`;
        appliedTransform = true;
      }

      if (mappingObject) {
        addInternalImport("primitives", "remap as remap$", usageLocation);
        transformBody += `\n      ...remap$(v, ${mappingObject})`;
        appliedTransform = true;
      }

      transformBody += `\n    };`;
    } else if (mappingObject) {
      // Case 2: Only mapping, no additional properties
      addInternalImport("primitives", "remap as remap$", usageLocation);
      transformBody = `return remap$(v, ${mappingObject});`;
      appliedTransform = true;
    }

    if (!appliedTransform) {
      return baseSchema;
    }

    // Generate isomorphic transform code using the z.transform helper
    const transformFn = `(v) => {\n    ${transformBody}\n  }`;
    return z.transform(baseSchema, transformFn);
  },
};

function templateBase64OutboundSchema(): string {
  const uint8ArraySchema = z.instanceof("Uint8Array", true);
  const stringTransform = z.transform(z.string(), "stringToBytes");
  return z.or(uint8ArraySchema, stringTransform, true);
}

function templateBase64InboundSchema(): string {
  const uint8ArraySchema = z.instanceof("Uint8Array", true);
  const stringTransform = z.transform(z.string(), "bytesFromBase64");
  return z.or(uint8ArraySchema, stringTransform, true);
}

registerTemplateFunc(
  "templateBase64OutboundSchema",
  templateBase64OutboundSchema,
);
registerTemplateFunc(
  "templateBase64InboundSchema",
  templateBase64InboundSchema,
);
