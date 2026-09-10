function templateUnionConstraintTerms(unionTypeDef: TypeDef): string {
  return unionTypeDef.AssociatedTypes.map((t) =>
    sanitizeType(t, false, unionTypeDef.OutputLocation, false),
  ).join(" | ");
}
registerTemplateFunc(
  "templateUnionConstraintTerms",
  templateUnionConstraintTerms,
);

function unionConstraintName(
  className: string,
  outputLocation: string,
): string {
  return findAvailableTypeName(className, outputLocation, [
    "Member",
    "Members",
    "Constraint",
    "MemberConstraint",
  ]);
}
registerTemplateFunc("unionConstraintName", unionConstraintName);

// Whether this union emits the single generic `New<Union>` constructor rather
// than one constructor per member. Requires the opt-in go.unionGenerics config,
// and falls back to per-member constructors when generics cannot express the
// union:
//   1. a member renders to a Go interface (`any`, `io.ReadCloser`), or
//   2. two members render to the same Go type, or
//   3. the discriminator mapping not a 1:1 pairing between discriminator values and member Go types.
// A discriminated union with a bijective mapping does use generics: the switch
// dispatches on the member Go type, so the value must be recoverable from the
// type. The switch sets the discriminator property on the member per case.
function isGenericUnion(unionTypeDef: TypeDef): boolean {
  if (
    unionTypeDef.Type.toString() !== "union" ||
    context.Global.Config.UnionGenerics !== true
  ) {
    return false;
  }

  const outputLocation = unionTypeDef.OutputLocation;

  const rendersToInterface = (t: TypeDef): boolean => {
    const kind = t.Type.toString();
    return (
      kind === "any" || kind === "request-stream" || kind === "response-stream"
    );
  };

  // AssociatedTypes must be non-interface and render to pairwise-distinct Go types
  const goTypes = new Set<string>();
  for (const member of unionTypeDef.AssociatedTypes) {
    if (rendersToInterface(member)) {
      return false;
    }
    goTypes.add(sanitizeType(member, false, outputLocation, false));
  }
  if (goTypes.size !== unionTypeDef.AssociatedTypes.length) {
    return false;
  }

  // Discriminated unions additionally require a bijective mapping between
  // discriminator values and member Go types, i.e.:
  // - no two discriminator values share a member Go type (injectivity)
  // - mapping set <==> AssociatedTypes set (surjectivity)
  // so the switch has a case for every value
  if (unionTypeDef.Discriminator && unionTypeDef.Discriminator.Mapping) {
    const mappingGoTypes = new Set<string>();
    for (const mapping of unionTypeDef.Discriminator.Mapping) {
      mappingGoTypes.add(
        sanitizeType(mapping.Type, false, outputLocation, false),
      );
    }

    if (
      mappingGoTypes.size !== unionTypeDef.Discriminator.Mapping.length ||
      mappingGoTypes.size !== goTypes.size
    ) {
      return false;
    }

    for (const t of goTypes) {
      if (!mappingGoTypes.has(t)) {
        return false;
      }
    }
  }

  return true;
}
registerTemplateFunc("isGenericUnion", isGenericUnion);

// Constructor name prefix for union types: "New" when generic constructors are
// enabled, otherwise the legacy "Create".
function unionConstructorPrefix(): string {
  return context.Global.Config.UnionGenerics === true ? "New" : "Create";
}
registerTemplateFunc("unionConstructorPrefix", unionConstructorPrefix);
