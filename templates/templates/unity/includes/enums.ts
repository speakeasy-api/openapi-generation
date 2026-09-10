// @ts-ignore
function sanitizeEnumValue(value: any, type: DataType): string {
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
