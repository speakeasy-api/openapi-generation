// @ts-ignore
function templateFieldAnnotations(fieldDef: FieldDef): string {
  let annotations = [];
  let pydanticFieldOptions = [];
  let metadataFields = [];

  if (fieldDef.Comments?.Deprecated) {
    pydanticFieldOptions.push(
      `deprecated=${templateDeprecated(fieldDef.Comments, "field")}`,
    );
  }

  for (const annotation of fieldDef.Annotations) {
    // This annotation is only used by TypeScript templates
    if (annotation.Type().toString() === "needsCasing") {
      continue;
    }

    switch (annotation.Type().toString()) {
      case "security":
        metadataFields.push(
          templateSecurityAnnotation(annotation as SecurityAnnotation),
        );
        break;
      case "json": {
        pydanticFieldOptions.push(
          ...templateJSONAnnotations(fieldDef, annotation as JSONAnnotation),
        );
        break;
      }
      case "param": {
        const { paramMetadata, fieldOptions } = templateParamAnnotations(
          fieldDef,
          annotation as ParamAnnotation,
        );

        metadataFields.push(paramMetadata);
        pydanticFieldOptions.push(...fieldOptions);
        break;
      }
      case "request":
        metadataFields.push(
          templateRequestAnnotation(annotation as RequestAnnotation),
        );
        break;
      case "multipartForm": {
        const { multipartMetadata, fieldOptions } =
          templateMultipartFormAnnotations(
            fieldDef,
            annotation as MultipartFormAnnotation,
          );

        metadataFields.push(multipartMetadata);
        pydanticFieldOptions.push(...fieldOptions);
        break;
      }
      case "form": {
        const { formMetadata, fieldOptions } = templateFormAnnotations(
          fieldDef,
          annotation as FormAnnotation,
        );

        metadataFields.push(formMetadata);
        pydanticFieldOptions.push(...fieldOptions);
        break;
      }
      case "encoding":
        break;
      case "response":
        return "";
      default:
        throw new Error(`unimplemented ${annotation.Type().toString()}`);
    }
  }

  if (
    fieldDef.Const &&
    !pydanticFieldOptions.some((o) => o.startsWith("alias="))
  ) {
    pydanticFieldOptions.push(
      `alias="${originalFieldName(fieldDef).replaceAll(`"`, `\\"`)}"`,
    );
  }

  if (pydanticFieldOptions.length > 0) {
    addImport("pydantic", "");
    annotations.push(
      `pydantic.Field(${[...new Set(pydanticFieldOptions)].join(", ")})`,
    );
  }

  if (metadataFields.length > 0) {
    addImport("utils", "FieldMetadata", true);
    annotations.push(`FieldMetadata(${metadataFields.join(", ")})`);
  }

  if (annotations.length == 0) {
    return "";
  }

  return [...new Set(annotations)].join(", ");
}

function templateTypeAnnotations(
  fieldDef: FieldDef,
  usageLocation: string,
  owningModelInfo: any,
): string {
  let annotations = [];

  let validator = getPydanticValidator(fieldDef.Type);
  if (validator) {
    annotations.push(validator);
  }

  if (fieldDef.Const) {
    let validator = getPydanticConstValidator(
      fieldDef,
      usageLocation,
      owningModelInfo,
    );
    if (validator) {
      annotations.push(validator);
    }
  }

  let serializer = getPydanticSerializer(fieldDef.Type);
  if (serializer) {
    annotations.push(serializer);
  }

  if (annotations.length == 0) {
    return "";
  }

  return [...new Set(annotations)].join(", ");
}

// @ts-ignore
function templateSecurityAnnotation(annotation: SecurityAnnotation): string {
  let options = [];

  if (annotation.Option) {
    options.push(`option=True`);
  }

  if (annotation.Scheme) {
    options.push(`scheme=True`);
    options.push(`scheme_type="${annotation.SecType}"`);

    if (annotation.SubType) {
      options.push(`sub_type="${annotation.SubType}"`);
    }

    if (annotation.Composite) {
      options.push(`composite=True`);
    }
  }

  if (annotation.FieldName) {
    options.push(`field_name="${annotation.FieldName}"`);
  }

  if (options.length > 0) {
    addImport("utils", "SecurityMetadata", true);
    return `security=SecurityMetadata(${options.join(", ")})`;
  }

  return `security=true`;
}

function getPydanticValidator(typeDef: TypeDef): string {
  const validator = getValidator(typeDef);
  if (!validator || typeDef.Type.toString() == "enum") {
    return "";
  }

  addImport("pydantic.functional_validators", "BeforeValidator");
  return `BeforeValidator(${validator})`;
}

function getPydanticConstValidator(
  fieldDef: FieldDef,
  usageLocation: string,
  owningModelInfo: any,
): string {
  if (!fieldDef.Const) {
    return "";
  }

  addImport("utils", "validate_const", true);
  let validator = `validate_const(${templateConstOrDefaultValue(fieldDef, {
    usageLocation: usageLocation,
    owningModelInfo: owningModelInfo,
  })})`;

  addImport("pydantic.functional_validators", "AfterValidator");
  return `AfterValidator(${validator})`;
}

function getValidator(typeDef: TypeDef): string {
  switch (typeDef.Type.toString()) {
    case "decimal":
      addImport("utils", "validate_decimal", true);
      return "validate_decimal";
    case "number":
      if (typeDef.Format !== "string") {
        break;
      }
      addImport("utils", "validate_float", true);
      return "validate_float";
    case "bigint":
      addImport("utils", "validate_int", true);
      return "validate_int";
    case "integer":
      if (typeDef.Format !== "string") {
        break;
      }
      addImport("utils", "validate_int", true);
      return "validate_int";
  }

  return "";
}

function getPydanticSerializer(typeDef: TypeDef): string {
  const serializer = getSerializer(typeDef);
  if (!serializer) {
    return "";
  }

  addImport("pydantic.functional_serializers", "PlainSerializer");
  return `PlainSerializer(${serializer})`;
}

function getSerializer(typeDef: TypeDef): string {
  switch (typeDef.Type.toString()) {
    case "decimal":
      addImport("utils", "serialize_decimal", true);
      return `serialize_decimal(${
        typeDef.Format == "string" ? "True" : "False"
      })`;
    case "number":
      if (typeDef.Format !== "string") {
        break;
      }
      addImport("utils", "serialize_float", true);
      return `serialize_float(${
        typeDef.Format == "string" ? "True" : "False"
      })`;
    case "bigint":
    case "integer":
      if (typeDef.Format !== "string") {
        break;
      }
      addImport("utils", "serialize_int", true);
      return `serialize_int(${typeDef.Format == "string" ? "True" : "False"})`;
    default:
      return "";
  }
}

// @ts-ignore
function templateJSONAnnotations(
  fieldDef: FieldDef,
  annotation: JSONAnnotation,
): string[] {
  const fieldOptions = [];

  const sanitizedFieldName = sanitizeFieldName(fieldDef.Name);
  if (
    annotation.FieldName != "-" &&
    (sanitizedFieldName != annotation.FieldName || fieldDef.Const)
  ) {
    fieldOptions.push(`alias="${annotation.FieldName.replaceAll(`"`, `\\"`)}"`);
  }

  if (annotation.Ignore) {
    fieldOptions.push(`exclude=True`);
  }

  return fieldOptions;
}

// @ts-ignore
function templateParamAnnotations(
  fieldDef: FieldDef,
  annotation: ParamAnnotation,
): { paramMetadata: string; fieldOptions: string[] } {
  let fieldOptions = [];

  const sanitizedFieldName = sanitizeFieldName(fieldDef.Name);
  if (sanitizedFieldName != annotation.Name) {
    fieldOptions.push(`alias="${annotation.Name.replaceAll(`"`, `\\"`)}"`);
  }

  const options = [];

  if (annotation.Serialization) {
    options.push(`serialization="${annotation.Serialization}"`);
  } else if (annotation.Style) {
    options.push(`style="${annotation.Style}"`);
    options.push(`explode=${annotation.Explode ? "True" : "False"}`);
  }

  let paramMetadata = `${annotation.ParamType.toLowerCase().replace(
    "param",
    "",
  )}=`;

  if (options.length > 0) {
    const metadataName = `${sanitizeClassName(annotation.ParamType)}Metadata`;

    addImport("utils", metadataName, true);
    paramMetadata += `${metadataName}(${options.join(", ")})`;
  } else {
    paramMetadata += "True";
  }

  return { paramMetadata, fieldOptions };
}

// @ts-ignore
function templateRequestAnnotation(annotation: RequestAnnotation): string {
  if (
    annotation.MediaType == "" ||
    annotation.MediaType == "application/octet-stream"
  ) {
    return `request=True`;
  }

  addImport("utils", "RequestMetadata", true);
  return `request=RequestMetadata(media_type="${annotation.MediaType}")`;
}

// @ts-ignore
function templateMultipartFormAnnotations(
  fieldDef: FieldDef,
  annotation: MultipartFormAnnotation,
): { multipartMetadata: string; fieldOptions: string[] } {
  const fieldOptions = [];

  const sanitizedFieldName = sanitizeFieldName(fieldDef.Name);
  if (sanitizedFieldName != annotation.Name) {
    fieldOptions.push(`alias="${annotation.Name.replaceAll(`"`, `\\"`)}"`);
  }

  let multipartMetadata = "multipart=";

  switch (true) {
    case annotation.File:
      addImport("utils", "MultipartFormMetadata", true);
      multipartMetadata += `MultipartFormMetadata(file=True)`;
      break;
    case annotation.Content:
      addImport("utils", "MultipartFormMetadata", true);
      multipartMetadata += `MultipartFormMetadata(content=True)`;
      break;
    case annotation.JSON:
      addImport("utils", "MultipartFormMetadata", true);
      multipartMetadata += `MultipartFormMetadata(json=True)`;
      break;
    default:
      multipartMetadata += "True";
  }

  return { multipartMetadata, fieldOptions };
}

// @ts-ignore
function templateFormAnnotations(
  fieldDef: FieldDef,
  annotation: FormAnnotation,
): { formMetadata: string; fieldOptions: string[] } {
  const fieldOptions = [];

  const sanitizedFieldName = sanitizeFieldName(fieldDef.Name);
  if (sanitizedFieldName != annotation.Name) {
    fieldOptions.push(`alias="${annotation.Name.replaceAll(`"`, `\\"`)}"`);
  }

  let formMetadata = "form=";

  switch (true) {
    case annotation.JSON:
      addImport("utils", "FormMetadata", true);
      formMetadata += `FormMetadata(json=True)`;
      break;
    case annotation.Style != "":
      addImport("utils", "FormMetadata", true);
      formMetadata += `FormMetadata(style="${annotation.Style}", explode=${
        annotation.Explode ? "True" : "False"
      })`;
      break;
    default:
      formMetadata += "True";
  }

  return { formMetadata, fieldOptions };
}
