function isRequestAnnotation(
  annotation: Annotation,
): annotation is RequestAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "json":
    case "param":
    case "multipartForm":
    case "form":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "request":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isParamAnnotation(
  annotation: Annotation,
): annotation is ParamAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "json":
    case "request":
    case "multipartForm":
    case "form":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "param":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isSecurityAnnotation(
  annotation: Annotation,
): annotation is SecurityAnnotation {
  switch (annotation.Type().toString()) {
    case "json":
    case "request":
    case "param":
    case "multipartForm":
    case "form":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "security":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isMultipartFormAnnotation(
  annotation: Annotation,
): annotation is MultipartFormAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "json":
    case "request":
    case "param":
    case "form":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "multipartForm":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isFormAnnotation(
  annotation: Annotation,
): annotation is FormAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "json":
    case "request":
    case "param":
    case "multipartForm":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "form":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isJSONAnnotation(
  annotation: Annotation,
): annotation is JSONAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "request":
    case "multipartForm":
    case "param":
    case "form":
    case "encoding":
    case "response":
    case "needsCasing":
      return false;
    case "json":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isEncodingAnnotation(
  annotation: Annotation,
): annotation is EncodingAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "request":
    case "multipartForm":
    case "param":
    case "form":
    case "json":
    case "response":
    case "needsCasing":
      return false;
    case "encoding":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function isResponseAnnotation(
  annotation: Annotation,
): annotation is ResponseAnnotation {
  switch (annotation.Type().toString()) {
    case "security":
    case "request":
    case "multipartForm":
    case "param":
    case "form":
    case "json":
    case "encoding":
    case "needsCasing":
      return false;
    case "response":
      return true;
    default:
      throw new Error(`unimplemented ${annotation.Type().toString()}`);
  }
}

function paramAnnotations(
  fieldDef: FieldDef,
  type: "header" | "queryParam" | "pathParam",
) {
  const annotations: ParamAnnotation[] = [];

  for (const annotation of fieldDef.Annotations) {
    if (isParamAnnotation(annotation)) {
      const paramAnnotation = annotation as ParamAnnotation;
      if (paramAnnotation.ParamType == type) {
        annotations.push(paramAnnotation);
      }
    }
  }

  return annotations;
}

function securityAnnotations(fieldDef: FieldDef) {
  const annotations: SecurityAnnotation[] = [];

  for (const annotation of fieldDef.Annotations) {
    if (isSecurityAnnotation(annotation)) {
      annotations.push(annotation as SecurityAnnotation);
    }
  }

  return annotations;
}

function requestAnnotation(fieldDef: FieldDef) {
  for (const annotation of fieldDef.Annotations) {
    if (isRequestAnnotation(annotation)) {
      return annotation;
    }
  }

  return null;
}

function hasAnnotation(
  fieldDef: FieldDef,
  annotationType: keyof SpecificAnnotations,
): boolean {
  return fieldDef.Annotations?.Has(annotationType);
}
registerTemplateFunc("hasAnnotation", hasAnnotation);
