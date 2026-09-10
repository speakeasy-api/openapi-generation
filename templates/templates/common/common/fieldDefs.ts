/**
 * Resolves the field name of a field as it is defined in the API spec. It will
 * fall back to the computed name if a cannonical name cannot be found.
 */
//@ts-ignore
function originalFieldName(field: FieldDef): string {
  return field.OriginalName || field.Name;
}
registerTemplateFunc("originalFieldName", originalFieldName);

function isMultipartFileField(fieldDef: FieldDef): boolean {
  if (fieldDef.Type.Type.toString() != "class") {
    return false;
  }

  const ann = fieldDef.Annotations?.Get("multipartForm");
  return Boolean(ann && isMultipartFormAnnotation(ann) && ann.File);
}
registerTemplateFunc("isMultipartFileField", isMultipartFileField);

function isParameterField(fieldDefLike: {
  Annotations?: Annotations;
}): boolean {
  const ann = fieldDefLike.Annotations?.Get("param");
  return ann && isParamAnnotation(ann);
}
registerTemplateFunc("isParameterField", isParameterField);

function getFieldContentType(fieldDefLike: {
  Annotations?: Annotations;
}): string {
  const ann = fieldDefLike.Annotations?.Get("encoding");
  if (!ann || !isEncodingAnnotation(ann)) {
    return "";
  }

  return ann.MediaType;
}
registerTemplateFunc("getFieldContentType", getFieldContentType);

function isSecurityField(fieldDefLike: { Annotations?: Annotations }): boolean {
  return Boolean(fieldDefLike.Annotations?.Has("security"));
}
registerTemplateFunc("isSecurityField", isSecurityField);

function isBase64FileInputField(fieldDef: FieldDef): boolean {
  return fieldDef.Type.Extensions?.Base64InputMode === "file";
}
registerTemplateFunc("isBase64FileInputField", isBase64FileInputField);

function isSecurityClassField(field: FieldDef): boolean {
  return (
    field.Type.Type.toString() === "class" &&
    field.Type.Fields.findIndex((f) => f.Annotations.Has("security")) >= 0
  );
}
registerTemplateFunc("isSecurityClassField", isSecurityClassField);
