/**
 * Envelope fields on response structs that should be skipped when finding content fields.
 * Used by pagination, streaming, and other response-inspection logic.
 */
const RESPONSE_ENVELOPE_FIELDS = new Set([
  "HTTPMeta",
  "ContentType",
  "StatusCode",
  "RawResponse",
  "Headers",
]);

/**
 * Envelope fields excluding Headers — used when Headers is not an envelope concern
 * (e.g., transform detection, test result type extraction).
 */
const RESPONSE_ENVELOPE_FIELDS_NO_HEADERS = new Set([
  "HTTPMeta",
  "ContentType",
  "StatusCode",
  "RawResponse",
]);

type InputClassType =
  | "JSONRequestBody"
  | "MultipartRequestBody"
  | "FormRequestBody"
  | "NotClass";

function getInputClassType(field: FieldDef): InputClassType {
  if (field.Type.Type.toString() !== "class") {
    return "NotClass";
  }

  if (field.Annotations?.Has("request")) {
    const requestAnno = field.Annotations.Get("request") as RequestAnnotation;
    switch (true) {
      case matchContentType(requestAnno.MediaType, "application/json"):
        return "JSONRequestBody";
      case matchContentType(requestAnno.MediaType, "multipart/form-data"):
        return "MultipartRequestBody";
      case matchContentType(
        requestAnno.MediaType,
        "application/x-www-form-urlencoded",
      ):
        return "FormRequestBody";
      default:
        throw new Error(
          `${requestAnno.MediaType} not implemented for flag type`,
        );
    }
  }

  // Class-type field without a request annotation (e.g., nested object params).
  // Callers handle this the same as non-class fields — metadata treats it as FlagKindJSON.
  return "NotClass";
}
registerTemplateFunc("getInputClassType", getInputClassType);

// @ts-ignore
function isPointerType(typeDef: TypeDef): boolean {
  return getPointerTypes().includes(typeDef.Type.toString());
}
registerTemplateFunc("isPointerType", isPointerType);

// @ts-ignore
function includeDecimal(): boolean {
  return isFeatureUsed("decimal");
}
registerTemplateFunc("includeDecimal", includeDecimal);
