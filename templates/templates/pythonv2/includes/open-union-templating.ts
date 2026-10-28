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

function isOpenDiscriminatedUnion(typeDef: TypeDef): boolean {
  return (
    typeDef.Type.toString() === "union" &&
    !!typeDef.Discriminator &&
    typeDef.IsUnionOpen &&
    typeDef.AssociatedTypes.length > 1
  );
}

function isArrayOrMap(typeDef: TypeDef): boolean {
  const kind = typeDef.Type.toString();
  return kind === "array" || kind === "map";
}

function containsOpenDiscriminatedUnion(typeDef: TypeDef): boolean {
  if (isArrayOrMap(typeDef)) {
    return containsOpenDiscriminatedUnion(typeDef.ItemType);
  }
  return isOpenDiscriminatedUnion(typeDef);
}

/**
 * Returns the open discriminated union a union member dispatches to: the member
 * itself, or the item type of an array or map member with non-nullable items.
 */
function openUnionDispatchTarget(typeDef: TypeDef): TypeDef | null {
  if (isOpenDiscriminatedUnion(typeDef)) {
    return typeDef;
  }
  if (
    isArrayOrMap(typeDef) &&
    !typeDef.ContainsNull &&
    isOpenDiscriminatedUnion(typeDef.ItemType)
  ) {
    return typeDef.ItemType;
  }
  return null;
}

/**
 * A non-discriminated union whose members include at least two open
 * discriminated unions (directly, or as array or map items). Unions with an
 * open union nested in nullable items or in more than one array/map level do
 * not qualify.
 */
function isUnionOfOpenUnions(typeDef: TypeDef): boolean {
  if (
    typeDef.Type.toString() !== "union" ||
    typeDef.Discriminator ||
    typeDef.AssociatedTypes.length < 2
  ) {
    return false;
  }

  let targets = 0;
  for (const associatedType of typeDef.AssociatedTypes) {
    if (openUnionDispatchTarget(associatedType)) {
      targets++;
    } else if (containsOpenDiscriminatedUnion(associatedType)) {
      return false;
    }
  }
  return targets >= 2;
}

/**
 * A union of open unions validates through dispatch_open_union_members unless
 * one of its open union members is a forward reference from owningType.
 */
function isOpenUnionDispatchEligible(
  typeDef: TypeDef,
  owningType: TypeDef,
): boolean {
  if (!isUnionOfOpenUnions(typeDef)) {
    return false;
  }
  for (const associatedType of typeDef.AssociatedTypes) {
    const target = openUnionDispatchTarget(associatedType);
    if (
      target &&
      (areCircular(owningType, associatedType) ||
        areCircular(owningType, target))
    ) {
      return false;
    }
  }
  return true;
}

function shouldTemplateOpenUnionDispatch(): boolean {
  for (const t of getAllTypes()) {
    if (isUnionOfOpenUnions(t)) {
      return true;
    }
  }
  return false;
}

registerTemplateFunc(
  "shouldTemplateOpenUnionDispatch",
  shouldTemplateOpenUnionDispatch,
);

function collectOpenUnions(
  typeDef: TypeDef,
  found: Map<string, TypeDef>,
  visitedUnions: Set<string>,
): void {
  if (isOpenDiscriminatedUnion(typeDef)) {
    found.set(typeDef.Name, typeDef);
  } else if (isArrayOrMap(typeDef)) {
    collectOpenUnions(typeDef.ItemType, found, visitedUnions);
  } else if (
    typeDef.Type.toString() === "union" &&
    !typeDef.Discriminator &&
    !visitedUnions.has(typeDef.Name)
  ) {
    visitedUnions.add(typeDef.Name);
    for (const associatedType of typeDef.AssociatedTypes) {
      collectOpenUnions(associatedType, found, visitedUnions);
    }
  }
}

/**
 * Maps each open discriminated union reachable from the members of a
 * non-discriminated union (directly, as array or map items at any depth, or
 * through nested non-discriminated unions) to the discriminator values of the
 * other reachable open unions as a Python literal, e.g.
 * (("kind", frozenset({"note", "pause"})),). Empty unless at least two open
 * unions are reachable.
 */
function openUnionSiblingVariants(unionTypeDef: TypeDef): Map<string, string> {
  const found = new Map<string, TypeDef>();
  collectOpenUnions(unionTypeDef, found, new Set<string>());
  const siblingVariants = new Map<string, string>();
  if (found.size < 2) {
    return siblingVariants;
  }
  for (const name of found.keys()) {
    const valuesByKey = new Map<string, Set<string>>();
    for (const [siblingName, sibling] of found) {
      if (siblingName === name) {
        continue;
      }
      const key = sibling.Discriminator.TypePropertyName;
      if (!valuesByKey.has(key)) {
        valuesByKey.set(key, new Set<string>());
      }
      for (const mapping of sibling.Discriminator.Mapping) {
        valuesByKey.get(key)!.add(mapping.Name);
      }
    }
    const entries = [...valuesByKey.keys()].sort().map((key) => {
      const values = [...valuesByKey.get(key)!]
        .sort()
        .map((value) => JSON.stringify(value))
        .join(", ");
      return `(${JSON.stringify(key)}, frozenset({${values}}))`;
    });
    siblingVariants.set(name, `(${entries.join(", ")},)`);
  }
  return siblingVariants;
}

function renderOpenUnionSiblingGuard(
  openUnionRef: string,
  openUnion: TypeDef,
  siblingVariants: string,
): string {
  addImport("typing_extensions", "Annotated");
  addImport("pydantic", "");
  addImport("functools", "partial");
  addImport("utils.unions", "decline_sibling_variants", true);

  return `Annotated[${openUnionRef}, pydantic.WrapValidator(partial(decline_sibling_variants, union_name=${JSON.stringify(
    sanitizeClassName(openUnion.Name),
  )}, sibling_variants=${siblingVariants}))]`;
}

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
