// @ts-ignore
function templateCursorExtraction(op: Operation): string {
  const pagination = op.Extensions.Pagination;
  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) return "";

  let cursorField: FieldDef | undefined;
  for (const field of op.Request.Field.Type.Fields) {
    if (field.Name === cursorInput.Name) {
      cursorField = field;
      break;
    }
    if (field.Type?.Fields) {
      const nested = field.Type.Fields.find((f) => f.Name === cursorInput.Name);
      if (nested) {
        cursorField = nested;
        break;
      }
    }
  }
  if (!cursorField) return "";

  const cursorType = sanitizeType(cursorField.Type, false);
  const accessor = stjAccessorForType(cursorType);

  if (cursorType === "string") {
    return `var nextCursor = __nextCursorEl.${accessor}();
    if (string.IsNullOrWhiteSpace(nextCursor))
    {
        return null;
    }`;
  }
  if (canFieldBeNull(cursorField)) {
    const nullableType = `${cursorType}${
      cursorField.Optional || cursorField.Nullable ? "?" : ""
    }`;
    return `${nullableType} nextCursorOrNull = __nextCursorEl.ValueKind == System.Text.Json.JsonValueKind.Null ? (${nullableType})null : __nextCursorEl.${accessor}();
    if (nextCursorOrNull == null)
    {
        return null;
    }
    ${cursorType} nextCursor = (${cursorType})nextCursorOrNull;`;
  }
  return `var nextCursor = __nextCursorEl.${accessor}();`;
}

function stjAccessorForType(t: string): string {
  switch (t) {
    case "string":
      return "GetString";
    case "long":
    case "long?":
      return "GetInt64";
    case "int":
    case "int?":
      return "GetInt32";
    case "double":
    case "double?":
      return "GetDouble";
    case "decimal":
    case "decimal?":
      return "GetDecimal";
    case "bool":
    case "bool?":
      return "GetBoolean";
    default:
      return "GetString";
  }
}

unregisterTemplateFunc("templateCursorExtraction");
registerTemplateFunc("templateCursorExtraction", templateCursorExtraction);
