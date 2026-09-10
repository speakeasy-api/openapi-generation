/* Open Union helpers for forward-compatible union parsing in C#.
 *
 * The Unknown sentinel needs two coordinated identifiers:
 *   - a C# member name on the union-type class (e.g. `Unknown`);
 *   - a string value carried by that member (e.g. `"UNKNOWN"`).
 *
 * Their collision domains differ. Values collide against the raw
 * discriminator-mapping names (or variant names for untagged unions) and are
 * deduped with the shared underscore-wrapping scheme (findUniqueUnknownValue).
 * Member names collide against the *sanitized* identifiers — C# sanitization
 * strips underscores, so "_UNKNOWN_" and "UNKNOWN" both sanitize to `Unknown`
 * and the member must be deduped independently with a numeric suffix.
 */

// @ts-ignore
function openUnionMemberNames(typeDef: TypeDef): string[] {
  if (typeDef.Discriminator?.Mapping) {
    return typeDef.Discriminator.Mapping.map((m) => sanitizeClassName(m.Name));
  }
  return (typeDef.AssociatedTypes ?? []).map((t) =>
    sanitizeClassName(sanitizeUnionTypeName(t)),
  );
}

/** Returns generated variant property names that can collide with helper
 * members added to the open-union wrapper. */
// @ts-ignore
function openUnionVariantPropertyNames(typeDef: TypeDef): string[] {
  const isErrorUnion = isUnionOfErrors(typeDef);
  return (typeDef.AssociatedTypes ?? []).map((t) => {
    const name = sanitizeFieldName(sanitizeUnionTypeName(t));
    return isErrorUnion ? `${name}Payload` : name;
  });
}

/** Returns the root for the coordinated open-union helper family:
 * `Unknown`, `UnknownRaw`, and `IsUnknown`. A single suffix is selected so
 * all three generated names are available in their respective classes. */
// @ts-ignore
function openUnionUnknownMemberName(typeDef: TypeDef): string {
  const unionMembers = new Set(openUnionMemberNames(typeDef));
  const variantProperties = new Set(openUnionVariantPropertyNames(typeDef));

  for (let suffix = 0; ; suffix++) {
    const root = suffix === 0 ? "Unknown" : `Unknown${suffix}`;
    if (
      !unionMembers.has(root) &&
      !variantProperties.has(`${root}Raw`) &&
      !variantProperties.has(`Is${root}`)
    ) {
      return root;
    }
  }
}

registerTemplateFunc("openUnionUnknownMemberName", openUnionUnknownMemberName);

/** Returns the raw string value carried by the Unknown sentinel, deduped
 * against discriminator-mapping names (or variant names for untagged unions)
 * via the shared underscore-wrapping scheme. */
// @ts-ignore
function openUnionUnknownValue(typeDef: TypeDef): string {
  if (typeDef.Discriminator?.Mapping) {
    return findUniqueUnknownValue(typeDef, typeDef.Discriminator.Mapping);
  }
  const variantMappings = (typeDef.AssociatedTypes ?? []).map((t) => ({
    Name: sanitizeUnionTypeName(t),
  }));
  // @ts-ignore
  return findUniqueUnknownValue(typeDef, variantMappings);
}

registerTemplateFunc("openUnionUnknownValue", openUnionUnknownValue);
