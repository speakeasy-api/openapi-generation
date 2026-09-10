/* Open Union helpers for forward-compatible union parsing in Go */

/**
 * Returns true if any type in the schema has IsUnionOpen set,
 * meaning we need to emit open union support code.
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

/**
 * Returns the Unknown enum constant name for a union type.
 * e.g. for unionType "VehicleType" returns "VehicleTypeUnknown"
 */
function openUnionUnknownEnumName(unionType: string): string {
  return `${unionType}Unknown`;
}

registerTemplateFunc("openUnionUnknownEnumName", openUnionUnknownEnumName);

/**
 * Returns the Go string literal for the unknown enum value.
 *
 * For tagged (discriminated) unions: uses findUniqueUnknownValue() to pick a
 * sentinel that doesn't collide with existing discriminator mappings.
 *
 * For untagged unions: uses a fixed "Unknown" sentinel — there are no
 * discriminator values to collide with.
 */
function openUnionUnknownEnumValue(
  typeDef: TypeDef,
  unionType: string,
  outputLocation: string,
): string {
  const stringType: GojaEnum<DataType> = {
    valueOf: () => "string",
    toString: () => "string",
  };

  if (!typeDef.Discriminator?.Mapping) {
    return sanitizeEnumValue("Unknown", stringType);
  }

  const rawValue = findUniqueUnknownValue(
    typeDef,
    typeDef.Discriminator.Mapping,
  );
  return sanitizeEnumValue(rawValue, stringType);
}

registerTemplateFunc("openUnionUnknownEnumValue", openUnionUnknownEnumValue);
