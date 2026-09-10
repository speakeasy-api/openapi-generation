/* Shared helpers for open discriminated unions across all language targets */

function getDiscriminatorDisplayName(mapping: DiscriminatorMapping): string {
  return mapping.DisplayName || mapping.Name;
}

const DEFAULT_UNKNOWN_DISCRIMINATOR_VALUE = "UNKNOWN";

/**
 * Finds a unique unknown value that doesn't conflict with any discriminator mapping.
 * Tries "UNKNOWN", "_UNKNOWN_", "__UNKNOWN__", etc. up to 10 underscores on each side.
 * Throws an error if no unique value can be found.
 */
// @ts-ignore
function findUniqueUnknownValue(
  typeDef: TypeDef,
  discriminatorMappings: DiscriminatorMapping[],
): string {
  const mappingNames = new Set(discriminatorMappings.map((m) => m.Name));

  for (let i = 0; i <= 10; i++) {
    const underscores = "_".repeat(i);
    const candidate = `${underscores}${DEFAULT_UNKNOWN_DISCRIMINATOR_VALUE}${underscores}`;
    if (!mappingNames.has(candidate)) {
      return candidate;
    }
  }

  throw new Error(
    `Cannot generate discriminated union for type "${typeDef.Name}": ` +
      `the discriminator mapping contains "UNKNOWN" and all underscore-wrapped variants ` +
      `(up to 10 underscores). Please modify your OpenAPI spec to use different discriminator values.`,
  );
}

let cachedHasAnyOpenUnion: boolean | null = null;

// Check if any union in the AST is open.
// @ts-ignore
function hasAnyOpenUnion(): boolean {
  if (cachedHasAnyOpenUnion !== null) {
    return cachedHasAnyOpenUnion;
  }
  cachedHasAnyOpenUnion = false;
  for (const [, models] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        if (t.Type == "union" && t.IsUnionOpen) {
          cachedHasAnyOpenUnion = true;
          return cachedHasAnyOpenUnion;
        }
      }
    }
  }
  return cachedHasAnyOpenUnion;
}

registerTemplateFunc("hasAnyOpenUnion", hasAnyOpenUnion);
