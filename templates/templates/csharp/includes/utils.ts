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
