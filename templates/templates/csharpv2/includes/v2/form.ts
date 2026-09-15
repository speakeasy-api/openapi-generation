function getFormAnnotation(field: FieldDef): FormAnnotation | null {
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() === "form")
      return annotation as FormAnnotation;
  }
  return null;
}

function getJsonAnnotationName(field: FieldDef): string | null {
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() === "json") {
      return (annotation as JSONAnnotation).FieldName;
    }
  }
  return null;
}

function hasFormFields(typeDef: TypeDef | undefined): boolean {
  return Boolean(typeDef?.Fields?.some((field) => getFormAnnotation(field)));
}

function hasJsonFormFields(typeDef: TypeDef | undefined): boolean {
  return Boolean(
    typeDef?.Fields?.some((field) => {
      const annotation = getFormAnnotation(field);
      return (
        annotation?.JSON ||
        (field.Type.Type.toString() === "class" &&
          hasJsonFormFields(field.Type))
      );
    }),
  );
}

function canUseGeneratedFormSerializer(typeDef: TypeDef | undefined): boolean {
  return Boolean(
    typeDef &&
      typeDef.Type.toString() === "class" &&
      hasFormFields(typeDef) &&
      !hasJsonFormFields(typeDef),
  );
}

function formSerializerRef(typeDef: TypeDef): string {
  return `global::${getModelNamespace(
    typeDef.OutputLocation,
    true,
  )}.${sanitizeClassName(typeDef.Name)}FormSerializer`;
}

function formWireName(
  field: FieldDef,
  annotation: FormAnnotation | null,
): string {
  return escapeString(
    annotation?.Name || getJsonAnnotationName(field) || field.Name,
  );
}

function formExplode(annotation: FormAnnotation | null): boolean {
  return annotation?.Explode ?? true;
}

function assertSupportedFormStyle(annotation: FormAnnotation | null): void {
  const style = annotation?.Style || "form";
  if (style !== "form") {
    throw new Error(`Unsupported form style for C# v2 AOT form body: ${style}`);
  }
}

// STJ source-gen derives its JsonMetadataServices converter accessor from the
// `TypeInfoPropertyName`, e.g. `BoolConverter` for `Bool`. JsonMetadataServices
// only defines converters under the .NET type names (BooleanConverter,
// Int32Converter, etc.), so emitting C# aliases (`bool`, `int`, …) in the
// JsonContext breaks the generated code. Map aliases to their .NET names here
// so both `typeof(...)` references and the property name resolve correctly.
const PRIMITIVE_ALIAS_TO_NET: Record<string, string> = {
  bool: "Boolean",
  byte: "Byte",
  sbyte: "SByte",
  char: "Char",
  short: "Int16",
  ushort: "UInt16",
  int: "Int32",
  uint: "UInt32",
  long: "Int64",
  ulong: "UInt64",
  float: "Single",
  double: "Double",
  decimal: "Decimal",
  string: "String",
  object: "Object",
};

function netTypeForCSharpAlias(name: string): string | null {
  return PRIMITIVE_ALIAS_TO_NET[name] ?? null;
}

function isDistinctNullableValueKind(typeDef: TypeDef): boolean {
  return isValueType(typeDef) && typeDef.Type?.toString() !== "enum";
}

// Renders a collection element for the JsonContext
function jsonContextItemType(typeDef: TypeDef): string {
  if (!typeDef.ItemType) {
    throw new Error(
      `jsonContextItemType: collection "${typeDef.Name}" (kind ${typeDef.Type}) has no ItemType`,
    );
  }
  const contextType = csharpTypeForJsonContext(typeDef.ItemType);

  // A nullable value-type element inside a List<>/Dictionary<> must be rendered as
  // `T?` in the JsonContext so the registered JsonTypeInfo<T> matches the model's
  // declared collection element type; otherwise typed AddJson<T> cannot infer T
  if (
    typeDef.ContainsNull &&
    isValueType(typeDef.ItemType) &&
    typeDef.ItemType.Type?.toString() !== "enum"
  ) {
    return `${contextType}?`;
  }

  return contextType;
}

function csharpTypeForJsonContext(typeDef: TypeDef): string {
  switch (typeDef.Type.toString()) {
    case "class":
    case "enum":
    case "union":
      return `${getModelNamespace(
        typeDef.OutputLocation,
        true,
      )}.${sanitizeClassName(typeDef.Name)}`;
    case "error":
      // Error TypeDefs emit two C# classes: a thrown wrapper extending
      // APIBaseException and a plain DTO `<X>Payload` carrying the wire shape.
      // Register the Payload — the wrapper's Exception contextType would force STJ
      // source-gen to walk RUC-annotated System.Exception surface.
      return `${getModelNamespace(
        typeDef.OutputLocation,
        true,
      )}.${sanitizeClassName(typeDef.Name)}Payload`;
    case "array":
      return `System.Collections.Generic.List<${jsonContextItemType(typeDef)}>`;
    case "map":
      return `System.Collections.Generic.Dictionary<string, ${jsonContextItemType(
        typeDef,
      )}>`;
    default: {
      const raw = sanitizeType(typeDef, false, "");
      const mapped = netTypeForCSharpAlias(raw);
      return mapped ? `System.${mapped}` : raw;
    }
  }
}

function scalarValueExpr(prop: string, field: FieldDef): string {
  if (field.Type.Type.toString() !== "enum") return prop;

  const enumType = field.Type.Enum?.Type?.Type?.toString();
  const supported =
    enumType === "string" ||
    enumType === "int32" ||
    enumType === "int64" ||
    enumType === "integer";
  if (!supported) return prop;

  if (field.Nullable || field.Optional) {
    prop += "?";
  }

  return field.Type.Enum?.Open ? `${prop}.Value` : `${prop}.Value()`;
}

function jsonTypeInfoName(typeDef: TypeDef): string {
  switch (typeDef.Type.toString()) {
    case "class":
    case "enum":
    case "union":
      return `${getJsonContextBucketPrefix(
        typeDef.OutputLocation,
      )}${sanitizeClassName(typeDef.Name)}`;
    case "error":
      // See csharpTypeForJsonContext: errors register the Payload DTO.
      return `${getJsonContextBucketPrefix(
        typeDef.OutputLocation,
      )}${sanitizeClassName(typeDef.Name)}Payload`;
  }

  // STJ source-gen derives JsonMetadataServices.<Name>Converter from
  // TypeInfoPropertyName. For primitive scalars the converter is named after
  // the .NET type only (e.g. BooleanConverter, Int32Converter), so
  // TypeInfoPropertyName must be the bare .NET type name. For arrays / maps
  // the converter is constructed at runtime by STJ regardless of name, but
  // the name must still be unique within the context — fall back to the
  // sanitized full type expression for those.
  const kind = typeDef.Type?.toString() ?? "";
  if (kind !== "array" && kind !== "map") {
    const raw = sanitizeType(typeDef, false, "");
    const mapped = netTypeForCSharpAlias(raw);
    if (mapped) return mapped;
    // Non-aliased primitives (DateTime, BigInteger, DateOnly, Guid, byte[], …):
    // strip namespace qualifiers + non-identifier characters. `byte[]` →
    // `ByteArray`, `int?` → `Int`, etc. Use sanitizeClassName so the result
    // is PascalCase and a valid C# identifier.
    const stripped = raw.replaceAll("global::", "").replace(/\[\]/g, " Array");
    const segments = stripped.split(/[^A-Za-z0-9_]+/).filter(Boolean);
    return sanitizeClassName(segments.join(" ")) || raw;
  }

  // Collections: csharpTypeForJsonContext renders nullable value-type elements as
  // `T?` (from the container's ContainsNull). Turn the `?` into a `Nullable` name
  // token so the element nullability is reflected in the property name
  // (`Dictionary<string, decimal?>` → `...DictionaryStringSystemDecimalNullable`)
  // and stays distinct from the non-nullable collection.
  const raw = csharpTypeForJsonContext(typeDef)
    .replaceAll("global::", "")
    .replaceAll("?", " Nullable ");
  return sanitizeClassName(raw.replaceAll(/[^A-Za-z0-9_]+/g, " "));
}

function jsonTypeInfoRef(typeDef: TypeDef): string {
  let outputLocation = typeDef.OutputLocation || "";
  const kind = typeDef.Type?.toString() ?? "";
  if (!outputLocation && (kind === "array" || kind === "map")) {
    outputLocation = typeDef.ItemType?.OutputLocation || "";
  }
  return `global::${getJsonContextNamespace(
    outputLocation,
  )}.${getJsonContextClassName(outputLocation)}.Default.${jsonTypeInfoName(
    typeDef,
  )}`;
}

function emitFormField(
  pairsVar: string,
  sourceExpr: string,
  field: FieldDef,
  parentType: TypeDef,
  forceAnnotation: FormAnnotation | null = null,
): string[] {
  const annotation = forceAnnotation ?? getFormAnnotation(field);
  assertSupportedFormStyle(annotation);

  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;
  const prop = wrappedFieldAccess(
    `${sourceExpr}.${sanitizeField(field, parentType)}`,
    field,
  );
  const name = formWireName(field, annotation);
  const explode = formExplode(annotation);

  if (annotation?.JSON) {
    return [
      `${fs}.AddJson(${pairsVar}, "${name}", ${prop}, ${jsonTypeInfoRef(
        field.Type,
      )});`,
    ];
  }

  switch (field.Type.Type.toString()) {
    case "array":
      return [`${fs}.AddList(${pairsVar}, "${name}", ${prop}, ${explode});`];
    case "map":
      return [`${fs}.AddMap(${pairsVar}, "${name}", ${prop}, ${explode});`];
    case "class":
      return emitFormObject(pairsVar, prop, field.Type, name, explode);
    default:
      return [
        `${fs}.AddScalar(${pairsVar}, "${name}", ${scalarValueExpr(
          prop,
          field,
        )});`,
      ];
  }
}

function isScalarFormField(
  field: FieldDef,
  annotation: FormAnnotation | null,
): boolean {
  if (annotation?.JSON) return false;

  switch (field.Type.Type.toString()) {
    case "array":
    case "map":
    case "class":
      return false;
    default:
      return true;
  }
}

function emitScalarFormFields(
  pairsVar: string,
  sourceExpr: string,
  fields: Array<{ field: FieldDef; annotation: FormAnnotation | null }>,
  parentType: TypeDef,
): string[] {
  if (fields.length === 0) return [];

  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;
  const entries = fields.map(({ field, annotation }) => {
    const name = formWireName(field, annotation);
    const prop = wrappedFieldAccess(
      `${sourceExpr}.${sanitizeField(field, parentType)}`,
      field,
    );
    return `    ("${name}", ${scalarValueExpr(prop, field)}),`;
  });

  return [
    `foreach (var (name, value) in new (string Name, object? Value)[]
{
${entries.join("\n")}
})
{
    ${fs}.AddScalar(${pairsVar}, name, value);
}`,
  ];
}

function emitFormFields(
  pairsVar: string,
  sourceExpr: string,
  fields: FieldDef[],
  parentType: TypeDef,
): string[] {
  const lines: string[] = [];
  let scalarFields: Array<{
    field: FieldDef;
    annotation: FormAnnotation | null;
  }> = [];

  const flushScalars = () => {
    lines.push(
      ...emitScalarFormFields(pairsVar, sourceExpr, scalarFields, parentType),
    );
    scalarFields = [];
  };

  for (const field of fields) {
    const annotation = getFormAnnotation(field);
    if (!annotation && !getJsonAnnotationName(field)) continue;

    if (isScalarFormField(field, annotation)) {
      assertSupportedFormStyle(annotation);
      scalarFields.push({ field, annotation });
      continue;
    }

    flushScalars();
    lines.push(
      ...emitFormField(pairsVar, sourceExpr, field, parentType, annotation),
    );
  }

  flushScalars();
  return lines;
}

function emitFormObject(
  pairsVar: string,
  sourceExpr: string,
  typeDef: TypeDef,
  name: string,
  explode: boolean,
): string[] {
  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;
  const nestedPairs = "__nestedForm";
  const nestedLines = emitFormFields(
    nestedPairs,
    sourceExpr,
    typeDef.Fields ?? [],
    typeDef,
  );

  if (nestedLines.length === 0) {
    return [`${fs}.AddScalar(${pairsVar}, "${name}", ${sourceExpr});`];
  }

  return [
    `if (${sourceExpr} != null)
{
    ${fs}.AddObjectFlatten(${pairsVar}, "${name}", ${nestedPairs} =>
    {
${indentString(nestedLines.join("\n"), 2)}
    }, ${explode});
}`,
  ];
}

function emitDirectFormBody(
  pairsVar: string,
  sourceExpr: string,
  bodyField: FieldDef,
): string[] {
  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;
  switch (bodyField.Type.Type.toString()) {
    case "map":
      return [
        `if (${sourceExpr} != null)
{
    foreach (var kv in ${sourceExpr})
    {
        ${pairsVar}.Add(new System.Collections.Generic.KeyValuePair<string?, string?>(kv.Key, global::${utilsNs}.Utilities.ValueToString(kv.Value)));
    }
}`,
      ];
    case "array":
      return [
        `if (${sourceExpr} != null)
{
    ${fs}.AddList(${pairsVar}, "${escapeString(
      bodyField.Name,
    )}", ${sourceExpr}, true);
}`,
      ];
    case "class":
      return emitFormFields(
        pairsVar,
        sourceExpr,
        bodyField.Type.Fields ?? [],
        bodyField.Type,
      );
    default:
      return [
        `${fs}.AddScalar(${pairsVar}, "${escapeString(
          bodyField.Name,
        )}", ${sourceExpr});`,
      ];
  }
}

// @ts-ignore
function templateApplyFormBody(op: Operation): string {
  if (!op.Request?.RequestBody) return "";

  const request = op.Request;
  const bodyField = request.RequestBody;
  const pairsVar = "__form";
  const requestType = request.Field?.Type;
  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;

  let sourceExpr = "request";
  let sourceType = bodyField.Type;

  if (requestType && hasFormFields(requestType)) {
    sourceType = requestType;
  } else if (request.Field && request.Field.Name !== bodyField.Name) {
    sourceExpr = `request.${sanitizeField(bodyField, requestType)}`;
  }

  const lines: string[] = [`if (request == null)`, `{`];

  if (!bodyField.Nullable && !bodyField.Optional) {
    lines.push(
      `    throw new System.ArgumentNullException(nameof(request), "request body is required");`,
    );
  } else {
    lines.push(`    httpRequest.Content = null;`);
  }

  lines.push(`}`, `else`, `{`, `    var ${pairsVar} = ${fs}.CreatePairs();`);

  const bodyLines =
    sourceType === requestType && hasFormFields(requestType)
      ? emitFormFields(
          pairsVar,
          sourceExpr,
          requestType.Fields ?? [],
          requestType,
        )
      : emitDirectFormBody(pairsVar, sourceExpr, bodyField);

  lines.push(...bodyLines.map((line) => `    ${line}`));
  lines.push(`    httpRequest.Content = ${fs}.FromPairs(${pairsVar});`, `}`);

  return lines.join("\n            ");
}

registerTemplateFunc("templateApplyFormBody", templateApplyFormBody);
