function templateEnums(typeDef: TypeDef): string {
  let enumNames = getEnumNames(typeDef);

  let enums = [];
  const descriptions = typeDef.Enum.Descriptions || {};

  typeDef.Enum.Values.forEach((value, index) => {
    const description = descriptions[`${value}`];
    if (description) {
      enums.push(...formatRubyEnumDescriptionComments(description));
    }

    enums.push(
      `${enumNames[index]} = new(${sanitizeEnumValue(
        value,
        typeDef.Enum.Type.Type,
      )})`,
    );
  });

  return enums.join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);

function formatRubyEnumDescriptionComments(description: string): string[] {
  return sanitizeComments(description, 0, true)
    .split("\n")
    .map((line) => line.trim());
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `'${value
        .toString()
        .replaceAll(/\\/g, "\\\\")
        .replaceAll(/'/g, "\\'")}'`;
    case "int32":
    case "integer":
      return templateNumericLiteral(value);
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

// @ts-ignore
function getEnumNames(t: TypeDef): string[] {
  if (t.Enum?.Names.length > 0) {
    return t.Enum.Names.map((n) => getEnumName(n));
  } else {
    return getEnumNamesFromValues(t.Enum?.Values);
  }
}
