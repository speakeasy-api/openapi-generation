// @ts-ignore
function templateAnnotations(fieldDef: FieldDef): string {
  let annotations = [];

  for (const annotation of fieldDef.Annotations) {
    switch (annotation.Type().toString()) {
      case "security":
        annotations.push(
          templateSecurityAnnotation(annotation as SecurityAnnotation),
        );
        break;
      case "json":
        annotations.push(
          templateJSONAnnotation(fieldDef, annotation as JSONAnnotation),
        );
        break;
      case "param":
        annotations.push(
          templateParamAnnotation(annotation as ParamAnnotation),
        );
        break;
      case "request":
        annotations.push(
          templateRequestAnnotation(annotation as RequestAnnotation),
        );
        break;
      case "multipartForm":
        annotations.push(
          templateMultipartFormAnnotation(
            annotation as MultipartFormAnnotation,
          ),
        );
        break;
      case "form":
        annotations.push(templateFormAnnotation(annotation as FormAnnotation));
        break;
      case "response":
        return "{}";
      case "needsCasing":
        // This annotation is only used by TypeScript templates
        break;
      case "encoding":
        // Encoding annotations are used for event-stream fields; handled elsewhere
        break;
      default:
        throw new Error(`unimplemented ${annotation.Type().toString()}`);
    }
  }

  if (annotations.length == 0) {
    return "{}";
  }

  return `{ ${annotations.join(", ")} }`;
}

registerTemplateFunc("templateAnnotations", templateAnnotations);

// @ts-ignore
function templateSecurityAnnotation(annotation: SecurityAnnotation): string {
  let options = [];
  if (annotation.Option) {
    options.push(`'option': true`);
  }
  if (annotation.Scheme) {
    options.push(`'scheme': true`);
    options.push(`'type': '${annotation.SecType}'`);

    if (annotation.SubType) {
      options.push(`'sub_type': '${annotation.SubType}'`);
    }
  }

  if (annotation.Composite) {
    options.push(`'composite': true`);
  }

  if (annotation.FieldName) {
    options.push(`'field_name': '${annotation.FieldName}'`);
  }
  return `'security': { ${options.join(", ")} }`;
}

function resolveJSONDecoderType(typeDef: TypeDef): TypeDef | null {
  const type = typeDef.Type.toString();

  if (type == "date-time" || type == "date" || typeDef.Enum) {
    return typeDef;
  }

  // Some schemas (for example anyOf with a single enum ref) are represented
  // as a wrapper union in the AST, so we need to unwrap to find the decoder.
  if (type == "union" && typeDef.AssociatedTypes.length == 1) {
    return resolveJSONDecoderType(typeDef.AssociatedTypes[0]);
  }

  return null;
}

// @ts-ignore
function templateJSONAnnotation(
  fieldDef: FieldDef,
  annotation: JSONAnnotation,
): string {
  let name = escapeString(annotation.FieldName);
  if (annotation.FieldName == "-") {
    if (fieldDef.IsAdditionalProperties) {
      name = sanitizeFieldName(fieldDef.Name);
    }
  }
  let formatAnnotations = "";
  let otherAnnotations = "";
  let optional = fieldDef.Optional ? "true" : "false";
  if (!fieldDef.Optional) {
    formatAnnotations += ", required: true";
  }
  const decoderType = resolveJSONDecoderType(fieldDef.Type);
  if (decoderType) {
    let decoder = "";
    const utilsPrefix = `::${sanitizeModuleName(
      context.Global.Config.Module,
    )}::Utils`;

    switch (decoderType.Type.toString()) {
      case "date-time":
        decoder = `${utilsPrefix}.datetime_from_iso_format(${optional})`;
        break;
      case "date":
        decoder = `${utilsPrefix}.date_from_iso_format(${optional})`;
        break;
      default:
        const enumType = sanitizeType(
          decoderType,
          false,
          false,
          decoderType.Scope,
        );
        if (decoderType.Enum?.Open) {
          decoder = `${utilsPrefix}.open_enum_from_string(${enumType}, ${optional})`;
        } else {
          decoder = `${utilsPrefix}.enum_from_string(${enumType}, ${optional})`;
        }
    }

    formatAnnotations += `, 'decoder': ${decoder}`;
  }
  if (fieldDef.Type.Type.toString() === "union") {
    if (fieldDef.Type.Discriminator) {
      otherAnnotations += `, 'discriminator': '${
        fieldDef.Type.Discriminator.TypePropertyName || "type"
      }'`;
      // Add discriminator mapping for inferred discriminators
      if (fieldDef.Type.Discriminator.Mapping) {
        const mappings = fieldDef.Type.Discriminator.Mapping.map((m) => {
          // Use the mapping Type directly instead of searching AssociatedTypes
          const typeName = sanitizeType(
            m.Type,
            false,
            false,
            fieldDef.Type.Scope,
          );
          return `'${m.Name}' => ${typeName}`;
        });
        if (mappings.length > 0) {
          otherAnnotations += `, 'discriminator_mapping': { ${mappings.join(
            ", ",
          )} }`;
        }
      }
    }
  }

  if (fieldDef.IsAdditionalProperties) {
    formatAnnotations += ", 'additional_properties': true";
  }

  return `'format_json': { 'letter_case': ::${sanitizeModuleName(
    context.Global.Config.Module,
  )}::Utils.field_name('${name}')${formatAnnotations} }${otherAnnotations}`;
}

// @ts-ignore
function templateParamAnnotation(annotation: ParamAnnotation): string {
  let serialization = "";
  if (annotation.Serialization) {
    serialization = `, 'serialization': '${annotation.Serialization}'`;
  } else if (annotation.Style) {
    serialization = `, 'style': '${annotation.Style}', 'explode': ${
      annotation.Explode ? "true" : "false"
    }`;
  }

  return `'${sanitizeFieldName(annotation.ParamType)}': { 'field_name': '${
    annotation.Name
  }'${serialization} }`;
}

// @ts-ignore
function templateRequestAnnotation(annotation: RequestAnnotation): string {
  return `'request': { 'media_type': '${annotation.MediaType}' }`;
}

// @ts-ignore
function templateMultipartFormAnnotation(
  annotation: MultipartFormAnnotation,
): string {
  if (annotation.File) {
    return `'multipart_form': { 'file': true, 'field_name': '${annotation.Name}' }`;
  }
  if (annotation.Content) {
    return `'multipart_form': { 'content': true }`;
  }
  let json = "";
  if (annotation.JSON) {
    json = ", 'json': true";
  }
  return `'multipart_form': { 'field_name': '${annotation.Name}'${json} }`;
}

// @ts-ignore
function templateFormAnnotation(annotation: FormAnnotation): string {
  let serialization = "";
  if (annotation.JSON) {
    serialization = ", 'json': true";
  } else if (annotation.Style) {
    serialization = `, 'style': '${annotation.Style}', 'explode': ${
      annotation.Explode ? "true" : "false"
    }`;
  }

  return `'form': { 'field_name': '${annotation.Name}'${serialization} }`;
}
