// SDK-side serializers (form/multipart/url) read request fields directly, so a
// wrapped field must be unwrapped to its inner T? via .ValueOrDefault.
// @ts-ignore
function wrappedFieldAccess(base: string, field: FieldDef): string {
  return needsOptionalNullableWrapper(field) ? `${base}.ValueOrDefault` : base;
}
