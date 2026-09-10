/**
 * Determines the SSE data encoding for a given type by inspecting the type
 * at template time. No AST-level encoding annotation is needed.
 *
 * - String/enum-string types → "string"
 * - Object/container types  → "application/json"
 * - Anything else           → "string"
 */
function sseDataTypeEncoding(typeDef: TypeDef): string {
  const t = typeDef.Type.toString();

  if (t === "string") {
    return "string";
  }

  if (t === "enum" && typeDef.Enum?.Type?.Type?.toString() === "string") {
    return "string";
  }

  // Union types: check if it's a mixed union (JSON + string variants)
  if (t === "union") {
    if (isSSEMixedUnion(typeDef)) {
      return "auto";
    }
    // Check if all variants are string-like
    const allString = typeDef.AssociatedTypes.every(
      (a) => sseDataTypeEncoding(a) === "string",
    );
    if (allString) {
      return "string";
    }
    return "application/json";
  }

  // Other object-like types: class, map, any, array, set
  if (["class", "map", "any", "array", "set"].includes(t)) {
    return "application/json";
  }

  return "string";
}
registerTemplateFunc("sseDataTypeEncoding", sseDataTypeEncoding);

/**
 * Returns true when the given TypeDef is a union that mixes JSON-serializable
 * variants with plain-string variants. Used at template time to generate
 * appropriate SSE data parsing code (e.g. try-JSON-then-fallback-to-string).
 */
function isSSEMixedUnion(typeDef: TypeDef): boolean {
  if (typeDef.Type.toString() !== "union") {
    return false;
  }

  let hasString = false;
  let hasNonString = false;

  for (const assoc of typeDef.AssociatedTypes) {
    if (sseDataTypeEncoding(assoc) === "string") {
      hasString = true;
    } else {
      hasNonString = true;
    }
  }

  return hasString && hasNonString;
}
registerTemplateFunc("isSSEMixedUnion", isSSEMixedUnion);

/**
 * Returns the resolved SSE-overload discriminator field/parameter name for an
 * operation. Throws if the operation does not have SSE overload enabled —
 * callers must gate on op.Extensions.SSEOverload presence before invoking.
 */
function sseStreamFieldName(op: Operation): string {
  if (!op.Extensions.SSEOverload) {
    throw new Error(
      `sseStreamFieldName called on operation without SSE overload: ${op.ID}`,
    );
  }
  return op.Extensions.SSEOverload.Name;
}
registerTemplateFunc("sseStreamFieldName", sseStreamFieldName);

/**
 * Reports whether the given field is the SSE-overload discriminator for the
 * operation. False when the operation has no SSE overload enabled.
 */
function isSSEStreamField(op: Operation, field: FieldDef): boolean {
  if (!op.Extensions.SSEOverload) {
    return false;
  }
  return sanitizeFieldName(field.Name) === op.Extensions.SSEOverload.Name;
}
registerTemplateFunc("isSSEStreamField", isSSEStreamField);
