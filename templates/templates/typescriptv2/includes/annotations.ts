// @ts-ignore
function templateAnnotations(
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  let annotations = [];
  let classTransformAnns = "";

  for (const annotation of fieldDef.Annotations) {
    switch (annotation.Type().toString()) {
      case "security":
        annotations.push(
          templateSecurityAnnotation(annotation as SecurityAnnotation),
        );
        break;
      case "json":
        classTransformAnns = templateJSONAnnotation(
          annotation as JSONAnnotation,
          fieldDef,
          usageLocation,
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
        return "";
      case "needsCasing":
        // This annotation is used by sanitization functions, not in metadata
        break;
      default:
        throw new Error(`unimplemented ${annotation.Type().toString()}`);
    }
  }

  let paramParts = [];

  if (annotations.length > 0) {
    paramParts.push(`data: "${annotations.join(", ")}"`);
  }

  let elemTypeAndDepth = getElemTypeAndDepth(fieldDef.Type, usageLocation, 1);
  if (elemTypeAndDepth) {
    paramParts.push(`elemType: ${elemTypeAndDepth.Type}`);
    if (elemTypeAndDepth.Depth > 1) {
      paramParts.push(`elemDepth: ${elemTypeAndDepth.Depth}`);
    }
  }

  let params = "";
  if (paramParts.length > 0) {
    params = `{ ${paramParts.join(", ")} }`;
  }

  return classTransformAnns
    ? `@SpeakeasyMetadata(${params})\n  ${classTransformAnns}`
    : `@SpeakeasyMetadata(${params})`;
}

registerTemplateFunc("templateAnnotations", templateAnnotations);

// @ts-ignore
function templateSecurityAnnotation(annotation: SecurityAnnotation): string {
  let name = escapeString(annotation.FieldName);

  let options = [];

  if (annotation.Option) {
    options.push("option=true");
  }
  if (annotation.Scheme) {
    options.push("scheme=true");
    options.push(`type=${annotation.SecType}`);

    if (annotation.SubType) {
      options.push(`subtype=${annotation.SubType}`);
    }
  }
  if (name) {
    options.push(`name=${name}`);
  }

  return `security, ${options.join(";")}`;
}

// @ts-ignore
function templateJSONAnnotation(
  annotation: JSONAnnotation,
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  if (annotation.Ignore) {
    return "@Exclude()";
  }

  const type = fieldDef.Type;

  const typeAnn = getTypeAnnotation(type, fieldDef.Optional, usageLocation);

  let transformMap = false;
  let transformArray = false;
  let childType: TypeDef = null;

  let transformAnn = "";
  if (type.Type.toString() === "map" || type.Type.toString() === "array") {
    let foundChild = false;
    let parentIncludesMap = type.Type.toString() === "map";

    childType = type.ItemType;

    while (!foundChild) {
      switch (childType.Type.toString()) {
        case "error":
        case "class":
          foundChild = true;

          if (parentIncludesMap) {
            if (type.Type.toString() === "array") {
              transformArray = true;
            } else {
              transformMap = true;
            }
          }
          break;
        case "array":
          childType = childType.ItemType;
          break;
        case "map":
          childType = childType.ItemType;
          parentIncludesMap = true;
          break;
        default:
          foundChild = true;
          break;
      }
    }
  }

  if (transformMap) {
    transformAnn = `\n  @Transform(({ value }) => {
      const obj: ${sanitizeType(type, fieldDef.Optional, usageLocation)} = {};
      for (const key in value) {
        obj[key] = objectToClass(value[key], ${sanitizeType(
          childType,
          false,
          usageLocation,
        )});
      }
      return obj;
    }, { toClassOnly: true })`;
  } else if (transformArray) {
    transformAnn = `\n  @Transform(({ value }) => {
      const arr: ${sanitizeType(type, fieldDef.Optional, usageLocation)} = [];
      for (const item of value) {
        arr.push(objectToClass(item, ${sanitizeType(
          childType,
          false,
          usageLocation,
        )}));
      }
      return arr;
    }, { toClassOnly: true })`;
  }

  if (type.Type.toString() === "date") {
    if (isLaxMode()) {
      transformAnn = `\n  @Transform(({ value }) => new Date(value), { toClassOnly: true })`;
    } else {
      transformAnn = `\n  @Type(() => String)\n  @Transform(({ value }) => new RFCDate(value), { toClassOnly: true })`;
    }
  } else if (type.Type.toString() === "date-time") {
    transformAnn = `\n  @Transform(({ value }) => new Date(value), { toClassOnly: true })`;
  }

  let name = escapeString(annotation.FieldName);
  return `@Expose({ name: "${name}" })${typeAnn}${transformAnn}`;
}

// @ts-ignore
function getTypeAnnotation(
  typeDef: TypeDef,
  optional: boolean,
  usageLocation: string,
): string {
  let typeAnn = "";
  switch (typeDef.Type.toString()) {
    case "array":
      return getTypeAnnotation(typeDef.ItemType, optional, usageLocation);
    case "error":
    case "class":
      typeAnn = `\n  @Type(() => ${sanitizeType(
        typeDef,
        optional,
        usageLocation,
      )})`;
      break;
  }

  return typeAnn;
}

// @ts-ignore
function templateParamAnnotation(annotation: ParamAnnotation): string {
  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.Serialization) {
    serialization = `serialization=${annotation.Serialization};`;
  }
  if (annotation.Style) {
    serialization = `style=${annotation.Style};explode=${
      annotation.Explode ? "true" : "false"
    };`;
  }

  return `${annotation.ParamType}, ${serialization}name=${name}`;
}

// @ts-ignore
function templateRequestAnnotation(annotation: RequestAnnotation): string {
  return `request, media_type=${annotation.MediaType}`;
}

// @ts-ignore
function templateMultipartFormAnnotation(
  annotation: MultipartFormAnnotation,
): string {
  let name = escapeString(annotation.Name);

  if (annotation.File) {
    return `multipart_form, file=true`;
  }
  if (annotation.Content) {
    return `multipart_form, content=true`;
  }
  let json = "";
  if (annotation.JSON) {
    json = ";json=true";
  }

  return `multipart_form, name=${name}${json}`;
}

// @ts-ignore
function templateFormAnnotation(annotation: FormAnnotation): string {
  let name = escapeString(annotation.Name);

  let serialization = "";
  if (annotation.JSON) {
    serialization = ";json=true";
  } else if (annotation.Style) {
    serialization = `;style=${annotation.Style};explode=${
      annotation.Explode ? "true" : "false"
    }`;
  }

  return `form, name=${name}${serialization}`;
}

type ElementInfo = {
  Type: string;
  Depth: number;
};

function getElemTypeAndDepth(
  typeDef: TypeDef,
  usageLocation: string,
  depth: number,
): ElementInfo | null {
  if (typeDef.IsContainer()) {
    if (typeDef.ItemType.IsContainer()) {
      return getElemTypeAndDepth(typeDef.ItemType, usageLocation, depth + 1);
    } else if (
      typeDef.ItemType.Type.toString() == "class" ||
      typeDef.ItemType.Type.toString() == "error"
    ) {
      return {
        Type: sanitizeClass(typeDef.ItemType, usageLocation, false),
        Depth: depth,
      };
    }
  }

  return null;
}
