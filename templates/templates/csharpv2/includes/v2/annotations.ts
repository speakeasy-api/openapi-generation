// @ts-ignore
function templateJsonPropertyAttribute(name: string): string {
  const escapedName = name.replaceAll(`\\`, `\\\\`).replaceAll(`"`, `\\"`);
  return `[System.Text.Json.Serialization.JsonPropertyName("${escapedName}")]`;
}

unregisterTemplateFunc("templateJsonPropertyAttribute");
registerTemplateFunc(
  "templateJsonPropertyAttribute",
  templateJsonPropertyAttribute,
);

// @ts-ignore
function getModelBaseImports(imports: Set<string>): Set<string> {
  // In v2, Utils namespace does not exist in the Models project.
  return imports;
}

// @ts-ignore
function getAnnotationImports(
  imports: Set<string>,
  fieldDef: FieldDef,
): Set<string> {
  // In v2, SpeakeasyMetadata is replaced by typed Roslyn-generator attributes.
  for (const annotation of fieldDef.Annotations) {
    switch (annotation.Type().toString()) {
      case "form":
      case "multipartForm":
        imports = addImport(imports, "Generators.Serialization", true);
        break;
      case "param":
      case "security":
      case "request":
        imports = addImport(imports, "Generators.Params", true);
        break;
    }
  }
  return imports;
}

// @ts-ignore
function getStjConverterFactory(typeDef: TypeDef): string {
  const utilsNs = getScopeNamespace("utils", true);

  // Descend through array/map containers to inspect the element type.
  // Factory-based converters handle BigInteger inside List<>/Dictionary<,>.
  let leaf = typeDef;
  while (
    leaf &&
    (leaf.Type?.toString() === "array" || leaf.Type?.toString() === "map")
  ) {
    leaf = leaf.ItemType;
  }
  if (!leaf) return "";

  const factory = (name: string) => `global::${utilsNs}.${name}`;

  const t = leaf.Type.toString();
  if (t === "bigint") {
    return factory(
      leaf.Format === "string"
        ? "BigIntegerStringJsonConverterFactory"
        : "BigIntegerJsonConverterFactory",
    );
  }
  if (t === "decimal") {
    return factory(
      leaf.Format === "string"
        ? "DecimalStringJsonConverterFactory"
        : "DecimalJsonConverterFactory",
    );
  }
  if (t === "date-time") {
    return factory("DateTimeJsonConverterFactory");
  }
  return "";
}

// @ts-ignore
function getStjConverterAttribute(typeDef: TypeDef): string {
  const factory = getStjConverterFactory(typeDef);
  return factory
    ? `[System.Text.Json.Serialization.JsonConverter(typeof(${factory}))]`
    : "";
}

// @ts-ignore
function templateJSONAnnotation(
  fieldDef: FieldDef,
  annotation: JSONAnnotation,
): string {
  if (annotation.Ignore && !fieldDef.IsAdditionalProperties) {
    return "[System.Text.Json.Serialization.JsonIgnore]";
  }

  const escaped = fieldDef.IsAdditionalProperties
    ? "additionalProperties"
    : jsonFieldNameLiteral(annotation);
  return `[System.Text.Json.Serialization.JsonPropertyName("${escaped}")]`;
}

function jsonFieldNameLiteral(annotation: JSONAnnotation): string {
  return escapeString(annotation.FieldName)
    .replaceAll(`\\`, `\\\\`)
    .replaceAll(`"`, `\\"`);
}

// @ts-ignore
function v2RequiredEligible(fieldDef: FieldDef): boolean {
  const annotation = getJsonAnnotation(fieldDef);
  return (
    annotation !== undefined &&
    !annotation.Ignore &&
    !fieldDef.Const &&
    !fieldDef.IsAdditionalProperties
  );
}

// @ts-ignore
function templateFieldAnnotations(fieldDef: FieldDef, indent: number): string {
  const resolved: string[] = [];

  const jsonIgnore = (condition: string) =>
    `[System.Text.Json.Serialization.JsonIgnore(Condition = System.Text.Json.Serialization.JsonIgnoreCondition.${condition})]`;

  let jsonIgnored = false;
  for (const annotation of getFieldAnnotations(fieldDef)) {
    const anno = templateFieldAnnotation(fieldDef, annotation);
    if (anno) {
      resolved.push(anno);
      if (anno === "[System.Text.Json.Serialization.JsonIgnore]") {
        jsonIgnored = true;
      }
    }
  }

  if (!jsonIgnored) {
    if (!fieldDef.Optional && v2RequiredEligible(fieldDef)) {
      resolved.push("[System.Text.Json.Serialization.JsonRequired]");
    }
    if (needsOptionalNullableWrapper(fieldDef)) {
      const factory = getStjConverterFactory(fieldDef.Type);
      const typeArgs = [sanitizeType(fieldDef.Type, true, "")];
      if (factory) {
        typeArgs.push(factory);
      }
      const converter = `global::${getModelNamespace(
        "",
      )}.OptionalNullableJsonConverter<${typeArgs.join(", ")}>`;
      resolved.push(
        `[System.Text.Json.Serialization.JsonConverter(typeof(${converter}))]`,
      );
      resolved.push(jsonIgnore("WhenWritingDefault"));
    } else {
      const stjAttr = getStjConverterAttribute(fieldDef.Type);
      if (stjAttr) {
        resolved.push(stjAttr);
      }

      if (
        (fieldDef.Optional || !fieldDef.Nullable) &&
        canFieldBeNull(fieldDef)
      ) {
        const emittedNullable =
          fieldDef.Nullable ||
          (!fieldDef.Const &&
            (fieldDef.Optional || fieldDef.Name === "ContentType"));
        const condition =
          isValueType(fieldDef.Type) && !emittedNullable
            ? "WhenWritingDefault"
            : "WhenWritingNull";
        resolved.push(jsonIgnore(condition));
      }
    }
  }

  if (fieldDef.Comments?.Deprecated) {
    resolved.unshift(
      templateDeprecationAnnotation(fieldDef.Comments, 0, "field", false),
    );
  }

  if (resolved.length == 0) {
    return "";
  }

  resolved.unshift("");

  return indentLines(resolved, indent);
}

unregisterTemplateFunc("templateFieldAnnotations");
registerTemplateFunc("templateFieldAnnotations", templateFieldAnnotations);

// @ts-ignore
function getJsonConverterType(_typeDef: TypeDef): string {
  // v2 serializers are generated by Roslyn generators instead of v1 Utils converters.
  return "";
}

// @ts-ignore
function templateSecurityAnnotation(annotation: SecurityAnnotation): string {
  const parts: string[] = [];
  if (annotation.Option) parts.push(`Option = true`);
  if (annotation.Scheme) {
    parts.push(`Scheme = true`);
    if (annotation.SecType) parts.push(`Type = "${annotation.SecType}"`);
    if (annotation.SubType) parts.push(`SubType = "${annotation.SubType}"`);
    if (annotation.Composite) parts.push(`Composite = true`);
  }
  const name = escapeString(annotation.FieldName);
  if (name) parts.push(`Name = "${name}"`);
  return `[SecurityParam(${parts.join(", ")})]`;
}

// @ts-ignore
function templateParamAnnotation(annotation: ParamAnnotation): string {
  const parts: string[] = [];
  const name = escapeString(annotation.Name);
  if (name) parts.push(`Name = "${name}"`);

  let attrName: string;
  switch (annotation.ParamType) {
    case "pathParam":
      attrName = "PathParam";
      if (annotation.Serialization) {
        parts.push(`Serialization = "${annotation.Serialization}"`);
      } else {
        if (annotation.Style && annotation.Style !== "simple")
          parts.push(`Style = "${annotation.Style}"`);
        if (annotation.Explode) parts.push(`Explode = true`);
      }
      break;
    case "queryParam":
      attrName = "QueryParam";
      if (annotation.Serialization) {
        parts.push(`Serialization = "${annotation.Serialization}"`);
      } else {
        if (annotation.Style && annotation.Style !== "form")
          parts.push(`Style = "${annotation.Style}"`);
        if (!annotation.Explode) parts.push(`Explode = false`);
      }
      break;
    case "header":
      attrName = "HeaderParam";
      if (annotation.Style && annotation.Style !== "simple")
        parts.push(`Style = "${annotation.Style}"`);
      if (annotation.Explode) parts.push(`Explode = true`);
      break;
    default:
      attrName = "PathParam";
  }

  return `[${attrName}(${parts.join(", ")})]`;
}

// @ts-ignore
function templateRequestAnnotation(annotation: RequestAnnotation): string {
  return `[RequestBody(MediaType = "${annotation.MediaType}")]`;
}

// @ts-ignore
function templateUnionFormAnnotation(discriminatorValue?: string): string {
  const annotations = [`[FormField()]`];

  if (discriminatorValue) {
    annotations.push(
      `[DiscriminatorValue(${sanitizeEnumValue(
        discriminatorValue,
        "string" as unknown as GojaEnum<DataType>,
      )})]`,
    );
  }

  return annotations.join("\n");
}

registerTemplateFunc(
  "templateUnionFormAnnotation",
  templateUnionFormAnnotation,
);

// @ts-ignore
function templateFormAnnotation(annotation: FormAnnotation): string {
  const parts: string[] = [];
  const name = escapeString(annotation.Name);
  if (name) parts.push(`Name = "${name}"`);
  if (annotation.JSON) parts.push(`IsJson = true`);
  if (annotation.Style && annotation.Style !== "form")
    parts.push(`Style = "${annotation.Style}"`);
  if (!annotation.Explode) parts.push(`Explode = false`);
  return `[FormField(${parts.join(", ")})]`;
}

// @ts-ignore
function templateMultipartFormAnnotation(
  annotation: MultipartFormAnnotation,
): string {
  const parts: string[] = [];
  if (annotation.File || annotation.Content) parts.push(`IsContent = true`);
  const name = escapeString(annotation.Name);
  if (name) parts.push(`Name = "${name}"`);
  if (annotation.JSON) parts.push(`IsJson = true`);
  return `[MultipartField(${parts.join(", ")})]`;
}

// @ts-ignore
function templateClassAnnotations(_typeDef: TypeDef, _indent: number): string {
  // [GenerateSerializer(SerializationFormat.Form)] previously triggered the
  // Roslyn source generator to emit `*FormSerializer.g.cs` helpers. Operation
  // handlers no longer call those — form bodies are inlined via
  // templateApplyFormBody, which uses `Utils.FormSerializer.AddJson` with
  // typed JsonTypeInfo<T> at the call site. The generated form serializers
  // were dead code that still produced AOT trim/analyzer warnings
  // (Serialize(value, value.GetType())). Disable the attribute so the
  // generator stops emitting them.
  return "";
}

unregisterTemplateFunc("templateClassAnnotations");
registerTemplateFunc("templateClassAnnotations", templateClassAnnotations);
