function isMultipartRequest(op: Operation): boolean {
  return getRequestMediaType(op).startsWith("multipart/");
}

registerTemplateFunc("isMultipartRequest", isMultipartRequest);

function usesSerialization(op: Operation, match: string): boolean {
  return op.SerializationMethod?.toString() === match;
}

registerTemplateFunc("usesSerialization", usesSerialization);
