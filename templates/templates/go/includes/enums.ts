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
