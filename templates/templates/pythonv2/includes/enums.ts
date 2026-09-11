function templateEnums(typeDef: TypeDef): string {
  const descriptions = typeDef.Enum.Descriptions || {};

  if (getEnumFormat(typeDef) == "union") {
    const lines: string[] = [];

    typeDef.Enum.Values.forEach((value) => {
      const description = descriptions[`${value}`];
      if (typeof description === "string" && description) {
        lines.push(...formatPythonEnumDescriptionComments(description));
      }

      lines.push(`${sanitizeEnumValue(value, typeDef.Enum.Type.Type)},`);
    });

    return lines.join("\n");
  } else {
    const enumNames = getEnumNames(typeDef);

    const enums: string[] = [];

    typeDef.Enum.Values.forEach((value, index) => {
      const description = descriptions[`${value}`];
      if (typeof description === "string" && description) {
        enums.push(
          ...formatPythonEnumDescriptionComments(description).map(
            (line) => `    ${line}`,
          ),
        );
      }

      enums.push(
        `    ${enumNames[index]} = ${sanitizeEnumValue(
          value,
          typeDef.Enum.Type.Type,
        )}`,
      );
    });

    return enums.join("\n");
  }
}

registerTemplateFunc("templateEnums", templateEnums);

function formatPythonEnumDescriptionComments(description: string): string[] {
  return sanitizeComments(description)
    .split("\n")
    .map((line) => {
      const trimmed = line.trim();
      if (trimmed.length === 0) {
        return "#";
      }

      return `# ${trimmed}`;
    });
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value.replaceAll("\\", "\\\\").replaceAll('"', '\\"')}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

function getEnumNames(t: TypeDef): string[] {
  if (t.Enum?.Names.length > 0) {
    return t.Enum.Names.map((n) => getEnumName(n));
  } else {
    return getEnumNamesFromValues(t.Enum?.Values);
  }
}

function templateUnrecognizedEnumType(typeDef: TypeDef): string {
  if (typeDef.Enum.Type.Type.toString() === "string") {
    addImport("types", "UnrecognizedStr", true);
    return "UnrecognizedStr";
  } else {
    addImport("types", "UnrecognizedInt", true);
    return "UnrecognizedInt";
  }
}
registerTemplateFunc(
  "templateUnrecognizedEnumType",
  templateUnrecognizedEnumType,
);
