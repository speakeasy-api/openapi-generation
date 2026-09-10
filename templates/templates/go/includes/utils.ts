// @ts-ignore
function resolveConflict(
  className: string,
  fieldName: string,
  alternative: string,
): string {
  if (className === fieldName) {
    return alternative;
  }

  return fieldName;
}

// Find a field in a request type by its original name
// @ts-ignore
function findPaginationField(
  requestType: TypeDef,
  fieldName: string,
): FieldDef | undefined {
  if (!requestType || !requestType.Fields) {
    return undefined;
  }

  return requestType.Fields.find((f) => originalFieldName(f) === fieldName);
}
registerTemplateFunc("findPaginationField", findPaginationField);

// Get the accessor for a pagination input field when parameters are flattened.
// For const fields, use the getter method on the request object.
// For regular fields, use the variable name directly.
// @ts-ignore
function getPaginationFlattenedAccessor(
  fields: FieldDef[],
  inputName: string,
): string {
  // Find the field with the matching name
  const field = fields.find((f) => originalFieldName(f) === inputName);

  if (field && field.Const) {
    // For const fields, use the getter method on the request variable
    // The request variable is always called "request" in the method body
    return `request.Get${sanitizeFieldName(inputName)}()`;
  }

  // For non-const fields, use the variable name directly
  return sanitizePrivateFieldName(inputName);
}
registerTemplateFunc(
  "getPaginationFlattenedAccessor",
  getPaginationFlattenedAccessor,
);

// Get the accessor for a pagination input field on the request struct when
// parameters are NOT flattened (maxMethodParams: 0). The field may be directly
// on the request struct (for params) or nested inside a body sub-field.
// @ts-ignore
function getPaginationNonFlattenedAccessor(
  requestFields: FieldDef[],
  inputName: string,
): string {
  // First check if the field is directly on the request struct
  const directField = requestFields.find(
    (f) => originalFieldName(f) === inputName,
  );
  if (directField) {
    return `request.${sanitizeFieldName(inputName)}`;
  }

  // Otherwise search nested body fields (fields with request annotation)
  for (const field of requestFields) {
    if (
      field.Type?.Type?.toString() === "class" &&
      field.Annotations?.Has("request")
    ) {
      const nestedFields = field.Type.Fields;
      if (nestedFields) {
        const nestedField = nestedFields.find(
          (f: FieldDef) => originalFieldName(f) === inputName,
        );
        if (nestedField) {
          return `request.${sanitizeFieldName(field.Name)}.${sanitizeFieldName(
            inputName,
          )}`;
        }
      }
    }
  }

  // Fallback to direct access
  return `request.${sanitizeFieldName(inputName)}`;
}
registerTemplateFunc(
  "getPaginationNonFlattenedAccessor",
  getPaginationNonFlattenedAccessor,
);

// Find the cursor field in the request type, searching nested body types if needed.
// @ts-ignore
function findPaginationCursorField(op: Operation): FieldDef | undefined {
  const pagination = op.Extensions?.Pagination;
  if (!pagination) return undefined;

  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) return undefined;

  const requestType = op.Request?.Field?.Type;
  if (!requestType?.Fields) return undefined;

  // Search top-level first
  const topLevel = requestType.Fields.find((f) => f.Name === cursorInput.Name);
  if (topLevel) return topLevel;

  // Search nested body fields
  for (const field of requestType.Fields) {
    if (field.Type?.Fields) {
      const nested = field.Type.Fields.find((f) => f.Name === cursorInput.Name);
      if (nested) return nested;
    }
  }
  return undefined;
}
registerTemplateFunc("findPaginationCursorField", findPaginationCursorField);

// Find the body wrapper field that contains the cursor input, if the cursor is nested.
// Returns undefined if the cursor is at the top level of the request type.
// @ts-ignore
function findPaginationCursorBodyWrapper(op: Operation): FieldDef | undefined {
  const pagination = op.Extensions?.Pagination;
  if (!pagination) return undefined;

  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) return undefined;

  const requestType = op.Request?.Field?.Type;
  if (!requestType?.Fields) return undefined;

  // If cursor is at top level, no wrapper needed
  const topLevel = requestType.Fields.find((f) => f.Name === cursorInput.Name);
  if (topLevel) return undefined;

  // Find which field contains the cursor
  for (const field of requestType.Fields) {
    if (field.Type?.Fields) {
      const nested = field.Type.Fields.find((f) => f.Name === cursorInput.Name);
      if (nested) return field;
    }
  }
  return undefined;
}
registerTemplateFunc(
  "findPaginationCursorBodyWrapper",
  findPaginationCursorBodyWrapper,
);

// Find the request-annotated wrapper field that a request-body pagination
// input is nested in, for operations whose parameters are NOT flattened.
// Returns undefined when no pagination input lives in the body; only
// requestBody inputs are considered so a same-named parameter cannot shadow
// the body field.
// @ts-ignore
function findPaginationBodyWrapper(op: Operation): FieldDef | undefined {
  const pagination = op.Extensions?.Pagination;
  if (!pagination) return undefined;

  const requestType = op.Request?.Field?.Type;
  if (!requestType?.Fields) return undefined;

  for (const kind of ["page", "offset", "cursor", "limit"]) {
    const input = getPaginationInput(pagination, kind);
    if (!input) continue;
    if (input.In.toString() !== "requestBody") continue;

    for (const field of requestType.Fields) {
      if (
        field.Annotations?.Has("request") &&
        field.Type?.Fields?.find((f) => originalFieldName(f) === input.Name)
      ) {
        return field;
      }
    }
  }
  return undefined;
}
registerTemplateFunc("findPaginationBodyWrapper", findPaginationBodyWrapper);

// Generate Go code to reconstruct a body wrapper field with the cursor updated.
// Used when the cursor pagination input is inside a nested body type.
// @ts-ignore
function templatePaginationBodyReconstruct(
  op: Operation,
  bodyField: FieldDef,
  sourceVarName: string,
): string {
  const pagination = op.Extensions?.Pagination;
  if (!pagination || !bodyField.Type?.Fields) return "";

  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) return "";

  const lines: string[] = [];
  lines.push(`${sanitizeType(bodyField.Type, false, "")}{`);

  for (const f of sortMethodParams(bodyField.Type.Fields)) {
    if (f.Const) continue;

    if (f.Name === cursorInput.Name) {
      lines.push(
        `            ${sanitizeFieldName(f.Name)}: ${templateFieldValueInit(
          f,
          "nCVal",
        )},`,
      );
    } else {
      const isOptionalNullable =
        f.Nullable &&
        f.Optional &&
        context.Global.Config.NullableOptionalWrapper;
      const accessor = `${sourceVarName}.${sanitizeFieldName(f.Name)}`;
      if (isOptionalNullable) {
        lines.push(`            ${sanitizeFieldName(f.Name)}: ${accessor},`);
      } else {
        lines.push(
          `            ${sanitizeFieldName(f.Name)}: ${templateFieldValueInit(
            f,
            accessor,
            false,
          )},`,
        );
      }
    }
  }

  lines.push(`        }`);
  return lines.join("\n");
}
registerTemplateFunc(
  "templatePaginationBodyReconstruct",
  templatePaginationBodyReconstruct,
);

// findPaginationFieldDeep moved to templates/templates/common/common/utils.ts

// @ts-ignore
function needsCustomJSONSerialization(
  typeDef: TypeDef,
  respectRequiredFields: boolean,
): boolean {
  if (respectRequiredFields) {
    return true;
  }

  if (typeDef.Type.toString() != "class") {
    return false;
  }

  if (
    typeDef.Extensions.TransformFromAPI ||
    typeDef.Extensions.TransformToAPI
  ) {
    return true;
  }

  if (typeDef.UsedInUnion) {
    return true;
  }

  // If includeEmptyObjects is enabled and there are optional fields that are
  // complex types (objects, arrays, maps), we need custom serialization
  const includeEmptyObjects = context.Global.Config.IncludeEmptyObjects;
  if (includeEmptyObjects) {
    const hasOptionalComplexFields = typeDef.Fields.some((field) => {
      if (!field.Optional || field.Default) {
        return false;
      }
      // Check if field type is complex (object, array, map, set)
      const fieldType = field.Type.Type.toString();
      return (
        fieldType === "class" ||
        fieldType === "array" ||
        fieldType === "map" ||
        fieldType === "set"
      );
    });
    if (hasOptionalComplexFields) {
      return true;
    }
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

    // Integer or number with string format need custom JSON serialization
    // to handle the integer:"string" and number:"string" struct tags
    if (
      (typeDef.Type.toString() == "integer" ||
        typeDef.Type.toString() == "number") &&
      typeDef.Format == "string"
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

  // Only template required fields if:
  // 1. RespectRequiredFields is enabled, OR
  // 2. UnionStrategy is "left-to-right" AND this type is used in a union
  const respectRequiredFields = context.Global.Config.RespectRequiredFields;
  const unionStrategy = context.Global.Config.UnionStrategy || "left-to-right";
  const isLeftToRight = unionStrategy === "left-to-right";

  if (!respectRequiredFields && !(isLeftToRight && typeDef.UsedInUnion)) {
    return "nil";
  }

  const requiredFields = [];

  for (const field of typeDef.Fields) {
    // field.Nullable condition was added because field.Optional is impacted by field.Nullable.
    // This was found out during #inc-2025-10-03-go-sdk-backwards-incompatible. We did not make a change in the ast because
    // other targets would also be impacted by a change in the AST. Long term we need to fix this in AST and add lot of tests around this.
    if (field.Optional || field.Nullable) {
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

// @ts-ignore
function isPointerType(typeDef: TypeDef): boolean {
  return getPointerTypes().includes(typeDef.Type.toString());
}
registerTemplateFunc("isPointerType", isPointerType);

function dereferenceAccessor(
  fieldDef: FieldDef,
  forcedPointer: boolean,
  accessor: string,
): string {
  switch (fieldDef.Type.Type.toString()) {
    case "map":
    case "array":
    case "set":
    case "bytes":
    case "decimal":
      break;
    default:
      if (forcedPointer) {
        return `*${accessor}`;
      }
  }

  return accessor;
}

// @ts-ignore
function inSameModel(
  modelName: string,
  outputLocation: string,
  typeDef: TypeDef,
): boolean {
  if (!typeDef.IsCustomType()) {
    return true;
  }

  if (!typeDef.ResolvedModel) {
    throw new Error(`Type ${typeDef.Name} is not resolved`);
  }

  return (
    `${outputLocation}${modelName}` ===
    `${typeDef.OutputLocation}${typeDef.ResolvedModel}`
  );
}

function canUseRawBody(operation: Operation): boolean {
  if (
    operation.Response.Responses.find((r) =>
      r.Content.find((c) => c.SerializationMethod === "eventstream"),
    )
  ) {
    return false;
  }

  return true;
}
registerTemplateFunc("canUseRawBody", canUseRawBody);

// @ts-ignore
function GetUserAgentHeader(): string {
  if (
    typeof context.Global.Config.WrapperName == "string" &&
    context.Global.Config.WrapperName
  ) {
    return context.Global.Config.WrapperName;
  }
  return "speakeasy-sdk/go";
}
registerTemplateFunc("GetUserAgentHeader", GetUserAgentHeader);

// 🔴🔴🔴
// TODO: Go doesn't currently support pulling defaults for parameters in the regular
// request building code of SDK methods so trying to use them for pagination
// will result in nasty inconsistencies.
// Reference: internal issue reference
// 🔴🔴🔴
// @ts-ignore
function findParamFieldByName(): FieldDef | undefined {
  return;
}

function getAllAcceptTypes(): string[] {
  const allAcceptTypes = new Set<string>();
  for (const operation of context.Local.Operations) {
    for (const type of operation.GetAcceptTypes()) {
      allAcceptTypes.add(type.split(";")[0]);
    }
  }
  for (const subSDK of context.Local.SubSDKs) {
    for (const operation of flattenOperationsPerSDK(subSDK)) {
      for (const type of operation.GetAcceptTypes()) {
        allAcceptTypes.add(type.split(";")[0]);
      }
    }
  }
  return Array.from(allAcceptTypes);
}

registerTemplateFunc("getAllAcceptTypes", getAllAcceptTypes);

function includeDecimal(): boolean {
  return isFeatureUsed("decimal");
}
registerTemplateFunc("includeDecimal", includeDecimal);
