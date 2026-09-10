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

// @ts-ignore
function templateEnums(typeDef: TypeDef): string {
  let enumNames = getEnumNames(typeDef);

  let enums = [];

  typeDef.Enum.Values.forEach((value, index) => {
    enums.push(
      `${enumNames[index]} ${sanitizeClassName(
        typeDef.Name,
      )} = ${sanitizeEnumValue(value, typeDef.Enum.Type.Type)}`,
    );
  });

  return enums.join("\n");
}

registerTemplateFunc("templateEnums", templateEnums);
