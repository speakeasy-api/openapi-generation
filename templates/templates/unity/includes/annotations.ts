// @ts-ignore
function templateAnnotations(fieldDef: FieldDef): string {
  let resolved = [`[SerializeField]`];
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
    let replacement = sanitizeDeprecationReplacement(
      fieldDef.Comments.DeprecationReplacement,
      "field",
    );

    if (replacement) {
      replacement = `. Use ${replacement} instead`;
    }

    resolved.unshift(
      `[Obsolete("This field will be removed in a future release, please migrate away from it as soon as possible${replacement}")]`,
    );
  }

  if (resolved.length == 0) {
    return "";
  }

  return resolved.join("\n        ");
}

registerTemplateFunc("templateAnnotations", templateAnnotations);

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
  }

  if (name) {
    options.push(`name=${name}`);
  }

  return `security:${options.join(",")}`;
}

function getJsonConverterType(typeDef: TypeDef): string {
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
  const fieldName = fieldDef.IsAdditionalProperties
    ? "additionalProperties"
    : escapeString(annotation.FieldName);

  const jsonProperties: string[] = [`"${fieldName}"`];
  const jsonConverters: string[] = [];

  switch (fieldDef.Type.Type.toString()) {
    case "array":
    case "map":
      const itemConverterType = getJsonConverterType(fieldDef.Type.ItemType);
      if (itemConverterType != "") {
        jsonProperties.push(`ItemConverterType = typeof(${itemConverterType})`);
      }
      break;
    case "date":
      jsonConverters.push("typeof(DateOnlyConverter)");
      break;
    default:
      const converterType = getJsonConverterType(fieldDef.Type);
      if (converterType != "") {
        jsonConverters.push(`typeof(${converterType})`);
      }
      break;
  }

  if (fieldDef.Nullable && !fieldDef.Optional) {
    jsonProperties.push("NullValueHandling = NullValueHandling.Include");
  }

  const attributes = [`[JsonProperty(${jsonProperties.join(", ")})]`];
  if (jsonConverters.length > 0) {
    attributes.push(`[JsonConverter(${jsonConverters.join(", ")})]`);
  }

  return attributes.join("\n        ");
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

  return `${annotation.ParamType}:${serialization}name=${name}`;
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
