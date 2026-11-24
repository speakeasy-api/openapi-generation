// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  let enumNames = getEnumNames(typeDef);

  let enums = [];

  const descriptions = typeDef.Enum.Descriptions || {};

  typeDef.Enum.Values.forEach((value, index) => {
    const description = descriptions[`${value}`];
    if (typeof description === "string" && description) {
      enums.push(
        ...formatGoEnumDescriptionComments(enumNames[index], description),
      );
    }

    enums.push(
      `${enumNames[index]} ${sanitizeClassName(
        typeDef.Name,
      )} = ${sanitizeEnumValue(value, typeDef.Enum.Type.Type)}`,
    );
  });

  return enums.join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);

// @ts-ignore
function templateEnumValueNames(typeDef: TypeDef): string {
  return getEnumNames(typeDef)
    .map((name) => `${name},`)
    .join("\n");
}

registerTemplateFunc("templateEnumValueNames", templateEnumValueNames);

type PackageIdentifiers = {
  taken: Set<string>;
  helpers: Map<string, string>;
};

function allocateEnumValuesFuncName(
  enumType: string,
  taken: Set<string>,
): string {
  let name = `${enumType}Values`;
  if (taken.has(name)) {
    name = `${enumType}EnumValues`;
    for (let i = 2; taken.has(name); i++) {
      name = `${enumType}EnumValues${i}`;
    }
  }
  taken.add(name);
  return name;
}

function unionPackageIdentifiers(
  union: TypeDef,
  className: string,
  outputLocation: string,
): string[] {
  const unionType = getUnionTypeName(className, outputLocation);
  const variants: string[] = [];
  const typeConstants: string[] = [];
  if (union.Discriminator) {
    for (const mapping of union.Discriminator.Mapping) {
      variants.push(sanitizeClassName(getDiscriminatorDisplayName(mapping)));
    }
    typeConstants.push(
      ...getUnionEnumNamesFromDiscriminator(
        unionType,
        union.Discriminator.Mapping,
        outputLocation,
      ),
    );
  } else {
    for (const assocType of union.AssociatedTypes) {
      variants.push(sanitizeClassName(sanitizeUnionTypeName(assocType)));
    }
    typeConstants.push(
      ...getUnionEnumNamesFromAssociatedTypes(
        unionType,
        union.AssociatedTypes,
        outputLocation,
      ),
    );
  }
  if (union.IsUnionOpen) {
    variants.push("Unknown");
    typeConstants.push(openUnionUnknownEnumName(unionType));
  }
  const names = [unionType, `New${className}`, ...typeConstants];
  for (const v of variants) {
    names.push(`New${className}${v}`, `Create${className}${v}`);
  }
  return names;
}

function packageIdentifiers(outputLocation: string): PackageIdentifiers {
  if (!context.GlobalComputed) {
    context.GlobalComputed = {};
  }
  if (!context.GlobalComputed._packageIdentifiers) {
    const byLocation = new Map<string, PackageIdentifiers>();
    for (const [location, models] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      const ids: PackageIdentifiers = { taken: new Set(), helpers: new Map() };
      const enums: string[] = [];
      for (const [, types] of sequencedMapEntries(models)) {
        for (const t of types) {
          const name = sanitizeClass(t, location, true, false);
          ids.taken.add(name);
          if (t.Type.toString() == "union") {
            unionPackageIdentifiers(t, name, location).forEach((n) =>
              ids.taken.add(n),
            );
          }
          if (t.Type.toString() == "enum" && t.Enum) {
            getEnumNames(t).forEach((c) => ids.taken.add(c));
            enums.push(name);
          }
        }
      }
      for (const name of enums.sort()) {
        ids.helpers.set(name, allocateEnumValuesFuncName(name, ids.taken));
      }
      byLocation.set(location, ids);
    }
    context.GlobalComputed._packageIdentifiers = byLocation;
  }
  let ids = context.GlobalComputed._packageIdentifiers.get(outputLocation);
  if (!ids) {
    ids = { taken: new Set(), helpers: new Map() };
    context.GlobalComputed._packageIdentifiers.set(outputLocation, ids);
  }
  return ids;
}

// @ts-ignore
function templateEnumValuesFuncName(
  typeDef: TypeDef,
  enumType: string,
): string {
  const ids = packageIdentifiers(typeDef.OutputLocation);
  let name = ids.helpers.get(enumType);
  if (!name) {
    getEnumNames(typeDef).forEach((c) => ids.taken.add(c));
    name = allocateEnumValuesFuncName(enumType, ids.taken);
    ids.helpers.set(enumType, name);
  }
  return name;
}

registerTemplateFunc("templateEnumValuesFuncName", templateEnumValuesFuncName);

function formatGoEnumDescriptionComments(
  enumName: string,
  description: string,
): string[] {
  const sanitized = sanitizeComments(description).split("\n");

  return sanitized.map((line, index) => {
    const trimmed = line.trim();

    if (index === 0) {
      const suffix = trimmed.length > 0 ? ` ${trimmed}` : "";
      return `// ${enumName}${suffix}`;
    }

    return trimmed.length > 0 ? `// ${trimmed}` : "//";
  });
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value.replaceAll(`\\`, `\\\\`).replaceAll(`"`, `\\"`)}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

registerTemplateFunc("sanitizeEnumValue", sanitizeEnumValue);
