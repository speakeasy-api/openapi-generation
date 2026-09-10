// @ts-ignore
function templateAnnotations(
  fieldDef: FieldDef,
  outputLocation: string,
  forceOptional: boolean = false,
): string {
  let annos = [];

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
          return templateJSONAnnotation(
            fieldDef,
            annotation as JSONAnnotation,
            outputLocation,
            forceOptional,
          );
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

    if (anno.startsWith("#")) {
      annos.push(anno);
    } else {
      annos.push(`#[SpeakeasyMetadata('${anno}')]`);
    }
  }

  if (annos.length == 0) {
    return "";
  }

  return annos.join("\n");
}

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
      case "response":
        return true;
    }
  }

  return false;
}

registerTemplateFunc("templateAnnotations", templateAnnotations);

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

    if (annotation.Composite) {
      options.push("composite=true");
    }
  }
  if (name) {
    options.push(`name=${name}`);
  }

  return `security:${options.join(",")}`;
}

function templateDict(dict: { [key: string]: string }): string {
  let entries = [];
  for (let [key, value] of Object.entries(dict)) {
    key = key.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
    entries.push(`'${key}' => '${value}'`);
  }
  return `[${entries.join(", ")}]`;
}

// @ts-ignore
function templateJSONAnnotation(
  fieldDef: FieldDef,
  annotation: JSONAnnotation,
  outputLocation: string,
  forceOptional: boolean = false,
): string {
  let name = escapeString(annotation.FieldName);

  let properties = [];

  if (annotation.FieldName == "-") {
    if (fieldDef.IsAdditionalProperties) {
      name = sanitizeFieldName(fieldDef.Name);
    } else {
      return "#[\\Speakeasy\\Serializer\\Annotation\\Exclude]\n";
    }
  }

  properties.push(
    `#[\\Speakeasy\\Serializer\\Annotation\\SerializedName('${name.replace(
      "'",
      "\\'",
    )}')]`,
  );

  switch (fieldDef.Type.Type.toString()) {
    case "class":
    case "enum":
    case "map":
    case "array":
      let type = sanitizeType(
        fieldDef.Type,
        fieldDef.Optional || forceOptional,
        fieldDef.Nullable,
        outputLocation,
        Qualification.ANNOTATION,
      );
      properties.push(
        `#[\\Speakeasy\\Serializer\\Annotation\\Type('${type}')]`,
      );
      break;
    case "any":
      properties.push(`#[\\Speakeasy\\Serializer\\Annotation\\Type('mixed')]`);
      break;
    case "union":
      let t = sanitizeType(
        fieldDef.Type,
        fieldDef.Optional,
        fieldDef.Nullable,
        outputLocation,
        Qualification.ANNOTATION,
      );
      properties.push(`#[\\Speakeasy\\Serializer\\Annotation\\Type('${t}')]`);

      break;
  }

  if (
    fieldDef.Type.Type.toString() === "union" &&
    fieldDef.Type.Discriminator
  ) {
    const unionMap = Object.fromEntries(
      fieldDef.Type.Discriminator.Mapping.map((mapping) => {
        return [
          mapping.Name,
          sanitizeType(
            mapping.Type,
            false,
            false,
            "",
            Qualification.ANNOTATION,
          ),
        ];
      }),
    );
    properties.push(
      `#[\\Speakeasy\\Serializer\\Annotation\\UnionDiscriminator(field: '${
        fieldDef.Type?.Discriminator?.TypePropertyName || "type"
      }', map: ${templateDict(unionMap)})]`,
    );
  }

  if (fieldDef.Optional) {
    properties.push("#[\\Speakeasy\\Serializer\\Annotation\\SkipWhenNull]");
  }

  if (
    (fieldDef.Type.Type.toString() === "bigint" ||
      fieldDef.Type.Type.toString() === "decimal") &&
    fieldDef.Type.Format == "string"
  ) {
    properties.push(
      `#[\\Speakeasy\\Serializer\\Annotation\\Accessor(getter: '${sanitizeFieldName(
        fieldDef.Name,
      )}Accessor')]`,
    );
  }

  return properties.join("\n");
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
  let extras = "";
  if (annotation.FieldType) {
    switch (annotation.FieldType.Type.toString()) {
      case "date":
        extras += ",dateTimeFormat=Y-m-d";
        break;
      case "date-time":
        extras += ",dateTimeFormat=Y-m-d\\TH:i:s.up";
        break;
      case "bigint":
      case "decimal":
        if (annotation.FieldType.Format == "string") {
          extras += ",serializeToString=true";
        }
    }
  }

  return `${annotation.ParamType}:${serialization}name=${name}${extras}`;
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
    return `multipartForm:file=true,name=${annotation.Name}`;
  }
  if (annotation.Content) {
    return `multipartForm:content=true`;
  }
  let json = "";
  if (annotation.JSON) {
    json = ",json=true";
  }
  let extras = "";
  if (annotation.FieldType) {
    switch (annotation.FieldType.Type.toString()) {
      case "date":
        extras += ",dateTimeFormat=Y-m-d";
        break;
      case "date-time":
        extras += ",dateTimeFormat=Y-m-d\\TH:i:s.up";
        break;
      case "bigint":
      case "decimal":
        if (annotation.FieldType.Format == "string") {
          extras += ",serializeToString=true";
        }
    }
  }

  return `multipartForm:name=${annotation.Name}${json}${extras}`;
}

// @ts-ignore
function templateFormAnnotation(annotation: FormAnnotation): string {
  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.JSON) {
    serialization = `,json=true`;
  } else if (annotation.Style) {
    serialization = `,style=${annotation.Style},explode=${
      annotation.Explode ? "true" : "false"
    }`;
  }
  let dateTimeFormat = "";
  if (annotation.FieldType) {
    switch (annotation.FieldType.Type.toString()) {
      case "date":
        dateTimeFormat = ",dateTimeFormat=Y-m-d";
        break;
      case "date-time":
        dateTimeFormat = ",dateTimeFormat=Y-m-d\\TH:i:s.up";
        break;
    }
  }

  return `form:name=${name}${serialization}${dateTimeFormat}`;
}
