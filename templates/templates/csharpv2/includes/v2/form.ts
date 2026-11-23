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

function formFieldName(
  field: FieldDef,
  annotation: FormAnnotation | null,
): string {
  return annotation?.Name || getJsonAnnotationName(field) || field.Name;
}

function formWireName(
  field: FieldDef,
  annotation: FormAnnotation | null,
): string {
  return escapeString(formFieldName(field, annotation));
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
function jsonContextItemType(
  typeDef: TypeDef,
  annotateNullableReferenceItems: boolean,
): string {
  if (!typeDef.ItemType) {
    throw new Error(
      `jsonContextItemType: collection "${typeDef.Name}" (kind ${typeDef.Type}) has no ItemType`,
    );
  }
  const contextType = csharpTypeForJsonContext(
    typeDef.ItemType,
    annotateNullableReferenceItems,
  );

  // A nullable value-type element inside a List<>/Dictionary<> must be rendered as
  // `T?` in the JsonContext so the registered JsonTypeInfo<T> matches the model's
  // declared collection element type; otherwise typed AddJson<T> cannot infer T
  if (typeDef.ContainsNull && isDistinctNullableValueKind(typeDef.ItemType)) {
    return `${contextType}?`;
  }

  if (
    annotateNullableReferenceItems &&
    typeDef.ContainsNull &&
    !isValueType(typeDef.ItemType)
  ) {
    return `${contextType}?`;
  }

  return contextType;
}

function csharpTypeForJsonContext(
  typeDef: TypeDef,
  annotateNullableReferenceItems = false,
): string {
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
      return `System.Collections.Generic.List<${jsonContextItemType(
        typeDef,
        annotateNullableReferenceItems,
      )}>`;
    case "map":
      return `System.Collections.Generic.Dictionary<string, ${jsonContextItemType(
        typeDef,
        annotateNullableReferenceItems,
      )}>`;
    default: {
      const raw = sanitizeType(typeDef, false, "");
      const mapped = netTypeForCSharpAlias(raw);
      return mapped ? `System.${mapped}` : raw;
    }
  }
}

function isWireEnum(typeDef: TypeDef | undefined): boolean {
  if (typeDef?.Type.toString() !== "enum") return false;
  const enumType = typeDef.Enum?.Type?.Type?.toString();
  return (
    enumType === "string" ||
    enumType === "int32" ||
    enumType === "int64" ||
    enumType === "integer"
  );
}

// Utilities.ValueToString cannot map a closed enum to its wire value without reflection
// (it would emit the member name) and does not know class-based open enums, so enum values,
// including collection items, are converted at the call site.
function wireValueExpr(
  access: string,
  typeDef: TypeDef,
  nullable: boolean,
): string {
  const kind = typeDef.Type.toString();
  if (isWireEnum(typeDef)) {
    const acc = nullable ? `${access}?` : access;
    return typeDef.Enum?.Open ? `${acc}.Value` : `${acc}.Value()`;
  }
  if ((kind !== "array" && kind !== "map") || !isWireEnum(typeDef.ItemType)) {
    return access;
  }
  return projectItemsExpr(access, typeDef, nullable, (v) =>
    wireValueExpr(v, typeDef.ItemType, typeDef.ContainsNull),
  );
}

function projectItemsExpr(
  access: string,
  collection: TypeDef,
  nullable: boolean,
  item: (itemAccess: string) => string,
): string {
  const linq = "global::System.Linq.Enumerable";
  const projected =
    collection.Type.toString() === "array"
      ? `${linq}.ToList(${linq}.Select(${access}, __e => ${item("__e")}))`
      : `${linq}.ToDictionary(${access}, __kv => __kv.Key, __kv => ${item(
          "__kv.Value",
        )})`;
  return nullable ? `(${access} == null ? null : ${projected})` : projected;
}

function isUnionType(typeDef: TypeDef | undefined): boolean {
  return typeDef?.Type.toString() === "union";
}

function formValueExpr(
  access: string,
  typeDef: TypeDef,
  nullable: boolean,
): string {
  const kind = typeDef.Type.toString();
  if ((kind === "array" || kind === "map") && isUnionType(typeDef.ItemType)) {
    return projectItemsExpr(access, typeDef, nullable, (v) =>
      formItemValueExpr(v, typeDef.ItemType, typeDef.ContainsNull),
    );
  }
  return wireValueExpr(access, typeDef, nullable);
}

function formItemValueExpr(
  access: string,
  itemType: TypeDef,
  nullable: boolean,
): string {
  return isUnionType(itemType)
    ? unionItemValueExpr(access, itemType, access, itemType)
    : wireValueExpr(access, itemType, nullable);
}

function unionVariants(
  typeDef: TypeDef,
): Array<{ property: string; variant: TypeDef }> {
  const variants = typeDef.Discriminator?.Mapping
    ? typeDef.Discriminator.Mapping.map((m) => m.Type)
    : typeDef.AssociatedTypes ?? [];
  const seen = new Set<string>();
  const result: Array<{ property: string; variant: TypeDef }> = [];
  for (const variant of variants) {
    const property = sanitizeFieldName(sanitizeUnionTypeName(variant));
    if (seen.has(property)) continue;
    seen.add(property);
    result.push({ property, variant });
  }
  return result;
}

function unknownPrimitiveExpr(access: string): string {
  const utilsNs = getScopeNamespace("utils", true);
  return `global::${utilsNs}.Utilities.UnknownPrimitiveToString(((global::${getModelNamespace(
    "",
  )}.IOpenUnion)${access}).GetUnknownRaw())`;
}

function unionItemValueExpr(
  access: string,
  unionType: TypeDef,
  itemAccess: string,
  itemType: TypeDef,
): string {
  const branches = unionVariants(unionType).map(({ property, variant }) => {
    const variantAccess = `${access}.${property}`;
    let value: string;
    switch (variant.Type.toString()) {
      case "union":
        value = unionItemValueExpr(
          variantAccess,
          variant,
          itemAccess,
          itemType,
        );
        break;
      case "array":
      case "map":
      case "class":
        value = `global::System.Text.Json.JsonSerializer.Serialize(${itemAccess}, ${jsonTypeInfoRef(
          itemType,
        )})`;
        break;
      default:
        value = wireValueExpr(variantAccess, variant, true);
    }
    return `${variantAccess} != null ? (object?)${value}`;
  });
  if (branches.length === 0 && !unionType.IsUnionOpen) return "(object?)null";
  const fallback = unionType.IsUnionOpen
    ? unknownPrimitiveExpr(access)
    : "null";
  return `(${access} == null ? null : ${[...branches, fallback].join(" : ")})`;
}

let unionLocalCounter = 0;

function resetUnionLocals(): void {
  unionLocalCounter = 0;
}

function emitUnionVariants(
  sourceExpr: string,
  typeDef: TypeDef,
  emitVariant: (access: string, variant: TypeDef) => string[],
  emitUnknown: ((unknownExpr: string) => string[]) | null,
  guarded: boolean,
): string[] {
  const isLocal = /^[A-Za-z_][A-Za-z0-9_]*$/.test(sourceExpr);
  const union = isLocal ? sourceExpr : `__union${unionLocalCounter++}`;
  const branches = unionVariants(typeDef).map(({ property, variant }) => {
    const access = `${union}.${property}`;
    return `if (${access} != null)
{
${indentString(emitVariant(access, variant).join("\n"), 1)}
}`;
  });
  if (typeDef.IsUnionOpen && emitUnknown) {
    branches.push(`{
${indentString(emitUnknown(unknownPrimitiveExpr(union)).join("\n"), 1)}
}`);
  }
  if (branches.length === 0) return [];
  const chain = branches.join("\nelse ");
  if (isLocal) return [chain];
  if (guarded) return [`var ${union} = ${sourceExpr};\n${chain}`];
  return [
    `var ${union} = ${sourceExpr};
if (${union} != null)
{
${indentString(chain, 1)}
}`,
  ];
}

function scalarValueExpr(prop: string, field: FieldDef): string {
  return wireValueExpr(prop, field.Type, field.Nullable || field.Optional);
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
  return jsonTypeInfoNameForContextType(csharpTypeForJsonContext(typeDef));
}

function jsonTypeInfoNameForContextType(contextType: string): string {
  const raw = contextType
    .replaceAll("global::", "")
    .replaceAll("?", " Nullable ");
  return sanitizeClassName(raw.replaceAll(/[^A-Za-z0-9_]+/g, " "));
}

function jsonTypeInfoRef(typeDef: TypeDef): string {
  return `global::${getJsonContextNamespace()}.JsonTypeInfos.${jsonTypeInfoName(
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
      return [
        `${fs}.AddList(${pairsVar}, "${name}", ${formValueExpr(
          prop,
          field.Type,
          true,
        )}, ${explode});`,
      ];
    case "map":
      return [
        `${fs}.AddMap(${pairsVar}, "${name}", ${formValueExpr(
          prop,
          field.Type,
          true,
        )}, ${explode});`,
      ];
    case "class":
      return emitFormObject(pairsVar, prop, field.Type, name, explode);
    case "union":
      return emitFormUnion(
        pairsVar,
        prop,
        field.Type,
        formFieldName(field, annotation),
        true,
      );
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
    case "union":
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
  typeDef: TypeDef,
  name: string,
  nullable = true,
  guarded = false,
): string[] {
  const utilsNs = getScopeNamespace("utils", true);
  const fs = `global::${utilsNs}.FormSerializer`;
  const guard = (body: string): string[] => [
    guarded
      ? body
      : `if (${sourceExpr} != null)
{
${indentString(body, 1)}
}`,
  ];
  switch (typeDef.Type.toString()) {
    case "map":
      return guard(`foreach (var kv in ${sourceExpr})
{
    ${pairsVar}.Add(new System.Collections.Generic.KeyValuePair<string?, string?>(kv.Key, global::${utilsNs}.Utilities.ValueToString(${formItemValueExpr(
      "kv.Value",
      typeDef.ItemType,
      typeDef.ContainsNull,
    )})));
}`);
    case "array":
      return guard(
        `${fs}.AddList(${pairsVar}, "${escapeString(name)}", ${formValueExpr(
          sourceExpr,
          typeDef,
          false,
        )}, true);`,
      );
    case "class":
      return emitFormFields(
        pairsVar,
        sourceExpr,
        typeDef.Fields ?? [],
        typeDef,
      );
    case "union":
      return emitFormUnion(pairsVar, sourceExpr, typeDef, name, false, guarded);
    default:
      return [
        `${fs}.AddScalar(${pairsVar}, "${escapeString(name)}", ${wireValueExpr(
          sourceExpr,
          typeDef,
          nullable,
        )});`,
      ];
  }
}

function emitFormUnion(
  pairsVar: string,
  sourceExpr: string,
  typeDef: TypeDef,
  name: string,
  withUnknown: boolean,
  guarded = false,
): string[] {
  const fs = `global::${getScopeNamespace("utils", true)}.FormSerializer`;
  return emitUnionVariants(
    sourceExpr,
    typeDef,
    (access, variant) =>
      isUnionType(variant)
        ? emitFormUnion(pairsVar, access, variant, name, withUnknown, true)
        : emitDirectFormBody(pairsVar, access, variant, name, true, true),
    withUnknown
      ? (unknownExpr) => [
          `${fs}.AddScalar(${pairsVar}, "${escapeString(
            name,
          )}", ${unknownExpr});`,
        ]
      : null,
    guarded,
  );
}

// @ts-ignore
function templateApplyFormBody(op: Operation): string {
  if (!op.Request?.RequestBody) return "";
  resetUnionLocals();

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

  const nullBranch =
    !bodyField.Nullable && !bodyField.Optional
      ? `throw new System.ArgumentNullException(nameof(request), "request body is required");`
      : `httpRequest.Content = null;`;

  const bodyLines =
    sourceType === requestType && hasFormFields(requestType)
      ? emitFormFields(
          pairsVar,
          sourceExpr,
          requestType.Fields ?? [],
          requestType,
        )
      : emitDirectFormBody(
          pairsVar,
          sourceExpr,
          bodyField.Type,
          bodyField.Name,
          bodyField.Nullable || bodyField.Optional,
        );

  return `if (request == null)
{
    ${nullBranch}
}
else
{
    var ${pairsVar} = ${fs}.CreatePairs();
${indentString(bodyLines.join("\n"), 1)}
    httpRequest.Content = ${fs}.FromPairs(${pairsVar});
}`;
}

registerTemplateFunc("templateApplyFormBody", templateApplyFormBody);
