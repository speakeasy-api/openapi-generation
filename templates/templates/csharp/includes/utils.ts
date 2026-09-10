// @ts-ignore
function isPointerType(typeDef: TypeDef): boolean {
  switch (typeDef.Type.toString()) {
    case "bytes":
    case "map":
    case "any":
    case "array":
    case "bigint":
    case "decimal":
      return true;
  }

  return false;
}

registerTemplateFunc("isPointerType", isPointerType);

function getResponseUnionName(
  unionType: TypeDef,
  content: ResponseBodyContent,
): string {
  return `${sanitizeClassName(unionType.Name)}.Create${sanitizeClassName(
    sanitizeUnionTypeName(content.Content.Type),
  )}`;
}

function getResponseNullUnionName(unionType: TypeDef): string {
  return `${sanitizeClassName(unionType.Name)}.CreateNull()`;
}
