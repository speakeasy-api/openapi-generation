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

function definedTypeNames(outputLocation: string): Set<string> {
  if (!context.GlobalComputed) {
    context.GlobalComputed = {};
  }
  if (!context.GlobalComputed._typeNamesByLocation) {
    const byLocation = new Map<string, Set<string>>();
    for (const [location, models] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      const names = new Set<string>();
      for (const [, types] of sequencedMapEntries(models)) {
        for (const t of types) {
          names.add(sanitizeClass(t, location, true, false));
        }
      }
      byLocation.set(location, names);
    }
    context.GlobalComputed._typeNamesByLocation = byLocation;
  }
  return (
    context.GlobalComputed._typeNamesByLocation.get(outputLocation) ??
    new Set<string>()
  );
}

// @ts-ignore
function templateEnumValuesFuncName(
  typeDef: TypeDef,
  enumType: string,
): string {
  const taken = definedTypeNames(typeDef.OutputLocation);
  let name = `${enumType}Values`;
  if (!taken.has(name)) {
    return name;
  }
  name = `${enumType}EnumValues`;
  for (let i = 2; taken.has(name); i++) {
    name = `${enumType}EnumValues${i}`;
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
