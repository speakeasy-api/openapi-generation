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

  let typeName = "";

  if (fieldDef.Type.Discriminator) {
    for (const mapping of fieldDef.Type.Discriminator.Mapping) {
      if (
        mapping.Type.Name === typeToUse.Name &&
        mapping.Type.Type.toString() === typeToUse.Type.toString()
      ) {
        typeName = sanitizeClassName(getDiscriminatorDisplayName(mapping));
        break;
      }
    }
  } else {
    typeName = sanitizeClassName(sanitizeUnionTypeName(typeToUse));
  }

  const ctorPrefix = unionConstructorPrefix();
  const ctorSuffix = isGenericUnion(fieldDef.Type) ? "" : typeName;

  const unionValue = [
    `${getAccessNamespace(
      "usage",
      fieldDef.Type.Scope.toString(),
      fieldDef.Type.OutputLocation,
    )}${ctorPrefix}${sanitizeClassName(fieldDef.Type.Name)}${ctorSuffix}(`,
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

  if (fieldDef.Type.OutputLocation) {
    addImport(fieldDef.Type.OutputLocation, true);
  } else {
    addImport(fieldDef.Type.Scope.toString());
  }

  if (fieldDef.Optional || fieldDef.Nullable) {
    addSDKPackageImport(true);
    return `${sanitizeSDKPackageName(true)}.Pointer(${unionValue})`;
  }

  return unionValue;
}
