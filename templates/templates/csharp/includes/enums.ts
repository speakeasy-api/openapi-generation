//@ts-ignore
function sanitizeEnumName(name: string): string {
  return sanitizeClassName(name, false);
}

registerTemplateFunc("sanitizeEnumName", sanitizeEnumName);

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value.replaceAll(`\\`, `\\\\`)}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

registerTemplateFunc("sanitizeEnumValue", sanitizeEnumValue);
