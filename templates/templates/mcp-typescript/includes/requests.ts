function getRequestBodyAccessor(
  operation: Operation,
  inputAccessor: string,
): string | undefined {
  const bodyDef = getRequestBody(operation);

  if (bodyDef == null) {
    return undefined;
  }

  return bodyDef.Type === "request"
    ? inputAccessor
    : sanitizeAccessor(
        inputAccessor,
        sanitizeFieldName(bodyDef.Value.Name),
        isRequestOptional(operation.Request),
      );
}

registerTemplateFunc("getRequestBodyAccessor", getRequestBodyAccessor);

function getRequestMediaType(op: Operation) {
  const annotations = op.Request?.RequestBody?.Annotations;
  if (!annotations) {
    return "";
  }

  const ann = annotations.Get("request");
  if (!ann || !isRequestAnnotation(ann)) {
    return "";
  }

  const mediaType = ann.MediaType;
  return mediaType === "*/*" ? "" : mediaType;
}

registerTemplateFunc("getRequestMediaType", getRequestMediaType);
