// @ts-ignore
function getParamAnnotation(fieldDef: FieldDef): ParamAnnotation | null {
  if (!fieldDef.Annotations) {
    return null;
  }
  return fieldDef.Annotations.Get("param");
}

registerTemplateFunc("getParamAnnotation", getParamAnnotation);

// @ts-ignore
function getOperationParameterName(fieldDef: FieldDef): string {
  const paramAnnot = getParamAnnotation(fieldDef);
  if (!paramAnnot) {
    return escapeString(fieldDef.Name);
  }
  return escapeString(paramAnnot.Name);
}

registerTemplateFunc("getOperationParameterName", getOperationParameterName);

// @ts-ignore
function templateAnnotations(
  fieldDef: FieldDef,
  trailingNewline: boolean = true,
): string {
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
          return templateJSONAnnotation(fieldDef);
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
          // TODO do we need to do anything here?
          return "";
        default:
          throw new Error(`unimplemented ${annotation.Type().toString()}`);
      }
    })();

    if (anno.startsWith("@")) {
      resolved.push(anno);
    } else if (anno.length > 0) {
      speakeasyMetadata.push(anno);
    }
  }

  if (speakeasyMetadata.length > 0) {
    resolved.push(
      `@${javaImportLocal("utils.SpeakeasyMetadata")}("${speakeasyMetadata.join(
        " ",
      )}")`,
    );
  }

  if (fieldDef.Comments?.Deprecated) {
    resolved.push(`@${javaImportDeprecated()}`);
  }

  if (resolved.length == 0) {
    return "";
  }

  return resolved
    .map((_) => "    " + _ + (trailingNewline ? "\n" : ""))
    .join("\n");
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
      options.push(`subtype=${annotation.SubType}`);
    }
  }

  if (annotation.Composite) {
    options.push(`composite`);
  }

  if (name) {
    options.push(`name=${name}`);
  }

  return `security:${options.join(",")}`;
}

// @ts-ignore
function templateJSONAnnotation(fieldDef: FieldDef): string {
  const name = escapeString(propertyName(fieldDef));

  const properties = [];

  if (fieldDef.IsAdditionalProperties) {
    properties.push(`@${javaImportJacksonAnn("JsonIgnore")}`);
  } else {
    if (!fieldDef.Optional && fieldDef.Nullable) {
      properties.push(
        `@${javaImportJacksonAnn("JsonInclude")}(${javaImportJacksonAnn(
          "JsonInclude.Include",
        )}.ALWAYS)`,
      );
    } else if (fieldDef.Optional || fieldDef.Nullable) {
      properties.push(
        `@${javaImportJacksonAnn("JsonInclude")}(${javaImportJacksonAnn(
          "JsonInclude.Include",
        )}.NON_ABSENT)`,
      );
    }
    const typ = fieldDef.Type.Type.toString();
    const format = fieldDef.Type.Format;

    properties.push(`@${javaImportJacksonAnn("JsonProperty")}("${name}")`);
    if ((typ == "decimal" || typ == "bigint") && format == "string") {
      properties.push(
        `@${javaImportJacksonAnn("JsonFormat")}(shape = ${javaImportJacksonAnn(
          "JsonFormat.Shape",
        )}.STRING)`,
      );
    }
  }
  return properties.join("\n    ");
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
  let encoding = "";
  if (annotation.AllowReserved) {
    encoding = "allowReserved=true,";
  }
  return `${annotation.ParamType}:${serialization}${encoding}name=${name}`;
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

function templateNullableAnnotation(type: string): string {
  return `@${javaImport("jakarta.annotation.Nullable")} ${javaImport(type)}`;
}
registerTemplateFunc("templateNullableAnnotation", templateNullableAnnotation);

function templateNonNullAnnotation(type: string): string {
  return `@${javaImport("jakarta.annotation.Nonnull")} ${javaImport(type)}`;
}
registerTemplateFunc("templateNonNullAnnotation", templateNonNullAnnotation);

function templateDeprecatedAnnotation(operation: Operation): string {
  if (operation.Comments?.Deprecated) {
    return `@${javaImportDeprecated()}`;
  }
  return "";
}
registerTemplateFunc(
  "templateDeprecatedAnnotation",
  templateDeprecatedAnnotation,
);

// Per-Builder visibility annotation, paired with @JsonPOJOBuilder, emitted
// only when exceedsJvmParamLimit() is true. Mirrors the global mapper
// baseline (auxiliary/JSON.java.stmpl sets every PropertyAccessor to
// Visibility.NONE) and deviates only by exposing setters, since
// @JsonPOJOBuilder needs Jackson to auto-detect Builder setter methods.
// All five PropertyAccessors must be set explicitly: @JsonAutoDetect does
// not inherit from the global mapper config — its own defaults are
// PUBLIC_ONLY (not NONE), so omitting any accessor would silently re-enable
// auto-detection for that accessor on this class.
function jacksonBuilderVisibilityOverride(): string {
  const ann = javaImport("com.fasterxml.jackson.annotation.JsonAutoDetect");
  return `@${ann}(setterVisibility = ${ann}.Visibility.ANY, getterVisibility = ${ann}.Visibility.NONE, isGetterVisibility = ${ann}.Visibility.NONE, fieldVisibility = ${ann}.Visibility.NONE, creatorVisibility = ${ann}.Visibility.NONE)`;
}

registerTemplateFunc(
  "jacksonBuilderVisibilityOverride",
  jacksonBuilderVisibilityOverride,
);
