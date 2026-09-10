function isPointerType(typeDef: TypeDef): boolean {
  switch (typeDef.Type.toString()) {
    case "bytes":
    case "map":
    case "any":
    case "array":
    case "set":
      return true;
  }

  return false;
}

registerTemplateFunc("isPointerType", isPointerType);

function getSecurityTypeSafe(security: FieldDef): TypeDef | null {
  if (!security) {
    return null;
  }

  return security.Type;
}
registerTemplateFunc("getSecurityTypeSafe", getSecurityTypeSafe);
