// @ts-ignore
function spreadAdditionalProperties(): boolean {
  return true;
}

unregisterTemplateFunc("spreadAdditionalProperties");
registerTemplateFunc("spreadAdditionalProperties", spreadAdditionalProperties);

// @ts-ignore
function templateAdditionalPropertiesClassAnnotation(): string {
  return "";
}

unregisterTemplateFunc("templateAdditionalPropertiesClassAnnotation");
registerTemplateFunc(
  "templateAdditionalPropertiesClassAnnotation",
  templateAdditionalPropertiesClassAnnotation,
);

// @ts-ignore
function templateAdditionalPropertiesMembers(): string {
  return "";
}

unregisterTemplateFunc("templateAdditionalPropertiesMembers");
registerTemplateFunc(
  "templateAdditionalPropertiesMembers",
  templateAdditionalPropertiesMembers,
);

const UNTYPED_ADDITIONAL_PROPERTIES_VALUE = "System.Text.Json.JsonElement";

function isUntypedAdditionalProperties(fieldDef: FieldDef): boolean {
  return (
    fieldDef.IsAdditionalProperties &&
    fieldDef.Type?.ItemType?.Type?.toString() === "any"
  );
}

// @ts-ignore
function sanitizeModelFieldType(
  fieldDef: FieldDef,
  optional: boolean,
  usageLocation: string,
): string {
  if (isUntypedAdditionalProperties(fieldDef)) {
    return `System.Collections.Generic.Dictionary<string, ${UNTYPED_ADDITIONAL_PROPERTIES_VALUE}>${
      optional ? "?" : ""
    }`;
  }
  return sanitizeType(fieldDef.Type, optional, usageLocation);
}

unregisterTemplateFunc("sanitizeModelFieldType");
registerTemplateFunc("sanitizeModelFieldType", sanitizeModelFieldType);

function getAdditionalPropertiesFieldDef(
  typeDef: TypeDef,
): FieldDef | undefined {
  const kind = typeDef.Type?.toString();
  if (kind !== "class" && kind !== "error") return undefined;
  return typeDef.Fields?.find((f: FieldDef) => f.IsAdditionalProperties);
}

// Source generation emits no metadata for the value type of a JSON-ignored member.
function collectAdditionalPropertiesValueTypes(
  typeDef: TypeDef,
  deps: Map<string, { Type: string; Name: string }>,
): void {
  const field = getAdditionalPropertiesFieldDef(typeDef);
  const item = field?.Type?.ItemType;
  if (!field || !item) return;

  const add = (type: string, name: string) => {
    if (!deps.has(type)) deps.set(type, { Type: type, Name: name });
  };
  if (isUntypedAdditionalProperties(field)) {
    add(UNTYPED_ADDITIONAL_PROPERTIES_VALUE, "JsonElement");
    return;
  }

  const valueType = jsonContextItemType(field.Type, false).replaceAll(
    "global::",
    "",
  );
  if (!SERIALIZABLE_TYPES.includes(item.Type?.toString() ?? "")) {
    add(
      valueType,
      valueType.endsWith("?")
        ? jsonTypeInfoNameForContextType(valueType)
        : jsonTypeInfoName(item),
    );
  }
  for (
    let t: TypeDef | undefined = item;
    t && (t.Type?.toString() === "array" || t.Type?.toString() === "map");
    t = t.ItemType
  ) {
    add(
      csharpTypeForJsonContext(t).replaceAll("global::", ""),
      jsonTypeInfoName(t),
    );
  }
}

interface AdditionalPropertiesType {
  Type: string;
  Member: string;
  ValueConverterFactory: string;
}

let _additionalPropertiesTypes: AdditionalPropertiesType[] | undefined;

function collectAdditionalPropertiesTypes(): AdditionalPropertiesType[] {
  if (_additionalPropertiesTypes) return _additionalPropertiesTypes;

  const result: AdditionalPropertiesType[] = [];
  forEachJsonObjectType((t, type) => {
    const field = getAdditionalPropertiesFieldDef(t);
    if (!field) return;
    const factory = getStjConverterFactory(field.Type);
    result.push({
      Type: type.startsWith("global::") ? type : `global::${type}`,
      Member: sanitizeFieldName(field.Name, sanitizeClassName(t.Name, false)),
      ValueConverterFactory: factory.endsWith("StringJsonConverterFactory")
        ? `new ${factory}()`
        : "",
    });
  });
  _additionalPropertiesTypes = result;
  return result;
}

// @ts-ignore
function hasAdditionalProperties(): boolean {
  return collectAdditionalPropertiesTypes().length > 0;
}

registerTemplateFunc("hasAdditionalProperties", hasAdditionalProperties);

// @ts-ignore
function templateValue(
  fieldDef: FieldDef,
  example?: any,
  parentHasExample?: boolean,
  additionalContext?: TemplateValueContext,
): string {
  if (!isUntypedAdditionalProperties(fieldDef)) {
    return templateValueCommon(
      fieldDef,
      example,
      parentHasExample,
      additionalContext,
    );
  }
  const values =
    example === undefined
      ? calculateExample(
          fieldDef,
          undefined,
          undefined,
          parentHasExample || false,
          additionalContext ?? {},
        )
      : example;
  if (
    values === null ||
    typeof values !== "object" ||
    Array.isArray(values) ||
    isExampleReferenceValue(values)
  ) {
    return templateValueCommon(
      fieldDef,
      example,
      parentHasExample,
      additionalContext,
    );
  }

  addUsageImport("System.Text.Json");
  const entries = Object.keys(values).map((key) =>
    templateMapValue(
      key,
      `JsonDocument.Parse(@"${JSON.stringify(values[key]).replaceAll(
        '"',
        '""',
      )}").RootElement`,
    ),
  );
  return [
    `new Dictionary<string, ${UNTYPED_ADDITIONAL_PROPERTIES_VALUE}>() {`,
    indentLines(processMapValues(entries), 1),
    "}",
  ].join("\n");
}

unregisterTemplateFunc("templateValue");
registerTemplateFunc("templateValue", templateValue);
