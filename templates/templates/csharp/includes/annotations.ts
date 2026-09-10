// @ts-ignore
function templateFieldAnnotations(fieldDef: FieldDef, indent: number): string {
  let resolved = [];
  let speakeasyMetadata = [];

  for (const annotation of fieldDef.Annotations) {
    // This annotation is only used by TypeScript templates
    if (annotation.Type().toString() === "needsCasing") {
      continue;
    }

    let anno = (() => {
      switch (annotation.Type().toString()) {
        case "security":
          return templateSecurityAnnotation(annotation as SecurityAnnotation);
        case "json":
          return templateJSONAnnotation(fieldDef, annotation as JSONAnnotation);
        case "param":
          return templateParamAnnotation(annotation as ParamAnnotation);
        case "request":
          return templateRequestAnnotation(annotation as RequestAnnotation);
        case "multipartForm":
          return templateMultipartFormAnnotation(
            annotation as MultipartFormAnnotation,
          );
        case "form":
          return templateFormAnnotation(annotation as FormAnnotation);
        case "response":
          return "";
        case "encoding":
          // For event-stream fields, encoding annotations are typically used to specify
          // how the data should be serialized (e.g., "application/json")
          // For C#, we can return an empty string as the encoding is handled elsewhere
          return "";
        default:
          throw new Error(`unimplemented ${annotation.Type().toString()}`);
      }
    })();

    if (anno.startsWith("[")) {
      resolved.push(anno);
    } else {
      speakeasyMetadata.push(anno);
    }
  }

  if (speakeasyMetadata.length > 0) {
    resolved.push(`[SpeakeasyMetadata("${speakeasyMetadata.join(" ")}")]`);
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

registerTemplateFunc("templateFieldAnnotations", templateFieldAnnotations);

// @ts-ignore
function containsSpeakeasyMetadataAnnotation(fieldDef: FieldDef): boolean {
  for (const annotation of fieldDef.Annotations) {
    switch (annotation.Type().toString()) {
      case "security":
        return true;
      case "param":
        return true;
      case "request":
        return true;
      case "multipartForm":
        return true;
      case "form":
        return true;
    }
  }

  return false;
}

// @ts-ignore
function templateSecurityAnnotation(annotation: SecurityAnnotation): string {
  let name = escapeString(annotation.FieldName);

  let options = [];

  if (annotation.Option) {
    options.push(`option=true`);
  }
  if (annotation.Scheme) {
    options.push(`scheme=true`);
    options.push(`type=${annotation.SecType}`);

    if (annotation.SubType) {
      options.push(`subType=${annotation.SubType}`);
    }

    if (annotation.Composite) {
      options.push("composite=true");
    }
  }

  if (name) {
    options.push(`name=${name}`);
  }

  return `security:${options.join(",")}`;
}

function getJsonConverterType(typeDef: TypeDef): string {
  // Only use these serializers for bigint and decimal values to be encoded as strings.
  if (typeDef.Format == "string") {
    switch (typeDef.Type.toString()) {
      case "decimal":
        return "DecimalStrConverter";
      case "bigint":
        return "BigIntStrConverter";
    }
  }

  return "";
}

// @ts-ignore
function needsCustomJsonConverter(fieldDef: FieldDef): boolean {
  switch (fieldDef.Type.Type.toString()) {
    case "array":
    case "map":
      return getJsonConverterType(fieldDef.Type.ItemType) != "";
    default:
      return getJsonConverterType(fieldDef.Type) != "";
  }
}

// @ts-ignore
function templateJSONAnnotation(
  fieldDef: FieldDef,
  annotation: JSONAnnotation,
): string {
  if (
    annotation.Ignore &&
    context.Global.Config.PresenceAwareJSONSerialization &&
    !fieldDef.IsAdditionalProperties
  ) {
    return "[JsonIgnore]";
  }

  // additionalProperties are emitted as a literal "additionalProperties" wire
  // key instead of being spread top-level (JsonExtensionData semantics). Revisit this
  // union scorer special-case when wire format is fixed.
  const fieldName = fieldDef.IsAdditionalProperties
    ? "additionalProperties"
    : escapeString(annotation.FieldName);

  const jsonProperties: string[] = [`"${fieldName}"`];
  const jsonConverters: string[] = [];

  let fieldType = fieldDef.Type.Type.toString();
  let itemLevelConverter = false;
  switch (fieldType) {
    case "array":
    case "map":
      const itemConverterType = getJsonConverterType(fieldDef.Type.ItemType);
      if (itemConverterType != "") {
        if (needsOptionalNullableWrapper(fieldDef)) {
          jsonConverters.push(`typeof(${itemConverterType})`);
          itemLevelConverter = true;
        } else {
          jsonProperties.push(
            `ItemConverterType = typeof(${itemConverterType})`,
          );
        }
      }
      break;
    default:
      const converterType = getJsonConverterType(fieldDef.Type);
      if (converterType != "") {
        jsonConverters.push(`typeof(${converterType})`);
      }
      break;
  }

  const presenceAware =
    context.Global.Config.PresenceAwareJSONSerialization && !annotation.Ignore;

  if (presenceAware) {
    const requiredAnno = ((f: FieldDef) => {
      const required = !f.Optional;
      const nullable = f.Nullable;
      if (required && !nullable) return "Newtonsoft.Json.Required.Always";
      if (required && nullable) return "Newtonsoft.Json.Required.AllowNull";
      if (!required && !nullable)
        return "Newtonsoft.Json.Required.DisallowNull";
      return "Newtonsoft.Json.Required.Default";
    })(fieldDef);
    jsonProperties.push(`Required = ${requiredAnno}`);
  }

  // Override global NullValueHandling.Ignore to write/read an explicit null
  const requiredNullable = !fieldDef.Optional && fieldDef.Nullable;
  const isUnion = fieldType === "union";
  if (needsOptionalNullableWrapper(fieldDef) || isUnion || requiredNullable) {
    const disallowNull =
      presenceAware && fieldDef.Optional && !fieldDef.Nullable;
    if (!disallowNull) {
      jsonProperties.push("NullValueHandling = NullValueHandling.Include");
    }
  }

  const attributes = [`[JsonProperty(${jsonProperties.join(", ")})]`];
  if (jsonConverters.length > 0) {
    let jsonConverterType = jsonConverters.join(", ");
    if (needsOptionalNullableWrapper(fieldDef)) {
      const wrapperConverter = itemLevelConverter
        ? "OptionalNullableItemJsonConverter"
        : "OptionalNullableJsonConverter";

      jsonConverterType = `typeof(${wrapperConverter}), new object[] { ${jsonConverterType} }`;
    }

    attributes.push(`[JsonConverter(${jsonConverterType})]`);
  }

  return attributes.join("\n");
}

// @ts-ignore
function templateParamAnnotation(annotation: ParamAnnotation): string {
  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.Serialization) {
    serialization = `serialization=${annotation.Serialization},`;
  } else if (annotation.Style) {
    serialization = `style=${annotation.Style},explode=${
      annotation.Explode ? "true" : "false"
    },`;
  }

  const allowReserved = annotation.AllowReserved ? ",allowReserved=true" : "";

  return `${annotation.ParamType}:${serialization}name=${name}${allowReserved}`;
}

// @ts-ignore
function templateRequestAnnotation(annotation: RequestAnnotation): string {
  return `request:mediaType=${annotation.MediaType}`;
}

// @ts-ignore
function templateMultipartFormAnnotation(
  annotation: MultipartFormAnnotation,
): string {
  if (annotation.File) {
    return `multipartForm:file,name=${annotation.Name}`;
  }
  if (annotation.Content) {
    return `multipartForm:content`;
  }
  let json = "";
  if (annotation.JSON) {
    json = ",json";
  }
  return `multipartForm:name=${annotation.Name}${json}`;
}

// @ts-ignore
function templateFormAnnotation(annotation: FormAnnotation): string {
  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.JSON) {
    serialization = `,json`;
  } else if (annotation.Style) {
    serialization = `,style=${annotation.Style},explode=${
      annotation.Explode ? "true" : "false"
    }`;
  }

  return `form:name=${name}${serialization}`;
}

// @ts-ignore
function templateDeprecationAnnotation(
  comments: CommentDef | null,
  indent: number,
  type: "method" | "field" | "class",
  addLeadingNewLine: boolean = true,
): string {
  if (!comments || !comments.Deprecated) {
    return "";
  }

  const prefix = addLeadingNewLine ? "\n" : "";

  const message =
    type == "class"
      ? comments.DeprecationMessage
      : `This ${type} will be removed in a future release, please migrate away from it as soon as possible`;

  const replacement = comments.DeprecationReplacement
    ? `. Use ${sanitizeDeprecationReplacement(
        comments.DeprecationReplacement,
        type,
      )} instead`
    : "";

  return indentString(
    `${prefix}[Obsolete("${message}${replacement}")]`,
    indent,
  );
}

registerTemplateFunc(
  "templateDeprecationAnnotation",
  templateDeprecationAnnotation,
);

//@ts-ignore
function templateUnionVariantAnnotations(): string {
  let annotations = '[SpeakeasyMetadata("form:explode=true")]';
  if (useSmartUnion()) {
    annotations = `[UnionVariant]\n${annotations}`;
  }

  return annotations;
}

registerTemplateFunc(
  "templateUnionVariantAnnotations",
  templateUnionVariantAnnotations,
);
