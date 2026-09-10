function isResponseModel(type: TypeDef): boolean {
  if (type.Type != "class") {
    return false;
  }

  for (const field of type.Fields) {
    if (field.Type.Type == "response") {
      return true;
    }
  }

  return false;
}
registerTemplateFunc("isResponseModel", isResponseModel);

/**
 * Returns true if the given type can be created by client code.
 * Models that pass this predicate should have usage snippets generated on their
 * docs pages for example.
 */
function canTemplateModelUsage(typeDef: TypeDef): boolean {
  switch (typeDef.Type.toString()) {
    case "string":
    case "integer":
    case "int32":
    case "bigint":
    case "number":
    case "float32":
    case "decimal":
    case "boolean":
    case "date":
    case "date-time":
    case "map":
    case "array":
    case "set":
    case "any":
    case "bytes":
    case "enum":
    case "response":
    case "request":
      return true;
    case "class":
      return typeDef.Fields.every(
        (f) => isPartOfCycle(f.Type) || canTemplateModelUsage(f.Type),
      );
    case "union":
      return typeDef.AssociatedTypes.every(
        (t) => isPartOfCycle(t) || canTemplateModelUsage(t),
      );
    case "error":
    case "request-stream":
    case "response-stream":
    case "jsonl":
    case "event-stream":
      return false;
    default:
      throw new Error(`Unrecognized type: ${typeDef.Type.toString()}`);
  }
}
