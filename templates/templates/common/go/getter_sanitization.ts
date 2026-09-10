// sanitizeGetterName returns a getter method name for a field that does not
// collide with any struct field name.
// @ts-ignore
function sanitizeGetterName(
  sanitizedFieldName: string,
  allFields: FieldDef[],
): string {
  const getterName = `Get${sanitizedFieldName}`;

  const allFieldNames = new Set<string>();
  for (const field of allFields) {
    allFieldNames.add(sanitizeFieldName(field.Name));
  }

  if (!allFieldNames.has(getterName)) {
    return getterName;
  }

  if (!allFieldNames.has(`${getterName}Value`)) {
    return `${getterName}Value`;
  }

  for (let i = 2; i < 100; i++) {
    const candidate = `${getterName}${i}`;
    if (!allFieldNames.has(candidate)) {
      return candidate;
    }
  }

  throw new Error(`Unable to resolve getter name conflict for "${getterName}"`);
}
registerTemplateFunc("sanitizeGetterName", sanitizeGetterName);
