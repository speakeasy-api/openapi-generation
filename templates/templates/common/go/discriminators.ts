type DiscriminatorGetter = DiscriminatorMapping & { Getter: string };

//@ts-ignore
function sanitizeDiscriminatorGetters(
  parentFields: FieldDef[],
  fieldName: string,
  mapping: DiscriminatorMapping[],
): DiscriminatorGetter[] {
  return mapping.map((m) => {
    let suffix = `${fieldName}${sanitizeClassName(
      getDiscriminatorDisplayName(m),
    )}`;
    if (
      parentFields.some((field) => sanitizeFieldName(field.Name) === suffix)
    ) {
      suffix += "Value";
    }

    return {
      ...m,
      Getter: `Get${suffix}`,
    };
  });
}

registerTemplateFunc(
  "sanitizeDiscriminatorGetters",
  sanitizeDiscriminatorGetters,
);

registerTemplateFunc(
  "getDiscriminatorDisplayName",
  getDiscriminatorDisplayName,
);
