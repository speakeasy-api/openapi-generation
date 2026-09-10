function getRequestBodyAccessor(
  operation: Operation,
  inputAccessor: string,
): string | undefined {
  const bodyDef = getRequestBody(operation);

  if (bodyDef == null) {
    return undefined;
  }

  if (bodyDef.Type === "request") {
    return inputAccessor;
  }

  // No-zod: payload = input (no schema rename), so accessor uses the TS
  // identifier (declareModelField). Zod mode: $outboundSchema renames TS →
  // wire, so payload has wire-named keys → accessor uses originalFieldName.
  const fieldName = isNoZod()
    ? declareModelField(bodyDef.Value).replace(/^"|"$/g, "")
    : originalFieldName(bodyDef.Value);

  return sanitizeAccessor(
    inputAccessor,
    fieldName,
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
