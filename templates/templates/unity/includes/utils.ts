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
