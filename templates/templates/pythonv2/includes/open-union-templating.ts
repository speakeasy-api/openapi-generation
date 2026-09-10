/* Open Discriminated Union helpers for forward-compatible union parsing in Python */

/**
 * Returns true if any type in the schema has IsUnionOpen set,
 * meaning we need to emit open discriminated union support code.
 */
function shouldTemplateOpenDiscriminatedUnion(): boolean {
  const allTypes = getAllTypes();
  for (const t of allTypes) {
    if (t.IsUnionOpen) {
      return true;
    }
  }
  return false;
}

registerTemplateFunc(
  "shouldTemplateOpenDiscriminatedUnion",
  shouldTemplateOpenDiscriminatedUnion,
);

/** Returns the Unknown fallback class name, e.g. "UnknownShape" */
function openUnionUnknownClassName(typeDef: TypeDef): string {
  return `Unknown${sanitizeClassName(typeDef.Name)}`;
}

registerTemplateFunc("openUnionUnknownClassName", openUnionUnknownClassName);

/** Returns the sanitized discriminator field name (snake_case) */
function openUnionDiscriminatorFieldName(typeDef: TypeDef): string {
  return sanitizeFieldName(typeDef.Discriminator.TypePropertyName);
}

registerTemplateFunc(
  "openUnionDiscriminatorFieldName",
  openUnionDiscriminatorFieldName,
);

/** Returns the raw discriminator property name for dict lookups */
function openUnionDiscriminatorPropertyName(typeDef: TypeDef): string {
  return typeDef.Discriminator.TypePropertyName;
}

registerTemplateFunc(
  "openUnionDiscriminatorPropertyName",
  openUnionDiscriminatorPropertyName,
);

/** Returns the variants dict name, e.g. "_SHAPE_VARIANTS" */
function openUnionVariantsDictName(typeDef: TypeDef): string {
  return `_${sanitizeFieldName(typeDef.Name).toUpperCase()}_VARIANTS`;
}

registerTemplateFunc("openUnionVariantsDictName", openUnionVariantsDictName);

/** Returns the parser function name, e.g. "_parse_shape" */
function openUnionParserFuncName(typeDef: TypeDef): string {
  return `_parse_${sanitizePrivateMethodName(typeDef.Name)}`;
}

registerTemplateFunc("openUnionParserFuncName", openUnionParserFuncName);

/**
 * Returns formatted dict entries for the variants mapping.
 * Each entry is on its own line, indented with 4 spaces:
 *     "circle": Circle,
 *     "rectangle": Rectangle,
 */
function openUnionVariantEntries(
  typeDef: TypeDef,
  usageLocation: string,
  owningModelInfo: PythonV2Type,
  suffixForTypes?: string,
): string {
  const entries: string[] = [];

  for (const mapping of typeDef.Discriminator.Mapping) {
    const memberTypeName = sanitizePydanticType(
      typeDefToFieldDef(mapping.Type),
      { optional: false, nullable: false },
      usageLocation,
      owningModelInfo,
      true,
      false,
      suffixForTypes,
    );
    entries.push(`    ${JSON.stringify(mapping.Name)}: ${memberTypeName},`);
  }

  return entries.join("\n");
}

registerTemplateFunc("openUnionVariantEntries", openUnionVariantEntries);

/**
 * Returns the pipe-separated return type annotation including the Unknown class.
 * e.g. "Circle | Rectangle | UnknownShape"
 */
function openUnionReturnType(
  typeDef: TypeDef,
  usageLocation: string,
  owningModelInfo: PythonV2Type,
  suffixForTypes?: string,
): string {
  const memberTypeNames: string[] = [];

  for (const mapping of typeDef.Discriminator.Mapping) {
    const memberTypeName = sanitizePydanticType(
      typeDefToFieldDef(mapping.Type),
      { optional: false, nullable: false },
      usageLocation,
      owningModelInfo,
      true,
      false,
      suffixForTypes,
    );
    memberTypeNames.push(memberTypeName);
  }

  const unknownClassName = `Unknown${sanitizeClassName(typeDef.Name)}`;
  return [...memberTypeNames, unknownClassName].join(" | ");
}

registerTemplateFunc("openUnionReturnType", openUnionReturnType);

/**
 * Side-effect function that adds all imports required by the open union template.
 * Returns empty string (used for side effects only).
 */
function openUnionAddImports(
  typeDef: TypeDef,
  usageLocation: string,
  owningModelInfo: PythonV2Type,
  suffixForTypes?: string,
): string {
  // Imports for the Unknown fallback model
  addImport("types", "BaseModel", true);
  addImport("pydantic", "ConfigDict");
  addImport("typing", "Any");
  addImport("typing", "Literal");

  // Trigger imports for each variant member type
  for (const mapping of typeDef.Discriminator.Mapping) {
    sanitizePydanticType(
      typeDefToFieldDef(mapping.Type),
      { optional: false, nullable: false },
      usageLocation,
      owningModelInfo,
      true,
      false,
      suffixForTypes,
    );
  }

  return "";
}

registerTemplateFunc("openUnionAddImports", openUnionAddImports);

/** Returns the sentinel discriminator value for the Unknown variant (e.g. "UNKNOWN") */
function openUnionUnknownDiscriminatorValue(typeDef: TypeDef): string {
  if (!typeDef.Discriminator || !typeDef.Discriminator.Mapping) {
    return DEFAULT_UNKNOWN_DISCRIMINATOR_VALUE;
  }
  return findUniqueUnknownValue(typeDef, typeDef.Discriminator.Mapping);
}

registerTemplateFunc(
  "openUnionUnknownDiscriminatorValue",
  openUnionUnknownDiscriminatorValue,
);
