function getDocGroupedType(type: TypeDef, docGroup: string): DocGroupedType {
  return {
    Type: type,
    DocGroup: docGroup,
  };
}

registerTemplateFunc("getDocGroupedType", getDocGroupedType);

function getDocGroupedOperation(
  operation: Operation,
  docGroup: string,
): DocGroupedOperation {
  return {
    Operation: operation,
    DocGroup: docGroup,
  };
}

registerTemplateFunc("getDocGroupedOperation", getDocGroupedOperation);
