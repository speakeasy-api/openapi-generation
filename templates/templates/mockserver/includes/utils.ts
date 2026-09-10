// @ts-ignore
function isPointerType(typeDef: TypeDef): boolean {
  switch (typeDef.Type.toString()) {
    case "bytes":
    case "map":
    case "any":
    case "array":
    case "set":
    case "bigint":
    case "decimal":
    case "event-stream":
    case "jsonl":
      return true;
  }

  return false;
}

registerTemplateFunc("isPointerType", isPointerType);

// @ts-ignore
function needsCustomJSONSerialization(typeDef: TypeDef): boolean {
  if (typeDef.Type.toString() != "class") {
    return false;
  }

  if (typeDef.UsedInUnion) {
    return true;
  }

  const typeNeeds = function (typeDef?: TypeDef): boolean {
    if (!typeDef) {
      return false;
    }

    if (
      typeDef.Type.toString() == "bigint" ||
      typeDef.Type.toString() == "decimal" ||
      typeDef.Type.toString() == "date" ||
      typeDef.Type.toString() == "date-time"
    ) {
      return true;
    }

    return false;
  };

  return typeDef.Fields.some((field) => {
    if (
      field.Const ||
      field.Default ||
      typeNeeds(field.Type) ||
      typeNeeds(field.Type.ItemType) ||
      field.IsAdditionalProperties
    ) {
      return true;
    }
  });
}
registerTemplateFunc(
  "needsCustomJSONSerialization",
  needsCustomJSONSerialization,
);

function templateRequiredFields(typeDef: TypeDef): string {
  if (typeDef.Type.toString() !== "class") {
    return "nil";
  }

  const requiredFields = [];

  for (const field of typeDef.Fields) {
    if (field.Optional) {
      continue;
    }

    requiredFields.push(originalFieldName(field));
  }

  if (requiredFields.length === 0) {
    return "nil";
  }

  return `[]string{${requiredFields.map((f) => `"${f}"`).join(", ")}}`;
}
registerTemplateFunc("templateRequiredFields", templateRequiredFields);
