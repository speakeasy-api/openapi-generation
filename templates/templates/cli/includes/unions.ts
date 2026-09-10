/**
 * CLI Union Flag Handling
 *
 * Handles discriminated and non-discriminated unions for CLI flag generation.
 * For discriminated unions, generates dot-notation flags for each variant.
 * For non-discriminated unions, generates a single JSON flag.
 */

/**
 * Check if a TypeDef represents a discriminated union.
 */
function isDiscriminatedUnion(typeDef: TypeDef): boolean {
  return (
    typeDef.Type.toString() === "union" &&
    typeDef.Discriminator !== undefined &&
    typeDef.Discriminator !== null &&
    typeDef.Discriminator.Mapping !== undefined &&
    typeDef.Discriminator.Mapping !== null &&
    typeDef.Discriminator.Mapping.length > 0
  );
}
registerTemplateFunc("isDiscriminatedUnion", isDiscriminatedUnion);

/**
 * Check if a TypeDef represents a non-discriminated union.
 */
function isNonDiscriminatedUnion(typeDef: TypeDef): boolean {
  return typeDef.Type.toString() === "union" && !isDiscriminatedUnion(typeDef);
}
registerTemplateFunc("isNonDiscriminatedUnion", isNonDiscriminatedUnion);

/**
 * Get the discriminator property name for a union.
 */
function getDiscriminatorKey(typeDef: TypeDef): string {
  if (!typeDef.Discriminator) {
    return "";
  }
  return typeDef.Discriminator.TypePropertyName;
}
registerTemplateFunc("getDiscriminatorKey", getDiscriminatorKey);

/**
 * Represents a variant in a discriminated union.
 */
interface UnionVariant {
  name: string; // The discriminator value (e.g., "circle", "rectangle")
  type: TypeDef; // The type definition for this variant
  flagName: string; // The sanitized flag name
}

/**
 * Get all variants from a discriminated union.
 */
function getUnionVariants(typeDef: TypeDef): UnionVariant[] {
  if (!typeDef.Discriminator) {
    return [];
  }

  return typeDef.Discriminator.Mapping.map((m) => ({
    name: m.Name,
    type: m.Type,
    flagName: sanitizeFlagName(m.Name),
  }));
}
registerTemplateFunc("getUnionVariants", getUnionVariants);

/**
 * Check if a variant type can be expanded into simple flags.
 * Checks if the variant's fields are simple enough for individual flags.
 */
function canExpandVariant(variantType: TypeDef): boolean {
  // Must be a class/object type
  if (variantType.Type.toString() !== "class") {
    return false;
  }

  const fields = variantType.Fields || [];
  for (const field of fields) {
    if (field.Const) continue;

    const fieldType = field.Type.Type.toString();

    // Reject complex types
    if (
      fieldType === "map" ||
      fieldType === "union" ||
      fieldType === "any" ||
      fieldType === "array"
    ) {
      return false;
    }

    // Allow primitives, enums, and simple nested objects (1 level)
    if (fieldType === "class") {
      // Check nested fields are all primitives
      const nestedFields = field.Type.Fields || [];
      for (const nf of nestedFields) {
        if (nf.Const) continue;
        const nfType = nf.Type.Type.toString();
        if (
          nfType === "class" ||
          nfType === "map" ||
          nfType === "union" ||
          nfType === "any"
        ) {
          return false;
        }
      }
    }
  }

  return true;
}
registerTemplateFunc("canExpandVariant", canExpandVariant);
