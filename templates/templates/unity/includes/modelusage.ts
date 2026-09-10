// @ts-ignore
function templateUnion(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let { type: typeToUse, example: selectedExample } = selectExampleUnionType(
    fieldDef,
    example,
    undefined,
    additionalContext,
  );

  let typeName = fieldDef.Type.Discriminator
    ? sanitizeClassName(typeToUse.Name)
    : sanitizeClassName(sanitizeUnionTypeName(typeToUse));

  const className = sanitizeClassName(fieldDef.Type.Name);

  return [
    `${className}.Create${typeName}(`,
    indentLines(
      [
        templateValue(
          typeDefToFieldDef(typeToUse, fieldDef),
          selectedExample,
          false,
          additionalContext,
        ),
      ],
      1,
    ) + ",",
    ")",
  ].join("\n");
}

// @ts-ignore
function allExceptionTypes(responses: SubResponse[]) {
  const errors = [];
  for (const response of responses) {
    if (response.Error) {
      for (const content of response.Content) {
        errors.push(content.Content.Type);
      }
    }
  }
  return errors;
}
registerTemplateFunc("allExceptionTypes", allExceptionTypes);
